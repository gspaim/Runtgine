package store_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/store"
)

func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestRunLifecycle(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)

	run := store.Run{RunID: "run-1", TaskID: "task-1", Status: store.StatusAccepted}
	if err := st.InsertRun(ctx, run, []byte(`{"task":"ir"}`)); err != nil {
		t.Fatal(err)
	}
	got, taskJSON, err := st.GetRun(ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.StatusAccepted || got.TaskID != "task-1" {
		t.Fatalf("run=%+v", got)
	}
	if string(taskJSON) != `{"task":"ir"}` {
		t.Fatalf("task_json=%s", taskJSON)
	}

	if err := st.UpdateRunStatus(ctx, "run-1", store.StatusRunning, ""); err != nil {
		t.Fatal(err)
	}
	got, _, _ = st.GetRun(ctx, "run-1")
	if got.Status != store.StatusRunning {
		t.Fatalf("status=%s", got.Status)
	}

	if _, _, err := st.GetRun(ctx, "missing"); !store.ErrNotFound(err) {
		t.Fatalf("missing run must be ErrNotFound, got %v", err)
	}
}

func TestPendingApprovalRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	if err := st.InsertRun(ctx, store.Run{RunID: "run-1", TaskID: "task-1", Status: store.StatusRunning}, nil); err != nil {
		t.Fatal(err)
	}
	if err := st.SetPendingApproval(ctx, "run-1", "s1", "fake.gate", "fake"); err != nil {
		t.Fatal(err)
	}
	got, _, _ := st.GetRun(ctx, "run-1")
	if got.Status != store.StatusWaitingApproval {
		t.Fatalf("status=%s", got.Status)
	}
	if got.PendingStepID != "s1" || got.PendingCapability != "fake.gate" || got.PendingPlayer != "fake" {
		t.Fatalf("pending=%+v", got)
	}
	if err := st.ClearPendingApproval(ctx, "run-1"); err != nil {
		t.Fatal(err)
	}
	got, _, _ = st.GetRun(ctx, "run-1")
	if got.PendingStepID != "" || got.PendingCapability != "" || got.PendingPlayer != "" {
		t.Fatalf("pending not cleared: %+v", got)
	}
}

func TestListRunsOrderedAndLimited(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	for _, id := range []string{"run-1", "run-2", "run-3"} {
		if err := st.InsertRun(ctx, store.Run{RunID: id, TaskID: "t", Status: store.StatusAccepted}, []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond) // distinct created_at ordering
	}
	runs, err := st.ListRuns(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Fatalf("limit ignored: %d", len(runs))
	}
	if runs[0].RunID != "run-3" {
		t.Fatalf("newest first expected run-3, got %s", runs[0].RunID)
	}
	if runs[0].TaskJSON == nil {
		t.Fatal("task_json must round-trip")
	}
}

func TestEventsRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	if err := st.InsertRun(ctx, store.Run{RunID: "run-1", TaskID: "task-1", Status: store.StatusAccepted}, nil); err != nil {
		t.Fatal(err)
	}

	step := "s1"
	e1, err := event.New(event.TypeRunStarted, "run-1", "task-1", nil, map[string]any{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
	e2, err := event.New(event.TypeStepStarted, "run-1", "task-1", &step, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AppendEvent(ctx, e1); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := st.AppendEvent(ctx, e2); err != nil {
		t.Fatal(err)
	}

	events, err := st.ListEvents(ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("events=%d", len(events))
	}
	if events[0].Type != event.TypeRunStarted || events[1].Type != event.TypeStepStarted {
		t.Fatalf("order=[%s %s]", events[0].Type, events[1].Type)
	}
	if events[0].Payload["k"] != "v" {
		t.Fatalf("payload=%v", events[0].Payload)
	}
	if events[1].StepID == nil || *events[1].StepID != "s1" {
		t.Fatalf("step_id=%v", events[1].StepID)
	}

	recent, err := st.ListRecentEvents(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 1 || recent[0].Type != event.TypeStepStarted {
		t.Fatalf("recent=%+v", recent)
	}
}

func TestStepOutputsUpsert(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	if err := st.SaveStepOutput(ctx, "run-1", "s1", "fake.echo", json.RawMessage(`{"n":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveStepOutput(ctx, "run-1", "s1", "fake.echo", json.RawMessage(`{"n":2}`)); err != nil {
		t.Fatal(err)
	}
	outs, err := st.ListStepOutputs(ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(outs) != 1 || outs[0].Capability != "fake.echo" || string(outs[0].Output) != `{"n":2}` {
		t.Fatalf("outputs=%+v", outs)
	}
}

func TestSubtasksAndChildRuns(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	if err := st.InsertRun(ctx, store.Run{RunID: "parent", TaskID: "t", Status: store.StatusAccepted}, nil); err != nil {
		t.Fatal(err)
	}
	if err := st.InsertRun(ctx, store.Run{RunID: "child", TaskID: "t", ParentRunID: "parent", Status: store.StatusSucceeded}, nil); err != nil {
		t.Fatal(err)
	}
	stt := store.Subtask{SubtaskID: "st-1", ParentRunID: "parent", TaskID: "t", Summary: "do x", SuggestedCapability: "fake.a"}
	if err := st.InsertSubtask(ctx, stt); err != nil {
		t.Fatal(err)
	}
	if err := st.SetSubtaskChildRun(ctx, "st-1", "child"); err != nil {
		t.Fatal(err)
	}
	subs, err := st.ListSubtasks(ctx, "parent")
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 1 || subs[0].ChildRunID != "child" {
		t.Fatalf("subtasks=%+v", subs)
	}
	children, err := st.ListChildRuns(ctx, "parent")
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 1 || children[0].RunID != "child" || children[0].Status != store.StatusSucceeded {
		t.Fatalf("children=%+v", children)
	}
}

func TestGraphNodesEdgesAndNeighbors(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	if err := st.UpsertGraphNode(ctx, "player", "fake", `{"kind":"deterministic"}`); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertGraphNode(ctx, "capability", "fake.echo", `{"policy":"allow"}`); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertGraphEdge(ctx, "provides", "player", "fake", "capability", "fake.echo", ""); err != nil {
		t.Fatal(err)
	}
	node, err := st.GetGraphNode(ctx, "player", "fake")
	if err != nil || node.AttrsJSON != `{"kind":"deterministic"}` {
		t.Fatalf("node=%+v err=%v", node, err)
	}

	out, err := st.QueryGraphNeighbors(ctx, "player", "fake", "", "out")
	if err != nil || len(out) != 1 || out[0].ID != "fake.echo" {
		t.Fatalf("out=%+v err=%v", out, err)
	}
	in, err := st.QueryGraphNeighbors(ctx, "capability", "fake.echo", "provides", "in")
	if err != nil || len(in) != 1 || in[0].ID != "fake" {
		t.Fatalf("in=%+v err=%v", in, err)
	}
	none, err := st.QueryGraphNeighbors(ctx, "capability", "fake.echo", "mentions", "in")
	if err != nil || len(none) != 0 {
		t.Fatalf("edge-kind filter failed: %+v err=%v", none, err)
	}
	nodes, err := st.ListGraphNodes(ctx)
	if err != nil || len(nodes) != 2 {
		t.Fatalf("nodes=%+v err=%v", nodes, err)
	}
	edges, err := st.ListGraphEdges(ctx)
	if err != nil || len(edges) != 1 || edges[0].Kind != "provides" || edges[0].AttrsJSON != "{}" {
		t.Fatalf("edges=%+v err=%v", edges, err)
	}
}

func TestClaimsLifecycle(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	if err := st.InsertRun(ctx, store.Run{RunID: "run-a", TaskID: "t", Status: store.StatusRunning}, nil); err != nil {
		t.Fatal(err)
	}
	if err := st.InsertRun(ctx, store.Run{RunID: "run-b", TaskID: "t", Status: store.StatusRunning}, nil); err != nil {
		t.Fatal(err)
	}

	holder, err := st.TryAcquireClaim(ctx, "run-a", "s1", "path", "data/out.txt", nil)
	if err != nil || holder != "" {
		t.Fatalf("first acquire holder=%q err=%v", holder, err)
	}
	// Same-run re-acquire is idempotent.
	holder, err = st.TryAcquireClaim(ctx, "run-a", "s1", "path", "data/out.txt", nil)
	if err != nil || holder != "" {
		t.Fatalf("re-acquire holder=%q err=%v", holder, err)
	}
	// Different run on an overlapping path conflicts.
	holder, err = st.TryAcquireClaim(ctx, "run-b", "s1", "path", "data/out.txt", func(aKind, aKey, bKind, bKey string) bool {
		return aKind == bKind && aKey == bKey
	})
	if err != nil || holder != "run-a" {
		t.Fatalf("conflict holder=%q err=%v", holder, err)
	}

	active, err := st.ListActiveClaims(ctx)
	if err != nil || len(active) != 1 || active[0].RunID != "run-a" {
		t.Fatalf("active=%+v err=%v", active, err)
	}

	released, err := st.ReleaseClaimsForRun(ctx, "run-a")
	if err != nil || len(released) != 1 || released[0].Key != "data/out.txt" {
		t.Fatalf("released=%+v err=%v", released, err)
	}
	active, _ = st.ListActiveClaims(ctx)
	if len(active) != 0 {
		t.Fatalf("claims not released: %+v", active)
	}
}

func TestSweepOrphanClaims(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	if err := st.InsertRun(ctx, store.Run{RunID: "run-live", TaskID: "t", Status: store.StatusRunning}, nil); err != nil {
		t.Fatal(err)
	}
	if err := st.InsertRun(ctx, store.Run{RunID: "run-dead", TaskID: "t", Status: store.StatusSucceeded}, nil); err != nil {
		t.Fatal(err)
	}
	strict := func(aKind, aKey, bKind, bKey string) bool {
		return aKind == bKind && aKey == bKey
	}
	if _, err := st.TryAcquireClaim(ctx, "run-live", "s1", "path", "live/out.txt", strict); err != nil {
		t.Fatal(err)
	}
	if _, err := st.TryAcquireClaim(ctx, "run-dead", "s1", "path", "dead/out.txt", strict); err != nil {
		t.Fatal(err)
	}
	n, err := st.SweepOrphanClaims(ctx)
	if err != nil || n != 1 {
		t.Fatalf("swept=%d err=%v", n, err)
	}
}

func TestHelpers(t *testing.T) {
	if store.ErrNotFound(nil) {
		t.Fatal("nil must not be ErrNotFound")
	}
	if got := store.FormatErr(nil); got != "" {
		t.Fatalf("FormatErr(nil)=%q", got)
	}
	if got := store.FormatErr(context.DeadlineExceeded); got == "" {
		t.Fatal("FormatErr must serialize non-nil errors")
	}
	if got := store.MustJSON(map[string]int{"a": 1}); got != `{"a":1}` {
		t.Fatalf("MustJSON=%s", got)
	}
	st := openStore(t)
	if err := st.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}
