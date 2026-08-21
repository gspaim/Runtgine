package llm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"github.com/gspaim/Runtgine/internal/core/task"
)

type recCompleter struct {
	calls int
}

func (r *recCompleter) Complete(context.Context, contextpack.Pack, json.RawMessage) (json.RawMessage, error) {
	r.calls++
	return json.RawMessage(`{"ok":true}`), nil
}

func TestSwitchCompleterUsesRoutedProvider(t *testing.T) {
	cfg := config.Config{
		LLMProviders: []config.LLMProvider{
			{ID: "openai-main", Kind: "openai-compat", DefaultModel: "gpt-4.1-mini"},
			{ID: "anthropic-main", Kind: "anthropic", DefaultModel: "claude-sonnet"},
		},
		LLMRouting: []config.LLMRouting{
			{Match: config.RoutingMatch{CapabilityPrefix: "pipeline.spec-review"}, ProviderID: "anthropic-main"},
		},
	}
	rec := &recCompleter{}
	sw := NewSwitchCompleter(cfg, HeuristicCompleter{})
	sw.cache["anthropic-main|claude-sonnet"] = rec
	pack := contextpack.Assemble(task.Task{TaskID: "t1", Intent: task.Intent{Summary: "review"}}, "s1", "pipeline.spec-review", nil)
	if _, err := sw.Complete(context.Background(), pack, nil); err != nil {
		t.Fatal(err)
	}
	if rec.calls != 1 {
		t.Fatalf("calls=%d", rec.calls)
	}
}
