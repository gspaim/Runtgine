package blast

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/graph"
)

func TestWalkEmptyAndWorkspaceIgnored(t *testing.T) {
	touches := []Touch{
		{Kind: "workspace", Key: "."},
		{Kind: "path", Key: "missing.md"},
	}
	got := Walk(graph.Snapshot{}, touches)
	if got == nil || len(got) != 0 {
		t.Fatalf("%v", got)
	}
}

func TestWalkSeedWithoutMentions(t *testing.T) {
	snap := graph.Snapshot{
		Nodes: []graph.Node{{Kind: graph.KindPath, ID: "notes.md"}},
	}
	got := Walk(snap, []Touch{{Kind: "path", Key: "notes.md"}, {Kind: "path", Key: "notes.md"}})
	if len(got) != 1 || got[0].Kind != graph.KindPath || got[0].Reason != ReasonSeed {
		t.Fatalf("%+v", got)
	}
}

func TestWalkMentionsHopAndSort(t *testing.T) {
	snap := graph.Snapshot{
		Nodes: []graph.Node{
			{Kind: graph.KindPath, ID: "notes.md"},
			{Kind: graph.KindPath, ID: "other.md"},
		},
		Edges: []graph.Edge{
			{
				Kind: graph.EdgeMentions,
				From: graph.Ref{Kind: graph.KindRun, ID: "prior-run"},
				To:   graph.Ref{Kind: graph.KindPath, ID: "notes.md"},
			},
			{
				Kind: graph.EdgeMentions,
				From: graph.Ref{Kind: graph.KindTask, ID: "task-1"},
				To:   graph.Ref{Kind: graph.KindPath, ID: "notes.md"},
			},
			{
				Kind: graph.EdgeProvides,
				From: graph.Ref{Kind: graph.KindPlayer, ID: "filesystem"},
				To:   graph.Ref{Kind: graph.KindCapability, ID: "fs.write"},
			},
			{
				Kind: graph.EdgeMentions,
				From: graph.Ref{Kind: graph.KindRun, ID: "other-run"},
				To:   graph.Ref{Kind: graph.KindPath, ID: "other.md"},
			},
		},
	}
	got := Walk(snap, []Touch{{Kind: "path", Key: "notes.md"}})
	if len(got) != 3 {
		t.Fatalf("len=%d %+v", len(got), got)
	}
	if got[0].Kind != graph.KindPath || got[0].ID != "notes.md" || got[0].Reason != ReasonSeed {
		t.Fatalf("seed=%+v", got[0])
	}
	if got[1].Kind != graph.KindTask || got[1].ID != "task-1" || got[1].Via != "path:notes.md" {
		t.Fatalf("task=%+v", got[1])
	}
	if got[2].Kind != graph.KindRun || got[2].ID != "prior-run" || got[2].Reason != ReasonMentions {
		t.Fatalf("run=%+v", got[2])
	}
}

func TestApplyWalkSnapshotErrorKeepsIR(t *testing.T) {
	rep := Report{Risk: RiskPath, Touches: []Touch{{Kind: "path", Key: "a.txt"}}, Affected: []Affected{{Kind: "stale"}}}
	ApplyWalk(&rep, graph.Snapshot{Nodes: []graph.Node{{Kind: graph.KindPath, ID: "a.txt"}}}, errors.New("db closed"))
	if rep.Risk != RiskPath || len(rep.Touches) != 1 {
		t.Fatalf("IR mutated: %+v", rep)
	}
	if rep.Affected == nil || len(rep.Affected) != 0 {
		t.Fatalf("affected=%v", rep.Affected)
	}
}

func TestAffectedJSONAlwaysArray(t *testing.T) {
	rep := Report{
		SchemaVersion:   SchemaVersion,
		Capabilities:    []string{},
		Touches:         []Touch{},
		PredictedClaims: []PredictedClaim{},
		Conflicts:       []Conflict{},
		Images:          []string{},
		Affected:        []Affected{},
		Risk:            RiskNone,
	}
	raw, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}
	arr, ok := obj["affected"].([]any)
	if !ok || arr == nil {
		t.Fatalf("affected missing or null: %s", raw)
	}
	if len(arr) != 0 {
		t.Fatalf("%s", raw)
	}
}
