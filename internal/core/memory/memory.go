package memory

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
)

const (
	KindDecision   = "decision"
	KindFailure    = "failure"
	KindHandoff    = "handoff"
	KindPreference = "preference"

	ValidityActive     = "active"
	ValiditySuperseded = "superseded"
	ValidityArchived   = "archived"

	CaptureOff      = "off"
	CaptureFailures = "failures"

	MaxTitleRunes = 200
	MaxBodyBytes  = 4096
	DefaultLimit  = 8
	SnippetRunes  = 240
)

// Service is the Project Memory Provider (G-123). It is not a Player.
type Service struct {
	Store *store.Store
	Log   *slog.Logger
}

// RecallQuery filters a Reader.Recall call. Text is required; Kind and
// Limit are optional (empty/zero fall back to defaults).
type RecallQuery struct {
	Text  string
	Kind  string
	Limit int
}

// CheckResult is the answer to a Reader.Check call: whether any
// `active` `failure` episode matches the pattern, and the first match
// when so.
type CheckResult struct {
	HasFailure bool
	Match      *Episode `json:"match,omitempty"`
}

// Reader is the read-only surface used by the Memory Player
// (`internal/players/memory`; G-180..G-186). It is satisfied by
// `*Service`. Implementations MUST degrade on error (caller treats
// err==nil but empty result as "no information"); the Player wraps
// every call to translate any error into a warning + empty result.
type Reader interface {
	Recall(ctx context.Context, q RecallQuery) ([]Hit, error)
	Check(ctx context.Context, pattern string) (CheckResult, error)
}

func New(st *store.Store, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{Store: st, Log: log}
}

