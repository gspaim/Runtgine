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
	if p.GraphHits.Items == nil || len(p.GraphHits.Items) != 0 {
		t.Fatalf("items=%v", p.GraphHits.Items)
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
