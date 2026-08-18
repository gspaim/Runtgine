package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/claim"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

func waitEvent(t *testing.T, core *api.Core, runID, typ string) {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		snap, err := core.GetRun(context.Background(), runID)
		if err == nil {
			for _, e := range snap.Events {
				if e.Type == typ {
					return
				}
			}
		}
		time.Sleep(15 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for event %s on %s", typ, runID)
}

func waitClaim(t *testing.T, core *api.Core, runID string, want store.Status) api.RunSnapshot {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	var snap api.RunSnapshot
	for time.Now().Before(deadline) {
		var err error
		snap, err = core.GetRun(context.Background(), runID)
		if err == nil && store.Status(snap.Status) == want {
			return snap
		}
		time.Sleep(15 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s (last=%s err=%s)", want, snap.Status, snap.Error)
	return snap
}

func newTask(t *testing.T, summary string, steps []task.Step) task.Task {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: summary},
		Steps:         steps,
	}
}

func writeStep(path, content string) task.Step {
	raw, _ := json.Marshal(map[string]any{"path": path, "content": content, "create_parents": true})
	return task.Step{StepID: "write", Capability: "fs.write", Input: raw}
}

func sleepStep(seconds int) task.Step {
	raw, _ := json.Marshal(map[string]any{
		"command":    []string{"sleep", strconv.Itoa(seconds)},
		"workdir":    ".",
		"timeout_ms": (seconds + 5) * 1000,
	})
	return task.Step{StepID: "hold", Capability: "shell.exec", Input: raw}
}

func TestClaimShellConcurrentHello(t *testing.T) {
	core := openPolicyCore(t, nil)
	a, err := core.SubmitTask(context.Background(), helloTask(t))
	if err != nil {
		t.Fatal(err)
	}
	b, err := core.SubmitTask(context.Background(), helloTask(t))
	if err != nil {
		t.Fatal(err)
	}
	waitClaim(t, core, a, store.StatusSucceeded)
	waitClaim(t, core, b, store.StatusSucceeded)
}

func TestClaimFSWriteSamePathConflict(t *testing.T) {
	core := openPolicyCore(t, nil)
	holder := newTask(t, "hold notes", []task.Step{
		writeStep("notes.md", "one"),
		sleepStep(2),
	})
	challenger := newTask(t, "write notes", []task.Step{writeStep("notes.md", "two")})
	a, err := core.SubmitTask(context.Background(), holder)
	if err != nil {
		t.Fatal(err)
	}
	waitEvent(t, core, a, event.TypeClaimAcquired)
	b, err := core.SubmitTask(context.Background(), challenger)
	if err != nil {
		t.Fatal(err)
	}
	snapB := waitClaim(t, core, b, store.StatusFailed)
	if !containsCode(snapB.Error, result.CodeClaimConflict) {
		t.Fatalf("b error=%s", snapB.Error)
	}
	if !containsCode(snapB.Error, a) {
		t.Fatalf("missing holder run id in %s", snapB.Error)
	}
	var started bool
	for _, e := range snapB.Events {
		if e.Type == event.TypeStepStarted {
			started = true
		}
		if e.Type == event.TypeClaimConflict {
			if e.Payload["holder_run_id"] != a {
				t.Fatalf("payload=%v", e.Payload)
			}
		}
	}
	if started {
		t.Fatal("challenger must not Execute")
	}
	waitClaim(t, core, a, store.StatusSucceeded)
}

func TestClaimFSWritePrefixConflict(t *testing.T) {
	core := openPolicyCore(t, nil)
	ctx := context.Background()
	if err := core.Store.InsertRun(ctx, store.Run{RunID: "holder-run", TaskID: "holder-task", Status: store.StatusRunning}, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if err := core.Runner.Claims.Acquire(ctx, "holder-run", "s1", claim.Resource{Kind: claim.KindPath, Key: "src"}); err != nil {
		t.Fatal(err)
	}
	b, err := core.SubmitTask(ctx, newTask(t, "write src/main", []task.Step{writeStep("src/main.go", "pkg")}))
	if err != nil {
		t.Fatal(err)
	}
	snapB := waitClaim(t, core, b, store.StatusFailed)
	if !containsCode(snapB.Error, result.CodeClaimConflict) || !containsCode(snapB.Error, "holder-run") {
		t.Fatalf("error=%s", snapB.Error)
	}
	var started bool
	for _, e := range snapB.Events {
		if e.Type == event.TypeStepStarted {
			started = true
		}
	}
	if started {
		t.Fatal("challenger must not Execute")
	}
}

func TestClaimDisjointPathsOK(t *testing.T) {
	core := openPolicyCore(t, nil)
	a, err := core.SubmitTask(context.Background(), newTask(t, "a", []task.Step{writeStep("a.txt", "a")}))
	if err != nil {
		t.Fatal(err)
	}
	b, err := core.SubmitTask(context.Background(), newTask(t, "b", []task.Step{writeStep("b.txt", "b")}))
	if err != nil {
		t.Fatal(err)
	}
	waitClaim(t, core, a, store.StatusSucceeded)
	waitClaim(t, core, b, store.StatusSucceeded)
}

func TestClaimGitCommitVsFSWrite(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}
	ws := t.TempDir()
	initGitRepo(t, ws)
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	core, err := api.Open(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = core.Close() })

	addRaw, _ := json.Marshal(map[string]any{"paths": []string{"README"}})
	holder := newTask(t, "git add hold", []task.Step{
		{StepID: "add", Capability: "git.add", Input: addRaw},
		sleepStep(2),
	})
	a, err := core.SubmitTask(context.Background(), holder)
	if err != nil {
		t.Fatal(err)
	}
	waitEvent(t, core, a, event.TypeClaimAcquired)
	b, err := core.SubmitTask(context.Background(), newTask(t, "fs", []task.Step{writeStep("x.txt", "x")}))
	if err != nil {
		t.Fatal(err)
	}
	snapB := waitClaim(t, core, b, store.StatusFailed)
	if !containsCode(snapB.Error, result.CodeClaimConflict) {
		t.Fatalf("error=%s", snapB.Error)
	}
	waitClaim(t, core, a, store.StatusSucceeded)
}

