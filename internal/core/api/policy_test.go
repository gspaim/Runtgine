package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

func openPolicyCore(t *testing.T, caps map[string]string) *api.Core {
	t.Helper()
	ws := t.TempDir()
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	cfg.ExecutionPolicy.Capabilities = caps
	core, err := api.Open(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = core.Close() })
	return core
}

func helloTask(t *testing.T) task.Task {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "policy hello"},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "shell.exec",
			Input:      json.RawMessage(`{"command":["echo","hello-runtgine"],"workdir":".","timeout_ms":5000}`),
		}},
	}
}

func waitStatus(t *testing.T, core *api.Core, runID string, want store.Status) api.RunSnapshot {
	t.Helper()
	deadline := time.Now().Add(4 * time.Second)
	var snap api.RunSnapshot
	for time.Now().Before(deadline) {
		var err error
		snap, err = core.GetRun(context.Background(), runID)
		if err == nil && store.Status(snap.Status) == want {
			return snap
		}
		time.Sleep(15 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s (last=%s err=%s)", want, snap.Status, snap.Error)
	return snap
}

func TestPolicyDefaultAllowHelloSucceeds(t *testing.T) {
	core := openPolicyCore(t, nil)
	runID, err := core.SubmitTask(context.Background(), helloTask(t))
	if err != nil {
		t.Fatal(err)
	}
	snap := waitStatus(t, core, runID, store.StatusSucceeded)
	if snap.PendingApproval != nil {
		t.Fatal("unexpected pending")
	}
}

func TestPolicyDenyRejectsWithoutExecute(t *testing.T) {
	core := openPolicyCore(t, map[string]string{"shell.exec": "deny"})
	_, err := core.SubmitTask(context.Background(), helloTask(t))
	if err == nil {
		t.Fatal("expected deny")
	}
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodePolicyDenied {
		t.Fatalf("got %#v", err)
	}
}

func TestPolicyUnknownCapabilityInConfigFailsOpen(t *testing.T) {
	ws := t.TempDir()
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	cfg.ExecutionPolicy.Capabilities = map[string]string{"nope.cap": "deny"}
	_, err := api.Open(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err == nil {
		t.Fatal("expected open error")
	}
}

func TestHITLGrantAndDeny(t *testing.T) {
	core := openPolicyCore(t, map[string]string{"shell.exec": "approval-required"})
	runID, err := core.SubmitTask(context.Background(), helloTask(t))
	if err != nil {
		t.Fatal(err)
	}
	snap := waitStatus(t, core, runID, store.StatusWaitingApproval)
	if snap.PendingApproval == nil || snap.PendingApproval.Capability != "shell.exec" {
		t.Fatalf("pending=%#v", snap.PendingApproval)
	}
	found := false
	for _, e := range snap.Events {
		if e.Type == event.TypeRunWaitingApproval {
			found = true
		}
	}
	if !found {
		t.Fatal("missing waiting event")
	}
	if err := core.ApproveRun(runID, runner.DecisionGrant); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, core, runID, store.StatusSucceeded)

	runID, err = core.SubmitTask(context.Background(), helloTask(t))
	if err != nil {
		t.Fatal(err)
	}
	waitStatus(t, core, runID, store.StatusWaitingApproval)
	if err := core.ApproveRun(runID, runner.DecisionDeny); err != nil {
		t.Fatal(err)
	}
	snap = waitStatus(t, core, runID, store.StatusFailed)
	if !containsCode(snap.Error, result.CodePolicyApprovalDenied) {
		t.Fatalf("error=%s", snap.Error)
	}
}

func TestDockerRunWaitsForHITL(t *testing.T) {
	core := openPolicyCore(t, nil)
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "docker run gated"},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "docker.run",
			Input:      json.RawMessage(`{"image":"alpine:3.19","argv":["echo","hi"]}`),
		}},
	}
	runID, err := core.SubmitTask(context.Background(), tk)
	if err != nil {
		t.Fatal(err)
	}
	snap := waitStatus(t, core, runID, store.StatusWaitingApproval)
	if snap.PendingApproval == nil || snap.PendingApproval.Capability != "docker.run" {
		t.Fatalf("pending=%#v", snap.PendingApproval)
	}
	if err := core.ApproveRun(runID, runner.DecisionDeny); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, core, runID, store.StatusFailed)
}

func TestApproveNotWaiting(t *testing.T) {
	core := openPolicyCore(t, nil)
	runID, err := core.SubmitTask(context.Background(), helloTask(t))
	if err != nil {
		t.Fatal(err)
	}
	waitStatus(t, core, runID, store.StatusSucceeded)
	err = core.ApproveRun(runID, runner.DecisionGrant)
	if err == nil {
		t.Fatal("expected not_waiting")
	}
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodePolicyNotWaiting {
		t.Fatalf("got %#v", err)
	}
}

func TestHITLRestartThenGrant(t *testing.T) {
	ws := t.TempDir()
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	cfg.ExecutionPolicy.Capabilities = map[string]string{"shell.exec": "approval-required"}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	core, err := api.Open(cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	runID, err := core.SubmitTask(context.Background(), helloTask(t))
	if err != nil {
		t.Fatal(err)
	}
	waitStatus(t, core, runID, store.StatusWaitingApproval)
	if err := core.Close(); err != nil {
		t.Fatal(err)
	}

	core2, err := api.Open(cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	defer core2.Close()
	snap, err := core2.GetRun(context.Background(), runID)
	if err != nil || snap.Status != string(store.StatusWaitingApproval) {
		t.Fatalf("after restart status=%s err=%v", snap.Status, err)
	}
	if err := core2.ApproveRun(runID, runner.DecisionGrant); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, core2, runID, store.StatusSucceeded)
}

func containsCode(errJSON, code string) bool {
	return strings.Contains(errJSON, code)
}
