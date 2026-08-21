package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds runtime settings.
// Precedence (P2 G-38): defaults < file < env < flags (flags applied by CLI).
type Config struct {
	WorkspaceRoot     string        `json:"workspace_root"`
	LogLevel          string        `json:"log_level"`
	MaxConcurrentRuns int           `json:"max_concurrent_runs"`
	LLMBackend        string        `json:"llm_backend"` // openai-compat | anthropic
	LLMAPIKey         string        `json:"-"`
	LLMBaseURL        string        `json:"llm_base_url"`
	LLMModel          string        `json:"llm_model"`
	AnthropicAPIKey   string        `json:"-"`
	AnthropicModel    string        `json:"anthropic_model"`
	LLMProviders      []LLMProvider `json:"llm_providers,omitempty"`
	LLMRouting        []LLMRouting  `json:"llm_routing,omitempty"`
	GitHubToken       string        `json:"-"`
	DBPath            string        `json:"-"`
	ExecutionPolicy   policyfile    `json:"execution_policy"`
	PolicyDefaultEnv  string        `json:"-"`
	Memory            Memory        `json:"memory"`
	Lessons           Lessons       `json:"lessons"`
	API               API           `json:"api"`
	Webhooks          []Webhook     `json:"webhooks,omitempty"`
	WebhookSecret     string        `json:"-"`
}

