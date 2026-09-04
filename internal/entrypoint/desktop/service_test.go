package desktop

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/graph"
	"github.com/gspaim/Runtgine/internal/core/lessons"
	"github.com/gspaim/Runtgine/internal/core/memory"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/task"
)

type fakeCore struct {
	compileCalls int
	submitted    int
	compileErr   error
	submitErr    error
	lastEP       string
	stream       chan event.Event
	runs         []api.RunSummary
	snapshot     api.RunSnapshot
	events       []event.Event
	config       api.ConfigSnapshot
	graph        graph.Snapshot
	graphRefresh int
	lessonRows   []lessons.Proposal
}

func (f *fakeCore) CompileIntent(_ context.Context, text, ep, _ string) (task.Task, string, error) {
	if f.compileErr != nil {
		return task.Task{}, "", f.compileErr
	}
	f.compileCalls++
	f.lastEP = ep
	capName := "shell.exec"
	method := "heuristic.shell"
	if strings.Contains(strings.ToLower(text), "git status") {
		capName = "git.status"
		method = "heuristic.git"
	}
	return task.Task{
		SchemaVersion: "0.1.0",
		Intent:        task.Intent{Summary: strings.TrimSpace(text)},
		Source:        task.Source{EntryPoint: ep},
		Steps:         []task.Step{{StepID: "s1", Capability: capName, Input: json.RawMessage(`{"workdir":"."}`)}},
	}, method, nil
}

func (f *fakeCore) SubmitIntent(ctx context.Context, text, ep, ref string) (string, string, error) {
	if f.submitErr != nil {
		return "", "", f.submitErr
	}
	tk, method, err := f.CompileIntent(ctx, text, ep, ref)
	if err != nil {
		return "", method, err
	}
	id, err := f.SubmitTask(ctx, tk)
	return id, method, err
}

func (f *fakeCore) SubmitTask(_ context.Context, tk task.Task) (string, error) {
	if f.submitErr != nil {
		return "", f.submitErr
	}
	f.submitted++
	f.lastEP = tk.Source.EntryPoint
	id := "run-wails-1"
	f.snapshot = api.RunSnapshot{RunID: id, Status: "running"}
	f.runs = []api.RunSummary{{RunID: id, Status: "running", Summary: tk.Intent.Summary, Source: tk.Source.EntryPoint}}
	return id, nil
}

func (f *fakeCore) GetRun(_ context.Context, runID string) (api.RunSnapshot, error) {
	if f.snapshot.RunID == runID {
		return f.snapshot, nil
	}
	return api.RunSnapshot{}, result.Runtime(result.CodeNotFound, "run not found", false, nil)
}

func (f *fakeCore) ListRuns(context.Context, int) ([]api.RunSummary, error) {
	return f.runs, nil
}

func (f *fakeCore) Subscribe(int) (<-chan event.Event, func()) {
	if f.stream == nil {
		f.stream = make(chan event.Event, 4)
	}
	return f.stream, func() {}
}

func (f *fakeCore) CancelRun(string) error { return nil }

func (f *fakeCore) ApproveRun(string, string) error { return nil }

func (f *fakeCore) ListRecentEvents(context.Context, int) ([]event.Event, error) {
	return f.events, nil
}

func (f *fakeCore) ConfigSnapshot() api.ConfigSnapshot {
	return f.config
}

func (f *fakeCore) GetGraphSnapshot(context.Context) (graph.Snapshot, error) {
	return f.graph, nil
}

func (f *fakeCore) RefreshGraph(context.Context) error {
	f.graphRefresh++
	return nil
}

func (f *fakeCore) ListLessons(_ context.Context, status string) ([]lessons.Proposal, error) {
	if status == "" {
		return f.lessonRows, nil
	}
	out := make([]lessons.Proposal, 0)
	for _, row := range f.lessonRows {
		if row.Status == status {
			out = append(out, row)
		}
	}
	return out, nil
}

