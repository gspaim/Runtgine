package git_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/players/git"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
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
	return dir
}

func execCap(t *testing.T, ws, cap string, input any) json.RawMessage {
	t.Helper()
	p := git.New()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: cap,
		Input:      raw,
		Workspace:  ws,
	})
	if err != nil {
		t.Fatalf("%s: %v", cap, err)
	}
	return out
}

func TestStatusCleanAndDirty(t *testing.T) {
	ws := initRepo(t)
	out := execCap(t, ws, git.CapStatus, map[string]any{"workdir": ".", "timeout_ms": 5000})
	var st struct {
		Branch    string   `json:"branch"`
		Porcelain []string `json:"porcelain"`
		Clean     bool     `json:"clean"`
	}
	if err := json.Unmarshal(out, &st); err != nil {
		t.Fatal(err)
	}
	if st.Branch == "" || !st.Clean || len(st.Porcelain) != 0 {
		t.Fatalf("clean status=%+v", st)
	}

	if err := os.WriteFile(filepath.Join(ws, "dirty.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out = execCap(t, ws, git.CapStatus, map[string]any{"workdir": "."})
	if err := json.Unmarshal(out, &st); err != nil {
		t.Fatal(err)
	}
	if st.Clean || len(st.Porcelain) == 0 {
		t.Fatalf("dirty status=%+v", st)
	}
}

func TestDiffLogAddCommit(t *testing.T) {
	ws := initRepo(t)
	if err := os.WriteFile(filepath.Join(ws, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = execCap(t, ws, git.CapAdd, map[string]any{"paths": []string{"a.txt"}})
	diff := execCap(t, ws, git.CapDiff, map[string]any{"staged": true, "paths": []string{"a.txt"}})
	var d struct {
		Diff string `json:"diff"`
	}
	if err := json.Unmarshal(diff, &d); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(d.Diff, "a.txt") && !strings.Contains(d.Diff, "+one") {
		t.Fatalf("unexpected diff %q", d.Diff)
	}

	commit := execCap(t, ws, git.CapCommit, map[string]any{"message": "add a.txt"})
	var c struct {
		Commit   string `json:"commit"`
		ExitCode int    `json:"exit_code"`
	}
	if err := json.Unmarshal(commit, &c); err != nil {
		t.Fatal(err)
	}
	if c.ExitCode != 0 || len(c.Commit) < 7 {
		t.Fatalf("commit=%+v", c)
	}

	logOut := execCap(t, ws, git.CapLog, map[string]any{"max": 5})
	var lg struct {
		Entries []struct {
			Hash    string `json:"hash"`
			Subject string `json:"subject"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(logOut, &lg); err != nil {
		t.Fatal(err)
	}
	if len(lg.Entries) < 2 || lg.Entries[0].Subject != "add a.txt" {
		t.Fatalf("log=%+v", lg.Entries)
	}
}

func TestPathEscapeRejected(t *testing.T) {
	ws := initRepo(t)
	err := git.ValidateStaticInput(ws, git.CapAdd, json.RawMessage(`{"paths":["../outside"]}`))
	if err == nil {
		t.Fatal("expected escape rejection")
	}
	p := git.New()
	_, err = p.Execute(context.Background(), registry.ExecRequest{
		Capability: git.CapAdd,
		Input:      json.RawMessage(`{"paths":["../outside"]}`),
		Workspace:  ws,
	})
	if err == nil {
		t.Fatal("expected execute rejection")
	}
}

func TestEmptyCommitMessageRejected(t *testing.T) {
	ws := initRepo(t)
	err := git.ValidateStaticInput(ws, git.CapCommit, json.RawMessage(`{"message":"  "}`))
	if err == nil {
		t.Fatal("expected empty message rejection")
	}
}

func TestCommitDisablesHooks(t *testing.T) {
	ws := initRepo(t)
	hooks := filepath.Join(ws, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	hook := filepath.Join(hooks, "pre-commit")
	script := "#!/bin/sh\necho HOOK_RAN > hook.out\nexit 1\n"
	if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "b.txt"), []byte("b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = execCap(t, ws, git.CapAdd, map[string]any{"paths": []string{"b.txt"}})
	_ = execCap(t, ws, git.CapCommit, map[string]any{"message": "no hooks"})
	if _, err := os.Stat(filepath.Join(ws, "hook.out")); err == nil {
		t.Fatal("pre-commit hook ran despite hooksPath=/dev/null")
	}
}

func TestUnknownCapability(t *testing.T) {
	ws := initRepo(t)
	p := git.New()
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: "git.push",
		Input:      json.RawMessage(`{}`),
		Workspace:  ws,
	})
	if err == nil {
		t.Fatal("expected unknown capability")
	}
}

func TestManifestCapabilities(t *testing.T) {
	m := git.New().Manifest()
	if m.Name != "git" {
		t.Fatalf("name=%s", m.Name)
	}
	want := map[string]bool{
		git.CapStatus: false, git.CapDiff: false, git.CapLog: false,
		git.CapAdd: false, git.CapCommit: false,
	}
	for _, c := range m.Capabilities {
		if _, ok := want[c.Name]; !ok {
			t.Fatalf("unexpected %s", c.Name)
		}
		want[c.Name] = true
	}
	for name, ok := range want {
		if !ok {
			t.Fatalf("missing %s", name)
		}
	}
}
