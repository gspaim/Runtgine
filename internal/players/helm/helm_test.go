package helm_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/helm"
)

func newChartWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	dir := filepath.Join(ws, "charts", "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("apiVersion: v2\nname: demo\nversion: 0.1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return ws
}

func asResultErr(t *testing.T, err error) result.Error {
	t.Helper()
	e, ok := err.(result.Error)
	if !ok {
		t.Fatalf("err is not result.Error: %v", err)
	}
	return e
}

func TestManifestNoInstall(t *testing.T) {
	m := helm.New().Manifest()
	if m.Name != "helm" {
		t.Fatalf("name=%s", m.Name)
	}
	found := 0
	for _, c := range m.Capabilities {
		switch c.Name {
		case "helm.install", "helm.upgrade", "helm.rollback", "helm.uninstall":
			t.Fatalf("%s must not be registered", c.Name)
		case helm.CapLint, helm.CapTemplate, helm.CapList, helm.CapStatus:
			found++
		}
	}
	if found != 4 {
		t.Fatalf("caps=%v", m.Capabilities)
	}
}

func TestUnknownCapability(t *testing.T) {
	err := helm.ValidateStaticInput(t.TempDir(), "helm.install", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeUnknownCapability {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestRejectsFlagRelease(t *testing.T) {
	err := helm.ValidateStaticInput(t.TempDir(), helm.CapStatus, json.RawMessage(`{"release":"--kubeconfig"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestChartRequiresMarker(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "charts", "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := helm.ValidateStaticInput(ws, helm.CapLint, json.RawMessage(`{"chart":"charts/demo"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestChartEscapeWorkspace(t *testing.T) {
	err := helm.ValidateStaticInput(t.TempDir(), helm.CapLint, json.RawMessage(`{"chart":"../outside"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestRejectsAbsChart(t *testing.T) {
	err := helm.ValidateStaticInput(t.TempDir(), helm.CapLint, json.RawMessage(`{"chart":"/etc"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
}

func TestFakeLint(t *testing.T) {
	ws := newChartWorkspace(t)
	p := helm.New()
	var sawDir string
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
		sawDir = dir
		saw = append([]string(nil), args...)
		return "==> Linting chart .\n1 chart(s) linted, 0 chart(s) failed", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: helm.CapLint,
		Input:      json.RawMessage(`{"chart":"charts/demo"}`),
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sawDir != ws {
		t.Fatalf("dir=%s want %s", sawDir, ws)
	}
	if len(saw) != 2 || saw[0] != "lint" || saw[1] != "charts/demo" {
		t.Fatalf("args=%v", saw)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["ok"] != true {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeLintFailureExits(t *testing.T) {
	ws := newChartWorkspace(t)
	p := helm.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
		return "", "Error: 1 chart(s) linted, 1 chart(s) failed", 1, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: helm.CapLint,
		Input:      json.RawMessage(`{"chart":"charts/demo"}`),
		Workspace:  ws,
	})
	if ve := asResultErr(t, err); ve.Code != result.CodePlayerError {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestFakeTemplate(t *testing.T) {
	ws := newChartWorkspace(t)
	p := helm.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return "apiVersion: v1\nkind: ConfigMap\n", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: helm.CapTemplate,
		Input:      json.RawMessage(`{"chart":"charts/demo","release":"web","namespace":"demo"}`),
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"template", "web", "charts/demo", "-n", "demo"}
	if len(saw) != len(want) {
		t.Fatalf("args=%v", saw)
	}
	for i, v := range want {
		if saw[i] != v {
			t.Fatalf("args=%v", saw)
		}
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["output"] != "apiVersion: v1\nkind: ConfigMap\n" || got["truncated"] != false {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeListJSON(t *testing.T) {
	ws := newChartWorkspace(t)
	p := helm.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `[{"name":"web","revision":1}]`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: helm.CapList,
		Input:      json.RawMessage(`{}`),
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	n := len(saw)
	if n < 2 || saw[n-2] != "-o" || saw[n-1] != "json" {
		t.Fatalf("args=%v", saw)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	releases, ok := got["releases"].([]any)
	if !ok || len(releases) != 1 {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeListBadJSON(t *testing.T) {
	ws := newChartWorkspace(t)
	p := helm.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
		return "not json", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: helm.CapList,
		Input:      json.RawMessage(`{}`),
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["releases"] != "not json" || got["truncated"] != true {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeStatusFail(t *testing.T) {
	ws := newChartWorkspace(t)
	p := helm.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
		return "", "Error: release: not found", 1, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: helm.CapStatus,
		Input:      json.RawMessage(`{"release":"web"}`),
		Workspace:  ws,
	})
	if ve := asResultErr(t, err); ve.Code != result.CodePlayerError {
		t.Fatalf("code=%s", ve.Code)
	}
}
