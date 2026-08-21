package router

import (
	"testing"

	"github.com/gspaim/Runtgine/internal/config"
)

func TestSelectLLMFallbackLegacy(t *testing.T) {
	got := SelectLLM(config.Config{LLMBackend: "anthropic", LLMModel: "claude"}, Request{Capability: "pipeline.tech-review"})
	if got.ProviderID != "anthropic" || got.ModelID != "claude" {
		t.Fatalf("%+v", got)
	}
}

func TestSelectLLMPrefixAndEffort(t *testing.T) {
	cfg := config.Config{
		LLMProviders: []config.LLMProvider{
			{ID: "openai-main", Kind: "openai-compat", DefaultModel: "gpt-4.1-mini"},
			{ID: "anthropic-main", Kind: "anthropic", DefaultModel: "claude-sonnet"},
		},
		LLMRouting: []config.LLMRouting{
			{Match: config.RoutingMatch{CapabilityPrefix: "pipeline.spec-review"}, ProviderID: "anthropic-main"},
			{Match: config.RoutingMatch{CapabilityPrefix: "pipeline.", EffortIn: []string{"S", "M"}}, ProviderID: "openai-main"},
			{Match: config.RoutingMatch{DifficultyGTE: 4}, ProviderID: "anthropic-main"},
		},
	}
	spec := SelectLLM(cfg, Request{Capability: "pipeline.spec-review", Effort: "S"})
	if spec.ProviderID != "anthropic-main" {
		t.Fatalf("spec-review -> %s", spec.ProviderID)
	}
	easy := SelectLLM(cfg, Request{Capability: "pipeline.tech-review", Effort: "S"})
	if easy.ProviderID != "openai-main" {
		t.Fatalf("easy tech-review -> %s", easy.ProviderID)
	}
	hard := SelectLLM(cfg, Request{Capability: "llm.chat", Difficulty: 5})
	if hard.ProviderID != "anthropic-main" {
		t.Fatalf("hard -> %s", hard.ProviderID)
	}
	none := SelectLLM(cfg, Request{Capability: "git.status"})
	if none.ProviderID != "openai-main" {
		t.Fatalf("fallback -> %s", none.ProviderID)
	}
}
