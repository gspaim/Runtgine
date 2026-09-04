package blast

import (
	"sort"

	"github.com/gspaim/Runtgine/internal/core/claim"
	"github.com/gspaim/Runtgine/internal/core/graph"
)

const (
	ReasonSeed     = "seed"
	ReasonMentions = "mentions"
)

type Affected struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Reason string `json:"reason"`
	Via    string `json:"via,omitempty"`
}

// ApplyWalk sets Affected from a graph snapshot. A snapshot error yields an
// empty list without changing the rest of the report (G-112).
func ApplyWalk(rep *Report, snap graph.Snapshot, snapErr error) {
	if rep == nil {
		return
	}
	if snapErr != nil {
		rep.Affected = []Affected{}
		return
	}
	rep.Affected = Walk(snap, rep.Touches)
}

// Walk returns 1-hop inbound mentions from unique path touches (G-113).
func Walk(snap graph.Snapshot, touches []Touch) []Affected {
	out := []Affected{}
	pathNodes := map[string]bool{}
	for _, n := range snap.Nodes {
		if n.Kind == graph.KindPath {
			pathNodes[n.ID] = true
		}
	}
	seenSeed := map[string]bool{}
	seeds := make([]string, 0)
	for _, th := range touches {
		if th.Kind != string(claim.KindPath) || th.Key == "" || seenSeed[th.Key] {
			continue
		}
		seenSeed[th.Key] = true
		seeds = append(seeds, th.Key)
	}

	seen := map[string]bool{}
	add := func(a Affected) {
		key := a.Kind + "\x00" + a.ID + "\x00" + a.Reason + "\x00" + a.Via
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, a)
	}

	for _, id := range seeds {
		if !pathNodes[id] {
			continue
		}
		add(Affected{Kind: graph.KindPath, ID: id, Reason: ReasonSeed})
		for _, e := range snap.Edges {
			if e.Kind != graph.EdgeMentions {
				continue
			}
			if e.To.Kind != graph.KindPath || e.To.ID != id {
				continue
			}
			if e.From.Kind != graph.KindRun && e.From.Kind != graph.KindTask {
				continue
			}
			add(Affected{
				Kind:   e.From.Kind,
				ID:     e.From.ID,
				Reason: ReasonMentions,
				Via:    "path:" + id,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		ri, rj := affectedKindRank(out[i].Kind), affectedKindRank(out[j].Kind)
		if ri != rj {
			return ri < rj
		}
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		if out[i].Reason != out[j].Reason {
			return out[i].Reason < out[j].Reason
		}
		return out[i].Via < out[j].Via
	})
	return out
}

func affectedKindRank(kind string) int {
	switch kind {
	case graph.KindPath:
		return 0
	case graph.KindTask:
		return 1
	case graph.KindRun:
		return 2
	default:
		return 9
	}
}
