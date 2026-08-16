package llm

import (
	"context"
	"encoding/json"

	"github.com/gspaim/Runtgine/internal/core/contextpack"
)

// Completer is the single LLM interface (G-25).
type Completer interface {
	Complete(ctx context.Context, pack contextpack.Pack, outputSchema json.RawMessage) (json.RawMessage, error)
}

// HeuristicCompleter is used when no API key is configured (tests / offline).
type HeuristicCompleter struct{}

func (HeuristicCompleter) Complete(_ context.Context, pack contextpack.Pack, _ json.RawMessage) (json.RawMessage, error) {
	switch pack.Step.Capability {
	case "pipeline.tech-review":
		return json.Marshal(map[string]any{
			"findings": []string{"Review intent: " + pack.Task.Summary},
			"risks":    []string{"offline heuristic — no LLM key"},
		})
	case "pipeline.spec-review":
		gaps := []string{}
		if pack.Task.Notes == "" {
			gaps = append(gaps, "intent.notes is empty")
		}
		return json.Marshal(map[string]any{
			"gaps":             gaps,
			"acceptance_hints": []string{"Define measurable acceptance for: " + pack.Task.Summary},
		})
	case "pipeline.decompose":
		return json.Marshal(map[string]any{
			"subtasks": []map[string]string{},
			"refined":  false,
		})
	default:
		return json.Marshal(map[string]any{"ok": true})
	}
}
