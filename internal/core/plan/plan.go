package plan

import (
	"encoding/json"

	"github.com/gspaim/Runtgine/internal/core/task"
	"github.com/google/uuid"
)

const SchemaVersion = "0.1.0"

type Plan struct {
	SchemaVersion string `json:"schema_version"`
	PlanID        string `json:"plan_id"`
	TaskID        string `json:"task_id"`
	RunID         string `json:"run_id"`
	Steps         []Step `json:"steps"`
}

type Step struct {
	StepID     string          `json:"step_id"`
	Capability string          `json:"capability"`
	Player     string          `json:"player"`
	Input      json.RawMessage `json:"input"`
	DependsOn  []string        `json:"depends_on,omitempty"`
	MaxRetries int             `json:"max_retries"`
}

func FromTask(t task.Task, runID string, resolve func(capability string) (player string, err error)) (Plan, error) {
	planID, err := uuid.NewV7()
	if err != nil {
		return Plan{}, err
	}
	steps := make([]Step, 0, len(t.Steps))
	for _, s := range t.Steps {
		player, err := resolve(s.Capability)
		if err != nil {
			return Plan{}, err
		}
		retries := 0
		if s.MaxRetries != nil {
			retries = *s.MaxRetries
		}
		steps = append(steps, Step{
			StepID:     s.StepID,
			Capability: s.Capability,
			Player:     player,
			Input:      s.Input,
			DependsOn:  append([]string{}, s.DependsOn...),
			MaxRetries: retries,
		})
	}
	return Plan{
		SchemaVersion: SchemaVersion,
		PlanID:        planID.String(),
		TaskID:        t.TaskID,
		RunID:         runID,
		Steps:         steps,
	}, nil
}
