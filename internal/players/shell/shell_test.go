package shell_test

import (
	"context"
	"encoding/json"
	"path/filepath"
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
