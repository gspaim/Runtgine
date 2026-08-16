package api_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/task"
	"github.com/google/uuid"
)

func openTestCore(t *testing.T) *api.Core {
	t.Helper()
	ws := t.TempDir()
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	if err := config.EnsureRuntimeDir(cfg); err != nil {
		t.Fatal(err)
	}
	core, err := api.Open(cfg, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = core.Close() })
	return core
}

func TestSubmitRejectsInvalidShellInput(t *testing.T) {
	core := openTestCore(t)
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "bad input"},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "shell.exec",
			Input:      json.RawMessage(`{"command":["echo"],"timeout_ms":"nope"}`),
		}},
	}
	_, err = core.SubmitTask(context.Background(), tk)
	if err == nil {
		t.Fatal("expected rejection")
	}
	var ve result.Error
	if !asValidation(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("got %v", err)
	}
}

func TestSubmitRejectsExtraShellProperty(t *testing.T) {
	core := openTestCore(t)
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "extra"},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "shell.exec",
			Input:      json.RawMessage(`{"command":["echo","hi"],"unexpected":true}`),
		}},
	}
	_, err = core.SubmitTask(context.Background(), tk)
	if err == nil {
		t.Fatal("expected rejection")
	}
}

func TestSubmitAcceptsHelloShape(t *testing.T) {
	core := openTestCore(t)
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "hello.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := task.ValidateDocument(raw); err != nil {
		t.Fatal(err)
	}
	tk, err := task.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	runID, err := core.SubmitTask(context.Background(), tk)
	if err != nil {
		t.Fatal(err)
	}
	if runID == "" {
		t.Fatal("empty run id")
	}
}

func asValidation(err error, dest *result.Error) bool {
	if err == nil {
		return false
	}
	if ve, ok := err.(result.Error); ok {
		*dest = ve
		return true
	}
	return false
}
