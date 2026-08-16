package api_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/store"
	"log/slog"
)

func TestPipelineRunOffline(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "main.go"), []byte("package main\nfunc Hello() {}\n"), 0o644)
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
	defer core.Close()

	tk, err := corepipe.NewTaskIR("Add greeting", "small change", "cli", "test")
	if err != nil {
		t.Fatal(err)
	}
	runID, err := core.SubmitTask(context.Background(), tk)
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		snap, err := core.GetRun(context.Background(), runID)
		if err != nil {
			t.Fatal(err)
		}
		switch store.Status(snap.Status) {
		case store.StatusSucceeded:
			if len(snap.Subtasks) == 0 {
				t.Fatal("expected persisted subtasks")
			}
			return
		case store.StatusFailed, store.StatusCancelled, store.StatusRejected:
			t.Fatalf("run %s: %s", snap.Status, snap.Error)
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timeout waiting for pipeline")
}