func (f *fakeCore) ApproveLesson(_ context.Context, id string) (memory.Episode, error) {
	for i, row := range f.lessonRows {
		if row.ID == id {
			if row.Status != lessons.StatusPending {
				return memory.Episode{}, result.Runtime(result.CodeInternal, "lesson is not pending", false, nil)
			}
			f.lessonRows[i].Status = lessons.StatusApproved
			return memory.Episode{ID: "ep-" + id, Title: row.Title}, nil
		}
	}
	return memory.Episode{}, result.Runtime(result.CodeNotFound, "lesson not found", false, nil)
}

func (f *fakeCore) RejectLesson(_ context.Context, id string) error {
	for i, row := range f.lessonRows {
		if row.ID == id {
			if row.Status != lessons.StatusPending {
				return result.Runtime(result.CodeInternal, "lesson is not pending", false, nil)
			}
			f.lessonRows[i].Status = lessons.StatusRejected
			return nil
		}
	}
	return result.Runtime(result.CodeNotFound, "lesson not found", false, nil)
}

type recEmit struct {
	name string
	n    int
}

func (r *recEmit) Emit(name string, _ any) {
	r.name = name
	r.n++
}

func TestPreviewGitStatusDoesNotSubmit(t *testing.T) {
	core := &fakeCore{}
	svc := NewService(core)
	prev, err := svc.CompileIntent("git status")
	if err != nil {
		t.Fatal(err)
	}
	if core.compileCalls != 1 || core.submitted != 0 {
		t.Fatalf("compile=%d submitted=%d", core.compileCalls, core.submitted)
	}
	if core.lastEP != "wails" {
		t.Fatalf("entry_point=%s", core.lastEP)
	}
	if prev.Method != "heuristic.git" {
		t.Fatalf("method=%s", prev.Method)
	}
	if !strings.Contains(string(prev.Task), "git.status") {
		t.Fatalf("task=%s", prev.Task)
	}
}

func TestSubmitIntentSetsWailsSource(t *testing.T) {
	core := &fakeCore{}
	svc := NewService(core)
	out, err := svc.SubmitIntent("git status")
	if err != nil {
		t.Fatal(err)
	}
	if out.RunID != "run-wails-1" {
		t.Fatalf("run_id=%s", out.RunID)
	}
	if core.submitted != 1 {
		t.Fatalf("submitted=%d", core.submitted)
	}
	if core.lastEP != "wails" {
		t.Fatalf("entry_point=%s", core.lastEP)
	}
}

func TestUnknownCapabilitySurfaces(t *testing.T) {
	core := &fakeCore{compileErr: result.Validation(result.CodeUnknownCapability, "no such capability", nil)}
	svc := NewService(core)
	_, err := svc.CompileIntent("teleport the repo")
	if err == nil {
		t.Fatal("expected error")
	}
	var re result.Error
	if !errorsAs(err, &re) || re.Code != result.CodeUnknownCapability {
		t.Fatalf("err=%v", err)
	}
	if core.submitted != 0 {
		t.Fatal("must not insert a Run")
	}
}

func TestJSONPreviewDoesNotSubmit(t *testing.T) {
	core := &fakeCore{}
	svc := NewService(core)
	raw := `{
	  "schema_version":"0.1.0",
	  "source":{"entry_point":"wails"},
	  "intent":{"summary":"git status"},
	  "steps":[{"step_id":"s1","capability":"git.status","input":{"workdir":"."}}]
	}`
	prev, err := svc.CompileTaskJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if core.submitted != 0 {
		t.Fatal("preview must not submit")
	}
	if prev.Method != "json" {
		t.Fatalf("method=%s", prev.Method)
	}
}

func TestEventForward(t *testing.T) {
	core := &fakeCore{}
	svc := NewService(core)
	rec := &recEmit{}
	svc.SetEmitter(rec)
	svc.Start()
	defer svc.Stop()
	core.stream <- event.Event{Type: event.TypeRunStarted, RunID: "run-wails-1"}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if rec.n > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if rec.n == 0 || rec.name != eventName {
		t.Fatalf("emits=%d name=%s", rec.n, rec.name)
	}
}

