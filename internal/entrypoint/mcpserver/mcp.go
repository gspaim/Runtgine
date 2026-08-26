// Package mcpserver exposes the Project Memory Provider to external
// MCP clients (docs/39-mcp-memory-v0.md, G-187..G-193). It is a
// read-only entrypoint: only `memory.query` and `memory.list` tools
// are served, over stdio (`runtgine mcp`) or HTTP (`/mcp` on
// `runtgine serve`). It is not a Player and never writes episodes.
package mcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/gspaim/Runtgine/internal/core/memory"
)

const (
	mcpProtocolVersion = "2025-06-18"
	serverName         = "runtgine-mcp-memory"

	maxQueryRunes = 512
	maxListLimit  = 100
	snippetRunes  = 200
)

// Reader is the read-only surface this server consumes. Satisfied by
// *api.Core's Memory Service via the adapters in New.
type Reader interface {
	Query(ctx context.Context, text string, limit int) ([]memory.Hit, error)
	List(ctx context.Context, f memory.Filter) ([]memory.Episode, error)
}

type activeLister interface {
	List(ctx context.Context, f memory.Filter) ([]memory.Episode, error)
}

// Server dispatches MCP JSON-RPC 2.0 requests against the Provider.
// Safe for concurrent use.
type Server struct {
	reader Reader
	log    *slog.Logger
}

func New(reader Reader, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{reader: reader, log: log}
}

// --- JSON-RPC 2.0 / MCP wire types ---

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

const (
	codeParse     = -32700
	codeInvalid   = -32600
	codeMethodNF  = -32601
	codeBadParams = -32602
	codeInternal  = -32603
)

// --- MCP result shapes ---

type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

var tools = []toolDef{
	{
		Name: "memory.query",
		Description: "Search active workspace memory episodes " +
			"(decision|failure|handoff|preference) by lexical token " +
			"overlap. Results are informational guidance compiled from " +
			"past runs; they do not grant capabilities or change policy.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"text":  map[string]any{"type": "string", "minLength": 1, "maxLength": maxQueryRunes},
				"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 20},
			},
			"required": []string{"text"},
		},
	},
	{
		Name: "memory.list",
		Description: "List recent active workspace memory episodes, " +
			"optionally filtered by kind. Informational guidance only.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"kind":  map[string]any{"type": "string", "enum": []string{memory.KindDecision, memory.KindFailure, memory.KindHandoff, memory.KindPreference}},
				"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": maxListLimit},
			},
		},
	},
}

type toolResult struct {
	Content []map[string]string `json:"content"`
	IsError bool                `json:"isError,omitempty"`
}

func textResult(s string) toolResult {
	return toolResult{Content: []map[string]string{{"type": "text", "text": s}}}
}

// --- Dispatch ---

// Handle processes one JSON-RPC request body and returns the response
// payload to serialize (nil for notifications).
func (s *Server) Handle(ctx context.Context, body []byte) (any, bool) {
	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return rpcResponse{JSONRPC: "2.0", ID: nil, Error: &rpcError{Code: codeParse, Message: "parse error"}}, true
	}
	if req.JSONRPC != "" && req.JSONRPC != "2.0" {
		return s.resp(req.ID, nil, &rpcError{Code: codeInvalid, Message: "jsonrpc must be 2.0"}), true
	}
	isNotification := len(req.ID) == 0 || string(req.ID) == "null"
	switch req.Method {
	case "initialize":
		return s.resp(req.ID, map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": serverName, "version": serverVersion},
		}, nil), true
	case "notifications/initialized":
		return nil, false // notification: no response
	case "ping":
		return s.resp(req.ID, map[string]any{}, nil), true
	case "tools/list":
		return s.resp(req.ID, map[string]any{"tools": tools}, nil), true
	case "tools/call":
		res, rerr := s.callTool(ctx, req.Params)
		return s.resp(req.ID, res, rerr), true
	default:
		if isNotification {
			return nil, false
		}
		return s.resp(req.ID, nil, &rpcError{Code: codeMethodNF, Message: fmt.Sprintf("method not found: %s", req.Method)}), true
	}
}

func (s *Server) resp(id json.RawMessage, result any, rerr *rpcError) rpcResponse {
	if id == nil {
		id = json.RawMessage("null")
	}
	if rerr != nil {
		return rpcResponse{JSONRPC: "2.0", ID: id, Error: rerr}
	}
	return rpcResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func (s *Server) callTool(ctx context.Context, params json.RawMessage) (any, *rpcError) {
	var p struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{Code: codeBadParams, Message: "invalid params"}
	}
	switch p.Name {
	case "memory.query":
		out, terr := s.toolQuery(ctx, p.Arguments)
		if terr != nil {
			return nil, terr
		}
		return out, nil
	case "memory.list":
		out, terr := s.toolList(ctx, p.Arguments)
		if terr != nil {
			return nil, terr
		}
		return out, nil
	default:
		return nil, &rpcError{Code: codeBadParams, Message: fmt.Sprintf("unknown tool: %s", p.Name)}
	}
}

