package pipeline_test

import (
	"testing"

	"github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/task"
)

func TestNewTaskIRLinear(t *testing.T) {
	tk, err := pipeline.NewTaskIR("Do X", "notes", "cli", "ref")
	if err != nil {
		t.Fatal(err)
	}
	if err := task.StructuralValidate(tk); err != nil {
		t.Fatal(err)
	}
	if len(tk.Steps) != 6 {
		t.Fatalf("steps=%d", len(tk.Steps))
	}
	if tk.Steps[0].Capability != pipeline.CapTechReview {
		t.Fatal(tk.Steps[0].Capability)
	}
}
