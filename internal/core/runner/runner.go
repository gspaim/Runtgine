package runner

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gspaim/Runtgine/internal/core/claim"
	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/graph"
	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/plan"
	"github.com/gspaim/Runtgine/internal/core/policy"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
	dockplayer "github.com/gspaim/Runtgine/internal/players/docker"
	"github.com/gspaim/Runtgine/internal/players/filesystem"
	"github.com/gspaim/Runtgine/internal/players/git"
	httpplayer "github.com/gspaim/Runtgine/internal/players/httpclient"
	"github.com/gspaim/Runtgine/internal/players/shell"
)

// GraphBridge syncs structural memory (G-65) and serves QueryHits (G-68). Optional.
type GraphBridge interface {
	SyncFromRun(ctx context.Context, runID string) error
	QueryHits(ctx context.Context, q graph.Query) graph.Hits
}

type Runner struct {
	Reg        *registry.Registry
	Bus        event.Bus
	Store      *store.Store
	Workspace  string
	LLMBackend string
	Log        *slog.Logger
	Graph      GraphBridge
	Policy     policy.Table
	Claims     *claim.Service

	mu       sync.Mutex
	cancels  map[string]context.CancelFunc
	sem      chan struct{}
	inflight sync.WaitGroup
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
	return r.SubmitChild(ctx, t, "")
}

