package jstest_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/players/jstest"
)

func fixtureWS(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	app := filepath.Join(ws, "app")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"name":"fixture","private":true,"scripts":{"test":"echo ok"}}`)
	if err := os.WriteFile(filepath.Join(app, "package.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return ws
}

func execWith(t *testing.T, p *jstest.Player, ws string, input any) (json.RawMessage, error) {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return p.Execute(context.Background(), registry.ExecRequest{
		Capability: jstest.CapTest,
		Input:      raw,
		Workspace:  ws,
	})
}

func TestValidateStaticInputRejectsUnknownCapability(t *testing.T) {
	if err := jstest.ValidateStaticInput("/tmp", "yarn.install", json.RawMessage(`{}`)); err == nil {
		t.Fatal("expected unknown capability")
	}
}

func TestValidateStaticInputMissingPackageJSON(t *testing.T) {
	ws := t.TempDir()
	app := filepath.Join(ws, "empty")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := jstest.ValidateStaticInput(ws, jstest.CapTest, json.RawMessage(`{"workdir":"./empty"}`)); err == nil {
		t.Fatal("expected missing package.json")
	}
}

func TestValidateStaticInputWorkdirEscape(t *testing.T) {
	if err := jstest.ValidateStaticInput("/tmp", jstest.CapTest, json.RawMessage(`{"workdir":"../escape"}`)); err == nil {
		t.Fatal("expected escape")
	}
}

func TestValidateStaticInputTimeoutRange(t *testing.T) {
	if err := jstest.ValidateStaticInput("/tmp", jstest.CapTest, json.RawMessage(`{"timeout_ms":1000000}`)); err == nil {
		t.Fatal("timeout out of range")
	}
}

func TestExecuteFakePass(t *testing.T) {
	ws := fixtureWS(t)
	p := jstest.New()
	p.SetExec(func(_ context.Context, _ string, _, args []string) (string, string, int, error) {
		if len(args) != 1 || args[0] != "test" {
			t.Fatalf("unexpected argv %v", args)
		}
		return "yarn test ok\n", "", 0, nil
	})
	raw, err := execWith(t, p, ws, map[string]any{"workdir": "./app"})
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		OK       bool   `json:"ok"`
		ExitCode int    `json:"exit_code"`
		Log      string `json:"log"`
		Script   string `json:"script"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK || out.ExitCode != 0 || out.Script != "echo ok" {
		t.Fatalf("%+v", out)
	}
}

func TestExecuteFakeFail(t *testing.T) {
	ws := fixtureWS(t)
	p := jstest.New()
	p.SetExec(func(_ context.Context, _ string, _, _ []string) (string, string, int, error) {
		return "Tests: 1 failed", "Test failed", 1, nil
	})
	_, err := execWith(t, p, ws, map[string]any{"workdir": "./app"})
	if err == nil {
		t.Fatal("expected player error")
	}
}

func TestExecuteExecError(t *testing.T) {
	ws := fixtureWS(t)
	p := jstest.New()
	p.SetExec(func(_ context.Context, _ string, _, _ []string) (string, string, int, error) {
		return "", "", -1, errors.New("boom")
	})
	if _, err := execWith(t, p, ws, map[string]any{"workdir": "./app"}); err == nil {
		t.Fatal("expected exec error")
	}
}

func TestExecuteMissingBinary(t *testing.T) {
	ws := fixtureWS(t)
	p := jstest.New()
	// Force defaultExec.LookPath to fail by clearing PATH.
	t.Setenv("PATH", "")
	if _, err := execWith(t, p, ws, map[string]any{"workdir": "./app"}); err == nil {
		t.Fatal("expected missing yarn")
	}
}

func TestManifestCapabilities(t *testing.T) {
	m := jstest.New().Manifest()
	if m.Name != "yarn" {
		t.Fatalf("name=%s", m.Name)
	}
	var hasTest bool
	for _, c := range m.Capabilities {
		if c.Name == jstest.CapTest {
			hasTest = true
		}
	}
	if !hasTest {
		t.Fatalf("missing %s", jstest.CapTest)
	}
}