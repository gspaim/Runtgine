package graph

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

type stubPlayer struct {
	name string
	caps []string
}

func (p stubPlayer) Manifest() registry.Manifest {
	cs := make([]registry.Capability, 0, len(p.caps))
	for _, c := range p.caps {
		cs = append(cs, registry.Capability{Name: c})
	}
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          p.name,
		Version:       "1.0.0",
		Kind:          registry.KindDeterministic,
		Capabilities:  cs,
	}
}

func (p stubPlayer) Execute(context.Context, registry.ExecRequest) (json.RawMessage, error) {
	return nil, nil
}

func openGraph(t *testing.T) *Service {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "runtgine.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return New(st, nil)
}

func TestUpsertIdempotent(t *testing.T) {
	g := openGraph(t)
	ctx := context.Background()
	if err := g.UpsertNode(ctx, KindPath, "cmd/runtgine/main.go", nil); err != nil {
		t.Fatal(err)
	}
	if err := g.UpsertNode(ctx, KindPath, "cmd/runtgine/main.go", nil); err != nil {
		t.Fatal(err)
	}
	if err := g.UpsertEdge(ctx, EdgeMentions, KindRun, "r1", KindPath, "cmd/runtgine/main.go", nil); err != nil {
		t.Fatal(err)
	}
	if err := g.UpsertEdge(ctx, EdgeMentions, KindRun, "r1", KindPath, "cmd/runtgine/main.go", nil); err != nil {
		t.Fatal(err)
	}
	snap, err := g.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Nodes) != 1 {
		t.Fatalf("nodes=%d", len(snap.Nodes))
	}
	if len(snap.Edges) != 1 {
		t.Fatalf("edges=%d", len(snap.Edges))
	}
}

func TestRefreshFromRegistry(t *testing.T) {
	g := openGraph(t)
	reg := registry.New()
	if err := reg.Register(stubPlayer{name: "shell", caps: []string{"shell.exec"}}); err != nil {
		t.Fatal(err)
	}
	if err := g.RefreshFromRegistry(context.Background(), reg); err != nil {
		t.Fatal(err)
	}
	n, err := g.GetNode(context.Background(), KindPlayer, "shell")
	if err != nil {
		t.Fatal(err)
	}
	if n.Attrs["version"] != "1.0.0" {
		t.Fatalf("attrs=%v", n.Attrs)
	}
	neigh, err := g.QueryNeighbors(context.Background(), KindPlayer, "shell", EdgeProvides, "out")
	if err != nil {
		t.Fatal(err)
	}
	if len(neigh) != 1 || neigh[0].Kind != KindCapability || neigh[0].ID != "shell.exec" {
		t.Fatalf("neighbors=%v", neigh)
	}
}

func TestSyncFromRunMentions(t *testing.T) {
	g := openGraph(t)
	ctx := context.Background()
	runID := mustV7(t)
	taskID := mustV7(t)
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        taskID,
		CreatedAt:     time.Now().UTC(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "search"},
		Steps:         []task.Step{{StepID: "repo-search", Capability: pipeline.CapRepoSearch, Input: json.RawMessage(`{}`)}},
	}
	raw, _ := json.Marshal(tk)
	if err := g.Store.InsertRun(ctx, store.Run{
		RunID: runID, TaskID: taskID, Status: store.StatusSucceeded,
	}, raw); err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(map[string]any{
		"paths":   []string{"internal/core/graph/graph.go"},
		"symbols": []string{"func New"},
	})
	if err := g.Store.SaveStepOutput(ctx, runID, "repo-search", pipeline.CapRepoSearch, out); err != nil {
		t.Fatal(err)
	}
	if err := g.SyncFromRun(ctx, runID); err != nil {
		t.Fatal(err)
	}
	if err := g.SyncFromRun(ctx, runID); err != nil {
		t.Fatal(err)
	}

	snap, err := g.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !hasNode(snap, KindRun, runID) || !hasNode(snap, KindTask, taskID) {
		t.Fatalf("missing run/task: %+v", snap.Nodes)
	}
	if !hasEdge(snap, EdgeInstanceOf, KindRun, runID, KindTask, taskID) {
		t.Fatal("missing instance_of")
	}
	if !hasEdge(snap, EdgeExecuted, KindRun, runID, KindCapability, pipeline.CapRepoSearch) {
		t.Fatal("missing executed")
	}
	if !hasEdge(snap, EdgeMentions, KindRun, runID, KindPath, "internal/core/graph/graph.go") {
		t.Fatal("missing path mention")
	}
	if !hasEdge(snap, EdgeMentions, KindRun, runID, KindSymbol, "func New") {
		t.Fatal("missing symbol mention")
	}

	var mentionEdges int
	for _, e := range snap.Edges {
		if e.Kind == EdgeMentions {
			mentionEdges++
		}
	}
	if mentionEdges != 2 {
		t.Fatalf("mention edges=%d", mentionEdges)
	}
}

func TestSyncFromRunChildOf(t *testing.T) {
	g := openGraph(t)
	ctx := context.Background()
	parent := mustV7(t)
	child := mustV7(t)
	taskID := mustV7(t)
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        taskID,
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "child"},
		Steps:         []task.Step{{StepID: "s1", Capability: "shell.exec", Input: json.RawMessage(`{}`)}},
	}
	raw, _ := json.Marshal(tk)
	if err := g.Store.InsertRun(ctx, store.Run{
		RunID: parent, TaskID: taskID, Status: store.StatusSucceeded,
	}, raw); err != nil {
		t.Fatal(err)
	}
	if err := g.Store.InsertRun(ctx, store.Run{
		RunID: child, TaskID: taskID, ParentRunID: parent, Status: store.StatusSucceeded,
	}, raw); err != nil {
		t.Fatal(err)
	}
	if err := g.SyncFromRun(ctx, child); err != nil {
		t.Fatal(err)
	}
	snap, err := g.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !hasEdge(snap, EdgeChildOf, KindRun, child, KindRun, parent) {
		t.Fatal("missing child_of")
	}
}

func mustV7(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return id.String()
}

func hasNode(snap Snapshot, kind, id string) bool {
	for _, n := range snap.Nodes {
		if n.Kind == kind && n.ID == id {
			return true
		}
	}
	return false
}

func hasEdge(snap Snapshot, kind, fromKind, fromID, toKind, toID string) bool {
	for _, e := range snap.Edges {
		if e.Kind == kind && e.From.Kind == fromKind && e.From.ID == fromID && e.To.Kind == toKind && e.To.ID == toID {
			return true
		}
	}
	return false
}
