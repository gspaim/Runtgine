package llm

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"github.com/gspaim/Runtgine/internal/core/router"
)

// SwitchCompleter routes Complete calls using cfg.llm_routing (G-147).
type SwitchCompleter struct {
	Cfg      config.Config
	mu       sync.Mutex
	cache    map[string]Completer
	fallback Completer
}

func NewSwitchCompleter(cfg config.Config, fallback Completer) *SwitchCompleter {
	if fallback == nil {
		fallback = HeuristicCompleter{}
	}
	return &SwitchCompleter{Cfg: cfg, cache: map[string]Completer{}, fallback: fallback}
}

func (s *SwitchCompleter) Complete(ctx context.Context, pack contextpack.Pack, outputSchema json.RawMessage) (json.RawMessage, error) {
	effort, difficulty := router.SignalsFromPriors(pack.PriorOutputs)
	route := router.SelectLLM(s.Cfg, router.Request{
		Capability: pack.Step.Capability,
		Effort:     effort,
		Difficulty: difficulty,
	})
	c := s.lookup(route)
	return c.Complete(ctx, pack, outputSchema)
}

func (s *SwitchCompleter) lookup(route router.Result) Completer {
	key := route.ProviderID + "|" + route.ModelID
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.cache[key]; ok {
		return c
	}
	c := completerFor(s.Cfg, route.ProviderID, route.ModelID)
	if c == nil {
		c = s.fallback
	}
	s.cache[key] = c
	return c
}

func completerFor(cfg config.Config, providerID, model string) Completer {
	for _, p := range cfg.LLMProviders {
		if p.ID != providerID {
			continue
		}
		if model == "" {
			model = p.DefaultModel
		}
		return CompleterFromKind(p.Kind, cfg, p.BaseURL, model)
	}
	return CompleterFromKind(providerID, cfg, cfg.LLMBaseURL, firstNonEmpty(model, cfg.LLMModel))
}

func CompleterFromKind(kind string, cfg config.Config, baseURL, model string) Completer {
	switch kind {
	case "anthropic":
		if cfg.AnthropicAPIKey != "" {
			return NewAnthropic(cfg.AnthropicAPIKey, firstNonEmpty(model, cfg.AnthropicModel))
		}
	default:
		if cfg.LLMAPIKey != "" {
			return NewOpenAICompat(firstNonEmpty(baseURL, cfg.LLMBaseURL), cfg.LLMAPIKey, firstNonEmpty(model, cfg.LLMModel))
		}
	}
	return HeuristicCompleter{}
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}
