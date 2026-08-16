package intent_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/intent"
	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/players/llm"
)

func TestCompileShellHeuristic(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "echo hello-intent"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicShell {
		t.Fatalf("method=%s", res.Method)
	}
	if len(res.Task.Steps) != 1 || res.Task.Steps[0].Capability != "shell.exec" {
		t.Fatalf("steps=%v", res.Task.Steps)
	}
	var in struct {
		Command []string `json:"command"`
	}
	if err := json.Unmarshal(res.Task.Steps[0].Input, &in); err != nil {
		t.Fatal(err)
	}
	if len(in.Command) < 2 || in.Command[0] != "echo" || in.Command[1] != "hello-intent" {
		t.Fatalf("command=%v", in.Command)
	}
}

func TestCompilePipelineHeuristic(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "revisar a arquitetura do workspace"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicPipeline {
		t.Fatalf("method=%s", res.Method)
	}
	if len(res.Task.Steps) != len(corepipe.Caps) {
		t.Fatalf("expected pipeline steps, got %d", len(res.Task.Steps))
	}
}

func TestCompileLLMFallbackShell(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "prepare a friendly greeting for the team"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodLLM {
		t.Fatalf("method=%s", res.Method)
	}
	if res.Task.Steps[0].Capability != "shell.exec" {
		t.Fatalf("cap=%s", res.Task.Steps[0].Capability)
	}
}

func TestCompileEmptyRejected(t *testing.T) {
	e := intent.New(nil)
	_, err := e.Compile(context.Background(), intent.Request{Text: "  "})
	if err == nil {
		t.Fatal("expected error")
	}
}
