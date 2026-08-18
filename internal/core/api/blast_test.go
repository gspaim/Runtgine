package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/blast"
	"github.com/gspaim/Runtgine/internal/core/claim"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

func TestBlastHelloEmptyAndNoRun(t *testing.T) {
	core := openPolicyCore(t, nil)
	rep, err := core.BlastTask(context.Background(), helloTask(t))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Risk != blast.RiskNone || len(rep.PredictedClaims) != 0 || len(rep.Touches) != 0 {
		t.Fatalf("%+v", rep)
	}
	found := false
	for _, c := range rep.Capabilities {
		if c == "shell.exec" {
			found = true
		}
	}
	if !found {
		t.Fatalf("caps=%v", rep.Capabilities)
	}
	runs, err := core.Store.ListRuns(context.Background(), 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 0 {
		t.Fatalf("blast created runs: %v", runs)
	}
	active, err := core.Store.ListActiveClaims(context.Background())
	if err != nil || len(active) != 0 {
		t.Fatalf("claims=%v err=%v", active, err)
	}
}

func TestBlastFSReadNoClaim(t *testing.T) {
	core := openPolicyCore(t, nil)
	if err := os.WriteFile(filepath.Join(core.Cfg.WorkspaceRoot, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"path": "a.txt"})
	rep, err := core.BlastTask(context.Background(), newTask(t, "read", []task.Step{
		{StepID: "r", Capability: "fs.read", Input: raw},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Risk != blast.RiskNone || len(rep.PredictedClaims) != 0 {
		t.Fatalf("%+v", rep)
	}
	if len(rep.Touches) != 1 || rep.Touches[0].Key != "a.txt" || rep.Touches[0].Mode != blast.ModeRead {
		t.Fatalf("touches=%v", rep.Touches)
	}
}

func TestBlastFSWritePredictedPath(t *testing.T) {
	core := openPolicyCore(t, nil)
	rep, err := core.BlastTask(context.Background(), newTask(t, "write", []task.Step{writeStep("notes.md", "x")}))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Risk != blast.RiskPath {
		t.Fatalf("risk=%s", rep.Risk)
	}
	if len(rep.PredictedClaims) != 1 || rep.PredictedClaims[0].Key != "notes.md" {
		t.Fatalf("claims=%v", rep.PredictedClaims)
	}
}

func TestBlastGitAddWorkspaceClaim(t *testing.T) {
	core := openPolicyCore(t, nil)
	if err := os.WriteFile(filepath.Join(core.Cfg.WorkspaceRoot, "README"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	addRaw, _ := json.Marshal(map[string]any{"paths": []string{"README"}})
	rep, err := core.BlastTask(context.Background(), newTask(t, "add", []task.Step{
		{StepID: "add", Capability: "git.add", Input: addRaw},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Risk != blast.RiskWorkspace {
		t.Fatalf("risk=%s", rep.Risk)
	}
	if len(rep.PredictedClaims) != 1 || rep.PredictedClaims[0].Kind != "workspace" {
		t.Fatalf("claims=%v", rep.PredictedClaims)
	}
	found := false
	for _, th := range rep.Touches {
		if th.Key == "README" && th.Mode == blast.ModeWrite {
			found = true
		}
	}
	if !found {
		t.Fatalf("touches=%v", rep.Touches)
	}
}

func TestBlastOverlayHolderNoAcquire(t *testing.T) {
	core := openPolicyCore(t, nil)
	ctx := context.Background()
	if err := core.Store.InsertRun(ctx, store.Run{RunID: "holder-run", TaskID: "t", Status: store.StatusRunning}, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if err := core.Runner.Claims.Acquire(ctx, "holder-run", "s1", claim.Resource{Kind: claim.KindPath, Key: "notes.md"}); err != nil {
		t.Fatal(err)
	}
	before, err := core.Store.ListActiveClaims(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := core.BlastTask(ctx, newTask(t, "write", []task.Step{writeStep("notes.md", "x")}))
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Conflicts) != 1 || rep.Conflicts[0].HolderRunID != "holder-run" {
		t.Fatalf("conflicts=%v", rep.Conflicts)
	}
	after, err := core.Store.ListActiveClaims(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("blast acquired a claim: before=%v after=%v", before, after)
	}
}

func TestBlastEscapePath(t *testing.T) {
	core := openPolicyCore(t, nil)
	_, err := core.BlastTask(context.Background(), newTask(t, "escape", []task.Step{writeStep("../secret", "x")}))
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("%#v", err)
	}
}

func TestBlastIgnoresPolicyDeny(t *testing.T) {
	core := openPolicyCore(t, map[string]string{"fs.write": "deny"})
	rep, err := core.BlastTask(context.Background(), newTask(t, "denied write", []task.Step{writeStep("z.txt", "z")}))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Risk != blast.RiskPath {
		t.Fatalf("risk=%s", rep.Risk)
	}
}
