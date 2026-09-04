package contextpack

import (
	"encoding/json"

	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

const (
	DefaultMaxChars         = 12000
	DefaultMaxFiles         = 40
	DefaultGraphMaxHits     = 20
	DefaultGraphMaxChars    = 4000
	DefaultMemoryMaxHits    = 8
	DefaultMemoryMaxChars   = 2000
	DefaultPlaybookMaxHits  = 2
	DefaultPlaybookMaxChars = 1500
)

type Pack struct {
	Task         TaskView           `json:"task"`
	Step         StepView           `json:"step"`
	PriorOutputs []store.StepOutput `json:"prior_outputs"`
	RepoHits     RepoHits           `json:"repo_hits"`
	GraphHits    GraphHits          `json:"graph_hits"`
	MemoryHits   MemoryHits         `json:"memory_hits"`
	PlaybookHits PlaybookHits       `json:"playbook_hits"`
	Budget       Budget             `json:"budget"`
}

type TaskView struct {
	TaskID  string `json:"task_id"`
	Summary string `json:"summary"`
	Notes   string `json:"notes,omitempty"`
}

type StepView struct {
	StepID     string `json:"step_id"`
	Capability string `json:"capability"`
}

type RepoHits struct {
	Paths   []string `json:"paths,omitempty"`
	Symbols []string `json:"symbols,omitempty"`
}

type GraphHit struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Reason string `json:"reason"`
	Score  int    `json:"score"`
}

type GraphHits struct {
	Items []GraphHit `json:"items"`
}

type MemoryHit struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Validity string `json:"validity"`
	Title    string `json:"title"`
	Snippet  string `json:"snippet"`
	Score    int    `json:"score"`
}

type MemoryHits struct {
	Items []MemoryHit `json:"items"`
}

type PlaybookHit struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

type PlaybookHits struct {
	Items []PlaybookHit `json:"items"`
}

type Budget struct {
	MaxChars         int `json:"max_chars"`
	MaxFiles         int `json:"max_files"`
	GraphMaxHits     int `json:"graph_max_hits"`
	GraphMaxChars    int `json:"graph_max_chars"`
	MemoryMaxHits    int `json:"memory_max_hits"`
	MemoryMaxChars   int `json:"memory_max_chars"`
	PlaybookMaxHits  int `json:"playbook_max_hits"`
	PlaybookMaxChars int `json:"playbook_max_chars"`
}

func Assemble(t task.Task, stepID, capability string, priors []store.StepOutput) Pack {
	p := Pack{
		Task:         TaskView{TaskID: t.TaskID, Summary: t.Intent.Summary, Notes: t.Intent.Notes},
		Step:         StepView{StepID: stepID, Capability: capability},
		PriorOutputs: priors,
		GraphHits:    GraphHits{Items: []GraphHit{}},
		MemoryHits:   MemoryHits{Items: []MemoryHit{}},
		PlaybookHits: PlaybookHits{Items: []PlaybookHit{}},
		Budget: Budget{
			MaxChars:         DefaultMaxChars,
			MaxFiles:         DefaultMaxFiles,
			GraphMaxHits:     DefaultGraphMaxHits,
			GraphMaxChars:    DefaultGraphMaxChars,
			MemoryMaxHits:    DefaultMemoryMaxHits,
			MemoryMaxChars:   DefaultMemoryMaxChars,
			PlaybookMaxHits:  DefaultPlaybookMaxHits,
			PlaybookMaxChars: DefaultPlaybookMaxChars,
		},
	}
	p.RepoHits = extractRepoHits(priors, p.Budget)
	p.PriorOutputs = truncatePriors(priors, p.Budget.MaxChars)
	return p
}

// WithSeededRepoHits copies Graph path/symbol hits into repo_hits when
// intra-run extraction left that field empty (G-138). Existing repo-search
// hits are not overwritten.
func WithSeededRepoHits(p Pack, items []GraphHit) Pack {
	if len(p.RepoHits.Paths) > 0 || len(p.RepoHits.Symbols) > 0 {
		return p
	}
	max := p.Budget.MaxFiles
	if max <= 0 {
		max = DefaultMaxFiles
	}
	var paths, symbols []string
	seenP := map[string]bool{}
	seenS := map[string]bool{}
	for _, it := range items {
		if it.ID == "" {
			continue
		}
		switch it.Kind {
		case "path":
			if seenP[it.ID] || len(paths) >= max {
				continue
			}
			seenP[it.ID] = true
			paths = append(paths, it.ID)
		case "symbol":
			if seenS[it.ID] || len(symbols) >= max {
				continue
			}
			seenS[it.ID] = true
			symbols = append(symbols, it.ID)
		}
	}
	p.RepoHits = RepoHits{Paths: paths, Symbols: symbols}
	return p
}

