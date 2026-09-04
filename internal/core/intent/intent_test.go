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

func TestCompilePlayerHeuristics(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	cases := []struct {
		text, method, cap string
	}{
		{"go test", intent.MethodHeuristicTest, "test.go"},
		{"go test ./...", intent.MethodHeuristicTest, "test.go"},
		{"roda os testes", intent.MethodHeuristicTest, "test.go"},
		{"rodar testes", intent.MethodHeuristicTest, "test.go"},
		{"run tests", intent.MethodHeuristicTest, "test.go"},
		{"npm test", intent.MethodHeuristicNPM, "npm.test"},
		{"npm run test", intent.MethodHeuristicNPM, "npm.test"},
		{"roda os testes npm", intent.MethodHeuristicNPM, "npm.test"},
		{"run npm tests", intent.MethodHeuristicNPM, "npm.test"},
		{"git status", intent.MethodHeuristicGit, "git.status"},
		{"git diff", intent.MethodHeuristicGit, "git.diff"},
		{"git log", intent.MethodHeuristicGit, "git.log"},
		{"docker ps", intent.MethodHeuristicDocker, "docker.ps"},
		{"helm lint charts/demo", intent.MethodHeuristicHelm, "helm.lint"},
		{"helm template charts/demo", intent.MethodHeuristicHelm, "helm.template"},
		{"helm list", intent.MethodHeuristicHelm, "helm.list"},
		{"helm status web", intent.MethodHeuristicHelm, "helm.status"},
		{"aws sts get-caller-identity", intent.MethodHeuristicAWS, "aws.sts-identity"},
		{"aws s3 ls", intent.MethodHeuristicAWS, "aws.s3-buckets"},
		{"aws s3 ls s3://data/logs", intent.MethodHeuristicAWS, "aws.s3-objects"},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			res, err := e.Compile(context.Background(), intent.Request{Text: tc.text})
			if err != nil {
				t.Fatal(err)
			}
			if res.Method != tc.method {
				t.Fatalf("method=%s want %s", res.Method, tc.method)
			}
			if len(res.Task.Steps) != 1 || res.Task.Steps[0].Capability != tc.cap {
				t.Fatalf("steps=%v", res.Task.Steps)
			}
		})
	}
}

func TestCompileGoTestBeatsShellPrefix(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "go test -count 1"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicTest || res.Task.Steps[0].Capability != "test.go" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
}

func TestCompileHelmChartInput(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "helm template charts/demo"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicHelm || res.Task.Steps[0].Capability != "helm.template" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
	var in struct {
		Chart string `json:"chart"`
	}
	if err := json.Unmarshal(res.Task.Steps[0].Input, &in); err != nil {
		t.Fatal(err)
	}
	if in.Chart != "charts/demo" {
		t.Fatalf("chart=%s", in.Chart)
	}
}

func TestCompileHelmInstallDoesNotMatchPlayer(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "helm install web charts/demo"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method == intent.MethodHeuristicHelm {
		t.Fatalf("install must not match the helm player heuristics (method=%s)", res.Method)
	}
}

func TestCompileAwsS3UriParsesStatically(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "aws s3 ls s3://data/logs/"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicAWS || res.Task.Steps[0].Capability != "aws.s3-objects" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
	var in struct {
		Bucket string `json:"bucket"`
		Prefix string `json:"prefix"`
	}
	if err := json.Unmarshal(res.Task.Steps[0].Input, &in); err != nil {
		t.Fatal(err)
	}
	if in.Bucket != "data" || in.Prefix != "logs" {
		t.Fatalf("bucket=%s prefix=%s", in.Bucket, in.Prefix)
	}
}

func TestCompileAwsMutantDoesNotMatchPlayer(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "aws s3 rm s3://data/x"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method == intent.MethodHeuristicAWS {
		t.Fatalf("s3 rm must not match the aws player heuristics (method=%s)", res.Method)
	}
}

func TestCompileNpmTestBeatsShellPrefix(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "npm test"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicNPM || res.Task.Steps[0].Capability != "npm.test" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
	var in struct {
		Workdir string `json:"workdir"`
	}
	if err := json.Unmarshal(res.Task.Steps[0].Input, &in); err != nil {
		t.Fatal(err)
	}
	if in.Workdir != "." {
		t.Fatalf("workdir=%q", in.Workdir)
	}
}

func TestCompilePytestBeatsShellPrefix(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "pytest tests/"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicPytest || res.Task.Steps[0].Capability != "pytest.run" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
}

func TestCompileYarnTestBeatsShellPrefix(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "yarn test"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicYarn || res.Task.Steps[0].Capability != "yarn.test" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
}

func TestCompileYarnInstallStillShell(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "yarn install"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicShell || res.Task.Steps[0].Capability != "shell.exec" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
}

func TestCompileNpmInstallStillShell(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "npm install"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicShell || res.Task.Steps[0].Capability != "shell.exec" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
}

func TestCompileGoBuildStillShell(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "go build"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicShell || res.Task.Steps[0].Capability != "shell.exec" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
}

func TestCompileGitPushStillShell(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "git push"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicShell || res.Task.Steps[0].Capability != "shell.exec" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
}

func TestCompilePlayerBeatsPipelineKeyword(t *testing.T) {
	e := intent.New(llm.HeuristicCompleter{})
	res, err := e.Compile(context.Background(), intent.Request{Text: "run tests then review the architecture"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Method != intent.MethodHeuristicTest || res.Task.Steps[0].Capability != "test.go" {
		t.Fatalf("method=%s cap=%s", res.Method, res.Task.Steps[0].Capability)
	}
}

func TestCompileEmptyRejected(t *testing.T) {
	e := intent.New(nil)
	_, err := e.Compile(context.Background(), intent.Request{Text: "  "})
	if err == nil {
		t.Fatal("expected error")
	}
}
