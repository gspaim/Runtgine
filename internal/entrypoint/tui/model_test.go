package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/blast"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/graph"
	"github.com/gspaim/Runtgine/internal/core/task"
)

type fakeCore struct {
	runs         []api.RunSummary
	snapshot     api.RunSnapshot
	events       []event.Event
	config       api.ConfigSnapshot
	graph        graph.Snapshot
	stream       chan event.Event
	cancelled    string
	approved     string
	refreshes    int
	refreshErr   error
	snapErr      error
	compileErr   error
	submitErr    error
	compileCalls int
	submitted    int
	hits         graph.Hits
	hitCalls     int
	blastRep     blast.Report
	blastErr     error
	blastCalls   int
}

func (f *fakeCore) ListRuns(context.Context, int) ([]api.RunSummary, error) {
	return f.runs, nil
}

func (f *fakeCore) GetRun(context.Context, string) (api.RunSnapshot, error) {
	return f.snapshot, nil
}

func (f *fakeCore) ListRecentEvents(context.Context, int) ([]event.Event, error) {
	return f.events, nil
}

func (f *fakeCore) ConfigSnapshot() api.ConfigSnapshot {
	return f.config
}

func (f *fakeCore) Subscribe(int) (<-chan event.Event, func()) {
	if f.stream == nil {
		f.stream = make(chan event.Event)
	}
	return f.stream, func() {}
}

func (f *fakeCore) CancelRun(runID string) error {
	f.cancelled = runID
	return nil
}

func (f *fakeCore) ApproveRun(runID, decision string) error {
	f.approved = runID + ":" + decision
	return nil
}

func (f *fakeCore) GetGraphSnapshot(context.Context) (graph.Snapshot, error) {
	if f.snapErr != nil {
		return graph.Snapshot{}, f.snapErr
	}
	return f.graph, nil
}

func (f *fakeCore) RefreshGraph(context.Context) error {
	f.refreshes++
	return f.refreshErr
}

