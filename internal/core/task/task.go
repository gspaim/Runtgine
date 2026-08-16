package task

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const SchemaVersion = "0.1.0"

type Task struct {
	SchemaVersion string         `json:"schema_version"`
	TaskID        string         `json:"task_id"`
	CreatedAt     time.Time      `json:"created_at"`
	Source        Source         `json:"source"`
	Intent        Intent         `json:"intent"`
	Steps         []Step         `json:"steps"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type Source struct {
	EntryPoint string `json:"entry_point"` // cli|tui|board|api|other
	Ref        string `json:"ref,omitempty"`
}

type Intent struct {
	Summary string `json:"summary"`
	Notes   string `json:"notes,omitempty"`
}

type Step struct {
	StepID     string          `json:"step_id"`
	Capability string          `json:"capability"`
	Input      json.RawMessage `json:"input"`
	DependsOn  []string        `json:"depends_on,omitempty"`
	MaxRetries *int            `json:"max_retries,omitempty"` // G-30 configurable per step
}

func NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func Parse(data []byte) (Task, error) {
	var t Task
	if err := json.Unmarshal(data, &t); err != nil {
		return t, fmt.Errorf("parse task ir: %w", err)
	}
	if t.TaskID == "" {
		id, err := NewID()
		if err != nil {
			return t, err
		}
		t.TaskID = id
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	if t.SchemaVersion == "" {
		t.SchemaVersion = SchemaVersion
	}
	if t.Metadata == nil {
		t.Metadata = map[string]any{}
	}
	for i := range t.Steps {
		if t.Steps[i].DependsOn == nil {
			t.Steps[i].DependsOn = []string{}
		}
	}
	return t, nil
}
