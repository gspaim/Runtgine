package task_test

import (
	"os"
	"path/filepath"
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
	if err := task.ValidateDocument(raw); err != nil {
		t.Fatal(err)
	}
	tk, err := task.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.IdentityValidate(tk); err != nil {
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

func TestRejectWrongSchemaVersion(t *testing.T) {
	raw := []byte(`{
	  "schema_version":"9.9.9",
	  "source":{"entry_point":"cli"},
	  "intent":{"summary":"x"},
	  "steps":[{"step_id":"s1","capability":"shell.exec","input":{"command":["echo","hi"]}}]
	}`)
	if err := task.ValidateDocument(raw); err == nil {
		t.Fatal("expected schema_version rejection")
	}
}

func TestRejectExtraProperty(t *testing.T) {
	raw := []byte(`{
	  "schema_version":"0.1.0",
	  "source":{"entry_point":"cli"},
	  "intent":{"summary":"x"},
	  "extra_field": true,
	  "steps":[{"step_id":"s1","capability":"shell.exec","input":{"command":["echo","hi"]}}]
	}`)
	if err := task.ValidateDocument(raw); err == nil {
		t.Fatal("expected additionalProperties rejection")
	}
}

func TestRejectNonV7TaskID(t *testing.T) {
	raw := []byte(`{
	  "schema_version":"0.1.0",
	  "task_id":"550e8400-e29b-41d4-a716-446655440000",
	  "source":{"entry_point":"cli"},
	  "intent":{"summary":"x"},
	  "steps":[{"step_id":"s1","capability":"shell.exec","input":{"command":["echo","hi"]}}]
	}`)
	if err := task.ValidateDocument(raw); err != nil {
		t.Fatal(err)
	}
	tk, err := task.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.IdentityValidate(tk); err == nil {
		t.Fatal("expected UUID v7 rejection")
	}
}

func TestParseGeneratesUUIDv7(t *testing.T) {
	raw := []byte(`{
	  "schema_version":"0.1.0",
	  "source":{"entry_point":"cli"},
	  "intent":{"summary":"x"},
	  "steps":[{"step_id":"s1","capability":"shell.exec","input":{"command":["echo","hi"]}}]
	}`)
	tk, err := task.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !task.IsUUIDv7(tk.TaskID) {
		t.Fatalf("generated task_id is not UUID v7: %s", tk.TaskID)
	}
	if err := task.IdentityValidate(tk); err != nil {
		t.Fatal(err)
	}
}

func TestAcceptWailsEntryPoint(t *testing.T) {
	raw := []byte(`{
	  "schema_version":"0.1.0",
	  "source":{"entry_point":"wails"},
	  "intent":{"summary":"git status"},
	  "steps":[{"step_id":"s1","capability":"git.status","input":{"workdir":"."}}]
	}`)
	if err := task.ValidateDocument(raw); err != nil {
		t.Fatal(err)
	}
	tk, err := task.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.StructuralValidate(tk); err != nil {
		t.Fatal(err)
	}
	if tk.Source.EntryPoint != "wails" {
		t.Fatalf("entry_point=%s", tk.Source.EntryPoint)
	}
}

func TestEmbeddedSchemaMatchesCanonical(t *testing.T) {
	canonical, err := os.ReadFile(filepath.Join("..", "..", "..", "schemas", "task-ir-v0.1.0.json"))
	if err != nil {
		t.Fatal(err)
	}
	embedded := task.EmbeddedTaskIRSchema()
	if string(canonical) != string(embedded) {
		t.Fatal("embedded Task IR schema diverges from schemas/task-ir-v0.1.0.json")
	}
}
