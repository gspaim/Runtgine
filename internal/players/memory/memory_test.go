package memory_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	providermem "github.com/gspaim/Runtgine/internal/core/memory"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/store"
	memplayer "github.com/gspaim/Runtgine/internal/players/memory"
)

// stubReader is a fake memory.Reader used to assert that the Player
// degrades on error and forwards results otherwise.
type stubReader struct {
	recallFn func(ctx context.Context, q providermem.RecallQuery) ([]providermem.Hit, error)
	checkFn  func(ctx context.Context, pattern string) (providermem.CheckResult, error)
}

func (s stubReader) Recall(ctx context.Context, q providermem.RecallQuery) ([]providermem.Hit, error) {
	return s.recallFn(ctx, q)
}

func (s stubReader) Check(ctx context.Context, pattern string) (providermem.CheckResult, error) {
	return s.checkFn(ctx, pattern)
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func newService(t *testing.T) *providermem.Service {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "runtgine.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return providermem.New(st, nil)
}

// --- Manifest / Validation ---------------------------------------------------

func TestManifestCapabilities(t *testing.T) {
	m := memplayer.New(nil, silentLogger()).Manifest()
	if m.Name != "memory" {
		t.Fatalf("name=%s", m.Name)
	}
	names := make(map[string]bool)
	for _, c := range m.Capabilities {
		names[c.Name] = true
	}
	if !names["memory.recall"] || !names["memory.check"] {
		t.Fatalf("caps=%v", names)
	}
}

func TestValidateStaticInputRejectsUnknownCapability(t *testing.T) {
	if err := memplayer.ValidateStaticInput("/tmp", "memory.write", mustMarshal(t, map[string]any{})); err == nil {
		t.Fatal("expected unknown capability error")
	}
}

func TestValidateStaticInputRecall(t *testing.T) {
	if err := memplayer.ValidateStaticInput("/tmp", "memory.recall", mustMarshal(t, map[string]any{})); err == nil {
		t.Fatal("missing query")
	}
	if err := memplayer.ValidateStaticInput("/tmp", "memory.recall", mustMarshal(t, map[string]any{"query": "x", "kind": "note"})); err == nil {
		t.Fatal("bad kind")
	}
	if err := memplayer.ValidateStaticInput("/tmp", "memory.recall", mustMarshal(t, map[string]any{"query": "x", "limit": 99})); err == nil {
		t.Fatal("limit out of range")
	}
	if err := memplayer.ValidateStaticInput("/tmp", "memory.recall", mustMarshal(t, map[string]any{"query": strings.Repeat("x", 600)})); err == nil {
		t.Fatal("query too long")
	}
	if err := memplayer.ValidateStaticInput("/tmp", "memory.recall", mustMarshal(t, map[string]any{"query": "ok", "limit": 5})); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
}

func TestValidateStaticInputCheck(t *testing.T) {
	if err := memplayer.ValidateStaticInput("/tmp", "memory.check", mustMarshal(t, map[string]any{})); err == nil {
		t.Fatal("missing pattern")
	}
	if err := memplayer.ValidateStaticInput("/tmp", "memory.check", mustMarshal(t, map[string]any{"pattern": strings.Repeat("x", 300)})); err == nil {
		t.Fatal("pattern too long")
	}
}

// --- Execute against a stub Reader -----------------------------------------

func TestExecuteRecallReturnsHits(t *testing.T) {
	svc := newService(t)
	if _, err := svc.Record(context.Background(), providermem.EpisodeInput{
		Kind: providermem.KindDecision, Title: "sqlite sidecar rejected", Body: "no ai-memory",
	}); err != nil {
		t.Fatal(err)
	}

	p := memplayer.New(svc, silentLogger())
	raw := execRecall(t, p, "sidecar", 5)
	var out struct {
		Hits []struct {
			ID    string `json:"id"`
			Kind  string `json:"kind"`
			Title string `json:"title"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Hits) != 1 || out.Hits[0].Kind != providermem.KindDecision {
		t.Fatalf("hits=%+v", out.Hits)
	}
}

func TestExecuteRecallDegradesOnReaderError(t *testing.T) {
	reader := stubReader{
		recallFn: func(_ context.Context, _ providermem.RecallQuery) ([]providermem.Hit, error) {
			return nil, errors.New("sqlite busy")
		},
		checkFn: func(_ context.Context, _ string) (providermem.CheckResult, error) {
			return providermem.CheckResult{}, nil
		},
	}
	p := memplayer.New(reader, silentLogger())
	raw := execRecall(t, p, "x", 5)
	var out struct {
		Hits []any `json:"hits"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Hits) != 0 {
		t.Fatalf("hits=%v", out.Hits)
	}
}

func TestExecuteCheckWithActiveFailure(t *testing.T) {
	svc := newService(t)
	ep, err := svc.Record(context.Background(), providermem.EpisodeInput{
		Kind: providermem.KindFailure, Title: "metadata IP denied", Body: "link-local blocked",
	})
	if err != nil {
		t.Fatal(err)
	}

	p := memplayer.New(svc, silentLogger())
	raw := execCheck(t, p, "metadata")
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		HasFailure bool `json:"has_failure"`
		Match      *struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"match"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if !out.HasFailure || out.Match == nil || out.Match.ID != ep.ID {
		t.Fatalf("check=%+v", out)
	}
}

func TestExecuteCheckNoMatch(t *testing.T) {
	svc := newService(t)
	_, _ = svc.Record(context.Background(), providermem.EpisodeInput{
		Kind: providermem.KindDecision, Title: "team uses REST",
	})
	p := memplayer.New(svc, silentLogger())
	raw := execCheck(t, p, "kubernetes")
	var out struct {
		HasFailure bool `json:"has_failure"`
		Match      any   `json:"match"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.HasFailure || out.Match != nil {
		t.Fatalf("expected no failure, got %+v", out)
	}
}

func TestExecuteCheckDegradesOnReaderError(t *testing.T) {
	reader := stubReader{
		recallFn: func(_ context.Context, _ providermem.RecallQuery) ([]providermem.Hit, error) {
			return nil, nil
		},
		checkFn: func(_ context.Context, _ string) (providermem.CheckResult, error) {
			return providermem.CheckResult{}, errors.New("store closed")
		},
	}
	p := memplayer.New(reader, silentLogger())
	raw := execCheck(t, p, "x")
	var out struct {
		HasFailure bool `json:"has_failure"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.HasFailure {
		t.Fatalf("HasFailure=%v", out.HasFailure)
	}
}

// --- Read-only guarantee ---------------------------------------------------

func TestManifestHasNoWriteCapability(t *testing.T) {
	m := memplayer.New(nil, silentLogger()).Manifest()
	for _, c := range m.Capabilities {
		if strings.Contains(c.Name, "record") || strings.Contains(c.Name, "supersede") || strings.Contains(c.Name, "archive") {
			t.Fatalf("unexpected write capability %s", c.Name)
		}
	}
}

func execRecall(t *testing.T, p *memplayer.Player, query string, limit int) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"query": query, "limit": limit})
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: memplayer.CapRecall,
		Input:      raw,
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return out
}

func execCheck(t *testing.T, p *memplayer.Player, pattern string) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"pattern": pattern})
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: memplayer.CapCheck,
		Input:      raw,
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return out
}