package graph

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"unicode"
)

const (
	DefaultHitLimit    = 20
	DefaultHitMaxChars = 4000

	scoreSeed       = 10
	scoreCapability = 8
	scoreMentions   = 5
	scoreExecuted   = 4
	scoreKeyword    = 2

	maxExecutedRuns = 5
)

// Query is the input to QueryHits (G-68).
type Query struct {
	Text           string
	SeedPaths      []string
	SeedSymbols    []string
	SeedCapability string
	Limit          int
	MaxChars       int
}

// Hit is one ranked structural fact for ContextPack.graph_hits.
type Hit struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Reason string `json:"reason"`
	Score  int    `json:"score"`
}

// Hits is the ranked result set (never nil Items after QueryHits).
type Hits struct {
	Items []Hit `json:"items"`
}

var hitStopwords = map[string]struct{}{
	"the": {}, "and": {}, "for": {}, "com": {}, "para": {},
	"uma": {}, "que": {}, "run": {}, "task": {},
}

// QueryHits returns deterministic ranked structural hits. Store failures
// degrade to empty Hits and are logged — never fatal to Runner/Intent.
func (s *Service) QueryHits(ctx context.Context, q Query) Hits {
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultHitLimit
	}
	maxChars := q.MaxChars
	if maxChars <= 0 {
		maxChars = DefaultHitMaxChars
	}

	best := map[string]Hit{} // key = kind\x00id
	add := func(kind, id, reason string, score int) {
		if kind == "" || id == "" || score < 0 {
			return
		}
		key := kind + "\x00" + id
		if prev, ok := best[key]; ok && prev.Score >= score {
			return
		}
		best[key] = Hit{Kind: kind, ID: id, Reason: reason, Score: score}
	}

	// 1. Direct seeds
	for _, p := range q.SeedPaths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, err := s.GetNode(ctx, KindPath, p); err != nil {
			continue
		}
		add(KindPath, p, "seed", scoreSeed)
		s.addMentionNeighbors(ctx, KindPath, p, add)
	}
	for _, sym := range q.SeedSymbols {
		sym = strings.TrimSpace(sym)
		if sym == "" {
			continue
		}
		if _, err := s.GetNode(ctx, KindSymbol, sym); err != nil {
			continue
		}
		add(KindSymbol, sym, "seed", scoreSeed)
		s.addMentionNeighbors(ctx, KindSymbol, sym, add)
	}

	// 3. Capability + recent executed runs
	capName := strings.TrimSpace(q.SeedCapability)
	if capName != "" {
		if _, err := s.GetNode(ctx, KindCapability, capName); err == nil {
			add(KindCapability, capName, "seed", scoreCapability)
			runs, err := s.QueryNeighbors(ctx, KindCapability, capName, EdgeExecuted, "in")
			if err != nil {
				s.Log.Warn("graph QueryHits neighbors failed", "err", err, "capability", capName)
			} else {
				sort.Slice(runs, func(i, j int) bool {
					// Prefer richer attrs/summary then id for stability when UpdatedAt missing on Node
					return runs[i].ID > runs[j].ID
				})
				// Re-fetch with updated_at ordering via store list of neighbors — Node lacks UpdatedAt.
				// Cap to maxExecutedRuns after stable sort by id desc (proxy for recency of UUID v7).
				n := maxExecutedRuns
				if len(runs) < n {
					n = len(runs)
				}
				for _, r := range runs[:n] {
					if r.Kind == KindRun {
						add(KindRun, r.ID, "executed", scoreExecuted)
					}
				}
			}
		}
	}

	// 4. Keywords against path/symbol/capability ids
	tokens := tokenizeHitText(q.Text)
	if len(tokens) > 0 {
		nodes, err := s.Store.ListGraphNodes(ctx)
		if err != nil {
			s.Log.Warn("graph QueryHits list nodes failed", "err", err)
		} else {
			for _, n := range nodes {
				switch n.Kind {
				case KindPath, KindSymbol, KindCapability:
				default:
					continue
				}
				lowerID := strings.ToLower(n.ID)
				for _, tok := range tokens {
					if strings.Contains(lowerID, tok) {
						add(n.Kind, n.ID, "keyword", scoreKeyword)
						break
					}
				}
			}
		}
	}

	items := make([]Hit, 0, len(best))
	for _, h := range best {
		items = append(items, h)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		return items[i].ID < items[j].ID
	})
	if len(items) > limit {
		items = items[:limit]
	}
	items = trimHitsByChars(items, maxChars)
	if items == nil {
		items = []Hit{}
	}
	return Hits{Items: items}
}

func (s *Service) addMentionNeighbors(ctx context.Context, kind, id string, add func(kind, id, reason string, score int)) {
	// Runs/tasks that mention this node (incoming mentions edges).
	sources, err := s.QueryNeighbors(ctx, kind, id, EdgeMentions, "in")
	if err != nil {
		s.Log.Warn("graph QueryHits mention sources failed", "err", err)
		return
	}
	for _, src := range sources {
		if src.Kind != KindRun && src.Kind != KindTask {
			continue
		}
		// Other paths/symbols mentioned by the same run/task.
		targets, err := s.QueryNeighbors(ctx, src.Kind, src.ID, EdgeMentions, "out")
		if err != nil {
			s.Log.Warn("graph QueryHits mention targets failed", "err", err)
			continue
		}
		for _, t := range targets {
			if t.Kind != KindPath && t.Kind != KindSymbol {
				continue
			}
			if t.Kind == kind && t.ID == id {
				continue
			}
			add(t.Kind, t.ID, "mentions", scoreMentions)
		}
	}
}

func tokenizeHitText(text string) []string {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return nil
	}
	var tok strings.Builder
	var out []string
	flush := func() {
		w := tok.String()
		tok.Reset()
		if len(w) < 3 {
			return
		}
		if _, stop := hitStopwords[w]; stop {
			return
		}
		out = append(out, w)
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' || r == '/' {
			tok.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}

func trimHitsByChars(items []Hit, maxChars int) []Hit {
	if maxChars <= 0 || len(items) == 0 {
		return items
	}
	for n := len(items); n > 0; n-- {
		slice := items[:n]
		b, err := json.Marshal(slice)
		if err != nil {
			return slice
		}
		if len(b) <= maxChars {
			return slice
		}
	}
	return []Hit{}
}