func TestListBoardRunsFiltersSource(t *testing.T) {
	core := &fakeCore{runs: []api.RunSummary{
		{RunID: "run-wails-1", Status: "succeeded", Source: "wails", Summary: "git status"},
		{RunID: "run-board-1", Status: "running", Source: "board", Summary: "from card"},
		{RunID: "run-tui-1", Status: "failed", Source: "tui", Summary: "ls"},
	}}
	svc := NewService(core)
	rows, err := svc.ListBoardRuns(20)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].RunID != "run-board-1" {
		t.Fatalf("board runs=%v", rows)
	}
}

func TestConfigSnapshotOmitsSecrets(t *testing.T) {
	core := &fakeCore{config: api.ConfigSnapshot{
		WorkspaceRoot:     "/tmp/ws",
		DBPath:            "/tmp/ws/.runtgine/runtgine.db",
		LogLevel:          "info",
		MaxConcurrentRuns: 2,
		LLMBackend:        "openai",
		LLMConnected:      true,
		GitHubConnected:   false,
		Precedence:        "defaults < config.json < env < CLI flags",
	}}
	svc := NewService(core)
	snap := svc.ConfigSnapshot()
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for k, v := range fields {
		lk := strings.ToLower(k)
		if strings.Contains(lk, "token") || strings.Contains(lk, "secret") || strings.HasSuffix(lk, "_key") || lk == "key" {
			t.Fatalf("secret-like field %s=%v", k, v)
		}
	}
	if _, ok := fields["llm_connected"]; !ok {
		t.Fatalf("expected llm_connected bool, json=%s", raw)
	}
	if strings.Contains(string(raw), "sk-") || strings.Contains(strings.ToLower(string(raw)), "ghp_") {
		t.Fatalf("looks like a leaked credential: %s", raw)
	}
}

func TestEventsGraphLessonsWired(t *testing.T) {
	core := &fakeCore{
		events: []event.Event{{Type: event.TypeRunStarted, RunID: "run-wails-1"}},
		graph:  graph.Snapshot{Nodes: []graph.Node{{Kind: "capability", ID: "git.status"}}},
		lessonRows: []lessons.Proposal{
			{ID: "les-1", Title: "Failure: git status", Status: lessons.StatusPending},
		},
	}
	svc := NewService(core)
	evs, err := svc.ListRecentEvents(10)
	if err != nil || len(evs) != 1 {
		t.Fatalf("events=%v err=%v", evs, err)
	}
	snap, err := svc.GetGraphSnapshot()
	if err != nil || len(snap.Nodes) != 1 {
		t.Fatalf("graph=%v err=%v", snap, err)
	}
	if err := svc.RefreshGraph(); err != nil || core.graphRefresh != 1 {
		t.Fatalf("refresh err=%v n=%d", err, core.graphRefresh)
	}
	pending, err := svc.ListLessons(lessons.StatusPending)
	if err != nil || len(pending) != 1 {
		t.Fatalf("lessons=%v err=%v", pending, err)
	}
	if err := svc.ApproveLesson("les-1"); err != nil {
		t.Fatal(err)
	}
	if core.lessonRows[0].Status != lessons.StatusApproved {
		t.Fatalf("status=%s", core.lessonRows[0].Status)
	}
	core.lessonRows[0].Status = lessons.StatusPending
	if err := svc.RejectLesson("les-1"); err != nil {
		t.Fatal(err)
	}
	if core.lessonRows[0].Status != lessons.StatusRejected {
		t.Fatalf("status=%s", core.lessonRows[0].Status)
	}
}

func errorsAs(err error, target *result.Error) bool {
	if err == nil {
		return false
	}
	re, ok := err.(result.Error)
	if !ok {
		return false
	}
	*target = re
	return true
}
