package graph

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

const (
	KindPlayer     = "player"
	KindCapability = "capability"
	KindTask       = "task"
	KindRun        = "run"
	KindPath       = "path"
	KindSymbol     = "symbol"

	EdgeProvides   = "provides"
	EdgeExecuted   = "executed"
	EdgeInstanceOf = "instance_of"
	EdgeMentions   = "mentions"
	EdgeChildOf    = "child_of"
)

type Node struct {
	Kind  string         `json:"kind"`
	ID    string         `json:"id"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

type Edge struct {
	Kind string `json:"kind"`
	From Ref    `json:"from"`
	To   Ref    `json:"to"`
}

type Ref struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type Snapshot struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Service is the Runtime Graph v0 (G-60+). Store owns SQL; this package owns kinds and sync.
type Service struct {
	Store *store.Store
	Log   *slog.Logger
}

func New(st *store.Store, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{Store: st, Log: log}
}

func (s *Service) UpsertNode(ctx context.Context, kind, id string, attrs map[string]any) error {
	return s.Store.UpsertGraphNode(ctx, kind, id, marshalAttrs(attrs))
}

func (s *Service) UpsertEdge(ctx context.Context, kind, fromKind, fromID, toKind, toID string, attrs map[string]any) error {
	return s.Store.UpsertGraphEdge(ctx, kind, fromKind, fromID, toKind, toID, marshalAttrs(attrs))
}

func (s *Service) GetNode(ctx context.Context, kind, id string) (Node, error) {
	n, err := s.Store.GetGraphNode(ctx, kind, id)
	if err != nil {
		return Node{}, err
	}
	return toNode(n), nil
}

func (s *Service) QueryNeighbors(ctx context.Context, kind, id, edgeKind, direction string) ([]Node, error) {
	nodes, err := s.Store.QueryGraphNeighbors(ctx, kind, id, edgeKind, direction)
	if err != nil {
		return nil, err
	}
	out := make([]Node, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, toNode(n))
	}
	return out, nil
}

func (s *Service) Snapshot(ctx context.Context) (Snapshot, error) {
	nodes, err := s.Store.ListGraphNodes(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	edges, err := s.Store.ListGraphEdges(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	snap := Snapshot{
		Nodes: make([]Node, 0, len(nodes)),
		Edges: make([]Edge, 0, len(edges)),
	}
	for _, n := range nodes {
		snap.Nodes = append(snap.Nodes, toNode(n))
	}
	for _, e := range edges {
		snap.Edges = append(snap.Edges, Edge{
			Kind: e.Kind,
			From: Ref{Kind: e.FromKind, ID: e.FromID},
			To:   Ref{Kind: e.ToKind, ID: e.ToID},
		})
	}
	return snap, nil
}

func (s *Service) RefreshFromRegistry(ctx context.Context, reg *registry.Registry) error {
	for _, m := range reg.Manifests() {
		if err := s.UpsertNode(ctx, KindPlayer, m.Name, map[string]any{
			"version": m.Version,
			"kind":    string(m.Kind),
		}); err != nil {
			return err
		}
		for _, c := range m.Capabilities {
			if err := s.UpsertNode(ctx, KindCapability, c.Name, nil); err != nil {
				return err
			}
			if err := s.UpsertEdge(ctx, EdgeProvides, KindPlayer, m.Name, KindCapability, c.Name, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) SyncFromRun(ctx context.Context, runID string) error {
	run, taskJSON, err := s.Store.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	var tk task.Task
	_ = json.Unmarshal(taskJSON, &tk)

	runAttrs := map[string]any{"status": string(run.Status)}
	if tk.Intent.Summary != "" {
		runAttrs["summary"] = tk.Intent.Summary
	}
	if err := s.UpsertNode(ctx, KindRun, run.RunID, runAttrs); err != nil {
		return err
	}
	taskAttrs := map[string]any{}
	if tk.Intent.Summary != "" {
		taskAttrs["summary"] = tk.Intent.Summary
	}
	if err := s.UpsertNode(ctx, KindTask, run.TaskID, taskAttrs); err != nil {
		return err
	}
	if err := s.UpsertEdge(ctx, EdgeInstanceOf, KindRun, run.RunID, KindTask, run.TaskID, nil); err != nil {
		return err
	}
	if run.ParentRunID != "" {
		if err := s.UpsertEdge(ctx, EdgeChildOf, KindRun, run.RunID, KindRun, run.ParentRunID, nil); err != nil {
			return err
		}
	}

	outputs, err := s.Store.ListStepOutputs(ctx, runID)
	if err != nil {
		return err
	}
	for _, o := range outputs {
		if o.Capability == "" {
			continue
		}
		if err := s.UpsertNode(ctx, KindCapability, o.Capability, nil); err != nil {
			return err
		}
		if err := s.UpsertEdge(ctx, EdgeExecuted, KindRun, run.RunID, KindCapability, o.Capability, nil); err != nil {
			return err
		}
		if o.Capability == pipeline.CapRepoSearch {
			if err := s.syncMentions(ctx, run.RunID, o.Output); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) syncMentions(ctx context.Context, runID string, raw json.RawMessage) error {
	var parsed struct {
		Paths   []string `json:"paths"`
		Symbols []string `json:"symbols"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil
	}
	for _, p := range parsed.Paths {
		if p == "" {
			continue
		}
		if err := s.UpsertNode(ctx, KindPath, p, nil); err != nil {
			return err
		}
		if err := s.UpsertEdge(ctx, EdgeMentions, KindRun, runID, KindPath, p, nil); err != nil {
			return err
		}
	}
	for _, sym := range parsed.Symbols {
		if sym == "" {
			continue
		}
		if err := s.UpsertNode(ctx, KindSymbol, sym, nil); err != nil {
			return err
		}
		if err := s.UpsertEdge(ctx, EdgeMentions, KindRun, runID, KindSymbol, sym, nil); err != nil {
			return err
		}
	}
	return nil
}

func marshalAttrs(attrs map[string]any) string {
	if len(attrs) == 0 {
		return "{}"
	}
	b, err := json.Marshal(attrs)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func toNode(n store.GraphNode) Node {
	attrs := map[string]any{}
	if n.AttrsJSON != "" && n.AttrsJSON != "{}" {
		_ = json.Unmarshal([]byte(n.AttrsJSON), &attrs)
	}
	out := Node{Kind: n.Kind, ID: n.ID}
	if len(attrs) > 0 {
		out.Attrs = attrs
	}
	return out
}
