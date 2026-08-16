package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
	"github.com/gspaim/Runtgine/internal/players/llm"
	pipeplayer "github.com/gspaim/Runtgine/internal/players/pipeline"
	"github.com/gspaim/Runtgine/internal/players/shell"
)

// Core is the Entry Point → Core API (G-07).
type Core struct {
	Cfg    config.Config
	Reg    *registry.Registry
	Bus    *event.MemoryBus
	Store  *store.Store
	Runner *runner.Runner
	Log    *slog.Logger
}

func Open(cfg config.Config, log *slog.Logger) (*Core, error) {
	if err := config.EnsureRuntimeDir(cfg); err != nil {
		return nil, err
	}
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	bus := event.NewMemoryBus()
	reg := registry.New()
	if err := reg.Register(shell.New()); err != nil {
		_ = st.Close()
		return nil, err
	}
	completer := llm.CompleterFromConfig(
		cfg.LLMBackend, cfg.LLMAPIKey, cfg.LLMBaseURL, cfg.LLMModel,
		cfg.AnthropicAPIKey, cfg.AnthropicModel,
	)
	if err := reg.Register(pipeplayer.NewWithRefine(completer)); err != nil {
		_ = st.Close()
		return nil, err
	}
	if err := reg.Register(llm.New(completer)); err != nil {
		_ = st.Close()
		return nil, err
	}
	r := runner.New(reg, bus, st, cfg.WorkspaceRoot, cfg.LLMBackend, cfg.MaxConcurrentRuns, log)
	return &Core{Cfg: cfg, Reg: reg, Bus: bus, Store: st, Runner: r, Log: log}, nil
}

func (c *Core) Close() error {
	return c.Store.Close()
}

func (c *Core) SubmitTask(ctx context.Context, t task.Task) (string, error) {
	res, err := c.Runner.Submit(ctx, t)
	if err != nil {
		return "", err
	}
	return res.RunID, nil
}

type ChildRunView struct {
	RunID  string `json:"run_id"`
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type RunSummary struct {
	RunID       string    `json:"run_id"`
	TaskID      string    `json:"task_id"`
	ParentRunID string    `json:"parent_run_id,omitempty"`
	Status      string    `json:"status"`
	Summary     string    `json:"summary"`
	Source      string    `json:"source"`
	SourceRef   string    `json:"source_ref,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ConfigSnapshot struct {
	WorkspaceRoot     string `json:"workspace_root"`
	DBPath            string `json:"db_path"`
	LogLevel          string `json:"log_level"`
	MaxConcurrentRuns int    `json:"max_concurrent_runs"`
	LLMBackend        string `json:"llm_backend"`
	LLMConnected      bool   `json:"llm_connected"`
	GitHubConnected   bool   `json:"github_connected"`
	Precedence        string `json:"precedence"`
}

type RunSnapshot struct {
	RunID       string          `json:"run_id"`
	TaskID      string          `json:"task_id"`
	ParentRunID string          `json:"parent_run_id,omitempty"`
	Status      string          `json:"status"`
	Error       string          `json:"error,omitempty"`
	Events      []event.Event   `json:"events"`
	Task        json.RawMessage `json:"task,omitempty"`
	Subtasks    []store.Subtask `json:"subtasks,omitempty"`
	ChildRuns   []ChildRunView  `json:"child_runs,omitempty"`
}

func (c *Core) GetRun(ctx context.Context, runID string) (RunSnapshot, error) {
	run, taskJSON, err := c.Store.GetRun(ctx, runID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RunSnapshot{}, result.Runtime(result.CodeInternal, "run not found", false, nil)
		}
		return RunSnapshot{}, err
	}
	evs, err := c.Store.ListEvents(ctx, runID)
	if err != nil {
		return RunSnapshot{}, err
	}
	subs, _ := c.Store.ListSubtasks(ctx, runID)
	children, _ := c.Store.ListChildRuns(ctx, runID)
	views := make([]ChildRunView, 0, len(children))
	for _, ch := range children {
		views = append(views, ChildRunView{RunID: ch.RunID, TaskID: ch.TaskID, Status: string(ch.Status)})
	}
	return RunSnapshot{
		RunID:       run.RunID,
		TaskID:      run.TaskID,
		ParentRunID: run.ParentRunID,
		Status:      string(run.Status),
		Error:       run.ErrorJSON,
		Events:      evs,
		Task:        taskJSON,
		Subtasks:    subs,
		ChildRuns:   views,
	}, nil
}

func (c *Core) ListRuns(ctx context.Context, limit int) ([]RunSummary, error) {
	records, err := c.Store.ListRuns(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]RunSummary, 0, len(records))
	for _, rec := range records {
		var parsed task.Task
		_ = json.Unmarshal(rec.TaskJSON, &parsed)
		out = append(out, RunSummary{
			RunID:       rec.RunID,
			TaskID:      rec.TaskID,
			ParentRunID: rec.ParentRunID,
			Status:      string(rec.Status),
			Summary:     parsed.Intent.Summary,
			Source:      parsed.Source.EntryPoint,
			SourceRef:   parsed.Source.Ref,
			CreatedAt:   rec.CreatedAt,
			UpdatedAt:   rec.UpdatedAt,
		})
	}
	return out, nil
}

func (c *Core) ListRecentEvents(ctx context.Context, limit int) ([]event.Event, error) {
	return c.Store.ListRecentEvents(ctx, limit)
}

func (c *Core) ConfigSnapshot() ConfigSnapshot {
	return ConfigSnapshot{
		WorkspaceRoot:     c.Cfg.WorkspaceRoot,
		DBPath:            c.Cfg.DBPath,
		LogLevel:          c.Cfg.LogLevel,
		MaxConcurrentRuns: c.Cfg.MaxConcurrentRuns,
		LLMBackend:        c.Cfg.LLMBackend,
		LLMConnected:      c.Cfg.LLMAPIKey != "" || c.Cfg.AnthropicAPIKey != "",
		GitHubConnected:   c.Cfg.GitHubToken != "",
		Precedence:        "defaults < config.json < env < CLI flags",
	}
}

func (c *Core) Subscribe(buffer int) (<-chan event.Event, func()) {
	return c.Bus.Subscribe(buffer)
}

func (c *Core) CancelRun(runID string) error {
	if err := c.Runner.Cancel(runID); err == nil {
		return nil
	}

	// A prior CLI process may have left a persisted non-terminal run without
	// an in-memory cancel function. Mark that stale run cancelled.
	run, _, err := c.Store.GetRun(context.Background(), runID)
	if err != nil {
		return err
	}
	switch run.Status {
	case store.StatusSucceeded, store.StatusFailed, store.StatusCancelled, store.StatusRejected:
		return result.Runtime(result.CodeCancelled, "run is already terminal", false, nil)
	}
	if err := c.Store.UpdateRunStatus(
		context.Background(), runID, store.StatusCancelled, "",
	); err != nil {
		return err
	}
	e, err := event.New(event.TypeRunCancelled, runID, run.TaskID, nil, map[string]any{
		"reason": "cancelled from persisted state",
	})
	if err != nil {
		return err
	}
	if err := c.Store.AppendEvent(context.Background(), e); err != nil {
		return err
	}
	c.Bus.Publish(e)
	return nil
}
