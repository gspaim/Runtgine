package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/graph"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

func TestGraphRefreshOnOpen(t *testing.T) {
	core := openTestCore(t)
	snap, err := core.GetGraphSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !snapshotHas(snap, graph.KindPlayer, "shell") {
		t.Fatalf("missing shell player: %+v", snap.Nodes)
	}
	if !snapshotHas(snap, graph.KindCapability, "shell.exec") {
		t.Fatal("missing shell.exec")
	}
	if !snapshotEdge(snap, graph.EdgeProvides, graph.KindPlayer, "shell", graph.KindCapability, "shell.exec") {
		t.Fatal("missing provides")
	}
	if !snapshotHas(snap, graph.KindPlayer, "http") {
		t.Fatal("missing http player")
	}
	if !snapshotHas(snap, graph.KindCapability, "http.get") || !snapshotHas(snap, graph.KindCapability, "http.head") {
		t.Fatal("missing http.get/http.head")
	}
	if !snapshotHas(snap, graph.KindPlayer, "test") {
		t.Fatal("missing test player")
	}
	if !snapshotHas(snap, graph.KindCapability, "test.go") {
		t.Fatal("missing test.go")
	}
}

func TestGraphSyncFromSuccessfulRun(t *testing.T) {
	core := openTestCore(t)
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "hello graph"},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "shell.exec",
			Input:      json.RawMessage(`{"command":["echo","hello-runtgine"],"workdir":".","timeout_ms":5000}`),
		}},
	}
	runID, err := core.SubmitTask(context.Background(), tk)
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, core, runID, store.StatusSucceeded)

	var snap graph.Snapshot
	deadline := time.Now().Add(2 * time.Second)
	for {
		var err error
		snap, err = core.GetGraphSnapshot(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if snapshotHas(snap, graph.KindRun, runID) &&
			snapshotEdge(snap, graph.EdgeExecuted, graph.KindRun, runID, graph.KindCapability, "shell.exec") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("graph did not sync run: %+v", snap)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !snapshotHas(snap, graph.KindRun, runID) {
		t.Fatal("missing run node")
	}
	if !snapshotHas(snap, graph.KindTask, tk.TaskID) {
		t.Fatal("missing task node")
	}
	if !snapshotEdge(snap, graph.EdgeInstanceOf, graph.KindRun, runID, graph.KindTask, tk.TaskID) {
		t.Fatal("missing instance_of")
	}
	if !snapshotEdge(snap, graph.EdgeExecuted, graph.KindRun, runID, graph.KindCapability, "shell.exec") {
		t.Fatal("missing executed")
	}
}

func TestQueryHitsAfterMentionsSync(t *testing.T) {
	core := openTestCore(t)
	ctx := context.Background()
	path := "internal/core/graph/hits.go"
	_ = core.Graph.UpsertNode(ctx, graph.KindPath, path, nil)
	_ = core.Graph.UpsertNode(ctx, graph.KindRun, "prior-run", nil)
	_ = core.Graph.UpsertEdge(ctx, graph.EdgeMentions, graph.KindRun, "prior-run", graph.KindPath, path, nil)

	hits := core.Graph.QueryHits(ctx, graph.Query{
		SeedPaths: []string{path},
		Text:      "hits graph",
	})
	found := false
	for _, h := range hits.Items {
		if h.Kind == graph.KindPath && h.ID == path && h.Reason == "seed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected seed path hit, got %+v", hits.Items)
	}
}

func TestGraphFailureDoesNotFailRun(t *testing.T) {
	core := openTestCore(t)
	core.Runner.Graph = failingGraph{}
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "graph sink fails"},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "shell.exec",
			Input:      json.RawMessage(`{"command":["echo","ok"],"workdir":".","timeout_ms":5000}`),
		}},
	}
	runID, err := core.SubmitTask(context.Background(), tk)
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, core, runID, store.StatusSucceeded)
}

func TestGraphRefreshIdempotent(t *testing.T) {
	core := openTestCore(t)
	ctx := context.Background()
	if err := core.RefreshGraph(ctx); err != nil {
		t.Fatal(err)
	}
	first, err := core.GetGraphSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := core.RefreshGraph(ctx); err != nil {
		t.Fatal(err)
	}
	second, err := core.GetGraphSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Nodes) != len(second.Nodes) || len(first.Edges) != len(second.Edges) {
		t.Fatalf("nodes %d→%d edges %d→%d", len(first.Nodes), len(second.Nodes), len(first.Edges), len(second.Edges))
	}
}

type failingGraph struct{}

func (failingGraph) SyncFromRun(context.Context, string) error {
	return errors.New("graph unavailable")
}

func (failingGraph) QueryHits(context.Context, graph.Query) graph.Hits {
	return graph.Hits{Items: []graph.Hit{}}
}

func waitTerminal(t *testing.T, core *api.Core, runID string, want store.Status) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		snap, err := core.GetRun(context.Background(), runID)
		if err != nil {
			t.Fatal(err)
		}
		st := store.Status(snap.Status)
		switch st {
		case want:
			return
		case store.StatusFailed, store.StatusCancelled, store.StatusRejected:
			t.Fatalf("run %s: %s", snap.Status, snap.Error)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timeout")
}

func snapshotHas(snap graph.Snapshot, kind, id string) bool {
	for _, n := range snap.Nodes {
		if n.Kind == kind && n.ID == id {
			return true
		}
	}
	return false
}

func snapshotEdge(snap graph.Snapshot, kind, fromKind, fromID, toKind, toID string) bool {
	for _, e := range snap.Edges {
		if e.Kind == kind && e.From.Kind == fromKind && e.From.ID == fromID && e.To.Kind == toKind && e.To.ID == toID {
			return true
		}
	}
	return false
}
