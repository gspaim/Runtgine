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
	WorkspaceRoot     string `json:"workspace_root"`
	LogLevel          string `json:"log_level"`
	MaxConcurrentRuns int    `json:"max_concurrent_runs"`
	LLMBackend        string `json:"llm_backend"` // openai-compat | anthropic (slice 2)
	DBPath            string `json:"-"`           // derived: workspace/.runtgine/runtgine.db
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
}

// EnsureRuntimeDir creates workspace/.runtgine if needed.
func EnsureRuntimeDir(cfg Config) error {
	dir := filepath.Join(cfg.WorkspaceRoot, ".runtgine")
	return os.MkdirAll(dir, 0o755)
}
