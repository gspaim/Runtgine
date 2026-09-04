package router

import (
	"encoding/json"
	"strings"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/store"
)

type Request struct {
	Capability string
	Effort     string
	Difficulty int
}

type Result struct {
	ProviderID string
	ModelID    string
}

// SelectLLM picks a provider/model. First matching rule of highest specificity
// wins. Empty providers fall back to legacy llm_backend.
func SelectLLM(cfg config.Config, req Request) Result {
	providers := cfg.LLMProviders
	if len(providers) == 0 {
		id := strings.TrimSpace(cfg.LLMBackend)
		if id == "" {
			id = "openai-compat"
		}
		return Result{ProviderID: id, ModelID: cfg.LLMModel}
	}
	def := providers[0]
	for _, p := range providers {
		if p.ID == cfg.LLMBackend {
			def = p
			break
		}
	}
	bestScore := -1
	best := Result{ProviderID: def.ID, ModelID: def.DefaultModel}
	for _, rule := range cfg.LLMRouting {
		score, ok := matchScore(rule.Match, req)
		if !ok || score < bestScore {
			continue
		}
		if score == bestScore && bestScore >= 0 {
			continue
		}
		model := rule.ModelID
		if model == "" {
			for _, p := range providers {
				if p.ID == rule.ProviderID {
					model = p.DefaultModel
					break
				}
			}
		}
		bestScore = score
		best = Result{ProviderID: rule.ProviderID, ModelID: model}
	}
	if best.ModelID == "" {
		for _, p := range providers {
			if p.ID == best.ProviderID {
				best.ModelID = p.DefaultModel
				break
			}
		}
	}
	return best
}

func matchScore(m config.RoutingMatch, req Request) (int, bool) {
	score := 0
	if m.Capability != "" {
		if m.Capability != req.Capability {
			return 0, false
		}
		score += 1000
	}
	if m.CapabilityPrefix != "" {
		if !strings.HasPrefix(req.Capability, m.CapabilityPrefix) {
			return 0, false
		}
		score += 100 + len(m.CapabilityPrefix)
	}
	if len(m.EffortIn) > 0 {
		ok := false
		for _, e := range m.EffortIn {
			if strings.EqualFold(e, req.Effort) {
				ok = true
				break
			}
		}
		if !ok {
			return 0, false
		}
		score += 10
	}
	if m.DifficultyGTE > 0 {
		if req.Difficulty < m.DifficultyGTE {
			return 0, false
		}
		score += 1
	}
	if score == 0 {
		return 0, false
	}
	return score, true
}

// SignalsFromPriors reads pipeline.effort / pipeline.difficulty JSON outputs.
func SignalsFromPriors(priors []store.StepOutput) (effort string, difficulty int) {
	for _, o := range priors {
		switch o.Capability {
		case "pipeline.effort":
			var parsed struct {
				Effort string `json:"effort"`
			}
			if json.Unmarshal(o.Output, &parsed) == nil {
				effort = parsed.Effort
			}
		case "pipeline.difficulty":
			var parsed struct {
				Difficulty int `json:"difficulty"`
			}
			if json.Unmarshal(o.Output, &parsed) == nil {
				difficulty = parsed.Difficulty
			}
		}
	}
	return effort, difficulty
}