func (f *fakeCore) CompileIntent(_ context.Context, text, _, _ string) (task.Task, string, error) {
	if f.compileErr != nil {
		return task.Task{}, "", f.compileErr
	}
	summary := strings.TrimSpace(text)
	if summary == "" {
		summary = "empty"
	}
	capName := "shell.exec"
	method := "heuristic.shell"
	lower := strings.ToLower(summary)
	if strings.Contains(lower, "git status") {
		capName = "git.status"
		method = "heuristic.git"
	}
	f.compileCalls++
	return task.Task{
		SchemaVersion: "0.1.0",
		Intent:        task.Intent{Summary: summary},
		Source:        task.Source{EntryPoint: "tui"},
		Steps:         []task.Step{{StepID: "s1", Capability: capName, Input: json.RawMessage(`{}`)}},
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
	f.submitted++
	id, _ := f.recordSubmit(tk)
	return id, method, nil
}

func (f *fakeCore) SubmitTask(_ context.Context, tk task.Task) (string, error) {
	if f.submitErr != nil {
		return "", f.submitErr
	}
	f.submitted++
	id, _ := f.recordSubmit(tk)
	return id, nil
}

func (f *fakeCore) QueryHits(_ context.Context, _ graph.Query) graph.Hits {
	f.hitCalls++
	if f.hits.Items == nil {
		return graph.Hits{Items: []graph.Hit{}}
	}
	return f.hits
}

func (f *fakeCore) BlastTask(_ context.Context, _ task.Task) (blast.Report, error) {
	f.blastCalls++
	if f.blastErr != nil {
		return blast.Report{}, f.blastErr
	}
	if f.blastRep.Risk == "" {
		return blast.Report{Risk: blast.RiskNone, Touches: []blast.Touch{}, Conflicts: []blast.Conflict{}, Affected: []blast.Affected{}}, nil
	}
	return f.blastRep, nil
}

func (f *fakeCore) recordSubmit(tk task.Task) (string, string) {
	id := fmt.Sprintf("run-%d", f.submitted)
	now := time.Now().UTC()
	raw, _ := json.Marshal(tk)
	f.runs = append([]api.RunSummary{{
		RunID: id, Status: "running", Summary: tk.Intent.Summary,
		Source: "tui", CreatedAt: now, UpdatedAt: now,
	}}, f.runs...)
	f.snapshot = api.RunSnapshot{RunID: id, Status: "running", Task: raw}
	return id, "json"
}

func fixture() *fakeCore {
	now := time.Now().UTC()
	taskJSON := json.RawMessage(`{
		"source":{"entry_point":"board","ref":"owner/repo#42"},
		"intent":{"summary":"Review orbital telemetry"},
		"steps":[
			{"step_id":"review","capability":"pipeline.tech-review","input":{}},
			{"step_id":"search","capability":"pipeline.repo-search","input":{},"depends_on":["review"]}
		]
	}`)
	step := "review"
	events := []event.Event{{
		Type: event.TypeStepStarted, TS: now, RunID: "019c-run-0001",
		StepID: &step, Payload: map[string]any{"player": "llm"},
	}}
	return &fakeCore{
		runs: []api.RunSummary{{
			RunID: "019c-run-0001", TaskID: "019c-task-0001", Status: "running",
			Summary: "Review orbital telemetry", Source: "board", SourceRef: "owner/repo#42",
			CreatedAt: now.Add(-time.Minute), UpdatedAt: now,
		}},
		snapshot: api.RunSnapshot{
			RunID: "019c-run-0001", TaskID: "019c-task-0001", Status: "running",
			Task: taskJSON, Events: events,
		},
		events: events,
		config: api.ConfigSnapshot{
			WorkspaceRoot: "/workspace", DBPath: "/workspace/.runtgine/runtgine.db",
			LogLevel: "info", MaxConcurrentRuns: 4, LLMBackend: "anthropic",
			LLMConnected: true, GitHubConnected: true,
			Precedence: "defaults < config.json < env < CLI flags",
		},
		graph: graph.Snapshot{
			Nodes: []graph.Node{
				{Kind: graph.KindPlayer, ID: "shell", Attrs: map[string]any{"version": "0.1.0"}},
				{Kind: graph.KindPlayer, ID: "git"},
				{Kind: graph.KindCapability, ID: "shell.exec"},
				{Kind: graph.KindCapability, ID: "git.status"},
			},
			Edges: []graph.Edge{
				{
					Kind: graph.EdgeProvides,
					From: graph.Ref{Kind: graph.KindPlayer, ID: "shell"},
					To:   graph.Ref{Kind: graph.KindCapability, ID: "shell.exec"},
				},
				{
					Kind: graph.EdgeProvides,
					From: graph.Ref{Kind: graph.KindPlayer, ID: "git"},
					To:   graph.Ref{Kind: graph.KindCapability, ID: "git.status"},
				},
			},
		},
	}
}

func loadedModel(t *testing.T) (Model, *fakeCore) {
	t.Helper()
	core := fixture()
	model, _ := New(core)
	msg := model.refreshCmd()()
	updated, _ := model.Update(msg)
	model = updated.(Model)
	gmsg := model.graphLoadCmd(false)()
	updated, _ = model.Update(gmsg)
	return updated.(Model), core
}

func TestResponsiveViewsRenderAllTabs(t *testing.T) {
	model, _ := loadedModel(t)
	for _, width := range []int{60, 100, 140} {
		for tab := 0; tab < tabCount; tab++ {
			model.width, model.height, model.tab = width, 30, tab
			content := model.View().Content
			if strings.TrimSpace(content) == "" {
				t.Fatalf("empty view at width %d tab %d", width, tab)
			}
		}
	}
}

func TestNavigationAndCancelConfirmation(t *testing.T) {
	model, core := loadedModel(t)
	model.tab = tabRuns

	updated, _ := model.Update(pressKey("enter"))
	model = updated.(Model)
	if model.tab != tabLive {
		t.Fatalf("enter selected tab %d, want LIVE", model.tab)
	}

	updated, cmd := model.Update(pressKey("c"))
	model = updated.(Model)
	if !model.confirm || cmd != nil {
		t.Fatal("first c must request confirmation")
	}

	updated, cmd = model.Update(pressKey("c"))
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("second c must issue cancellation")
	}
	cancelResult := cmd()
	updated, _ = model.Update(cancelResult)
	model = updated.(Model)
	if core.cancelled != "019c-run-0001" || model.confirm {
		t.Fatalf("cancelled=%q confirm=%v", core.cancelled, model.confirm)
	}
}