func TestClaimHITLDoesNotHoldThenConflictOnGrant(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}
	ws := t.TempDir()
	initGitRepo(t, ws)
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	cfg.ExecutionPolicy.Capabilities = map[string]string{"fs.write": "approval-required"}
	core, err := api.Open(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = core.Close() })

	waiting := newTask(t, "gated write", []task.Step{writeStep("notes.md", "late")})
	a, err := core.SubmitTask(context.Background(), waiting)
	if err != nil {
		t.Fatal(err)
	}
	waitClaim(t, core, a, store.StatusWaitingApproval)
	active, err := core.Store.ListActiveClaims(context.Background())
	if err != nil || len(active) != 0 {
		t.Fatalf("HITL must not claim: %v %v", active, err)
	}

	addRaw, _ := json.Marshal(map[string]any{"paths": []string{"README"}})
	holder := newTask(t, "hold workspace", []task.Step{
		{StepID: "add", Capability: "git.add", Input: addRaw},
		sleepStep(2),
	})
	b, err := core.SubmitTask(context.Background(), holder)
	if err != nil {
		t.Fatal(err)
	}
	waitEvent(t, core, b, event.TypeClaimAcquired)

	if err := core.ApproveRun(a, runner.DecisionGrant); err != nil {
		t.Fatal(err)
	}
	snapA := waitClaim(t, core, a, store.StatusFailed)
	if !containsCode(snapA.Error, result.CodeClaimConflict) {
		t.Fatalf("after grant error=%s", snapA.Error)
	}
	waitClaim(t, core, b, store.StatusSucceeded)
}

func TestClaimDenyNeverAcquires(t *testing.T) {
	core := openPolicyCore(t, map[string]string{"fs.write": "deny"})
	_, err := core.SubmitTask(context.Background(), newTask(t, "denied", []task.Step{writeStep("z.txt", "z")}))
	if err == nil {
		t.Fatal("expected deny")
	}
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodePolicyDenied {
		t.Fatalf("got %#v", err)
	}
	active, err := core.Store.ListActiveClaims(context.Background())
	if err != nil || len(active) != 0 {
		t.Fatalf("deny acquired %v err=%v", active, err)
	}
}

func TestClaimOrphanSweepOnOpen(t *testing.T) {
	ws := t.TempDir()
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	core, err := api.Open(cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := core.Store.InsertRun(ctx, store.Run{RunID: "dead-run", TaskID: "dead-task", Status: store.StatusFailed}, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := core.Store.TryAcquireClaim(ctx, "dead-run", "s1", "path", "notes.md", func(aKind, aKey, bKind, bKey string) bool {
		return aKind == bKind && aKey == bKey
	}); err != nil {
		t.Fatal(err)
	}
	if err := core.Close(); err != nil {
		t.Fatal(err)
	}

	core2, err := api.Open(cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = core2.Close() })
	active, err := core2.Store.ListActiveClaims(ctx)
	if err != nil || len(active) != 0 {
		t.Fatalf("orphans remain: %v err=%v", active, err)
	}
	runID, err := core2.SubmitTask(ctx, newTask(t, "after sweep", []task.Step{writeStep("notes.md", "ok")}))
	if err != nil {
		t.Fatal(err)
	}
	waitClaim(t, core2, runID, store.StatusSucceeded)
}

func TestClaimDockerRunMountRequired(t *testing.T) {
	core := openPolicyCore(t, map[string]string{"docker.run": "allow"})
	ctx := context.Background()
	if err := core.Store.InsertRun(ctx, store.Run{RunID: "ws-holder", TaskID: "t", Status: store.StatusRunning}, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if err := core.Runner.Claims.Acquire(ctx, "ws-holder", "s1", claim.Workspace()); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"image": "alpine:3.19", "mount_workspace": true, "argv": []string{"true"}})
	b, err := core.SubmitTask(ctx, newTask(t, "docker mount", []task.Step{
		{StepID: "run", Capability: "docker.run", Input: raw},
	}))
	if err != nil {
		t.Fatal(err)
	}
	snapB := waitClaim(t, core, b, store.StatusFailed)
	if !containsCode(snapB.Error, result.CodeClaimConflict) {
		t.Fatalf("docker mount should claim workspace: %s", snapB.Error)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "README")
	run("commit", "-m", "init")
}
