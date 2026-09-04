package runner_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/claim"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/policy"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

type fakePlayer struct {
	name string
	caps []registry.Capability
	fn   func(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error)
}

func (f *fakePlayer) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          f.name,
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities:  f.caps,
	}
}

func (f *fakePlayer) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	return f.fn(ctx, req)
}

func openCap(name, schema string) registry.Capability {
	return registry.Capability{Name: name, InputSchema: json.RawMessage(schema)}
}

func newTestRunner(t *testing.T, players ...*fakePlayer) (*runner.Runner, *event.MemoryBus, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	reg := registry.New()
	bus := event.NewMemoryBus()
	for _, p := range players {
		if err := reg.Register(p); err != nil {
			t.Fatal(err)
		}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := runner.New(reg, bus, st, t.TempDir(), "", 2, log)
	return r, bus, st
}

func mkTask(t *testing.T, entryPoint string, steps ...task.Step) task.Task {
	t.Helper()
	id, err := task.NewID()
	if err != nil {
		t.Fatal(err)
	}
	return task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id,
		CreatedAt:     time.Now().UTC(),
		Source:        task.Source{EntryPoint: entryPoint},
		Intent:        task.Intent{Summary: "runner test"},
		Steps:         steps,
	}
}

func waitForStatus(t *testing.T, st *store.Store, runID string, want store.Status) store.Run {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		run, _, err := st.GetRun(context.Background(), runID)
		if err == nil && run.Status == want {
			return run
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("run %s never reached status %s", runID, want)
	return store.Run{}
}

func storedEvents(t *testing.T, st *store.Store, runID string) []event.Event {
	t.Helper()
	events, err := st.ListEvents(context.Background(), runID)
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func hasEvent(events []event.Event, typ string) bool {
	for _, e := range events {
		if e.Type == typ {
			return true
		}
	}
	return false
}

func eventPayload(events []event.Event, typ string) map[string]any {
	for _, e := range events {
		if e.Type == typ {
			return e.Payload
		}
	}
	return nil
}

func TestSubmitSucceedsEndToEnd(t *testing.T) {
	p := &fakePlayer{
		name: "fake",
		caps: []registry.Capability{openCap("fake.echo", `{"type":"object"}`)},
		fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) {
			return json.RawMessage(`{"echoed":true}`), nil
		},
	}
	r, bus, st := newTestRunner(t, p)
	ch, unsub := bus.Subscribe(64)
	defer unsub()

	res, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.echo", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	if err != nil {
		t.Fatal(err)
	}
	r.WaitIdle()

	run, _, err := st.GetRun(context.Background(), res.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != store.StatusSucceeded {
		t.Fatalf("status=%s", run.Status)
	}

	// Drain the bus until the terminal event (earlier events arrive first).
	deadline := time.After(2 * time.Second)
delivered:
	for {
		select {
		case got := <-ch:
			if got.Type == event.TypeRunSucceeded {
				break delivered
			}
		case <-deadline:
			t.Fatal("bus did not deliver run.succeeded")
		}
	}

	events := storedEvents(t, st, res.RunID)
	for _, typ := range []string{
		event.TypeTaskAccepted, event.TypeRunPlanned, event.TypeRunStarted,
		event.TypeStepStarted, event.TypeStepSucceeded, event.TypeRunSucceeded,
	} {
		if !hasEvent(events, typ) {
			t.Fatalf("missing event %s in %v", typ, events)
		}
	}
	outs, err := st.ListStepOutputs(context.Background(), res.RunID)
	if err != nil || len(outs) != 1 || string(outs[0].Output) != `{"echoed":true}` {
		t.Fatalf("outputs=%+v err=%v", outs, err)
	}
}

func TestSubmitRejectsUnknownCapability(t *testing.T) {
	r, _, st := newTestRunner(t,
		&fakePlayer{name: "fake", caps: []registry.Capability{openCap("fake.echo", `{"type":"object"}`)},
			fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) { return nil, nil }},
	)
	_, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "missing.cap", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	var ve result.Error
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodeUnknownCapability {
		t.Fatalf("err=%v", err)
	}
	runs, listErr := st.ListRuns(context.Background(), 100)
	if listErr != nil || len(runs) != 0 {
		t.Fatalf("rejected task must not create a run: %d rows err=%v", len(runs), listErr)
	}
}

func TestSubmitRejectsInvalidInput(t *testing.T) {
	p := &fakePlayer{
		name: "fake",
		caps: []registry.Capability{openCap("fake.echo", `{"type":"object","required":["text"],"additionalProperties":false,"properties":{"text":{"type":"string"}}}`)},
		fn:   func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) { return nil, nil },
	}
	r, _, _ := newTestRunner(t, p)
	_, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.echo", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	var ve result.Error
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("err=%v", err)
	}
}

