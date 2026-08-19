package contextpack

import (
	"encoding/json"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/task"
)

func TestAssembleIncludesEmptyGraphHits(t *testing.T) {
	tk := task.Task{TaskID: "t1", Intent: task.Intent{Summary: "s"}}
	p := Assemble(tk, "s1", "pipeline.tech-review", nil)
	if p.Budget.GraphMaxHits != DefaultGraphMaxHits || p.Budget.GraphMaxChars != DefaultGraphMaxChars {
		t.Fatalf("budget=%+v", p.Budget)
	}
	if p.Budget.MemoryMaxHits != DefaultMemoryMaxHits || p.Budget.MemoryMaxChars != DefaultMemoryMaxChars {
		t.Fatalf("memory budget=%+v", p.Budget)
	}
	if p.GraphHits.Items == nil || len(p.GraphHits.Items) != 0 {
		t.Fatalf("items=%v", p.GraphHits.Items)
	}
	if p.MemoryHits.Items == nil || len(p.MemoryHits.Items) != 0 {
		t.Fatalf("memory items=%v", p.MemoryHits.Items)
	}
	raw, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["graph_hits"]; !ok {
		t.Fatal("missing graph_hits")
	}
	if _, ok := m["memory_hits"]; !ok {
		t.Fatal("missing memory_hits")
	}
	budget, ok := m["budget"].(map[string]any)
	if !ok {
		t.Fatal("missing budget")
	}
	if _, ok := budget["memory_max_hits"]; !ok {
		t.Fatal("missing memory_max_hits")
	}
}

func TestWithGraphHitsCaps(t *testing.T) {
	p := Assemble(task.Task{TaskID: "t"}, "s", "c", nil)
	p.Budget.GraphMaxHits = 2
	items := []GraphHit{
		{Kind: "path", ID: "a", Reason: "seed", Score: 10},
		{Kind: "path", ID: "b", Reason: "seed", Score: 9},
		{Kind: "path", ID: "c", Reason: "seed", Score: 8},
	}
	p = WithGraphHits(p, items)
	if len(p.GraphHits.Items) != 2 {
		t.Fatalf("got %d", len(p.GraphHits.Items))
	}
	if p.GraphHits.Items[0].ID != "a" || p.GraphHits.Items[1].ID != "b" {
		t.Fatalf("%v", p.GraphHits.Items)
	}

	p.Budget.GraphMaxHits = 10
	p.Budget.GraphMaxChars = 50
	p = WithGraphHits(p, items)
	b, _ := json.Marshal(p.GraphHits.Items)
	if len(b) > 50 {
		t.Fatalf("chars %d %s", len(b), b)
	}
}

func TestWithMemoryHitsCaps(t *testing.T) {
	p := Assemble(task.Task{TaskID: "t"}, "s", "c", nil)
	p.Budget.MemoryMaxHits = 2
	items := []MemoryHit{
		{ID: "a", Kind: "decision", Validity: "active", Title: "A", Score: 10},
		{ID: "b", Kind: "decision", Validity: "active", Title: "B", Score: 9},
		{ID: "c", Kind: "decision", Validity: "active", Title: "C", Score: 8},
	}
	p = WithMemoryHits(p, items)
	if len(p.MemoryHits.Items) != 2 {
		t.Fatalf("got %d", len(p.MemoryHits.Items))
	}
	if p.MemoryHits.Items[0].ID != "a" || p.MemoryHits.Items[1].ID != "b" {
		t.Fatalf("%v", p.MemoryHits.Items)
	}

	p.Budget.MemoryMaxHits = 10
	p.Budget.MemoryMaxChars = 80
	p = WithMemoryHits(p, items)
	b, _ := json.Marshal(p.MemoryHits.Items)
	if len(b) > 80 {
		t.Fatalf("chars %d %s", len(b), b)
	}
}
