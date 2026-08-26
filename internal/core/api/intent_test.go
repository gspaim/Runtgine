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

func TestCompileIntentTemplate(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, ".runtgine", "templates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{
  "id": "verify",
  "title": "Verify",
  "steps": [
    {"step_id": "status", "capability": "git.status", "input": {"workdir": "."}},
    {"step_id": "test", "capability": "test.go", "input": {"workdir": "."}, "depends_on": ["status"]}
  ]
}`)
	if err := os.WriteFile(filepath.Join(dir, "verify.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	core, err := api.Open(cfg, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = core.Close() })

	tk, method, err := core.CompileIntent(context.Background(), "run template verify", "cli", "test")
	if err != nil {
		t.Fatal(err)
	}
	if method != intent.MethodHeuristicTemplate {
		t.Fatalf("method=%s", method)
	}
	if len(tk.Steps) != 2 || tk.Metadata["template"] != "verify" {
		t.Fatalf("task=%+v", tk)
	}
	if _, _, err := core.CompileIntent(context.Background(), "run template missing", "cli", "test"); err == nil {
		t.Fatal("expected unknown template")
	}
	snap, err := core.GetGraphSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range snap.Nodes {
		if n.Kind == "template" && n.ID == "verify" {
			found = true
		}
	}
	if !found {
		t.Fatal("graph missing template node")
	}
}
