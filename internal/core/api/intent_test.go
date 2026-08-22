package api_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/intent"
	"github.com/gspaim/Runtgine/internal/core/store"
)

func TestSubmitIntentShell(t *testing.T) {
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
	defer core.Close()

	runID, method, err := core.SubmitIntent(context.Background(), "echo hello-intent", "cli", "test")
	if err != nil {
		t.Fatal(err)
	}
	if method != intent.MethodHeuristicShell {
		t.Fatalf("method=%s", method)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		snap, err := core.GetRun(context.Background(), runID)
		if err != nil {
			t.Fatal(err)
		}
		switch store.Status(snap.Status) {
		case store.StatusSucceeded:
			return
		case store.StatusFailed, store.StatusCancelled, store.StatusRejected:
			t.Fatalf("run %s: %s", snap.Status, snap.Error)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timeout")
}

func TestCompileIntentGoTest(t *testing.T) {
	core := openTestCore(t)
	tk, method, err := core.CompileIntent(context.Background(), "go test", "cli", "test")
	if err != nil {
		t.Fatal(err)
	}
	if method != intent.MethodHeuristicTest {
		t.Fatalf("method=%s", method)
	}
	if len(tk.Steps) != 1 || tk.Steps[0].Capability != "test.go" {
		t.Fatalf("steps=%v", tk.Steps)
	}
}

func TestCompileIntentNpmTest(t *testing.T) {
	core := openTestCore(t)
	tk, method, err := core.CompileIntent(context.Background(), "npm test", "cli", "test")
	if err != nil {
		t.Fatal(err)
	}
	if method != intent.MethodHeuristicNPM {
		t.Fatalf("method=%s", method)
	}
	if len(tk.Steps) != 1 || tk.Steps[0].Capability != "npm.test" {
		t.Fatalf("steps=%v", tk.Steps)
	}
}