type LLMProvider struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	BaseURL      string `json:"base_url,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
}

type LLMRouting struct {
	Match      RoutingMatch `json:"match"`
	ProviderID string       `json:"provider_id"`
	ModelID    string       `json:"model_id,omitempty"`
}

type RoutingMatch struct {
	Capability       string   `json:"capability,omitempty"`
	CapabilityPrefix string   `json:"capability_prefix,omitempty"`
	EffortIn         []string `json:"effort_in,omitempty"`
	DifficultyGTE    int      `json:"difficulty_gte,omitempty"`
}

type Lessons struct {
	Capture string `json:"capture"`
}

type API struct {
	Listen       string `json:"listen"`
	Token        string `json:"-"`
	MaxBodyBytes int    `json:"max_body_bytes,omitempty"`
}

type Webhook struct {
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// Memory is the on-disk memory object (G-127).
type Memory struct {
	Capture string `json:"capture"`
}

const (
	MemoryCaptureOff       = "off"
	MemoryCaptureFailures  = "failures"
	LessonsCaptureOff      = "off"
	LessonsCaptureFailures = "failures"
	DefaultAPIListen       = "127.0.0.1:7420"
	DefaultAPIMaxBody      = 1 << 20
)

// policyfile is the on-disk execution_policy object (G-82).
type policyfile struct {
	Default      string            `json:"default"`
	Capabilities map[string]string `json:"capabilities"`
}

func Defaults() Config {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return Config{
		WorkspaceRoot:     cwd,
		LogLevel:          "info",
		MaxConcurrentRuns: 4,
		LLMBackend:        "openai-compat",
		Memory:            Memory{Capture: MemoryCaptureOff},
		Lessons:           Lessons{Capture: LessonsCaptureOff},
		API:               API{Listen: DefaultAPIListen, MaxBodyBytes: DefaultAPIMaxBody},
	}
}

// Load merges defaults, optional config file, and env.
func Load(workspaceOverride string) (Config, error) {
	cfg := Defaults()
	if workspaceOverride != "" {
		cfg.WorkspaceRoot = workspaceOverride
	}
	if v := os.Getenv("RUNTGINE_WORKSPACE"); v != "" && workspaceOverride == "" {
		cfg.WorkspaceRoot = v
	}

	abs, err := filepath.Abs(cfg.WorkspaceRoot)
	if err != nil {
		return cfg, fmt.Errorf("workspace: %w", err)
	}
	cfg.WorkspaceRoot = abs

	filePath := filepath.Join(cfg.WorkspaceRoot, ".runtgine", "config.json")
	if b, err := os.ReadFile(filePath); err == nil {
		var fromFile Config
		if err := json.Unmarshal(b, &fromFile); err != nil {
			return cfg, fmt.Errorf("config file: %w", err)
		}
		mergeFile(&cfg, fromFile)
	} else if !os.IsNotExist(err) {
		return cfg, fmt.Errorf("config file: %w", err)
	}

	applyEnv(&cfg)

	if cfg.Memory.Capture == "" {
		cfg.Memory.Capture = MemoryCaptureOff
	}
	switch cfg.Memory.Capture {
	case MemoryCaptureOff, MemoryCaptureFailures:
	default:
		return cfg, fmt.Errorf("memory.capture: want off|failures, got %q", cfg.Memory.Capture)
	}
	if cfg.Lessons.Capture == "" {
		cfg.Lessons.Capture = LessonsCaptureOff
	}
	switch cfg.Lessons.Capture {
	case LessonsCaptureOff, LessonsCaptureFailures:
	default:
		return cfg, fmt.Errorf("lessons.capture: want off|failures, got %q", cfg.Lessons.Capture)
	}
	if cfg.API.Listen == "" {
		cfg.API.Listen = DefaultAPIListen
	}
	if cfg.API.MaxBodyBytes <= 0 {
		cfg.API.MaxBodyBytes = DefaultAPIMaxBody
	}

	cfg.DBPath = filepath.Join(cfg.WorkspaceRoot, ".runtgine", "runtgine.db")
	return cfg, nil
}

func mergeFile(dst *Config, src Config) {
	if src.LogLevel != "" {
		dst.LogLevel = src.LogLevel
	}
	if src.MaxConcurrentRuns > 0 {
		dst.MaxConcurrentRuns = src.MaxConcurrentRuns
	}
	if src.LLMBackend != "" {
		dst.LLMBackend = src.LLMBackend
	}
	if src.LLMBaseURL != "" {
		dst.LLMBaseURL = src.LLMBaseURL
	}
	if src.LLMModel != "" {
		dst.LLMModel = src.LLMModel
	}
	if src.AnthropicModel != "" {
		dst.AnthropicModel = src.AnthropicModel
	}
	if src.ExecutionPolicy.Default != "" {
		dst.ExecutionPolicy.Default = src.ExecutionPolicy.Default
	}
	if len(src.ExecutionPolicy.Capabilities) > 0 {
		if dst.ExecutionPolicy.Capabilities == nil {
			dst.ExecutionPolicy.Capabilities = map[string]string{}
		}
		for k, v := range src.ExecutionPolicy.Capabilities {
			dst.ExecutionPolicy.Capabilities[k] = v
		}
	}
	if src.Memory.Capture != "" {
		dst.Memory.Capture = src.Memory.Capture
	}
	if src.Lessons.Capture != "" {
		dst.Lessons.Capture = src.Lessons.Capture
	}
	if len(src.LLMProviders) > 0 {
		dst.LLMProviders = append([]LLMProvider(nil), src.LLMProviders...)
	}
	if len(src.LLMRouting) > 0 {
		dst.LLMRouting = append([]LLMRouting(nil), src.LLMRouting...)
	}
	if src.API.Listen != "" {
		dst.API.Listen = src.API.Listen
	}
	if src.API.MaxBodyBytes > 0 {
		dst.API.MaxBodyBytes = src.API.MaxBodyBytes
	}
	if len(src.Webhooks) > 0 {
		dst.Webhooks = append([]Webhook(nil), src.Webhooks...)
	}
	// workspace_root in file is ignored if env/flag set; allow file to set if still default cwd-only? skip for safety
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("RUNTGINE_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("RUNTGINE_MAX_CONCURRENT_RUNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxConcurrentRuns = n
		}
	}
	if v := os.Getenv("RUNTGINE_LLM_BACKEND"); v != "" {
		cfg.LLMBackend = v
	}
	if v := firstEnv("RUNTGINE_LLM_API_KEY", "OPENAI_API_KEY"); v != "" {
		cfg.LLMAPIKey = v
	}
	if v := os.Getenv("RUNTGINE_LLM_BASE_URL"); v != "" {
		cfg.LLMBaseURL = v
	}
	if v := os.Getenv("RUNTGINE_LLM_MODEL"); v != "" {
		cfg.LLMModel = v
	}
	if v := firstEnv("RUNTGINE_ANTHROPIC_API_KEY", "ANTHROPIC_API_KEY"); v != "" {
		cfg.AnthropicAPIKey = v
	}
	if v := os.Getenv("RUNTGINE_ANTHROPIC_MODEL"); v != "" {
		cfg.AnthropicModel = v
	}
	if v := firstEnv("RUNTGINE_GITHUB_TOKEN", "GITHUB_TOKEN"); v != "" {
		cfg.GitHubToken = v
	}
	if v := os.Getenv("RUNTGINE_POLICY_DEFAULT"); v != "" {
		cfg.PolicyDefaultEnv = v
	}
	if v := os.Getenv("RUNTGINE_MEMORY_CAPTURE"); v != "" {
		cfg.Memory.Capture = v
	}
	if v := os.Getenv("RUNTGINE_LESSONS_CAPTURE"); v != "" {
		cfg.Lessons.Capture = v
	}
	if v := os.Getenv("RUNTGINE_API_LISTEN"); v != "" {
		cfg.API.Listen = v
	}
	if v := os.Getenv("RUNTGINE_API_TOKEN"); v != "" {
		cfg.API.Token = v
	}
	if v := os.Getenv("RUNTGINE_API_MAX_BODY_BYTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.API.MaxBodyBytes = n
		}
	}
	if v := os.Getenv("RUNTGINE_WEBHOOK_SECRET"); v != "" {
		cfg.WebhookSecret = v
	}
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// EnsureRuntimeDir creates workspace/.runtgine if needed.
func EnsureRuntimeDir(cfg Config) error {
	dir := filepath.Join(cfg.WorkspaceRoot, ".runtgine")
	return os.MkdirAll(dir, 0o755)
}
