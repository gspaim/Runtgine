package task_test

import (
	"testing"

	"github.com/gspaim/Runtgine/internal/core/task"
)

func TestStructuralValidateOK(t *testing.T) {
	raw := []byte(`{
	  "schema_version":"0.1.0",
	  "task_id":"01900000-0000-7000-8000-000000000001",
	  "created_at":"2026-08-16T00:00:00Z",
	  "source":{"entry_point":"cli"},
	  "intent":{"summary":"x"},
	  "steps":[{"step_id":"s1","capability":"shell.exec","input":{"command":["echo","hi"]}}]
	}`)
	tk, err := task.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.StructuralValidate(tk); err != nil {
		t.Fatal(err)
	}
}

func TestCycleDetected(t *testing.T) {
	raw := []byte(`{
	  "schema_version":"0.1.0",
	  "source":{"entry_point":"cli"},
	  "intent":{"summary":"x"},
	  "steps":[
	    {"step_id":"a","capability":"shell.exec","input":{"command":["true"]},"depends_on":["b"]},
	    {"step_id":"b","capability":"shell.exec","input":{"command":["true"]},"depends_on":["a"]}
	  ]
	}`)
	tk, err := task.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.StructuralValidate(tk); err == nil {
		t.Fatal("expected cycle error")
	}
}
