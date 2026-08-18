package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/claim"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/graph"
	"github.com/gspaim/Runtgine/internal/core/intent"
	"github.com/gspaim/Runtgine/internal/core/policy"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
	dockerplayer "github.com/gspaim/Runtgine/internal/players/docker"
	"github.com/gspaim/Runtgine/internal/players/filesystem"
	gitplayer "github.com/gspaim/Runtgine/internal/players/git"
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
	Intent *intent.Engine
	Graph  *graph.Service
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
	if err := reg.Register(gitplayer.New()); err != nil {
		_ = st.Close()
		return nil, err
	}
	if err := reg.Register(filesystem.New()); err != nil {
		_ = st.Close()
		return nil, err
	}
	if err := reg.Register(dockerplayer.New()); err != nil {
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
	g := graph.New(st, log)
	tab, err := policy.Compile(policy.FileConfig{
		Default:      cfg.ExecutionPolicy.Default,
		Capabilities: cfg.ExecutionPolicy.Capabilities,
	}, cfg.PolicyDefaultEnv, reg)
	if err != nil {
		_ = st.Close()
		return nil, err
	}
	r := runner.New(reg, bus, st, cfg.WorkspaceRoot, cfg.LLMBackend, cfg.MaxConcurrentRuns, log)
	r.Graph = g
	r.Policy = tab
	claims := claim.New(st)
	if err := claims.SweepOrphans(context.Background()); err != nil {
		if log != nil {
			log.Warn("claim sweep failed", "err", err)
		}
	}
	r.Claims = claims
	if err := g.RefreshFromRegistry(context.Background(), reg); err != nil {
		if log != nil {
			log.Warn("graph refresh failed", "err", err)
		}
	}
	eng := intent.New(completer)
	eng.Graph = g
	return &Core{
		Cfg:    cfg,
		Reg:    reg,
		Bus:    bus,
		Store:  st,
		Runner: r,
		Intent: eng,
		Graph:  g,
		Log:    log,
	}, nil
}

func (c *Core) Close() error {
	if c.Runner != nil {
		c.Runner.WaitIdle()
	}
	return c.Store.Close()
}

func (c *Core) SubmitTask(ctx context.Context, t task.Task) (string, error) {
	res, err := c.Runner.Submit(ctx, t)
	if err != nil {
		return "", err
	}
	return res.RunID, nil
}

// CompileIntent translates natural language into Task IR (G-51).
func (c *Core) CompileIntent(ctx context.Context, text, entryPoint, ref string) (task.Task, string, error) {
	res, err := c.Intent.Compile(ctx, intent.Request{
		Text:       text,
		EntryPoint: entryPoint,
		Ref:        ref,
	})
	if err != nil {
		return task.Task{}, "", err
	}
	return res.Task, res.Method, nil
}

// SubmitIntent compiles NL intent and submits the resulting Task IR.
func (c *Core) SubmitIntent(ctx context.Context, text, entryPoint, ref string) (string, string, error) {
	tk, method, err := c.CompileIntent(ctx, text, entryPoint, ref)
	if err != nil {
		return "", "", err
	}
	runID, err := c.SubmitTask(ctx, tk)
	if err != nil {
		return "", method, err
	}
	return runID, method, nil
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

type PendingApproval struct {
	StepID     string `json:"step_id"`
	Capability string `json:"capability"`
	Player     string `json:"player"`
}

type RunSnapshot struct {
	RunID           string           `json:"run_id"`
	TaskID          string           `json:"task_id"`
	ParentRunID     string           `json:"parent_run_id,omitempty"`
	Status          string           `json:"status"`
	Error           string           `json:"error,omitempty"`
	PendingApproval *PendingApproval `json:"pending_approval,omitempty"`
	Events          []event.Event    `json:"events"`
	Task            json.RawMessage  `json:"task,omitempty"`
	Subtasks        []store.Subtask  `json:"subtasks,omitempty"`
	ChildRuns       []ChildRunView   `json:"child_runs,omitempty"`
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
	snap := RunSnapshot{
		RunID:       run.RunID,
		TaskID:      run.TaskID,
		ParentRunID: run.ParentRunID,
		Status:      string(run.Status),
		Error:       run.ErrorJSON,
		Events:      evs,
		Task:        taskJSON,
		Subtasks:    subs,
		ChildRuns:   views,
	}
	if run.Status == store.StatusWaitingApproval && run.PendingStepID != "" {
		snap.PendingApproval = &PendingApproval{
			StepID:     run.PendingStepID,
			Capability: run.PendingCapability,
			Player:     run.PendingPlayer,
		}
	}
	return snap, nil
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

func (c *Core) ApproveRun(runID, decision string) error {
	return c.Runner.Approve(context.Background(), runID, decision)
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
	if err := c.Store.ClearPendingApproval(context.Background(), runID); err != nil {
		return err
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
	c.Runner.ReleaseClaims(runID, run.TaskID)
	c.syncGraph(runID)
	return nil
}

func (c *Core) GetGraphSnapshot(ctx context.Context) (graph.Snapshot, error) {
	return c.Graph.Snapshot(ctx)
}

func (c *Core) RefreshGraph(ctx context.Context) error {
	return c.Graph.RefreshFromRegistry(ctx, c.Reg)
}

func (c *Core) QueryGraphNeighbors(ctx context.Context, kind, id, edgeKind, direction string) ([]graph.Node, error) {
	return c.Graph.QueryNeighbors(ctx, kind, id, edgeKind, direction)
}

func (c *Core) syncGraph(runID string) {
	if c.Graph == nil || runID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Graph.SyncFromRun(ctx, runID); err != nil && c.Log != nil {
		c.Log.Warn("graph sync failed", "run_id", runID, "err", err)
	}
}
