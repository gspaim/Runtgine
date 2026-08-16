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
)

type fakeCore struct {
	runs      []api.RunSummary
	snapshot  api.RunSnapshot
	events    []event.Event
	config    api.ConfigSnapshot
	stream    chan event.Event
	cancelled string
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
	}
}

func loadedModel(t *testing.T) (Model, *fakeCore) {
	t.Helper()
	core := fixture()
	model, _ := New(core)
	msg := model.refreshCmd()()
	updated, _ := model.Update(msg)
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

func key(value string) tea.KeyPressMsg {
	switch value {
	case "enter":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	default:
		r := []rune(value)[0]
		return tea.KeyPressMsg(tea.Key{Text: value, Code: r})
	}
}
