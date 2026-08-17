package graph

import (
	"context"
	"encoding/json"
	"testing"
)

func TestQueryHitsSeedAndMentions(t *testing.T) {
	g := openGraph(t)
	ctx := context.Background()
	_ = g.UpsertNode(ctx, KindPath, "a.go", nil)
	_ = g.UpsertNode(ctx, KindPath, "b.go", nil)
	_ = g.UpsertNode(ctx, KindRun, "run-1", nil)
	_ = g.UpsertEdge(ctx, EdgeMentions, KindRun, "run-1", KindPath, "a.go", nil)
	_ = g.UpsertEdge(ctx, EdgeMentions, KindRun, "run-1", KindPath, "b.go", nil)

	hits := g.QueryHits(ctx, Query{SeedPaths: []string{"a.go"}, Limit: 20})
	by := map[string]Hit{}
	for _, h := range hits.Items {
		by[h.Kind+"/"+h.ID] = h
	}
	if h, ok := by["path/a.go"]; !ok || h.Reason != "seed" || h.Score != scoreSeed {
		t.Fatalf("seed hit=%v ok=%v", h, ok)
	}
	if h, ok := by["path/b.go"]; !ok || h.Reason != "mentions" || h.Score != scoreMentions {
		t.Fatalf("mentions hit=%v ok=%v", h, ok)
	}
}

func TestQueryHitsCapabilityAndKeyword(t *testing.T) {
	g := openGraph(t)
	ctx := context.Background()
	_ = g.UpsertNode(ctx, KindCapability, "pipeline.tech-review", nil)
	_ = g.UpsertNode(ctx, KindRun, "run-z", nil)
	_ = g.UpsertNode(ctx, KindRun, "run-a", nil)
	_ = g.UpsertEdge(ctx, EdgeExecuted, KindRun, "run-z", KindCapability, "pipeline.tech-review", nil)
	_ = g.UpsertEdge(ctx, EdgeExecuted, KindRun, "run-a", KindCapability, "pipeline.tech-review", nil)
	_ = g.UpsertNode(ctx, KindPath, "internal/review/handler.go", nil)

	hits := g.QueryHits(ctx, Query{
		Text:           "please review handler",
		SeedCapability: "pipeline.tech-review",
		Limit:          20,
	})
	var sawCap, sawRun, sawKeyword bool
	for _, h := range hits.Items {
		switch {
		case h.Kind == KindCapability && h.ID == "pipeline.tech-review":
			sawCap = h.Score == scoreCapability
		case h.Kind == KindRun && h.Reason == "executed":
			sawRun = true
		case h.Kind == KindPath && h.Reason == "keyword" && h.ID == "internal/review/handler.go":
			sawKeyword = h.Score == scoreKeyword
		}
	}
	if !sawCap || !sawRun || !sawKeyword {
		t.Fatalf("cap=%v run=%v keyword=%v items=%v", sawCap, sawRun, sawKeyword, hits.Items)
	}
}

func TestQueryHitsDedupKeepsHigherScore(t *testing.T) {
	g := openGraph(t)
	ctx := context.Background()
	_ = g.UpsertNode(ctx, KindPath, "shared.go", nil)
	hits := g.QueryHits(ctx, Query{
		Text:      "shared.go file",
		SeedPaths: []string{"shared.go"},
	})
	n := 0
	for _, h := range hits.Items {
		if h.Kind == KindPath && h.ID == "shared.go" {
			n++
			if h.Score != scoreSeed {
				t.Fatalf("expected seed score, got %v", h)
			}
		}
	}
	if n != 1 {
		t.Fatalf("dup count=%d items=%v", n, hits.Items)
	}
}

func TestQueryHitsLimitAndMaxChars(t *testing.T) {
	g := openGraph(t)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		id := "file-" + string(rune('a'+i)) + ".go"
		_ = g.UpsertNode(ctx, KindPath, id, nil)
	}
	hits := g.QueryHits(ctx, Query{Text: ".go", Limit: 3})
	if len(hits.Items) != 3 {
		t.Fatalf("limit got %d", len(hits.Items))
	}
	// Force tiny char budget
	hits = g.QueryHits(ctx, Query{Text: ".go", Limit: 10, MaxChars: 40})
	b, _ := json.Marshal(hits.Items)
	if len(b) > 40 {
		t.Fatalf("max chars exceeded: %d %s", len(b), b)
	}
}

func TestQueryHitsMissingSeedsEmpty(t *testing.T) {
	g := openGraph(t)
	hits := g.QueryHits(context.Background(), Query{SeedPaths: []string{"nope.go"}})
	if len(hits.Items) != 0 {
		t.Fatalf("expected empty, got %v", hits.Items)
	}
}