func TestSubmitRejectsBadEntryPoint(t *testing.T) {
	r, _, _ := newTestRunner(t,
		&fakePlayer{name: "fake", caps: []registry.Capability{openCap("fake.echo", `{"type":"object"}`)},
			fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) { return nil, nil }},
	)
	_, err := r.Submit(context.Background(), mkTask(t, "bogus",
		task.Step{StepID: "s1", Capability: "fake.echo", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	var ve result.Error
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodeSchema {
		t.Fatalf("err=%v", err)
	}
}

func TestPolicyDenyRejectsAtAdmission(t *testing.T) {
	p := &fakePlayer{
		name: "fake",
		caps: []registry.Capability{openCap("fake.echo", `{"type":"object"}`)},
		fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) {
			t.Fatal("denied player must not execute")
			return nil, nil
		},
	}
	r, _, _ := newTestRunner(t, p)
	r.Policy = policy.Table{Caps: map[string]policy.Verb{"fake.echo": policy.Deny}}

	_, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.echo", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	var ve result.Error
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodePolicyDenied {
		t.Fatalf("err=%v", err)
	}
}

func TestApprovalFlowGrant(t *testing.T) {
	calls := 0
	gate := openCap("fake.gate", `{"type":"object"}`)
	gate.ExecutionPolicy = string(policy.ApprovalRequired)
	p := &fakePlayer{
		name: "fake",
		caps: []registry.Capability{gate},
		fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) {
			calls++
			return json.RawMessage(`{}`), nil
		},
	}
	r, _, st := newTestRunner(t, p)

	res, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.gate", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	if err != nil {
		t.Fatal(err)
	}
	r.WaitIdle()

	run := waitForStatus(t, st, res.RunID, store.StatusWaitingApproval)
	if run.PendingStepID != "s1" || run.PendingCapability != "fake.gate" {
		t.Fatalf("pending=%+v", run)
	}
	events := storedEvents(t, st, res.RunID)
	if !hasEvent(events, event.TypeRunWaitingApproval) {
		t.Fatalf("missing waiting_approval in %v", events)
	}
	if calls != 0 {
		t.Fatalf("player must not run before approval (calls=%d)", calls)
	}

	if err := r.Approve(context.Background(), res.RunID, runner.DecisionGrant); err != nil {
		t.Fatal(err)
	}
	r.WaitIdle()
	waitForStatus(t, st, res.RunID, store.StatusSucceeded)
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
	events = storedEvents(t, st, res.RunID)
	if !hasEvent(events, event.TypeRunApprovalGranted) {
		t.Fatalf("missing approval_granted in %v", events)
	}
}

func TestApprovalFlowDeny(t *testing.T) {
	gate := openCap("fake.gate", `{"type":"object"}`)
	gate.ExecutionPolicy = string(policy.ApprovalRequired)
	p := &fakePlayer{
		name: "fake",
		caps: []registry.Capability{gate},
		fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) {
			t.Fatal("denied step must not execute")
			return nil, nil
		},
	}
	r, _, st := newTestRunner(t, p)

	res, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.gate", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	if err != nil {
		t.Fatal(err)
	}
	r.WaitIdle()
	waitForStatus(t, st, res.RunID, store.StatusWaitingApproval)

	if err := r.Approve(context.Background(), res.RunID, runner.DecisionDeny); err != nil {
		t.Fatal(err)
	}
	waitForStatus(t, st, res.RunID, store.StatusFailed)
	events := storedEvents(t, st, res.RunID)
	if !hasEvent(events, event.TypeRunApprovalDenied) || !hasEvent(events, event.TypeRunFailed) {
		t.Fatalf("events=%v", events)
	}
}

func TestApproveValidation(t *testing.T) {
	r, _, st := newTestRunner(t,
		&fakePlayer{name: "fake", caps: []registry.Capability{openCap("fake.echo", `{"type":"object"}`)},
			fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) { return json.RawMessage(`{}`), nil }},
	)
	if err := r.Approve(context.Background(), "missing-run", runner.DecisionGrant); err == nil {
		t.Fatal("unknown run must fail")
	} else if ve, ok := asResultErr(err); !ok || ve.Code != result.CodeNotFound {
		t.Fatalf("err=%v", err)
	}

	res, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.echo", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	if err != nil {
		t.Fatal(err)
	}
	r.WaitIdle()
	waitForStatus(t, st, res.RunID, store.StatusSucceeded)

	if err := r.Approve(context.Background(), res.RunID, runner.DecisionGrant); err == nil {
		t.Fatal("approve on succeeded run must fail")
	} else if ve, ok := asResultErr(err); !ok || ve.Code != result.CodePolicyNotWaiting {
		t.Fatalf("err=%v", err)
	}
	if err := r.Approve(context.Background(), res.RunID, "bogus"); err == nil {
		t.Fatal("bogus decision must fail")
	}
}

func TestStepRetryThenSucceed(t *testing.T) {
	calls := 0
	p := &fakePlayer{
		name: "fake",
		caps: []registry.Capability{openCap("fake.echo", `{"type":"object"}`)},
		fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) {
			calls++
			if calls == 1 {
				return nil, result.Runtime(result.CodePlayerError, "transient", true, nil)
			}
			return json.RawMessage(`{"ok":true}`), nil
		},
	}
	r, _, st := newTestRunner(t, p)
	retries := 1
	res, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.echo", Input: json.RawMessage(`{}`), DependsOn: []string{}, MaxRetries: &retries}))
	if err != nil {
		t.Fatal(err)
	}
	r.WaitIdle()
	waitForStatus(t, st, res.RunID, store.StatusSucceeded)
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
	payload := eventPayload(storedEvents(t, st, res.RunID), event.TypeStepSucceeded)
	if payload["attempts"] != float64(2) {
		t.Fatalf("attempts=%v", payload["attempts"])
	}
}