func TestApproveDenyKeysOnWaitingRun(t *testing.T) {
	model, core := loadedModel(t)
	core.runs[0].Status = "waiting_approval"
	core.snapshot.Status = "waiting_approval"
	core.snapshot.PendingApproval = &api.PendingApproval{StepID: "s1", Capability: "shell.exec", Player: "shell"}
	model.runs = core.runs
	model.snapshot = core.snapshot
	model.tab = tabRuns

	updated, cmd := model.Update(pressKey("a"))
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("a must approve")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)
	if core.approved != "019c-run-0001:grant" {
		t.Fatalf("approved=%q", core.approved)
	}

	updated, cmd = model.Update(pressKey("d"))
	if cmd == nil {
		t.Fatal("d must deny")
	}
	_ = cmd()
}

func TestNoColorAndConfigMasksSecrets(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("OPENAI_API_KEY", "super-secret-value")
	model, _ := loadedModel(t)
	model.tab = tabConfig
	content := model.View().Content

	if strings.Contains(content, "super-secret-value") {
		t.Fatal("config view leaked a secret")
	}
	if strings.Contains(content, "\x1b[38;") || strings.Contains(content, "\x1b[48;") {
		t.Fatalf("NO_COLOR view contains ANSI color sequences: %q", content)
	}
	if !strings.Contains(content, "secret masked") {
		t.Fatal("config view must state that configured secrets are masked")
	}
}

func TestEventsFilterInput(t *testing.T) {
	model, _ := loadedModel(t)
	model.tab = tabEvents

	updated, _ := model.Update(pressKey("/"))
	model = updated.(Model)
	if !model.filtering {
		t.Fatal("/ should enable event filtering")
	}
	updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Text: "step", Code: 's'}))
	model = updated.(Model)
	if model.filter != "step" {
		t.Fatalf("filter=%q, want step", model.filter)
	}
}

func TestTabCycleIncludesGraph(t *testing.T) {
	model, _ := loadedModel(t)
	model.tab = tabEvents

	updated, _ := model.Update(pressKey("tab"))
	model = updated.(Model)
	if model.tab != tabGraph {
		t.Fatalf("tab from EVENTS got %s, want GRAPH", tabNames[model.tab])
	}
	updated, _ = model.Update(pressKey("tab"))
	model = updated.(Model)
	if model.tab != tabConfig {
		t.Fatalf("tab from GRAPH got %s, want CONFIG", tabNames[model.tab])
	}
	updated, _ = model.Update(pressKey("shift+tab"))
	model = updated.(Model)
	if model.tab != tabGraph {
		t.Fatalf("shift+tab from CONFIG got %s, want GRAPH", tabNames[model.tab])
	}
}

func TestGraphListsShellProvides(t *testing.T) {
	model, _ := loadedModel(t)
	model.tab = tabGraph
	model.width, model.height = 140, 30
	selectGraphNode(t, &model, graph.KindPlayer, "shell")
	content := model.View().Content
	if !strings.Contains(content, "GRAPH") {
		t.Fatal("GRAPH tab missing title")
	}
	if !strings.Contains(content, "player") || !strings.Contains(content, "shell") {
		t.Fatalf("GRAPH missing player/shell: %s", content)
	}
	if !strings.Contains(content, "provides") || !strings.Contains(content, "shell.exec") {
		t.Fatalf("GRAPH detail missing provides shell.exec: %s", content)
	}
}

