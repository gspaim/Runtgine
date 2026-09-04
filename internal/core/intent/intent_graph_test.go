package intent_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"github.com/gspaim/Runtgine/internal/core/graph"
	"github.com/gspaim/Runtgine/internal/core/intent"
	"github.com/gspaim/Runtgine/internal/core/memory"
	"github.com/gspaim/Runtgine/internal/players/llm"
)

type countingGraph struct {
	n atomic.Int32
}

func (c *countingGraph) QueryHits(context.Context, graph.Query) graph.Hits {
	c.n.Add(1)
	return graph.Hits{Items: []graph.Hit{{
		Kind: graph.KindPath, ID: "from-graph.go", Reason: "keyword", Score: 2,
	}}}
}

type captureCompleter struct {
	last contextpack.Pack
}

func (c *captureCompleter) Complete(_ context.Context, pack contextpack.Pack, _ json.RawMessage) (json.RawMessage, error) {
	c.last = pack
	return json.RawMessage(`{"summary":"hi","route":"shell","shell_command":["echo","hi"]}`), nil
}

func TestHeuristicPlayerDoesNotQueryGraph(t *testing.T) {
	g := &countingGraph{}
	e := intent.New(llm.HeuristicCompleter{})
	e.Graph = g
	res, err := e.Compile(context.Background(), intent.Request{Text: "go test"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicTest {
		t.Fatalf("method=%s", res.Method)
	}
	if g.n.Load() != 0 {
		t.Fatalf("QueryHits called %d times", g.n.Load())
	}
}

func TestHeuristicShellDoesNotQueryGraph(t *testing.T) {
	g := &countingGraph{}
	m := &countingMemory{}
	e := intent.New(llm.HeuristicCompleter{})
	e.Graph = g
	e.Memory = m
	res, err := e.Compile(context.Background(), intent.Request{Text: "echo hello-intent"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicShell {
		t.Fatalf("method=%s", res.Method)
	}
	if g.n.Load() != 0 {
		t.Fatalf("QueryHits called %d times", g.n.Load())
	}
	if m.n.Load() != 0 {
		t.Fatalf("Memory.Query called %d times", m.n.Load())
	}
}

func TestLLMPathQueriesGraph(t *testing.T) {
	g := &countingGraph{}
	cap := &captureCompleter{}
	e := intent.New(cap)
	e.Graph = g
	res, err := e.Compile(context.Background(), intent.Request{Text: "prepare a friendly greeting for the team"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodLLM {
		t.Fatalf("method=%s", res.Method)
	}
	if g.n.Load() != 1 {
		t.Fatalf("QueryHits called %d times", g.n.Load())
	}
	if len(cap.last.GraphHits.Items) != 1 || cap.last.GraphHits.Items[0].ID != "from-graph.go" {
		t.Fatalf("pack hits=%v", cap.last.GraphHits.Items)
	}
	if len(cap.last.RepoHits.Paths) != 1 || cap.last.RepoHits.Paths[0] != "from-graph.go" {
		t.Fatalf("seeded repo_hits=%v", cap.last.RepoHits)
	}
}

type countingMemory struct {
	n atomic.Int32
}

func (c *countingMemory) Query(context.Context, string, int) ([]memory.Hit, error) {
	c.n.Add(1)
	return []memory.Hit{{
		Episode: memory.Episode{
			ID:       "from-memory",
			Kind:     memory.KindDecision,
			Validity: memory.ValidityActive,
			Title:    "Prefer lexical memory",
			Body:     "active episode",
		},
		Score: 3,
	}}, nil
}

func TestHeuristicShellDoesNotQueryMemory(t *testing.T) {
	m := &countingMemory{}
	e := intent.New(llm.HeuristicCompleter{})
	e.Memory = m
	res, err := e.Compile(context.Background(), intent.Request{Text: "echo hi"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicShell {
		t.Fatalf("method=%s", res.Method)
	}
	if m.n.Load() != 0 {
		t.Fatalf("Memory.Query called %d times", m.n.Load())
	}
}

func TestLLMPathQueriesMemory(t *testing.T) {
	m := &countingMemory{}
	cap := &captureCompleter{}
	e := intent.New(cap)
	e.Memory = m
	res, err := e.Compile(context.Background(), intent.Request{Text: "prepare a friendly greeting for the team"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodLLM {
		t.Fatalf("method=%s", res.Method)
	}
	if m.n.Load() != 1 {
		t.Fatalf("Memory.Query called %d times", m.n.Load())
	}
	if len(cap.last.MemoryHits.Items) != 1 || cap.last.MemoryHits.Items[0].ID != "from-memory" {
		t.Fatalf("pack memory hits=%v", cap.last.MemoryHits.Items)
	}
}
