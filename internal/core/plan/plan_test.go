package plan_test

import (
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/plan"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/task"
)

func mkTask(t *testing.T) task.Task {
	t.Helper()
	id, err := task.NewID()
	if err != nil {
		t.Fatal(err)
	}
	retries := 3
	return task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id,
		CreatedAt:     time.Now().UTC(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "plan test"},
		Steps: []task.Step{
			{StepID: "s1", Capability: "fake.a", Input: []byte(`{}`)},
			{StepID: "s2", Capability: "fake.b", Input: []byte(`{}`), DependsOn: []string{"s1"}, MaxRetries: &retries},
		},
	}
}

func TestFromTaskResolvesAndCopiesSteps(t *testing.T) {
	tk := mkTask(t)
	resolved := map[string]string{}
	p, err := plan.FromTask(tk, "run-1", func(capability string) (string, error) {
		name := "player-" + capability
		resolved[capability] = name
		return name, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.SchemaVersion != plan.SchemaVersion {
		t.Fatalf("schema=%s", p.SchemaVersion)
	}
	if p.PlanID == "" || p.TaskID != tk.TaskID || p.RunID != "run-1" {
		t.Fatalf("plan=%+v", p)
	}
	if len(p.Steps) != 2 {
		t.Fatalf("steps=%+v", p.Steps)
	}
	if p.Steps[0].Player != "player-fake.a" {
		t.Fatalf("player=%s", p.Steps[0].Player)
	}
	if p.Steps[1].MaxRetries != 3 {
		t.Fatalf("max_retries=%d", p.Steps[1].MaxRetries)
	}
	if p.Steps[0].MaxRetries != 0 {
		t.Fatalf("nil max_retries must map to 0, got %d", p.Steps[0].MaxRetries)
	}
	if len(p.Steps[1].DependsOn) != 1 || p.Steps[1].DependsOn[0] != "s1" {
		t.Fatalf("depends_on=%v", p.Steps[1].DependsOn)
	}
	if resolved["fake.a"] == "" || resolved["fake.b"] == "" {
		t.Fatalf("resolver did not see all capabilities: %v", resolved)
	}
}

func TestFromTaskPropagatesResolveError(t *testing.T) {
	tk := mkTask(t)
	_, err := plan.FromTask(tk, "run-1", func(capability string) (string, error) {
		return "", result.Validation(result.CodeUnknownCapability, "unregistered", nil)
	})
	if err == nil {
		t.Fatal("expected resolve error")
	}
	var ve result.Error
	if !asErr(err, &ve) || ve.Code != result.CodeUnknownCapability {
		t.Fatalf("err=%v", err)
	}
}

func asErr(err error, ve *result.Error) bool {
	e, ok := err.(result.Error)
	if ok {
		*ve = e
	}
	return ok
}
