package api_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/store"
)

func TestTUIQueriesAndPersistedCancellation(t *testing.T) {
	workspace := t.TempDir()
	cfg := config.Defaults()
	cfg.WorkspaceRoot = workspace
	cfg.DBPath = filepath.Join(workspace, ".runtgine", "runtgine.db")
	cfg.LLMAPIKey = "must-not-be-exposed"
	cfg.GitHubToken = "must-not-be-exposed"
	core, err := api.Open(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer core.Close()

	taskJSON := []byte(`{
		"source":{"entry_point":"board","ref":"owner/repo#7"},
		"intent":{"summary":"Inspect issue"},
		"steps":[{"step_id":"review","capability":"pipeline.tech-review","input":{}}]
	}`)
	run := store.Run{RunID: "run-stale", TaskID: "task-stale", Status: store.StatusRunning}
	if err := core.Store.InsertRun(context.Background(), run, taskJSON); err != nil {
		t.Fatal(err)
	}

	runs, err := core.ListRuns(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].Summary != "Inspect issue" || runs[0].Source != "board" {
		t.Fatalf("unexpected summaries: %#v", runs)
	}

	if err := core.CancelRun(run.RunID); err != nil {
		t.Fatal(err)
	}
	snapshot, err := core.GetRun(context.Background(), run.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Status != string(store.StatusCancelled) {
		t.Fatalf("status=%q, want cancelled", snapshot.Status)
	}
	if len(snapshot.Events) != 1 || snapshot.Events[0].Type != event.TypeRunCancelled {
		t.Fatalf("unexpected events: %#v", snapshot.Events)
	}

	recent, err := core.ListRecentEvents(context.Background(), 10)
	if err != nil || len(recent) != 1 {
		t.Fatalf("recent events=%d err=%v", len(recent), err)
	}

	configJSON, err := json.Marshal(core.ConfigSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if string(configJSON) == "" || containsSecret(string(configJSON)) {
		t.Fatalf("unsafe config snapshot: %s", configJSON)
	}
}

func containsSecret(value string) bool {
	return strings.Contains(value, "must-not-be-exposed") ||
		strings.Contains(value, "api_key") ||
		strings.Contains(value, "token")
}
