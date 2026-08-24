// Package memory hosts the Memory Player (G-180..G-186). It exposes a
// read-only surface over the Project Memory Provider
// (`internal/core/memory`) via two capabilities:
//
//   - memory.recall: lexical search over active episodes.
//   - memory.check: returns whether any active failure episode
//     matches a pattern.
//
// The Player never writes, supersedes, or archives episodes. When the
// underlying Reader returns an error, the step degrades to a
// successful empty result with a slog.Warn (G-184): the Run must not
// fail because memory was unavailable.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	providermem "github.com/gspaim/Runtgine/internal/core/memory"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
)

const (
	CapRecall = "memory.recall"
	CapCheck  = "memory.check"

	MaxQueryRunes     = 512
	MaxPatternRunes   = 256
	MaxRecallLimit    = 20
	DefaultRecallLim  = 5
	MaxSnippetBytes   = 1024
)

// Player is a deterministic, read-only Player over Project Memory.
type Player struct {
	reader providermem.Reader
	log    *slog.Logger
}

func New(r providermem.Reader, log *slog.Logger) *Player {
	if log == nil {
		log = slog.Default()
	}
	if r == nil {
		r = NopReader{}
	}
	return &Player{reader: r, log: log}
}

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "memory",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapRecall,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["query"],
  "properties":{
    "query":{"type":"string","minLength":1,"maxLength":512},
    "kind":{"type":"string","enum":["decision","failure","handoff","preference"]},
    "limit":{"type":"integer","minimum":1,"maximum":20}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["hits"],
  "properties":{
    "hits":{"type":"array","items":{"type":"object"}}
  }
}`),
			},
			{
				Name: CapCheck,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["pattern"],
  "properties":{
    "pattern":{"type":"string","minLength":1,"maxLength":256}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["has_failure"],
  "properties":{
    "has_failure":{"type":"boolean"},
    "match":{"type":"object"}
  }
}`),
			},
		},
	}
}

type recallInput struct {
	Query string `json:"query"`
	Kind  string `json:"kind"`
	Limit int    `json:"limit"`
}

type checkInput struct {
	Pattern string `json:"pattern"`
}

type hitOut struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet"`
	CreatedAt string `json:"created_at"`
}

type recallOutput struct {
	Hits []hitOut `json:"hits"`
}

type matchOut struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type checkOutput struct {
	HasFailure bool     `json:"has_failure"`
	Match      *matchOut `json:"match,omitempty"`
}

// ValidateStaticInput enforces ranges and required fields per
// capability. It does not call the Provider.
func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapRecall:
		var in recallInput
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid "+CapRecall+" input: "+err.Error(), nil)
		}
		q := strings.TrimSpace(in.Query)
		if q == "" {
			return result.Validation(result.CodeInvalidInput, "memory.recall: query must be non-empty", nil)
		}
		if utf8RuneCount(q) > MaxQueryRunes {
			return result.Validation(result.CodeInvalidInput, "memory.recall: query exceeds 512 runes", map[string]any{"runes": utf8RuneCount(q)})
		}
		if in.Kind != "" && !validKind(in.Kind) {
			return result.Validation(result.CodeInvalidInput, "memory.recall: unknown kind", map[string]any{"kind": in.Kind})
		}
		if in.Limit != 0 && (in.Limit < 1 || in.Limit > MaxRecallLimit) {
			return result.Validation(result.CodeInvalidInput, "memory.recall: limit must be between 1 and 20", map[string]any{"limit": in.Limit})
		}
		return nil
	case CapCheck:
		var in checkInput
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid "+CapCheck+" input: "+err.Error(), nil)
		}
		p := strings.TrimSpace(in.Pattern)
		if p == "" {
			return result.Validation(result.CodeInvalidInput, "memory.check: pattern must be non-empty", nil)
		}
		if utf8RuneCount(p) > MaxPatternRunes {
			return result.Validation(result.CodeInvalidInput, "memory.check: pattern exceeds 256 runes", map[string]any{"runes": utf8RuneCount(p)})
		}
		return nil
	default:
		return result.Validation(result.CodeUnknownCapability, "memory player cannot validate "+capability, nil)
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	switch req.Capability {
	case CapRecall:
		return p.executeRecall(ctx, req.Input)
	case CapCheck:
		return p.executeCheck(ctx, req.Input)
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "memory player cannot execute "+req.Capability, nil)
	}
}

func (p *Player) executeRecall(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in recallInput
	_ = json.Unmarshal(raw, &in)
	limit := in.Limit
	if limit == 0 {
		limit = DefaultRecallLim
	}
	hits, err := p.reader.Recall(ctx, providermem.RecallQuery{
		Text:  strings.TrimSpace(in.Query),
		Kind:  in.Kind,
		Limit: limit,
	})
	if err != nil {
		p.log.Warn("memory.recall degraded", "err", err.Error())
		hits = nil
	}
	out := recallOutput{Hits: make([]hitOut, 0, len(hits))}
	for _, h := range hits {
		out.Hits = append(out.Hits, hitOut{
			ID:        h.ID,
			Kind:      h.Kind,
			Title:     h.Title,
			Snippet:   truncateRunes(h.Body, MaxSnippetBytes),
			CreatedAt: h.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return json.Marshal(out)
}

func (p *Player) executeCheck(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in checkInput
	_ = json.Unmarshal(raw, &in)
	res, err := p.reader.Check(ctx, strings.TrimSpace(in.Pattern))
	if err != nil {
		p.log.Warn("memory.check degraded", "err", err.Error())
		res = providermem.CheckResult{}
	}
	out := checkOutput{HasFailure: res.HasFailure}
	if res.Match != nil {
		out.Match = &matchOut{ID: res.Match.ID, Title: res.Match.Title}
	}
	return json.Marshal(out)
}

func validKind(k string) bool {
	switch k {
	case providermem.KindDecision, providermem.KindFailure, providermem.KindHandoff, providermem.KindPreference:
		return true
	}
	return false
}

func utf8RuneCount(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func truncateRunes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	data := []byte(s[:maxBytes])
	for len(data) > 0 && (data[len(data)-1]&0xC0) == 0x80 {
		data = data[:len(data)-1]
	}
	if len(data) == 0 {
		return ""
	}
	return fmt.Sprintf("%s...", string(data))
}