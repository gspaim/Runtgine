package runner

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/plan"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
	"github.com/google/uuid"
)

type Runner struct {
	Reg        *registry.Registry
	Bus        event.Bus
	Store      *store.Store
	Workspace  string
	LLMBackend string
	Log        *slog.Logger

	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	sem     chan struct{}
}

func New(reg *registry.Registry, bus event.Bus, st *store.Store, workspace, llmBackend string, maxConcurrent int, log *slog.Logger) *Runner {
	if maxConcurrent < 1 {
		maxConcurrent = 4
	}
	if log == nil {
		log = slog.Default()
	}
	return &Runner{
		Reg:        reg,
		Bus:        bus,
		Store:      st,
		Workspace:  workspace,
		LLMBackend: llmBackend,
		Log:        log,
		cancels:    map[string]context.CancelFunc{},
		sem:        make(chan struct{}, maxConcurrent),
	}
}

type SubmitResult struct {
	RunID string
}

func (r *Runner) Submit(ctx context.Context, t task.Task) (SubmitResult, error) {
	if err := task.StructuralValidate(t); err != nil {
		_ = r.emit("", t.TaskID, nil, event.TypeTaskRejected, map[string]any{
			"code":    result.CodeSchema,
			"message": err.Error(),
		})
		return SubmitResult{}, result.Validation(result.CodeSchema, err.Error(), nil)
	}
	for _, s := range t.Steps {
		if !r.Reg.HasCapability(s.Capability) {
			msg := "capability \"" + s.Capability + "\" is not registered"
			_ = r.emit("", t.TaskID, nil, event.TypeTaskRejected, map[string]any{
				"code":    result.CodeUnknownCapability,
				"message": msg,
			})
			return SubmitResult{}, result.Validation(result.CodeUnknownCapability, msg, nil)
		}
	}

	runID, err := uuid.NewV7()
	if err != nil {
		return SubmitResult{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	rid := runID.String()

	taskJSON, _ := json.Marshal(t)
	if err := r.Store.InsertRun(ctx, store.Run{RunID: rid, TaskID: t.TaskID, Status: store.StatusAccepted}, taskJSON); err != nil {
		return SubmitResult{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	_ = r.emit(rid, t.TaskID, nil, event.TypeTaskAccepted, map[string]any{"steps": len(t.Steps)})

	p, err := plan.FromTask(t, rid, func(cap string) (string, error) {
		name, _, err := r.Reg.Resolve(cap, r.LLMBackend)
		return name, err
	})
	if err != nil {
		_ = r.Store.UpdateRunStatus(ctx, rid, store.StatusFailed, store.FormatErr(err))
		return SubmitResult{}, err
	}
	_ = r.Store.UpdateRunStatus(ctx, rid, store.StatusPlanned, "")
	_ = r.emit(rid, t.TaskID, nil, event.TypeRunPlanned, map[string]any{"plan_id": p.PlanID})

	runCtx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.cancels[rid] = cancel
	r.mu.Unlock()

	go r.execute(runCtx, t, p)

	return SubmitResult{RunID: rid}, nil
}

func (r *Runner) Cancel(runID string) error {
	r.mu.Lock()
	cancel, ok := r.cancels[runID]
	r.mu.Unlock()
	if !ok {
		return result.Runtime(result.CodeInternal, "run not active or unknown", false, nil)
	}
	cancel()
	return nil
}

func (r *Runner) execute(ctx context.Context, t task.Task, p plan.Plan) {
	r.sem <- struct{}{}
	defer func() { <-r.sem }()
	defer func() {
		r.mu.Lock()
		delete(r.cancels, p.RunID)
		r.mu.Unlock()
	}()

	_ = r.Store.UpdateRunStatus(ctx, p.RunID, store.StatusRunning, "")
	_ = r.emit(p.RunID, t.TaskID, nil, event.TypeRunStarted, nil)

	order, err := task.TopoOrder(t.Steps)
	if err != nil {
		r.fail(ctx, p, err)
		return
	}
	byID := map[string]plan.Step{}
	for _, s := range p.Steps {
		byID[s.StepID] = s
	}

	for _, sid := range order {
		if ctx.Err() != nil {
			r.cancelled(ctx, p)
			return
		}
		s := byID[sid]
		stepID := s.StepID
		_ = r.emit(p.RunID, t.TaskID, &stepID, event.TypeStepStarted, map[string]any{
			"capability": s.Capability,
			"player":     s.Player,
		})

		player, ok := r.Reg.Get(s.Player)
		if !ok {
			err := result.Runtime(result.CodeInternal, "player missing: "+s.Player, false, nil)
			_ = r.emit(p.RunID, t.TaskID, &stepID, event.TypeStepFailed, map[string]any{"error": err.Error()})
			r.fail(ctx, p, err)
			return
		}

		var lastErr error
		var out json.RawMessage
		attempts := s.MaxRetries + 1
		start := time.Now()
		for attempt := 1; attempt <= attempts; attempt++ {
			if ctx.Err() != nil {
				r.cancelled(ctx, p)
				return
			}
			out, lastErr = player.Execute(ctx, registry.ExecRequest{
				Capability: s.Capability,
				Input:      s.Input,
				Workspace:  r.Workspace,
				RunID:      p.RunID,
				TaskID:     t.TaskID,
				StepID:     s.StepID,
			})
			if lastErr == nil {
				_ = r.emit(p.RunID, t.TaskID, &stepID, event.TypeStepSucceeded, map[string]any{
					"duration_ms": time.Since(start).Milliseconds(),
					"attempts":    attempt,
					"output":      jsonRaw(out),
				})
				break
			}
			var re result.Error
			retryable := false
			if errors.As(lastErr, &re) {
				retryable = re.Retryable
			}
			if attempt >= attempts || !retryable {
				_ = r.emit(p.RunID, t.TaskID, &stepID, event.TypeStepFailed, map[string]any{
					"error":       lastErr.Error(),
					"duration_ms": time.Since(start).Milliseconds(),
					"attempts":    attempt,
				})
				r.fail(ctx, p, lastErr)
				return
			}
			r.Log.Warn("step retry", "run_id", p.RunID, "step_id", stepID, "attempt", attempt, "err", lastErr)
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}

	_ = r.Store.UpdateRunStatus(ctx, p.RunID, store.StatusSucceeded, "")
	_ = r.emit(p.RunID, t.TaskID, nil, event.TypeRunSucceeded, nil)
}

func (r *Runner) fail(ctx context.Context, p plan.Plan, err error) {
	_ = r.Store.UpdateRunStatus(ctx, p.RunID, store.StatusFailed, store.FormatErr(err))
	_ = r.emit(p.RunID, p.TaskID, nil, event.TypeRunFailed, map[string]any{"error": err.Error()})
}

func (r *Runner) cancelled(ctx context.Context, p plan.Plan) {
	_ = r.Store.UpdateRunStatus(ctx, p.RunID, store.StatusCancelled, "")
	_ = r.emit(p.RunID, p.TaskID, nil, event.TypeRunCancelled, nil)
}

func (r *Runner) emit(runID, taskID string, stepID *string, typ string, payload map[string]any) error {
	e, err := event.New(typ, runID, taskID, stepID, payload)
	if err != nil {
		return err
	}
	r.Log.Info("event", "type", e.Type, "run_id", runID, "task_id", taskID, "event_id", e.EventID)
	r.Bus.Publish(e)
	if r.Store != nil && runID != "" {
		_ = r.Store.AppendEvent(context.Background(), e)
	}
	return nil
}

func jsonRaw(b json.RawMessage) any {
	if len(b) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	return v
}
