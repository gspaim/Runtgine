package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/memory"
)

type stubReader struct {
	hits     []memory.Hit
	episodes []memory.Episode
	err      error
}

func (s *stubReader) Query(_ context.Context, _ string, _ int) ([]memory.Hit, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.hits, nil
}

func (s *stubReader) List(_ context.Context, f memory.Filter) ([]memory.Episode, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([]memory.Episode, 0, len(s.episodes))
	for _, e := range s.episodes {
		if f.Kind != "" && e.Kind != f.Kind {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func newTestServer(r Reader) *Server {
	return New(r, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
}

func req(t *testing.T, id int, method string, params any) []byte {
	t.Helper()
	m := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		m["params"] = params
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

type rpcResp struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func dispatch(t *testing.T, s *Server, body []byte) rpcResp {
	t.Helper()
	out, ok := s.Handle(context.Background(), body)
	if !ok || out == nil {
		t.Fatal("expected a response")
	}
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var resp rpcResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("invalid response %s: %v", raw, err)
	}
	return resp
}

func TestInitialize(t *testing.T) {
	s := newTestServer(&stubReader{})
	resp := dispatch(t, s, req(t, 1, "initialize", map[string]any{
		"protocolVersion": mcpProtocolVersion,
		"capabilities":    map[string]any{},
	}))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	var res struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
	}
	if err := json.Unmarshal(resp.Result, &res); err != nil {
		t.Fatal(err)
	}
	if res.ProtocolVersion == "" || !strings.Contains(res.ServerInfo.Name, "runtgine") {
		t.Fatalf("bad initialize result: %+v", res)
	}
}

func TestToolsListOnlyReadOnly(t *testing.T) {
	s := newTestServer(&stubReader{})
	resp := dispatch(t, s, req(t, 2, "tools/list", nil))
	var res struct {
		Tools []toolDef `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Tools) != 2 {
		t.Fatalf("want exactly 2 tools, got %d", len(res.Tools))
	}
	for _, tool := range res.Tools {
		switch tool.Name {
		case "memory.query", "memory.list":
		default:
			t.Fatalf("unexpected tool %q", tool.Name)
		}
	}
}

func TestUnknownToolRejected(t *testing.T) {
	s := newTestServer(&stubReader{})
	resp := dispatch(t, s, req(t, 3, "tools/call", map[string]any{
		"name":      "memory.record",
		"arguments": map[string]any{},
	}))
	if resp.Error == nil {
		t.Fatal("expected error for write tool")
	}
}

func TestQueryWithHits(t *testing.T) {
	r := &stubReader{hits: []memory.Hit{
		{Episode: memory.Episode{ID: "a1", Kind: "failure", Title: "deploy quebrou", Body: "docker build falhou por tag duplicada", CreatedAt: time.Now().UTC()}, Score: 2},
	}}
	s := newTestServer(r)
	resp := dispatch(t, s, req(t, 4, "tools/call", map[string]any{
		"name":      "memory.query",
		"arguments": map[string]any{"text": "deploy"},
	}))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	var res struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Content) != 1 || res.Content[0].Type != "text" {
		t.Fatalf("bad content: %+v", res.Content)
	}
	var payload struct {
		Hits []map[string]any `json:"hits"`
	}
	if err := json.Unmarshal([]byte(res.Content[0].Text), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Hits) != 1 {
		t.Fatalf("want 1 hit, got %d", len(payload.Hits))
	}
	hit := payload.Hits[0]
	for _, key := range []string{"id", "kind", "title", "snippet", "score"} {
		if _, ok := hit[key]; !ok {
			t.Fatalf("hit missing %q: %+v", key, hit)
		}
	}
}

func TestQueryValidationErrors(t *testing.T) {
	s := newTestServer(&stubReader{})
	for _, text := range []string{"", strings.Repeat("x", maxQueryRunes+1)} {
		args := map[string]any{"text": text}
		resp := dispatch(t, s, req(t, 5, "tools/call", map[string]any{"name": "memory.query", "arguments": args}))
		if resp.Error == nil {
			t.Fatalf("expected validation error for text=%q", text)
		}
	}
}

func TestListFilterByKind(t *testing.T) {
	r := &stubReader{episodes: []memory.Episode{
		{ID: "f1", Kind: "failure", Title: "falha A", CreatedAt: time.Now().UTC()},
		{ID: "d1", Kind: "decision", Title: "decisao B", CreatedAt: time.Now().UTC()},
	}}
	s := newTestServer(r)
	resp := dispatch(t, s, req(t, 6, "tools/call", map[string]any{
		"name":      "memory.list",
		"arguments": map[string]any{"kind": "failure"},
	}))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	var res struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &res); err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Episodes []map[string]any `json:"episodes"`
	}
	if err := json.Unmarshal([]byte(res.Content[0].Text), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Episodes) != 1 || payload.Episodes[0]["id"] != "f1" {
		t.Fatalf("want only failure episode, got %+v", payload.Episodes)
	}
}

func TestListInvalidKindRejected(t *testing.T) {
	// The Provider itself validates kind; a bad kind must surface as an
	// RPC error, not an empty list.
	r := &stubReader{episodes: []memory.Episode{{ID: "x", Kind: "decision"}}}
	s := newTestServer(&listKindGuard{inner: r})
	resp := dispatch(t, s, req(t, 7, "tools/call", map[string]any{
		"name":      "memory.list",
		"arguments": map[string]any{"kind": "transcript"},
	}))
	if resp.Error == nil {
		t.Fatal("expected validation error for invalid kind")
	}
}

// listKindGuard mirrors the Provider's List validation for the stub.
type listKindGuard struct{ inner Reader }

func (g *listKindGuard) Query(ctx context.Context, text string, limit int) ([]memory.Hit, error) {
	return g.inner.Query(ctx, text, limit)
}

func (g *listKindGuard) List(ctx context.Context, f memory.Filter) ([]memory.Episode, error) {
	switch f.Kind {
	case "", memory.KindDecision, memory.KindFailure, memory.KindHandoff, memory.KindPreference:
	default:
		return nil, errors.New("validation.invalid_input")
	}
	return g.inner.List(ctx, f)
}

func TestProviderErrorDegrades(t *testing.T) {
	s := newTestServer(&stubReader{err: errors.New("store: busy")})
	for i, call := range []struct {
		name      string
		arguments map[string]any
		wantEmpty string
	}{
		{"memory.query", map[string]any{"text": "deploy"}, `"hits":[]`},
		{"memory.list", map[string]any{}, `"episodes":[]`},
	} {
		resp := dispatch(t, s, req(t, 10+i, "tools/call", map[string]any{"name": call.name, "arguments": call.arguments}))
		if resp.Error != nil {
			t.Fatalf("%s: provider failure must degrade, got RPC error %+v", call.name, resp.Error)
		}
		var res struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.Unmarshal(resp.Result, &res); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(res.Content[0].Text, call.wantEmpty) {
			t.Fatalf("%s: want empty payload containing %s, got %s", call.name, call.wantEmpty, res.Content[0].Text)
		}
	}
}

func TestUnknownMethod(t *testing.T) {
	s := newTestServer(&stubReader{})
	resp := dispatch(t, s, req(t, 11, "resources/list", nil))
	if resp.Error == nil || resp.Error.Code != codeMethodNF {
		t.Fatalf("want method-not-found, got %+v", resp.Error)
	}
}

func TestServeStdioRoundtrip(t *testing.T) {
	s := newTestServer(&stubReader{})
	var out bytes.Buffer
	in := strings.Join([]string{
		string(req(t, 1, "initialize", map[string]any{})),
		string(req(t, 2, "tools/list", nil)),
	}, "\n") + "\n"
	if err := s.ServeStdio(context.Background(), strings.NewReader(in), &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 responses, got %d: %s", len(lines), out.String())
	}
	if !strings.Contains(lines[1], `"memory.query"`) || !strings.Contains(lines[1], `"memory.list"`) {
		t.Fatalf("tools/list missing read-only tools: %s", lines[1])
	}
}

func TestServeHTTPTypes(t *testing.T) {
	s := newTestServer(&stubReader{})
	// GET rejected.
	get := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	getRec := httptest.NewRecorder()
	s.ServeHTTP(getRec, get)
	if getRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET /mcp must be 405, got %d", getRec.Code)
	}
	// POST initialize works.
	post := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(req(t, 1, "initialize", map[string]any{})))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, post)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /mcp must be 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("bad content-type: %s", ct)
	}
}

func TestNotificationNoResponse(t *testing.T) {
	s := newTestServer(&stubReader{})
	body := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	if _, ok := s.Handle(context.Background(), []byte(body)); ok {
		t.Fatal("notification must not produce a response")
	}
}