func (r *Runner) SubmitChild(ctx context.Context, t task.Task, parentRunID string) (SubmitResult, error) {
	if err := r.validateAdmission(t); err != nil {
		return SubmitResult{}, err
	}

	runID, err := uuid.NewV7()
	if err != nil {
		return SubmitResult{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	rid := runID.String()

	taskJSON, _ := json.Marshal(t)
	if err := r.Store.InsertRun(ctx, store.Run{RunID: rid, TaskID: t.TaskID, ParentRunID: parentRunID, Status: store.StatusAccepted}, taskJSON); err != nil {
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

	r.inflight.Add(1)
	go func() {
		defer r.inflight.Done()
		r.execute(runCtx, t, p, "")
	}()

	return SubmitResult{RunID: rid}, nil
}

// WaitIdle blocks until every in-flight execute (including graph sync) has returned.
func (r *Runner) WaitIdle() {
	r.inflight.Wait()
}

// ValidateTaskIR runs admission checks used by SubmitTask *before* policy
// (schema, registry, static player input). BlastTask uses this subset.
func (r *Runner) ValidateTaskIR(t task.Task) error {
	return r.validateTaskIR(t)
}

func (r *Runner) validateAdmission(t task.Task) error {
	if err := r.validateTaskIR(t); err != nil {
		return err
	}
	return r.checkAdmissionPolicy(t)
}

func (r *Runner) validateTaskIR(t task.Task) error {
	raw, err := json.Marshal(t)
	if err != nil {
		return r.reject(t.TaskID, result.CodeSchema, "marshal task ir: "+err.Error())
	}
	if err := task.ValidateDocument(raw); err != nil {
		return r.reject(t.TaskID, result.CodeSchema, err.Error())
	}
	if err := task.IdentityValidate(t); err != nil {
		return r.reject(t.TaskID, result.CodeSchema, err.Error())
	}
	if err := task.StructuralValidate(t); err != nil {
		return r.reject(t.TaskID, result.CodeSchema, err.Error())
	}
	for _, s := range t.Steps {
		if !r.Reg.HasCapability(s.Capability) {
			return r.reject(t.TaskID, result.CodeUnknownCapability,
				"capability \""+s.Capability+"\" is not registered")
		}
		if err := r.Reg.ValidateInput(s.Capability, s.Input); err != nil {
			var ve result.Error
			if errors.As(err, &ve) {
				return r.reject(t.TaskID, ve.Code, ve.Message)
			}
			return r.reject(t.TaskID, result.CodeInvalidInput, err.Error())
		}
		if s.Capability == shell.CapExec {
			if err := shell.ValidateStaticInput(r.Workspace, s.Input); err != nil {
				var ve result.Error
				if errors.As(err, &ve) {
					return r.reject(t.TaskID, ve.Code, ve.Message)
				}
				return r.reject(t.TaskID, result.CodeInvalidInput, err.Error())
			}
		}
		switch s.Capability {
		case git.CapStatus, git.CapDiff, git.CapLog, git.CapAdd, git.CapCommit:
			if err := git.ValidateStaticInput(r.Workspace, s.Capability, s.Input); err != nil {
				var ve result.Error
				if errors.As(err, &ve) {
					return r.reject(t.TaskID, ve.Code, ve.Message)
				}
				return r.reject(t.TaskID, result.CodeInvalidInput, err.Error())
			}
		case filesystem.CapRead, filesystem.CapWrite, filesystem.CapList, filesystem.CapStat:
			if err := filesystem.ValidateStaticInput(r.Workspace, s.Capability, s.Input); err != nil {
				var ve result.Error
				if errors.As(err, &ve) {
					return r.reject(t.TaskID, ve.Code, ve.Message)
				}
				return r.reject(t.TaskID, result.CodeInvalidInput, err.Error())
			}
		case dockplayer.CapPS, dockplayer.CapInspect, dockplayer.CapLogs, dockplayer.CapRun, dockplayer.CapBuild:
			if err := dockplayer.ValidateStaticInput(r.Workspace, s.Capability, s.Input); err != nil {
				var ve result.Error
				if errors.As(err, &ve) {
					return r.reject(t.TaskID, ve.Code, ve.Message)
				}
				return r.reject(t.TaskID, result.CodeInvalidInput, err.Error())
			}
		case httpplayer.CapGet, httpplayer.CapHead:
			if err := httpplayer.ValidateStaticInput(r.Workspace, s.Capability, s.Input); err != nil {
				var ve result.Error
				if errors.As(err, &ve) {
					return r.reject(t.TaskID, ve.Code, ve.Message)
				}
				return r.reject(t.TaskID, result.CodeInvalidInput, err.Error())
			}
		}
	}
	return nil
}

func (r *Runner) checkAdmissionPolicy(t task.Task) error {
	for _, s := range t.Steps {
		manifest, err := policy.ParseVerb(r.Reg.ManifestPolicy(s.Capability))
		if err != nil {
			return r.reject(t.TaskID, result.CodePolicyDenied, err.Error())
		}
		verb := r.Policy.Verb(s.Capability, manifest)
		if verb == policy.Deny {
			err := policy.DeniedError(s.Capability)
			_ = r.emit("", t.TaskID, nil, event.TypeTaskRejected, map[string]any{
				"code":    err.Code,
				"message": err.Message,
			})
			return err
		}
	}
	return nil
}

func (r *Runner) reject(taskID, code, message string) error {
	_ = r.emit("", taskID, nil, event.TypeTaskRejected, map[string]any{
		"code":    code,
		"message": message,
	})
	return result.Validation(code, message, nil)
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

const (
	DecisionGrant = "grant"
	DecisionDeny  = "deny"
)

func (r *Runner) Approve(ctx context.Context, runID, decision string) error {
	run, taskJSON, err := r.Store.GetRun(ctx, runID)
	if err != nil {
		return result.Runtime(result.CodeInternal, "run not found", false, nil)
	}
	if run.Status != store.StatusWaitingApproval {
		return policy.NotWaitingError()
	}
	stepID := run.PendingStepID
	capName := run.PendingCapability
	player := run.PendingPlayer

	switch decision {
	case DecisionDeny:
		err := policy.ApprovalDeniedError(capName)
		_ = r.emit(runID, run.TaskID, &stepID, event.TypeRunApprovalDenied, map[string]any{
			"step_id":    stepID,
			"capability": capName,
			"player":     player,
		})
		_ = r.Store.ClearPendingApproval(ctx, runID)
		_ = r.Store.UpdateRunStatus(ctx, runID, store.StatusFailed, store.FormatErr(err))
		_ = r.emit(runID, run.TaskID, nil, event.TypeRunFailed, map[string]any{"error": err.Error()})
		r.releaseClaims(runID, run.TaskID)
		r.syncGraph(runID)
		return nil
	case DecisionGrant:
		_ = r.emit(runID, run.TaskID, &stepID, event.TypeRunApprovalGranted, map[string]any{
			"step_id":    stepID,
			"capability": capName,
			"player":     player,
		})
		var t task.Task
		if err := json.Unmarshal(taskJSON, &t); err != nil {
			return result.Runtime(result.CodeInternal, "corrupt task json", false, nil)
		}
		p, err := plan.FromTask(t, runID, func(cap string) (string, error) {
			name, _, err := r.Reg.Resolve(cap, r.LLMBackend)
			return name, err
		})
		if err != nil {
			return err
		}
		skip := stepID
		_ = r.Store.ClearPendingApproval(ctx, runID)
		runCtx, cancel := context.WithCancel(context.Background())
		r.mu.Lock()
		r.cancels[runID] = cancel
		r.mu.Unlock()
		r.inflight.Add(1)
		go func() {
			defer r.inflight.Done()
			r.execute(runCtx, t, p, skip)
		}()
		return nil
	default:
		return result.Runtime(result.CodeInternal, "decision must be grant or deny", false, nil)
	}
}

func (r *Runner) execute(ctx context.Context, t task.Task, p plan.Plan, skipApprovalFor string) {
	r.sem <- struct{}{}
	defer func() { <-r.sem }()
	defer func() {
		r.mu.Lock()
		delete(r.cancels, p.RunID)
		r.mu.Unlock()
	}()

	_ = r.Store.UpdateRunStatus(ctx, p.RunID, store.StatusRunning, "")
	if skipApprovalFor == "" {
		_ = r.emit(p.RunID, t.TaskID, nil, event.TypeRunStarted, nil)
	}

	order, err := task.TopoOrder(t.Steps)
	if err != nil {
		r.fail(ctx, p, err)
		return
	}
	byID := map[string]plan.Step{}
	for _, s := range p.Steps {
		byID[s.StepID] = s
	}

	done := map[string]store.StepOutput{}
	if priorsList, err := r.Store.ListStepOutputs(ctx, p.RunID); err == nil {
		for _, o := range priorsList {
			done[o.StepID] = o
		}
	}

	var priors []store.StepOutput
	for _, sid := range order {
		if ctx.Err() != nil {
			r.cancelled(ctx, p)
			return
		}
		s := byID[sid]
		if prev, ok := done[sid]; ok {
			priors = append(priors, prev)
			continue
		}

		manifest, perr := policy.ParseVerb(r.Reg.ManifestPolicy(s.Capability))
		if perr != nil {
			r.fail(ctx, p, result.Runtime(result.CodeInternal, perr.Error(), false, nil))
			return
		}
		verb := r.Policy.Verb(s.Capability, manifest)
		if verb == policy.Deny {
			err := policy.DeniedError(s.Capability)
			r.fail(ctx, p, err)
			return
		}
		if verb == policy.ApprovalRequired && sid != skipApprovalFor {
			_ = r.Store.SetPendingApproval(ctx, p.RunID, s.StepID, s.Capability, s.Player)
			payload := map[string]any{
				"run_id":     p.RunID,
				"step_id":    s.StepID,
				"capability": s.Capability,
				"player":     s.Player,
			}
			_ = r.emit(p.RunID, t.TaskID, &s.StepID, event.TypeRunWaitingApproval, payload)
			return
		}

		if err := r.acquireStep(ctx, p, t, s); err != nil {
			r.fail(ctx, p, err)
			return
		}

		stepID := s.StepID
		pack := contextpack.Assemble(t, s.StepID, s.Capability, priors)
		pack = r.attachGraphHits(ctx, pack, t)
		packJSON, _ := contextpack.Marshal(pack)
		_ = r.emit(p.RunID, t.TaskID, &stepID, event.TypeStepStarted, map[string]any{
			"capability":    s.Capability,
			"player":        s.Player,
			"context_bytes": len(packJSON),
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
				Capability:   s.Capability,
				Input:        s.Input,
				Workspace:    r.Workspace,
				RunID:        p.RunID,
				TaskID:       t.TaskID,
				StepID:       s.StepID,
				Context:      packJSON,
				PriorOutputs: priors,
			})
			if lastErr == nil {
				_ = r.Store.SaveStepOutput(context.Background(), p.RunID, s.StepID, s.Capability, out)
				priors = append(priors, store.StepOutput{StepID: s.StepID, Capability: s.Capability, Output: out})
				if s.Capability == corepipe.CapDecompose {
					r.persistAndSpawn(t, p.RunID, out)
				}
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

	_ = r.Store.ClearPendingApproval(ctx, p.RunID)
	_ = r.Store.UpdateRunStatus(ctx, p.RunID, store.StatusSucceeded, "")
	_ = r.emit(p.RunID, t.TaskID, nil, event.TypeRunSucceeded, nil)
	r.releaseClaims(p.RunID, t.TaskID)
	r.syncGraph(p.RunID)
}

func (r *Runner) fail(ctx context.Context, p plan.Plan, err error) {
	_ = r.Store.UpdateRunStatus(ctx, p.RunID, store.StatusFailed, store.FormatErr(err))
	_ = r.emit(p.RunID, p.TaskID, nil, event.TypeRunFailed, map[string]any{"error": err.Error()})
	r.releaseClaims(p.RunID, p.TaskID)
	r.syncGraph(p.RunID)
}

func (r *Runner) cancelled(ctx context.Context, p plan.Plan) {
	_ = r.Store.UpdateRunStatus(ctx, p.RunID, store.StatusCancelled, "")
	_ = r.emit(p.RunID, p.TaskID, nil, event.TypeRunCancelled, nil)
	r.releaseClaims(p.RunID, p.TaskID)
	r.syncGraph(p.RunID)
}

func (r *Runner) syncGraph(runID string) {
	if r.Graph == nil || runID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.Graph.SyncFromRun(ctx, runID); err != nil {
		r.Log.Warn("graph sync failed", "run_id", runID, "err", err)
	}
}

func (r *Runner) attachGraphHits(ctx context.Context, pack contextpack.Pack, t task.Task) contextpack.Pack {
	if r.Graph == nil {
		return pack
	}
	text := t.Intent.Summary
	if t.Intent.Notes != "" && len(t.Intent.Notes) < 500 {
		if text != "" {
			text += " "
		}
		text += t.Intent.Notes
	}
	hits := r.Graph.QueryHits(ctx, graph.Query{
		Text:           text,
		SeedPaths:      pack.RepoHits.Paths,
		SeedSymbols:    pack.RepoHits.Symbols,
		SeedCapability: pack.Step.Capability,
		Limit:          pack.Budget.GraphMaxHits,
		MaxChars:       pack.Budget.GraphMaxChars,
	})
	items := make([]contextpack.GraphHit, 0, len(hits.Items))
	for _, h := range hits.Items {
		items = append(items, contextpack.GraphHit{
			Kind: h.Kind, ID: h.ID, Reason: h.Reason, Score: h.Score,
		})
	}
	return contextpack.WithGraphHits(pack, items)
}

func (r *Runner) acquireStep(ctx context.Context, p plan.Plan, t task.Task, s plan.Step) error {
	if r.Claims == nil {
		return nil
	}
	res, ok, err := claim.Required(s.Capability, s.Input)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if err := r.Claims.Acquire(ctx, p.RunID, s.StepID, res); err != nil {
		var ve result.Error
		if errors.As(err, &ve) && ve.Code == result.CodeClaimConflict {
			holder, _ := ve.Details["holder_run_id"].(string)
			_ = r.emit(p.RunID, t.TaskID, &s.StepID, event.TypeClaimConflict, map[string]any{
				"kind":          string(res.Kind),
				"key":           res.Key,
				"holder_run_id": holder,
			})
		}
		return err
	}
	_ = r.emit(p.RunID, t.TaskID, &s.StepID, event.TypeClaimAcquired, map[string]any{
		"kind":       string(res.Kind),
		"key":        res.Key,
		"step_id":    s.StepID,
		"capability": s.Capability,
	})
	return nil
}

func (r *Runner) ReleaseClaims(runID, taskID string) {
	r.releaseClaims(runID, taskID)
}

func (r *Runner) releaseClaims(runID, taskID string) {
	if r.Claims == nil || runID == "" {
		return
	}
	released, err := r.Claims.ReleaseAll(context.Background(), runID)
	if err != nil {
		r.Log.Warn("claim release failed", "run_id", runID, "err", err)
		return
	}
	for _, res := range released {
		_ = r.emit(runID, taskID, nil, event.TypeClaimReleased, map[string]any{
			"kind": string(res.Kind),
			"key":  res.Key,
		})
	}
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

func (r *Runner) persistAndSpawn(parent task.Task, parentRunID string, decomposeOut json.RawMessage) {
	var parsed struct {
		Subtasks []struct {
			SubtaskID           string `json:"subtask_id"`
			Summary             string `json:"summary"`
			SuggestedCapability string `json:"suggested_capability"`
			Notes               string `json:"notes"`
		} `json:"subtasks"`
	}
	if err := json.Unmarshal(decomposeOut, &parsed); err != nil {
		return
	}
	for _, st := range parsed.Subtasks {
		if st.SubtaskID == "" {
			id, err := uuid.NewV7()
			if err != nil {
				continue
			}
			st.SubtaskID = id.String()
		}
		rec := store.Subtask{
			SubtaskID:           st.SubtaskID,
			ParentRunID:         parentRunID,
			TaskID:              parent.TaskID,
			Summary:             st.Summary,
			SuggestedCapability: st.SuggestedCapability,
			Notes:               st.Notes,
		}
		_ = r.Store.InsertSubtask(context.Background(), rec)
		child, ok := childTask(parent, st.Summary, st.SuggestedCapability, st.Notes)
		if !ok {
			continue
		}
		res, err := r.SubmitChild(context.Background(), child, parentRunID)
		if err != nil {
			r.Log.Warn("child run failed to submit", "err", err, "subtask", st.SubtaskID)
			continue
		}
		_ = r.Store.SetSubtaskChildRun(context.Background(), st.SubtaskID, res.RunID)
	}
}

func childTask(parent task.Task, summary, capName, notes string) (task.Task, bool) {
	if capName == "" {
		capName = "shell.exec"
	}
	if strings.HasPrefix(capName, "pipeline.") {
		return task.Task{}, false
	}
	input := json.RawMessage(`{}`)
	if capName == "shell.exec" {
		msg := summary
		if msg == "" {
			msg = parent.Intent.Summary
		}
		b, _ := json.Marshal(map[string]any{
			"command":    []string{"echo", msg},
			"workdir":    ".",
			"timeout_ms": 5000,
		})
		input = b
	}
	id, err := task.NewID()
	if err != nil {
		return task.Task{}, false
	}
	return task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id,
		CreatedAt:     time.Now().UTC(),
		Source:        task.Source{EntryPoint: parent.Source.EntryPoint, Ref: parent.Source.Ref},
		Intent:        task.Intent{Summary: summary, Notes: notes},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: capName,
			Input:      input,
		}},
		Metadata: map[string]any{"child": true},
	}, true
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