func TestGraphFilterHidesPlayers(t *testing.T) {
	model, _ := loadedModel(t)
	model.tab = tabGraph
	model.width = 140

	updated, _ := model.Update(pressKey("/"))
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Text: "capability", Code: 'c'}))
	model = updated.(Model)
	if model.graphFilter != "capability" {
		t.Fatalf("graphFilter=%q", model.graphFilter)
	}
	if model.filter != "" {
		t.Fatalf("EVENTS filter leaked into GRAPH: %q", model.filter)
	}
	nodes := model.filteredGraphNodes()
	if len(nodes) == 0 {
		t.Fatal("filter capability hid every node")
	}
	for _, n := range nodes {
		if n.Kind != graph.KindCapability {
			t.Fatalf("filter leaked kind %s id %s", n.Kind, n.ID)
		}
	}
	content := model.View().Content
	if strings.Contains(content, "player         git") || strings.Contains(content, "player         shell") {
		t.Fatalf("filtered GRAPH still lists players: %s", content)
	}
}

func TestGraphRefreshCallsCore(t *testing.T) {
	model, core := loadedModel(t)
	model.tab = tabGraph
	if core.refreshes != 0 {
		t.Fatalf("snapshot load must not RefreshGraph, got %d", core.refreshes)
	}
	updated, cmd := model.Update(pressKey("r"))
	if cmd == nil {
		t.Fatal("r on GRAPH must refresh")
	}
	msg := cmd()
	updated, _ = updated.Update(msg)
	model = updated.(Model)
	if core.refreshes != 1 {
		t.Fatalf("refreshes=%d, want 1", core.refreshes)
	}
	if !model.graphLoaded {
		t.Fatal("GRAPH snapshot not loaded after r")
	}
}

func TestGraphNarrowAndNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("RUNTGINE_ASCII", "1")
	model, _ := loadedModel(t)
	model.theme = DetectTheme()
	model.tab = tabGraph
	model.width, model.height = 70, 24
	content := model.View().Content
	if strings.TrimSpace(content) == "" {
		t.Fatal("narrow GRAPH view empty")
	}
	if !strings.Contains(content, "GRAPH") {
		t.Fatal("narrow GRAPH missing title")
	}
	if !strings.Contains(content, "player") || !strings.Contains(content, "shell") {
		t.Fatalf("narrow GRAPH missing kind+id: %s", content)
	}
	if strings.Contains(content, " --- ") {
		t.Fatal("narrow GRAPH drew a horizontal edge diagram")
	}
	if strings.Contains(content, "\x1b[38;") || strings.Contains(content, "\x1b[48;") {
		t.Fatalf("NO_COLOR GRAPH contains ANSI color: %q", content)
	}
}

func selectGraphNode(t *testing.T, model *Model, kind, id string) {
	t.Helper()
	nodes := model.filteredGraphNodes()
	for i, n := range nodes {
		if n.Kind == kind && n.ID == id {
			model.graphSelected = i
			return
		}
	}
	t.Fatalf("node %s %s not in GRAPH list", kind, id)
}

func pressKey(value string) tea.KeyPressMsg {
	switch value {
	case "enter":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	case "tab":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	case "shift+tab":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift})
	case "esc":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})
	case "backspace":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace})
	case "ctrl+p":
		return tea.KeyPressMsg(tea.Key{Code: 'p', Mod: tea.ModCtrl})
	case "ctrl+enter":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter, Mod: tea.ModCtrl})
	case "ctrl+j":
		return tea.KeyPressMsg(tea.Key{Code: 'j', Mod: tea.ModCtrl})
	case "ctrl+b":
		return tea.KeyPressMsg(tea.Key{Code: 'b', Mod: tea.ModCtrl})
	default:
		r := []rune(value)[0]
		return tea.KeyPressMsg(tea.Key{Text: value, Code: r})
	}
}

func TestIntentIsFirstTab(t *testing.T) {
	model, _ := loadedModel(t)
	if model.tab != tabIntent || tabNames[0] != "INTENT" {
		t.Fatalf("first tab=%s", tabNames[model.tab])
	}
	if !strings.Contains(model.View().Content, "INTENT") {
		t.Fatal("INTENT missing from default view")
	}
	updated, _ := model.Update(pressKey("shift+tab"))
	model = updated.(Model)
	if model.tab != tabConfig {
		t.Fatalf("shift+tab from INTENT got %s, want CONFIG", tabNames[model.tab])
	}
}

