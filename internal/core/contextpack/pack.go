package contextpack

import (
	"encoding/json"

	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

const (
	DefaultMaxChars      = 12000
	DefaultMaxFiles      = 40
	DefaultGraphMaxHits  = 20
	DefaultGraphMaxChars = 4000
)

type Pack struct {
	Task         TaskView           `json:"task"`
	Step         StepView           `json:"step"`
	PriorOutputs []store.StepOutput `json:"prior_outputs"`
	RepoHits     RepoHits           `json:"repo_hits"`
	GraphHits    GraphHits          `json:"graph_hits"`
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

type Budget struct {
	MaxChars      int `json:"max_chars"`
	MaxFiles      int `json:"max_files"`
	GraphMaxHits  int `json:"graph_max_hits"`
	GraphMaxChars int `json:"graph_max_chars"`
}

func Assemble(t task.Task, stepID, capability string, priors []store.StepOutput) Pack {
	p := Pack{
		Task:         TaskView{TaskID: t.TaskID, Summary: t.Intent.Summary, Notes: t.Intent.Notes},
		Step:         StepView{StepID: stepID, Capability: capability},
		PriorOutputs: priors,
		GraphHits:    GraphHits{Items: []GraphHit{}},
		Budget: Budget{
			MaxChars:      DefaultMaxChars,
			MaxFiles:      DefaultMaxFiles,
			GraphMaxHits:  DefaultGraphMaxHits,
			GraphMaxChars: DefaultGraphMaxChars,
		},
	}
	p.RepoHits = extractRepoHits(priors, p.Budget)
	p.PriorOutputs = truncatePriors(priors, p.Budget.MaxChars)
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

func capStrings(in []string, n int) []string {
	if n <= 0 || len(in) <= n {
		return in
	}
	return in[:n]
}

func Marshal(p Pack) (json.RawMessage, error) {
	return json.Marshal(p)
}