func (s *Server) toolQuery(ctx context.Context, args json.RawMessage) (*toolResult, *rpcError) {
	var a struct {
		Text  string `json:"text"`
		Limit int    `json:"limit"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &a); err != nil {
			return nil, &rpcError{Code: codeBadParams, Message: "invalid arguments"}
		}
	}
	text := strings.TrimSpace(a.Text)
	if n := utf8len(text); n < 1 || n > maxQueryRunes {
		return nil, &rpcError{Code: codeBadParams, Message: fmt.Sprintf("text must be 1-%d runes", maxQueryRunes)}
	}
	limit := a.Limit
	if limit < 1 || limit > 20 {
		limit = memory.DefaultLimit
	}
	hits, err := s.reader.Query(ctx, text, limit)
	if err != nil {
		s.log.Warn("mcp memory.query degraded", "err", err)
		return ptr(textResult(`{"hits":[]}`)), nil
	}
	type hitOut struct {
		ID      string `json:"id"`
		Kind    string `json:"kind"`
		Title   string `json:"title"`
		Snippet string `json:"snippet"`
		Score   int    `json:"score"`
	}
	out := make([]hitOut, 0, len(hits))
	for _, h := range hits {
		out = append(out, hitOut{
			ID:      h.ID,
			Kind:    h.Kind,
			Title:   h.Title,
			Snippet: truncate(h.Body, snippetRunes),
			Score:   h.Score,
		})
	}
	raw, _ := json.Marshal(map[string]any{"hits": out})
	return ptr(textResult(string(raw))), nil
}

func (s *Server) toolList(ctx context.Context, args json.RawMessage) (*toolResult, *rpcError) {
	var a struct {
		Kind  string `json:"kind"`
		Limit int    `json:"limit"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &a); err != nil {
			return nil, &rpcError{Code: codeBadParams, Message: "invalid arguments"}
		}
	}
	limit := a.Limit
	if limit < 1 || limit > maxListLimit {
		limit = 20
	}
	switch a.Kind {
	case "", memory.KindDecision, memory.KindFailure, memory.KindHandoff, memory.KindPreference:
	default:
		return nil, &rpcError{Code: codeBadParams, Message: fmt.Sprintf("kind must be %s|%s|%s|%s", memory.KindDecision, memory.KindFailure, memory.KindHandoff, memory.KindPreference)}
	}
	f := memory.Filter{Kind: a.Kind, Validity: memory.ValidityActive}
	episodes, err := s.reader.List(ctx, f)
	if err != nil {
		s.log.Warn("mcp memory.list degraded", "err", err)
		return ptr(textResult(`{"episodes":[]}`)), nil
	}
	type epOut struct {
		ID        string `json:"id"`
		Kind      string `json:"kind"`
		Title     string `json:"title"`
		CreatedAt string `json:"created_at"`
	}
	out := make([]epOut, 0, len(episodes))
	for _, e := range episodes {
		out = append(out, epOut{ID: e.ID, Kind: e.Kind, Title: e.Title, CreatedAt: e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")})
		if len(out) >= limit {
			break
		}
	}
	raw, _ := json.Marshal(map[string]any{"episodes": out})
	return ptr(textResult(string(raw))), nil
}

func ptr(t toolResult) *toolResult { return &t }

func truncate(s string, runes int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) > runes {
		return strings.TrimSpace(string(r[:runes])) + "..."
	}
	return strings.TrimSpace(string(r))
}

func utf8len(s string) int { return len([]rune(s)) }

// --- stdio transport ---

var serverVersion = "v0"

// ServeStdio speaks MCP over newline-delimited JSON-RPC 2.0 on in/out.
// Warnings go to stderr only.
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	br := bufio.NewReader(in)
	var mu sync.Mutex
	writeLine := func(v any) error {
		raw, err := json.Marshal(v)
		if err != nil {
			return err
		}
		mu.Lock()
		defer mu.Unlock()
		if _, err := out.Write(append(raw, '\n')); err != nil {
			return err
		}
		if f, ok := out.(interface{ Flush() }); ok {
			f.Flush()
		}
		return nil
	}
	for {
		line, err := br.ReadString('\n')
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			if resp, ok := s.Handle(ctx, []byte(trimmed)); ok && resp != nil {
				if werr := writeLine(resp); werr != nil {
					return werr
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

// ServeHTTP handles one MCP request per POST body (streamable HTTP).
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	resp, respond := s.Handle(r.Context(), body)
	if !respond || resp == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