func TestIntentPreviewGitStatus(t *testing.T) {
	model, core := loadedModel(t)
	model.tab = tabIntent
	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "git status", Code: 'g'}))
	model = updated.(Model)
	if model.intentDraft != "git status" {
		t.Fatalf("draft=%q", model.intentDraft)
	}
	updated, cmd := model.Update(pressKey("ctrl+p"))
	if cmd == nil {
		t.Fatal("ctrl+p must preview")
	}
	updated, _ = updated.Update(cmd())
	model = updated.(Model)
	if core.compileCalls == 0 {
		t.Fatal("preview did not CompileIntent")
	}
	if core.submitted != 0 {
		t.Fatal("preview must not submit")
	}
	if !strings.Contains(model.intentPreview, "git.status") {
		t.Fatalf("preview=%s", model.intentPreview)
	}
	if model.intentMethod != "heuristic.git" {
		t.Fatalf("method=%s", model.intentMethod)
	}
}

func TestIntentSubmitGoesToLive(t *testing.T) {
	model, core := loadedModel(t)
	model.tab = tabIntent
	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "git status", Code: 'g'}))
	model = updated.(Model)
	updated, cmd := model.Update(pressKey("ctrl+enter"))
	if cmd == nil {
		t.Fatal("ctrl+enter must submit")
	}
	updated, cmd = updated.Update(cmd())
	model = updated.(Model)
	if model.tab != tabLive {
		t.Fatalf("tab=%s after submit", tabNames[model.tab])
	}
	if cmd == nil {
		t.Fatal("submit should refresh selected run")
	}
	updated, _ = model.Update(cmd())
	model = updated.(Model)
	if core.submitted != 1 {
		t.Fatalf("submitted=%d", core.submitted)
	}
	if model.snapshot.RunID == "" || model.snapshot.RunID == "019c-run-0001" {
		t.Fatalf("selected run %q", model.snapshot.RunID)
	}
}

func TestIntentJSONToggleAndResizeNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	model, _ := loadedModel(t)
	model.theme = DetectTheme()
	model.tab = tabIntent
	updated, _ := model.Update(pressKey("ctrl+j"))
	model = updated.(Model)
	if !model.intentJSON {
		t.Fatal("ctrl+j should enable JSON mode")
	}
	model.width, model.height = 70, 24
	content := model.View().Content
	if !strings.Contains(content, "INTENT") {
		t.Fatal("narrow INTENT missing title")
	}
	if !strings.Contains(content, "JSON") {
		t.Fatal("JSON mode not shown")
	}
	if strings.Contains(content, "\x1b[38;") || strings.Contains(content, "\x1b[48;") {
		t.Fatalf("NO_COLOR INTENT contains ANSI color")
	}
}

func TestHelpOverlayToggle(t *testing.T) {
	model, _ := loadedModel(t)
	updated, _ := model.Update(pressKey("?"))
	model = updated.(Model)
	if !model.helpOpen {
		t.Fatal("? should open help")
	}
	content := model.View().Content
	if !strings.Contains(content, "HELP") {
		t.Fatalf("help overlay missing HELP: %s", content)
	}
	updated, _ = model.Update(pressKey("esc"))
	model = updated.(Model)
	if model.helpOpen {
		t.Fatal("esc should close help")
	}
}

func TestLiveEmptyHits(t *testing.T) {
	model, _ := loadedModel(t)
	model.tab = tabLive
	content := model.View().Content
	if !strings.Contains(content, "No hits.") {
		t.Fatalf("LIVE missing empty hits: %s", content)
	}
}

func TestLiveShowsContextPackHits(t *testing.T) {
	model, core := loadedModel(t)
	core.snapshot.Events[0].Payload["graph_hits"] = []map[string]any{
		{"kind": "path", "id": "docs/46-tui-v1.md", "score": 10},
	}
	model.snapshot = core.snapshot
	model.tab = tabLive
	content := model.View().Content
	if !strings.Contains(content, "docs/46-tui-v1.md") {
		t.Fatalf("LIVE missing graph hit: %s", content)
	}
	if strings.Contains(content, "No hits.") {
		t.Fatal("LIVE still shows empty hits")
	}
}

