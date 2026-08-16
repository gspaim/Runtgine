package shell_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/players/shell"
)

func TestShellExecEcho(t *testing.T) {
	p := shell.New()
	ws := t.TempDir()
	in, _ := json.Marshal(map[string]any{
		"command":    []string{"echo", "hi"},
		"workdir":    ".",
		"timeout_ms": 5000,
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: shell.CapExec,
		Input:      in,
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if int(got["exit_code"].(float64)) != 0 {
		t.Fatalf("exit=%v", got["exit_code"])
	}
}

func TestWorkdirOutsideWorkspaceRejected(t *testing.T) {
	p := shell.New()
	ws := t.TempDir()
	outside := filepath.Dir(ws)
	in, _ := json.Marshal(map[string]any{
		"command": []string{"echo", "x"},
		"workdir": outside,
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: shell.CapExec,
		Input:      in,
		Workspace:  ws,
	})
	if err == nil {
		t.Fatal("expected sandbox error")
	}
}

func TestWorkdirSymlinkEscapeRejected(t *testing.T) {
	ws := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(ws, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	in, _ := json.Marshal(map[string]any{
		"command": []string{"echo", "x"},
		"workdir": "escape",
	})
	if err := shell.ValidateStaticInput(ws, in); err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}

func TestOmittedEnvDoesNotInheritTokens(t *testing.T) {
	p := shell.New()
	ws := t.TempDir()
	const secret = "secret-should-not-leak"
	t.Setenv("GITHUB_TOKEN", secret)
	t.Setenv("RUNTGINE_LLM_API_KEY", secret)
	in, _ := json.Marshal(map[string]any{
		"command":    []string{"env"},
		"workdir":    ".",
		"timeout_ms": 5000,
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: shell.CapExec,
		Input:      in,
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	stdout, _ := got["stdout"].(string)
	if strings.Contains(stdout, secret) {
		t.Fatal("child inherited secret from parent env")
	}
	if !strings.Contains(stdout, "PATH=") {
		t.Fatal("expected PATH in minimal inherited env")
	}
}

func TestExplicitEnvOnly(t *testing.T) {
	p := shell.New()
	ws := t.TempDir()
	t.Setenv("GITHUB_TOKEN", "secret-should-not-leak")
	in, _ := json.Marshal(map[string]any{
		"command":    []string{"env"},
		"workdir":    ".",
		"timeout_ms": 5000,
		"env":        map[string]string{"FOO": "bar"},
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: shell.CapExec,
		Input:      in,
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	stdout, _ := got["stdout"].(string)
	if !strings.Contains(stdout, "FOO=bar") {
		t.Fatalf("expected FOO=bar, got %q", stdout)
	}
	if strings.Contains(stdout, "GITHUB_TOKEN=") {
		t.Fatal("explicit env must not inherit host secrets")
	}
	if strings.Contains(stdout, "PATH=") {
		t.Fatal("explicit env should not inherit PATH")
	}
}
