package memory_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gspaim/Runtgine/internal/core/memory"
	"github.com/gspaim/Runtgine/internal/core/store"
)

func openMem(t *testing.T) *memory.Service {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "runtgine.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return memory.New(st, nil)
}

func TestRecordListQuery(t *testing.T) {
	ctx := context.Background()
	s := openMem(t)
	a, err := s.Record(ctx, memory.EpisodeInput{
		Kind:  memory.KindDecision,
		Title: "Use SQLite for project memory",
		Body:  "Same DB as Core. No sidecar.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.ID == "" || a.Validity != memory.ValidityActive || a.Kind != memory.KindDecision {
		t.Fatalf("%+v", a)
	}
	b, err := s.Record(ctx, memory.EpisodeInput{
		Kind:  memory.KindFailure,
		Title: "HTTP GET to metadata IP denied",
		Body:  "link-local blocked",
	})
	if err != nil {
		t.Fatal(err)
	}

	listed, err := s.List(ctx, memory.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 {
		t.Fatalf("list=%d", len(listed))
	}

	hits, err := s.Query(ctx, "sqlite memory sidecar", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].ID != a.ID {
		t.Fatalf("hits=%+v", hits)
	}
	if hits[0].Score < 2 {
		t.Fatalf("score=%d", hits[0].Score)
	}

	hidden, err := s.Query(ctx, "metadata denied", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(hidden) != 1 || hidden[0].ID != b.ID {
		t.Fatalf("failure hits=%+v", hidden)
	}
}

func TestQueryHidesSupersededAndArchived(t *testing.T) {
	ctx := context.Background()
	s := openMem(t)
	old, err := s.Record(ctx, memory.EpisodeInput{
		Kind:  memory.KindDecision,
		Title: "Prefer REST over GraphQL",
		Body:  "team default",
	})
	if err != nil {
		t.Fatal(err)
	}
	neu, err := s.Supersede(ctx, old.ID, memory.EpisodeInput{
		Kind:  memory.KindDecision,
		Title: "Prefer GraphQL for the board API",
		Body:  "replaces REST default",
	})
	if err != nil {
		t.Fatal(err)
	}
	if neu.Validity != memory.ValidityActive {
		t.Fatalf("new validity=%s", neu.Validity)
	}

	listed, err := s.List(ctx, memory.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	var foundOld bool
	for _, e := range listed {
		if e.ID == old.ID {
			foundOld = true
			if e.Validity != memory.ValiditySuperseded || e.SuccessorID != neu.ID {
				t.Fatalf("old=%+v", e)
			}
		}
	}
	if !foundOld {
		t.Fatal("old episode missing from list")
	}

	hits, err := s.Query(ctx, "prefer GraphQL REST board", 8)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range hits {
		if h.ID == old.ID {
			t.Fatalf("superseded still queried: %+v", h)
		}
		if h.Validity != memory.ValidityActive {
			t.Fatalf("non-active hit %+v", h)
		}
	}
	if len(hits) == 0 || hits[0].ID != neu.ID {
		t.Fatalf("expected successor, got %+v", hits)
	}

	archived, err := s.Archive(ctx, neu.ID)
	if err != nil {
		t.Fatal(err)
	}
	if archived.Validity != memory.ValidityArchived {
		t.Fatalf("archived=%+v", archived)
	}
	hits, err = s.Query(ctx, "GraphQL board", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("archived still queried: %+v", hits)
	}
}

func TestSupersedeTransactional(t *testing.T) {
	ctx := context.Background()
	s := openMem(t)
	_, err := s.Supersede(ctx, "00000000-0000-0000-0000-000000000000", memory.EpisodeInput{
		Kind:  memory.KindDecision,
		Title: "ghost successor",
	})
	if err == nil {
		t.Fatal("expected missing predecessor")
	}
	listed, err := s.List(ctx, memory.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 0 {
		t.Fatalf("rolled back insert leaked: %+v", listed)
	}
}

func TestQueryTieBreakCreatedAtThenID(t *testing.T) {
	ctx := context.Background()
	s := openMem(t)
	first, err := s.Record(ctx, memory.EpisodeInput{Kind: memory.KindHandoff, Title: "alpha token", Body: "token"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Record(ctx, memory.EpisodeInput{Kind: memory.KindHandoff, Title: "beta token", Body: "token"})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := s.Query(ctx, "token", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits=%d", len(hits))
	}
	if hits[0].Score != hits[1].Score {
		t.Fatalf("expected equal scores %+v", hits)
	}
	if !hits[0].CreatedAt.After(hits[1].CreatedAt) && hits[0].ID != second.ID {
		t.Fatalf("expected newer first: %s then %s (first=%s second=%s)", hits[0].ID, hits[1].ID, first.ID, second.ID)
	}
	if hits[0].ID != second.ID {
		t.Fatalf("tie-break created_at desc: got %s want %s", hits[0].ID, second.ID)
	}
}

func TestRecordRejectsInvalid(t *testing.T) {
	ctx := context.Background()
	s := openMem(t)
	if _, err := s.Record(ctx, memory.EpisodeInput{Kind: "note", Title: "x"}); err == nil {
		t.Fatal("bad kind")
	}
	if _, err := s.Record(ctx, memory.EpisodeInput{Kind: memory.KindDecision, Title: ""}); err == nil {
		t.Fatal("empty title")
	}
	long := strings.Repeat("á", memory.MaxTitleRunes+1)
	if utf8.RuneCountInString(long) <= memory.MaxTitleRunes {
		t.Fatal("fixture")
	}
	if _, err := s.Record(ctx, memory.EpisodeInput{Kind: memory.KindDecision, Title: long}); err == nil {
		t.Fatal("title too long")
	}
	if _, err := s.Record(ctx, memory.EpisodeInput{
		Kind:  memory.KindDecision,
		Title: "ok",
		Body:  strings.Repeat("x", memory.MaxBodyBytes+1),
	}); err == nil {
		t.Fatal("body too long")
	}
}

func TestTokenizeLowerSplit(t *testing.T) {
	got := memory.Tokenize("Hello, SQLite_v0!")
	want := []string{"hello", "sqlite", "v0"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