func TestIntentPreviewQueriesHits(t *testing.T) {
	model, core := loadedModel(t)
	core.hits = graph.Hits{Items: []graph.Hit{{Kind: "capability", ID: "git.status", Score: 8}}}
	model.tab = tabIntent
	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "git status", Code: 'g'}))
	model = updated.(Model)
	updated, cmd := model.Update(pressKey("ctrl+p"))
	if cmd == nil {
		t.Fatal("ctrl+p must preview")
	}
	updated, _ = updated.Update(cmd())
	model = updated.(Model)
	if core.hitCalls == 0 {
		t.Fatal("preview must QueryHits")
	}
	if !strings.Contains(model.View().Content, "git.status") {
		t.Fatalf("INTENT preview missing hits: %s", model.View().Content)
	}
}

func TestIntentBlastDoesNotSubmit(t *testing.T) {
	model, core := loadedModel(t)
	core.blastRep = blast.Report{Risk: blast.RiskPath, Touches: []blast.Touch{{Kind: "path", Key: "notes.md", Capability: "fs.write", Mode: blast.ModeWrite}}}
	model.tab = tabIntent
	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "git status", Code: 'g'}))
	model = updated.(Model)
	updated, cmd := model.Update(pressKey("ctrl+b"))
	if cmd == nil {
		t.Fatal("ctrl+b must blast")
	}
	updated, _ = updated.Update(cmd())
	model = updated.(Model)
	if core.blastCalls != 1 {
		t.Fatalf("blastCalls=%d", core.blastCalls)
	}
	if core.submitted != 0 {
		t.Fatal("blast must not submit")
	}
	if !model.blastOpen || model.blastRep == nil || model.blastRep.Risk != blast.RiskPath {
		t.Fatalf("blast panel not open: open=%v rep=%v", model.blastOpen, model.blastRep)
	}
	content := model.View().Content
	if !strings.Contains(content, "PATH") && !strings.Contains(content, "path") {
		t.Fatalf("INTENT blast missing risk: %s", content)
	}
}

func TestLiveBlastAndGraphIgnoresB(t *testing.T) {
	model, core := loadedModel(t)
	core.blastRep = blast.Report{Risk: blast.RiskNone}
	model.tab = tabLive
	updated, cmd := model.Update(pressKey("b"))
	if cmd == nil {
		t.Fatal("b on LIVE must blast")
	}
	updated, _ = updated.Update(cmd())
	model = updated.(Model)
	if core.blastCalls != 1 {
		t.Fatalf("LIVE blastCalls=%d", core.blastCalls)
	}

	model.tab = tabGraph
	updated, cmd = model.Update(pressKey("b"))
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(blastMsg); ok {
			t.Fatal("GRAPH must not start BlastTask")
		}
	}
	updated, cmd = model.Update(pressKey("ctrl+b"))
	if cmd != nil {
		if _, ok := cmd().(blastMsg); ok {
			t.Fatal("GRAPH ctrl+b must not blast")
		}
	}
}

func TestNarrowHitsBlastNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("RUNTGINE_ASCII", "1")
	model, core := loadedModel(t)
	model.theme = DetectTheme()
	core.snapshot.Events[0].Payload["graph_hits"] = []map[string]any{
		{"kind": "path", "id": "README.md", "score": 4},
	}
	model.snapshot = core.snapshot
	rep := blast.Report{Risk: blast.RiskWorkspace}
	model.blastRep = &rep
	model.blastOpen = true
	model.tab = tabLive
	model.width, model.height = 70, 24
	content := model.View().Content
	if strings.TrimSpace(content) == "" {
		t.Fatal("narrow LIVE empty")
	}
	if !strings.Contains(content, "README.md") {
		t.Fatalf("narrow LIVE missing hits: %s", content)
	}
	if !strings.Contains(content, "WORKSPACE") && !strings.Contains(content, "workspace") {
		t.Fatalf("narrow LIVE missing risk text: %s", content)
	}
	if strings.Contains(content, "\x1b[38;") || strings.Contains(content, "\x1b[48;") {
		t.Fatalf("NO_COLOR LIVE contains ANSI color")
	}
}
