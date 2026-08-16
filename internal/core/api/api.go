package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

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

type RunSnapshot struct {
	RunID       string            `json:"run_id"`
	TaskID      string            `json:"task_id"`
	ParentRunID string            `json:"parent_run_id,omitempty"`
	Status      string            `json:"status"`
	Error       string            `json:"error,omitempty"`
	Events      []event.Event     `json:"events"`
	Task        json.RawMessage   `json:"task,omitempty"`
	Subtasks    []store.Subtask   `json:"subtasks,omitempty"`
	ChildRuns   []ChildRunView    `json:"child_runs,omitempty"`
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

func (c *Core) Subscribe(buffer int) (<-chan event.Event, func()) {
	return c.Bus.Subscribe(buffer)
}

func (c *Core) CancelRun(runID string) error {
	return c.Runner.Cancel(runID)
}
