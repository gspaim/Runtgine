package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gspaim/Runtgine/internal/core/contextpack"
	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
)

type Player struct {
	Completer Completer
}

func New(c Completer) *Player {
	if c == nil {
		c = HeuristicCompleter{}
	}
	return &Player{Completer: c}
}

func (p *Player) Manifest() registry.Manifest {
	obj := json.RawMessage(`{"type":"object","additionalProperties":false}`)
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "llm",
		Version:       "0.1.0",
		Kind:          registry.KindAI,
		Capabilities: []registry.Capability{
			{Name: corepipe.CapTechReview, InputSchema: obj, OutputSchema: json.RawMessage(`{"type":"object","required":["findings","risks"]}`)},
			{Name: corepipe.CapSpecReview, InputSchema: obj, OutputSchema: json.RawMessage(`{"type":"object","required":["gaps","acceptance_hints"]}`)},
		},
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var pack contextpack.Pack
	if len(req.Context) > 0 {
		_ = json.Unmarshal(req.Context, &pack)
	} else {
		pack = contextpack.Pack{
			Step: contextpack.StepView{StepID: req.StepID, Capability: req.Capability},
		}
	}
	schema := outputSchema(req.Capability)
	out, err := p.completeJSON(ctx, pack, schema)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (p *Player) completeJSON(ctx context.Context, pack contextpack.Pack, schema json.RawMessage) (json.RawMessage, error) {
	var last error
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := p.Completer.Complete(ctx, pack, schema)
		if err != nil {
			last = err
			continue
		}
		raw = json.RawMessage(strings.TrimSpace(string(raw)))
		if !json.Valid(raw) {
			last = result.Runtime(result.CodePlayerError, "LLM returned non-JSON", true, nil)
			continue
		}
		return raw, nil
	}
	if last == nil {
		last = result.Runtime(result.CodePlayerError, "LLM complete failed", true, nil)
	}
	return nil, last
}

func outputSchema(cap string) json.RawMessage {
	switch cap {
	case corepipe.CapTechReview:
		return json.RawMessage(`{"type":"object","required":["findings","risks"]}`)
	case corepipe.CapSpecReview:
		return json.RawMessage(`{"type":"object","required":["gaps","acceptance_hints"]}`)
	case corepipe.CapDecompose:
		return json.RawMessage(`{"type":"object","required":["subtasks"]}`)
	default:
		return json.RawMessage(`{"type":"object"}`)
	}
}

func systemPrompt(schema json.RawMessage) string {
	return fmt.Sprintf("You are a Runtgine pipeline assistant. Reply with JSON only matching this schema: %s", string(schema))
}

func userPrompt(pack contextpack.Pack) string {
	b, _ := json.Marshal(pack)
	return "ContextPack:\n" + string(b)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func CompleterFromConfig(backend, openaiKey, openaiBase, openaiModel, anthropicKey, anthropicModel string) Completer {
	switch strings.ToLower(backend) {
	case "anthropic":
		if anthropicKey != "" {
			return NewAnthropic(anthropicKey, anthropicModel)
		}
	default: // openai-compat
		if openaiKey != "" {
			return NewOpenAICompat(openaiBase, openaiKey, openaiModel)
		}
	}
	return HeuristicCompleter{}
}
