package blast

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/claim"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

func TestTouchedAndRequiredTables(t *testing.T) {
	hits, err := Touched("shell.exec", json.RawMessage(`{"command":["echo","hi"]}`))
	if err != nil || len(hits) != 0 {
		t.Fatalf("shell: %v %v", hits, err)
	}
	hits, err = Touched("http.get", json.RawMessage(`{"url":"https://example.com/"}`))
	if err != nil || len(hits) != 0 {
		t.Fatalf("http.get: %v %v", hits, err)
	}
	hits, err = Touched("test.go", json.RawMessage(`{"packages":["./..."]}`))
	if err != nil || len(hits) != 0 {
		t.Fatalf("test.go: %v %v", hits, err)
	}
	hits, err = Touched("fs.read", json.RawMessage(`{"path":"a.txt"}`))
	if err != nil || len(hits) != 1 || hits[0].Mode != ModeRead || hits[0].Resource.Key != "a.txt" {
		t.Fatalf("fs.read: %+v err=%v", hits, err)
	}
	hits, err = Touched("fs.write", json.RawMessage(`{"path":"notes.md"}`))
	if err != nil || len(hits) != 1 || hits[0].Mode != ModeWrite {
		t.Fatalf("fs.write: %+v err=%v", hits, err)
	}
	hits, err = Touched("git.add", json.RawMessage(`{"paths":["README","src/a.go"]}`))
	if err != nil || len(hits) != 2 || hits[0].Resource.Key != "README" {
		t.Fatalf("git.add: %+v err=%v", hits, err)
	}
	hits, err = Touched("git.commit", json.RawMessage(`{"message":"m"}`))
	if err != nil || len(hits) != 1 || hits[0].Resource.Kind != claim.KindWorkspace {
		t.Fatalf("git.commit: %+v err=%v", hits, err)
	}
	hits, err = Touched("docker.run", json.RawMessage(`{"image":"alpine:3.19"}`))
	if err != nil || len(hits) != 0 {
		t.Fatalf("docker.run no mount: %+v err=%v", hits, err)
	}
	hits, err = Touched("docker.run", json.RawMessage(`{"image":"alpine:3.19","mount_workspace":true}`))
	if err != nil || len(hits) != 1 || hits[0].Resource.Kind != claim.KindWorkspace {
		t.Fatalf("docker.run mount: %+v err=%v", hits, err)
	}
	if _, err := Touched("fs.write", json.RawMessage(`{"path":"../out"}`)); err == nil {
		t.Fatal("expected escape")
	}
}

func TestRiskOf(t *testing.T) {
	if RiskOf(nil) != RiskNone {
		t.Fatal("empty")
	}
	if RiskOf([]claim.Resource{{Kind: claim.KindPath, Key: "a.txt"}}) != RiskPath {
		t.Fatal("path")
	}
	if RiskOf([]claim.Resource{
		{Kind: claim.KindPath, Key: "a.txt"},
		claim.Workspace(),
	}) != RiskWorkspace {
		t.Fatal("workspace wins")
	}
}

func TestAnalyzeHelloEmpty(t *testing.T) {
	rep, err := Analyze([]task.Step{{
		StepID:     "s1",
		Capability: "shell.exec",
		Input:      json.RawMessage(`{"command":["echo","hello-runtgine"]}`),
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Risk != RiskNone || len(rep.PredictedClaims) != 0 || len(rep.Touches) != 0 {
		t.Fatalf("%+v", rep)
	}
	if len(rep.Capabilities) != 1 || rep.Capabilities[0] != "shell.exec" {
		t.Fatalf("caps=%v", rep.Capabilities)
	}
	if rep.Conflicts == nil || rep.Images == nil || rep.Affected == nil {
		t.Fatal("empty slices must be non-nil")
	}
	if len(rep.Affected) != 0 {
		t.Fatalf("affected=%v", rep.Affected)
	}
}

func TestAnalyzeGitAddVsWrite(t *testing.T) {
	rep, err := Analyze([]task.Step{
		{StepID: "add", Capability: "git.add", Input: json.RawMessage(`{"paths":["README"]}`)},
		{StepID: "w", Capability: "fs.write", Input: json.RawMessage(`{"path":"notes.md"}`)},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Risk != RiskWorkspace {
		t.Fatalf("risk=%s", rep.Risk)
	}
	if len(rep.PredictedClaims) != 2 {
		t.Fatalf("claims=%v", rep.PredictedClaims)
	}
	if rep.PredictedClaims[0].Kind != "workspace" || rep.PredictedClaims[1].Key != "notes.md" {
		t.Fatalf("claims=%v", rep.PredictedClaims)
	}
	found := false
	for _, th := range rep.Touches {
		if th.Key == "README" && th.Mode == ModeWrite {
			found = true
		}
	}
	if !found {
		t.Fatalf("touches=%v", rep.Touches)
	}
}

func TestOverlayConflicts(t *testing.T) {
	pred := []claim.Resource{{Kind: claim.KindPath, Key: "notes.md"}}
	active := []store.ResourceClaim{{RunID: "holder-run", Kind: "path", Key: "notes.md"}}
	got := Overlay(pred, active)
	if len(got) != 1 || got[0].HolderRunID != "holder-run" {
		t.Fatalf("%v", got)
	}
	if len(Overlay(pred, nil)) != 0 {
		t.Fatal("no active")
	}
}

func TestAnalyzeEscapeError(t *testing.T) {
	_, err := Analyze([]task.Step{{
		StepID: "s1", Capability: "fs.write", Input: json.RawMessage(`{"path":"../secret"}`),
	}}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("%#v", err)
	}
}

func TestAnalyzeImages(t *testing.T) {
	rep, err := Analyze([]task.Step{{
		StepID:     "run",
		Capability: "docker.run",
		Input:      json.RawMessage(`{"image":"alpine:3.19"}`),
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Images) != 1 || rep.Images[0] != "alpine:3.19" {
		t.Fatalf("images=%v", rep.Images)
	}
	if rep.Risk != RiskNone {
		t.Fatalf("no mount → none, got %s", rep.Risk)
	}
}
