package templates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkipsInvalidAndDuplicates(t *testing.T) {
	dir := t.TempDir()
	valid := `{
  "id": "verify",
  "title": "Verify",
  "steps": [{"step_id": "s1", "capability": "git.status", "input": {}}]
}`
	if err := os.WriteFile(filepath.Join(dir, "a.json"), []byte(valid), 0o644); err != nil {
		t.Fatal(err)
	}
	dup := `{
  "id": "verify",
  "title": "Dup",
  "steps": [{"step_id": "s1", "capability": "test.go", "input": {}}]
}`
	if err := os.WriteFile(filepath.Join(dir, "b.json"), []byte(dup), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "extra.json"), []byte(`{
  "id": "x",
  "title": "X",
  "remote": "git://nope",
  "steps": [{"step_id": "s1", "capability": "git.status"}]
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, warns, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "verify" {
		t.Fatalf("got=%+v", got)
	}
	if len(warns) < 2 {
		t.Fatalf("warns=%v", warns)
	}
}

func TestLoadMissingDir(t *testing.T) {
	got, warns, err := Load(filepath.Join(t.TempDir(), "nope"))
	if err != nil || len(got) != 0 || len(warns) != 0 {
		t.Fatalf("got=%v warns=%v err=%v", got, warns, err)
	}
}

func TestParseRejectsCycleAndBadID(t *testing.T) {
	_, err := parse([]byte(`{
  "id": "bad id",
  "title": "X",
  "steps": [{"step_id": "s1", "capability": "git.status"}]
}`))
	if err == nil {
		t.Fatal("expected bad id")
	}
	_, err = parse([]byte(`{
  "id": "cyc",
  "title": "X",
  "steps": [
    {"step_id": "a", "capability": "git.status", "depends_on": ["b"]},
    {"step_id": "b", "capability": "test.go", "depends_on": ["a"]}
  ]
}`))
	if err == nil {
		t.Fatal("expected cycle")
	}
}

func TestCompileCopiesSteps(t *testing.T) {
	tpl := Template{
		ID:    "verify",
		Title: "Verify workspace",
		Steps: []Step{
			{StepID: "status", Capability: "git.status"},
			{StepID: "test", Capability: "test.go", DependsOn: []string{"status"}},
		},
	}
	tk, err := Compile(tpl, "cli", "test", "run template verify")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Intent.Summary != "Verify workspace" {
		t.Fatalf("summary=%s", tk.Intent.Summary)
	}
	if tk.Metadata["template"] != "verify" {
		t.Fatalf("meta=%v", tk.Metadata)
	}
	if len(tk.Steps) != 2 || tk.Steps[1].Capability != "test.go" {
		t.Fatalf("steps=%v", tk.Steps)
	}
	if string(tk.Steps[0].Input) != "{}" {
		t.Fatalf("input=%s", tk.Steps[0].Input)
	}
	var round Template
	raw, _ := json.Marshal(tpl)
	if err := json.Unmarshal(raw, &round); err != nil {
		t.Fatal(err)
	}
}

func TestLookup(t *testing.T) {
	list := []Template{{ID: "verify", Title: "V", Steps: []Step{{StepID: "s1", Capability: "git.status"}}}}
	if _, ok := Lookup(list, "verify"); !ok {
		t.Fatal("missing")
	}
	if _, ok := Lookup(list, "nope"); ok {
		t.Fatal("unexpected")
	}
}