func TestNonRetryableFailureFailsRun(t *testing.T) {
	p := &fakePlayer{
		name: "fake",
		caps: []registry.Capability{openCap("fake.echo", `{"type":"object"}`)},
		fn: func(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) {
			return nil, result.Runtime(result.CodePlayerError, "boom", false, nil)
		},
	}
	r, _, st := newTestRunner(t, p)
	res, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.echo", Input: json.RawMessage(`{}`), DependsOn: []string{}}))
	if err != nil {
		t.Fatal(err)
	}
	r.WaitIdle()
	run := waitForStatus(t, st, res.RunID, store.StatusFailed)
	if !strings.Contains(run.ErrorJSON, "boom") {
		t.Fatalf("error_json=%s", run.ErrorJSON)
	}
	if !hasEvent(storedEvents(t, st, res.RunID), event.TypeRunFailed) {
		t.Fatal("missing run.failed event")
	}
}

func TestCancelBetweenSteps(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	var once sync.Once
	p := &fakePlayer{name: "fake", caps: []registry.Capability{
		openCap("fake.wait", `{"type":"object"}`),
		openCap("fake.echo", `{"type":"object"}`),
	}}
	p.fn = func(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
		if req.Capability == "fake.wait" {
			once.Do(func() { close(started) })
			<-release
		}
		return json.RawMessage(`{}`), nil
	}
	r, _, st := newTestRunner(t, p)

	res, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fake.wait", Input: json.RawMessage(`{}`), DependsOn: []string{}},
		task.Step{StepID: "s2", Capability: "fake.echo", Input: json.RawMessage(`{}`), DependsOn: []string{"s1"}}))
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("first step never started")
	}
	if err := r.Cancel(res.RunID); err != nil {
		t.Fatal(err)
	}
	close(release)
	r.WaitIdle()

	waitForStatus(t, st, res.RunID, store.StatusCancelled)
	if !hasEvent(storedEvents(t, st, res.RunID), event.TypeRunCancelled) {
		t.Fatal("missing run.cancelled event")
	}
	if err := r.Cancel(res.RunID); err == nil {
		t.Fatal("cancel on finished run must fail")
	}
}

func TestClaimsConflictFailsSecondRun(t *testing.T) {
	startedA := make(chan struct{})
	releaseA := make(chan struct{})
	var onceA sync.Once
	p := &fakePlayer{
		name: "fake-fs",
		caps: []registry.Capability{openCap("fs.write", `{"type":"object","required":["path"],"properties":{"path":{"type":"string"}}}`)},
	}
	p.fn = func(_ context.Context, req registry.ExecRequest) (json.RawMessage, error) {
		if strings.Contains(string(req.Input), `"data/out.txt"`) {
			onceA.Do(func() { close(startedA) })
			<-releaseA
		}
		return json.RawMessage(`{}`), nil
	}
	r, _, st := newTestRunner(t, p)
	r.Claims = claim.New(st)
	// The runner validates fs.write with the real filesystem player
	// rules: the parent directory must exist in the workspace.
	if err := os.MkdirAll(filepath.Join(r.Workspace, "data"), 0o755); err != nil {
		t.Fatal(err)
	}

	resA, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fs.write", Input: json.RawMessage(`{"path":"data/out.txt"}`), DependsOn: []string{}}))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-startedA:
	case <-time.After(2 * time.Second):
		t.Fatal("run A never started (claim not acquired?)")
	}

	resB, err := r.Submit(context.Background(), mkTask(t, "cli",
		task.Step{StepID: "s1", Capability: "fs.write", Input: json.RawMessage(`{"path":"data/out.txt"}`), DependsOn: []string{}}))
	if err != nil {
		t.Fatal(err)
	}
	waitForStatus(t, st, resB.RunID, store.StatusFailed)

	bEvents := storedEvents(t, st, resB.RunID)
	if !hasEvent(bEvents, event.TypeClaimConflict) {
		t.Fatalf("missing claim.conflict in %v", bEvents)
	}
	conflict := eventPayload(bEvents, event.TypeClaimConflict)
	if conflict["holder_run_id"] != resA.RunID {
		t.Fatalf("holder=%v want %s", conflict["holder_run_id"], resA.RunID)
	}

	close(releaseA)
	r.WaitIdle()
	waitForStatus(t, st, resA.RunID, store.StatusSucceeded)

	active, err := st.ListActiveClaims(context.Background())
	if err != nil || len(active) != 0 {
		t.Fatalf("active claims after success: %+v err=%v", active, err)
	}
}

func asErr(err error, ve *result.Error) bool {
	e, ok := err.(result.Error)
	if ok {
		*ve = e
	}
	return ok
}

func asResultErr(err error) (result.Error, bool) {
	e, ok := err.(result.Error)
	return e, ok
}