// WithGraphHits attaches ranked structural hits and applies graph budgets.
// Items should already be score-sorted (highest first); lowest scores are dropped first.
func WithGraphHits(p Pack, items []GraphHit) Pack {
	if p.Budget.GraphMaxHits <= 0 {
		p.Budget.GraphMaxHits = DefaultGraphMaxHits
	}
	if p.Budget.GraphMaxChars <= 0 {
		p.Budget.GraphMaxChars = DefaultGraphMaxChars
	}
	if items == nil {
		items = []GraphHit{}
	}
	if len(items) > p.Budget.GraphMaxHits {
		items = items[:p.Budget.GraphMaxHits]
	}
	items = trimGraphHitsByChars(items, p.Budget.GraphMaxChars)
	p.GraphHits = GraphHits{Items: items}
	return p
}

// WithMemoryHits attaches ranked episodic hits and applies memory budgets.
// Items should already be score-sorted (highest first); lowest scores are dropped first.
func WithMemoryHits(p Pack, items []MemoryHit) Pack {
	if p.Budget.MemoryMaxHits <= 0 {
		p.Budget.MemoryMaxHits = DefaultMemoryMaxHits
	}
	if p.Budget.MemoryMaxChars <= 0 {
		p.Budget.MemoryMaxChars = DefaultMemoryMaxChars
	}
	if items == nil {
		items = []MemoryHit{}
	}
	if len(items) > p.Budget.MemoryMaxHits {
		items = items[:p.Budget.MemoryMaxHits]
	}
	items = trimMemoryHitsByChars(items, p.Budget.MemoryMaxChars)
	p.MemoryHits = MemoryHits{Items: items}
	return p
}

func extractRepoHits(priors []store.StepOutput, b Budget) RepoHits {
	var hits RepoHits
	for _, o := range priors {
		if o.Capability != "pipeline.repo-search" {
			continue
		}
		var parsed struct {
			Paths   []string `json:"paths"`
			Symbols []string `json:"symbols"`
		}
		if err := json.Unmarshal(o.Output, &parsed); err != nil {
			continue
		}
		hits.Paths = capStrings(parsed.Paths, b.MaxFiles)
		hits.Symbols = capStrings(parsed.Symbols, b.MaxFiles)
	}
	return hits
}

func truncatePriors(priors []store.StepOutput, maxChars int) []store.StepOutput {
	if maxChars <= 0 {
		return priors
	}
	used := 0
	out := make([]store.StepOutput, 0, len(priors))
	for _, o := range priors {
		n := len(o.Output)
		if used+n > maxChars {
			remain := maxChars - used
			if remain < 0 {
				remain = 0
			}
			if remain == 0 {
				break
			}
			cut := append(json.RawMessage(nil), o.Output[:remain]...)
			o.Output = cut
			out = append(out, o)
			break
		}
		out = append(out, o)
		used += n
	}
	return out
}

func trimGraphHitsByChars(items []GraphHit, maxChars int) []GraphHit {
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
	return []GraphHit{}
}

func trimMemoryHitsByChars(items []MemoryHit, maxChars int) []MemoryHit {
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
	return []MemoryHit{}
}

func capStrings(in []string, n int) []string {
	if n <= 0 || len(in) <= n {
		return in
	}
	return in[:n]
}

func WithPlaybookHits(p Pack, items []PlaybookHit) Pack {
	if p.Budget.PlaybookMaxHits <= 0 {
		p.Budget.PlaybookMaxHits = DefaultPlaybookMaxHits
	}
	if p.Budget.PlaybookMaxChars <= 0 {
		p.Budget.PlaybookMaxChars = DefaultPlaybookMaxChars
	}
	if items == nil {
		items = []PlaybookHit{}
	}
	if len(items) > p.Budget.PlaybookMaxHits {
		items = items[:p.Budget.PlaybookMaxHits]
	}
	used := 0
	out := items[:0]
	for _, it := range items {
		n := len(it.Snippet)
		if used+n > p.Budget.PlaybookMaxChars {
			break
		}
		out = append(out, it)
		used += n
	}
	p.PlaybookHits = PlaybookHits{Items: out}
	return p
}

func Marshal(p Pack) (json.RawMessage, error) {
	return json.Marshal(p)
}
