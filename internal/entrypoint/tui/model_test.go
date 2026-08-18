package tui

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/graph"
)

type fakeCore struct {
	runs       []api.RunSummary
	snapshot   api.RunSnapshot
	events     []event.Event
	config     api.ConfigSnapshot
	graph      graph.Snapshot
	stream     chan event.Event
	cancelled  string
	approved   string
	refreshes  int
	refreshErr error
	snapErr    error
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

	updated, _ := model.Update(key("enter"))
	model = updated.(Model)
	if model.tab != tabLive {
		t.Fatalf("enter selected tab %d, want LIVE", model.tab)
	}

	updated, cmd := model.Update(key("c"))
	model = updated.(Model)
	if !model.confirm || cmd != nil {
		t.Fatal("first c must request confirmation")
	}

	updated, cmd = model.Update(key("c"))
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

	updated, cmd := model.Update(key("a"))
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

	updated, cmd = model.Update(key("d"))
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

	updated, _ := model.Update(key("/"))
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

	updated, _ := model.Update(key("tab"))
	model = updated.(Model)
	if model.tab != tabGraph {
		t.Fatalf("tab from EVENTS got %s, want GRAPH", tabNames[model.tab])
	}
	updated, _ = model.Update(key("tab"))
	model = updated.(Model)
	if model.tab != tabConfig {
		t.Fatalf("tab from GRAPH got %s, want CONFIG", tabNames[model.tab])
	}
	updated, _ = model.Update(key("shift+tab"))
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

	updated, _ := model.Update(key("/"))
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
	updated, cmd := model.Update(key("r"))
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

func key(value string) tea.KeyPressMsg {
	switch value {
	case "enter":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	case "tab":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	case "shift+tab":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift})
	default:
		r := []rune(value)[0]
		return tea.KeyPressMsg(tea.Key{Text: value, Code: r})
	}
}
