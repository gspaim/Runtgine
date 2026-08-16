package pipeline

import (
	"encoding/json"
	"time"

	"github.com/gspaim/Runtgine/internal/core/task"
)

const (
	CapTechReview  = "pipeline.tech-review"
	CapSpecReview  = "pipeline.spec-review"
	CapRepoSearch  = "pipeline.repo-search"
	CapEffort      = "pipeline.effort"
	CapDifficulty  = "pipeline.difficulty"
	CapDecompose   = "pipeline.decompose"
)

var Caps = []string{
	CapTechReview, CapSpecReview, CapRepoSearch, CapEffort, CapDifficulty, CapDecompose,
}

func emptyInput() json.RawMessage { return json.RawMessage(`{}`) }

// NewTaskIR builds the linear six-step pipeline Task IR (G-22).
func NewTaskIR(summary, notes, entryPoint, ref string) (task.Task, error) {
	id, err := task.NewID()
	if err != nil {
		return task.Task{}, err
	}
	if entryPoint == "" {
		entryPoint = "cli"
	}
	steps := []task.Step{
		{StepID: "tech-review", Capability: CapTechReview, Input: emptyInput()},
		{StepID: "spec-review", Capability: CapSpecReview, Input: emptyInput(), DependsOn: []string{"tech-review"}},
		{StepID: "repo-search", Capability: CapRepoSearch, Input: emptyInput(), DependsOn: []string{"spec-review"}},
		{StepID: "effort", Capability: CapEffort, Input: emptyInput(), DependsOn: []string{"repo-search"}},
		{StepID: "difficulty", Capability: CapDifficulty, Input: emptyInput(), DependsOn: []string{"effort"}},
		{StepID: "decompose", Capability: CapDecompose, Input: emptyInput(), DependsOn: []string{"difficulty"}},
	}
	return task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id,
		CreatedAt:     time.Now().UTC(),
		Source:        task.Source{EntryPoint: entryPoint, Ref: ref},
		Intent:        task.Intent{Summary: summary, Notes: notes},
		Steps:         steps,
		Metadata:      map[string]any{"template": "board-pipeline-v0"},
	}, nil
}
