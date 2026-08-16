package pipeline_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/registry"
	pipeplayer "github.com/gspaim/Runtgine/internal/players/pipeline"
)

func TestRepoSearch(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "x.go"), []byte("package p\nfunc Foo() {}\n"), 0o644)
	p := pipeplayer.New()
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: corepipe.CapRepoSearch,
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Paths   []string `json:"paths"`
		Symbols []string `json:"symbols"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Paths) == 0 {
		t.Fatal("expected paths")
	}
}