type Episode struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	Validity    string    `json:"validity"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	RunID       string    `json:"run_id,omitempty"`
	TaskID      string    `json:"task_id,omitempty"`
	SuccessorID string    `json:"successor_id,omitempty"`
}

type EpisodeInput struct {
	Kind   string `json:"kind"`
	Title  string `json:"title"`
	Body   string `json:"body,omitempty"`
	RunID  string `json:"run_id,omitempty"`
	TaskID string `json:"task_id,omitempty"`
}

type Filter struct {
	Kind     string `json:"kind,omitempty"`
	Validity string `json:"validity,omitempty"`
}

type Hit struct {
	Episode
	Score int `json:"score"`
}

func (s *Service) Record(ctx context.Context, in EpisodeInput) (Episode, error) {
	if err := validateInput(in); err != nil {
		return Episode{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Episode{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	row := store.MemoryEpisodeRow{
		ID:        id.String(),
		Kind:      in.Kind,
		Validity:  ValidityActive,
		Title:     strings.TrimSpace(in.Title),
		Body:      in.Body,
		CreatedAt: time.Now().UTC(),
		RunID:     strings.TrimSpace(in.RunID),
		TaskID:    strings.TrimSpace(in.TaskID),
	}
	if err := s.Store.InsertMemoryEpisode(ctx, row); err != nil {
		return Episode{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	return fromRow(row), nil
}

func (s *Service) List(ctx context.Context, f Filter) ([]Episode, error) {
	if f.Kind != "" && !validKind(f.Kind) {
		return nil, result.Validation(result.CodeInvalidInput, "unknown memory kind", map[string]any{"kind": f.Kind})
	}
	if f.Validity != "" && !validValidity(f.Validity) {
		return nil, result.Validation(result.CodeInvalidInput, "unknown memory validity", map[string]any{"validity": f.Validity})
	}
	rows, err := s.Store.ListMemoryEpisodes(ctx, f.Kind, f.Validity)
	if err != nil {
		return nil, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	out := make([]Episode, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

// Query returns active episodes ranked by lexical token overlap (G-125).
func (s *Service) Query(ctx context.Context, text string, limit int) ([]Hit, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}
	rows, err := s.Store.ListActiveMemoryEpisodes(ctx)
	if err != nil {
		return []Hit{}, result.Runtime(result.CodeInternal, err.Error(), true, nil)
	}
	tokens := Tokenize(text)
	hits := make([]Hit, 0, len(rows))
	for _, row := range rows {
		hay := strings.ToLower(row.Title + " " + row.Body)
		score := 0
		for _, tok := range tokens {
			if strings.Contains(hay, tok) {
				score++
			}
		}
		if score == 0 {
			continue
		}
		hits = append(hits, Hit{Episode: fromRow(row), Score: score})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		if !hits[i].CreatedAt.Equal(hits[j].CreatedAt) {
			return hits[i].CreatedAt.After(hits[j].CreatedAt)
		}
		return hits[i].ID < hits[j].ID
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	if hits == nil {
		hits = []Hit{}
	}
	return hits, nil
}

// Recall satisfies Reader.Recall. Wraps Query with the optional kind
// filter; identical ranking/score semantics (G-125).
func (s *Service) Recall(ctx context.Context, q RecallQuery) ([]Hit, error) {
	hits, err := s.Query(ctx, q.Text, q.Limit)
	if err != nil {
		return nil, err
	}
	if q.Kind == "" {
		return hits, nil
	}
	out := make([]Hit, 0, len(hits))
	for _, h := range hits {
		if h.Kind == q.Kind {
			out = append(out, h)
		}
	}
	return out, nil
}

// Check satisfies Reader.Check. Returns HasFailure=true when at
// least one `active` `failure` episode's title or body contains
// any token from `pattern`. Match is the most recent matching
// episode (created_at desc, mirroring the Query tie-break).
func (s *Service) Check(ctx context.Context, pattern string) (CheckResult, error) {
	rows, err := s.Store.ListActiveMemoryEpisodes(ctx)
	if err != nil {
		return CheckResult{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	tokens := Tokenize(pattern)
	if len(tokens) == 0 {
		return CheckResult{HasFailure: false}, nil
	}
	var best *Episode
	bestAt := time.Time{}
	for _, row := range rows {
		if row.Kind != KindFailure {
			continue
		}
		hay := strings.ToLower(row.Title + " " + row.Body)
		matched := false
		for _, tok := range tokens {
			if strings.Contains(hay, tok) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		ep := fromRow(row)
		if best == nil || ep.CreatedAt.After(bestAt) {
			best = &ep
			bestAt = ep.CreatedAt
		}
	}
	if best == nil {
		return CheckResult{HasFailure: false}, nil
	}
	return CheckResult{HasFailure: true, Match: best}, nil
}

func (s *Service) Supersede(ctx context.Context, oldID string, in EpisodeInput) (Episode, error) {
	oldID = strings.TrimSpace(oldID)
	if oldID == "" {
		return Episode{}, result.Validation(result.CodeInvalidInput, "supersede requires an episode id", nil)
	}
	if err := validateInput(in); err != nil {
		return Episode{}, err
	}
	old, err := s.Store.GetMemoryEpisode(ctx, oldID)
	if err != nil {
		if store.ErrNotFound(err) {
			return Episode{}, result.Validation(result.CodeInvalidInput, "episode not found", map[string]any{"id": oldID})
		}
		return Episode{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	if old.Validity != ValidityActive {
		return Episode{}, result.Validation(result.CodeInvalidInput, "only active episodes can be superseded", map[string]any{
			"id":       oldID,
			"validity": old.Validity,
		})
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Episode{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	neu := store.MemoryEpisodeRow{
		ID:        id.String(),
		Kind:      in.Kind,
		Validity:  ValidityActive,
		Title:     strings.TrimSpace(in.Title),
		Body:      in.Body,
		CreatedAt: time.Now().UTC(),
		RunID:     strings.TrimSpace(in.RunID),
		TaskID:    strings.TrimSpace(in.TaskID),
	}
	if err := s.Store.SupersedeMemoryEpisode(ctx, oldID, neu); err != nil {
		if store.ErrNotFound(err) {
			return Episode{}, result.Validation(result.CodeInvalidInput, "episode not found", map[string]any{"id": oldID})
		}
		return Episode{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	return fromRow(neu), nil
}

func (s *Service) Archive(ctx context.Context, id string) (Episode, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Episode{}, result.Validation(result.CodeInvalidInput, "archive requires an episode id", nil)
	}
	if _, err := s.Store.GetMemoryEpisode(ctx, id); err != nil {
		if store.ErrNotFound(err) {
			return Episode{}, result.Validation(result.CodeInvalidInput, "episode not found", map[string]any{"id": id})
		}
		return Episode{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	if err := s.Store.ArchiveMemoryEpisode(ctx, id); err != nil {
		if store.ErrNotFound(err) {
			return Episode{}, result.Validation(result.CodeInvalidInput, "episode not found", map[string]any{"id": id})
		}
		return Episode{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	row, err := s.Store.GetMemoryEpisode(ctx, id)
	if err != nil {
		return Episode{}, result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	return fromRow(row), nil
}

func Tokenize(text string) []string {
	text = strings.ToLower(text)
	seen := map[string]struct{}{}
	var out []string
	var b strings.Builder
	flush := func() {
		tok := b.String()
		b.Reset()
		if tok == "" {
			return
		}
		if _, ok := seen[tok]; ok {
			return
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}

func ClampTitle(s string) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > MaxTitleRunes {
		return string(runes[:MaxTitleRunes])
	}
	return s
}

func ClampBody(s string) string {
	if len(s) <= MaxBodyBytes {
		return s
	}
	b := []byte(s[:MaxBodyBytes])
	for len(b) > 0 && !utf8.Valid(b) {
		b = b[:len(b)-1]
	}
	return string(b)
}

func Snippet(body string) string {
	runes := []rune(body)
	if len(runes) <= SnippetRunes {
		return body
	}
	return string(runes[:SnippetRunes]) + "..."
}

func fromRow(row store.MemoryEpisodeRow) Episode {
	return Episode{
		ID:          row.ID,
		Kind:        row.Kind,
		Validity:    row.Validity,
		Title:       row.Title,
		Body:        row.Body,
		CreatedAt:   row.CreatedAt.UTC(),
		RunID:       row.RunID,
		TaskID:      row.TaskID,
		SuccessorID: row.SuccessorID,
	}
}

func validateInput(in EpisodeInput) error {
	if !validKind(in.Kind) {
		return result.Validation(result.CodeInvalidInput, "kind must be decision|failure|handoff|preference", map[string]any{"kind": in.Kind})
	}
	title := strings.TrimSpace(in.Title)
	n := utf8.RuneCountInString(title)
	if n < 1 || n > MaxTitleRunes {
		return result.Validation(result.CodeInvalidInput, "title must be 1–200 runes", map[string]any{"runes": n})
	}
	if len(in.Body) > MaxBodyBytes {
		return result.Validation(result.CodeInvalidInput, "body must be at most 4096 UTF-8 bytes", map[string]any{"bytes": len(in.Body)})
	}
	if in.Body != "" && !utf8.ValidString(in.Body) {
		return result.Validation(result.CodeInvalidInput, "body must be valid UTF-8", nil)
	}
	return nil
}

func validKind(k string) bool {
	switch k {
	case KindDecision, KindFailure, KindHandoff, KindPreference:
		return true
	}
	return false
}

func validValidity(v string) bool {
	switch v {
	case ValidityActive, ValiditySuperseded, ValidityArchived:
		return true
	}
	return false
}
