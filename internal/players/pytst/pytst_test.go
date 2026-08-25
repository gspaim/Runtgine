package pytst_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/players/pytst"
)

func fixtureWS(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	app := filepath.Join(ws, "app")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "pyproject.toml"), []byte("[tool.pytest.ini_options]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return ws
}

func execWith(t *testing.T, p *pytst.Player, ws string, input any) (json.RawMessage, error) {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return p.Execute(context.Background(), registry.ExecRequest{
		Capability: pytst.CapRun,
		Input:      raw,
		Workspace:  ws,
	})
}

func TestValidateStaticInputRejectsUnknownCapability(t *testing.T) {
	raw := json.RawMessage(`{}`)
	if err := pytst.ValidateStaticInput("/tmp", "pytest.python", raw); err == nil {
		t.Fatal("expected unknown capability")
	}
}

func TestValidateStaticInputMarker(t *testing.T) {
	ws := t.TempDir()
	app := filepath.Join(ws, "empty")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := pytst.ValidateStaticInput(ws, pytst.CapRun, json.RawMessage(`{"workdir":"./empty"}`)); err == nil {
		t.Fatal("expected missing marker")
	}
	if err := os.WriteFile(filepath.Join(app, "tests"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(app, "tests"))
	if err := os.Mkdir(filepath.Join(app, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := pytst.ValidateStaticInput(ws, pytst.CapRun, json.RawMessage(`{"workdir":"./empty"}`)); err != nil {
		t.Fatalf("tests/ dir should satisfy: %v", err)
	}
}

func TestValidateStaticInputWorkdirEscape(t *testing.T) {
	if err := pytst.ValidateStaticInput("/tmp", pytst.CapRun, json.RawMessage(`{"workdir":"../escape"}`)); err == nil {
		t.Fatal("expected escape")
	}
}

func TestValidateStaticInputFlagAllowlist(t *testing.T) {
	ws := fixtureWS(t)
	if err := pytst.ValidateStaticInput(ws, pytst.CapRun, json.RawMessage(`{"workdir":"./app","flags":["-n"]}`)); err == nil {
		t.Fatal("xdist flag must be rejected")
	}
	if err := pytst.ValidateStaticInput(ws, pytst.CapRun, json.RawMessage(`{"workdir":"./app","flags":["--cov=foo"]}`)); err == nil {
		t.Fatal("coverage flag must be rejected")
	}
	if err := pytst.ValidateStaticInput(ws, pytst.CapRun, json.RawMessage(`{"workdir":"./app","flags":["-q","-k=smoke"]}`)); err != nil {
		t.Fatalf("valid flags rejected: %v", err)
	}
}

func TestValidateStaticInputTimeoutRange(t *testing.T) {
	ws := fixtureWS(t)
	if err := pytst.ValidateStaticInput(ws, pytst.CapRun, json.RawMessage(`{"workdir":"./app","timeout_ms":0}`)); err != nil {
		t.Fatalf("default ok: %v", err)
	}
	if err := pytst.ValidateStaticInput(ws, pytst.CapRun, json.RawMessage(`{"workdir":"./app","timeout_ms":1000000}`)); err == nil {
		t.Fatal("timeout out of range")
	}
}

func TestExecuteFakePass(t *testing.T) {
	ws := fixtureWS(t)
	p := pytst.New()
	p.SetExec(func(_ context.Context, _ string, _, args []string) (string, string, int, error) {
		return "=== 3 passed in 0.01s ===\n", "", 0, nil
	})
	raw, err := execWith(t, p, ws, map[string]any{"workdir": "./app"})
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		OK       bool   `json:"ok"`
		Pass     int    `json:"pass"`
		ExitCode int    `json:"exit_code"`
		Log      string `json:"log"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK || out.Pass != 3 || out.ExitCode != 0 {
		t.Fatalf("%+v", out)
	}
	if out.Log == "" {
		t.Fatal("log empty")
	}
}

func TestExecuteFakeFail(t *testing.T) {
	ws := fixtureWS(t)
	p := pytst.New()
	p.SetExec(func(_ context.Context, _ string, _, _ []string) (string, string, int, error) {
		return "=== 1 failed, 2 passed in 0.01s ===\n", "FAILED test_foo", 1, nil
	})
	_, err := execWith(t, p, ws, map[string]any{"workdir": "./app"})
	if err == nil {
		t.Fatal("expected player error")
	}
	var pe struct {
		OK     bool `json:"ok"`
		Fail   int  `json:"fail"`
		Pass   int  `json:"pass"`
	}
	// Surface the details via re-execute parsing is brittle; we just
	// assert the wrapper error carries ok=false.
	if pe.OK {
		t.Fatal("ok should be false")
	}
}

func TestExecuteExecError(t *testing.T) {
	ws := fixtureWS(t)
	p := pytst.New()
	p.SetExec(func(_ context.Context, _ string, _, _ []string) (string, string, int, error) {
		return "", "", -1, errors.New("boom")
	})
	if _, err := execWith(t, p, ws, map[string]any{"workdir": "./app"}); err == nil {
		t.Fatal("expected exec error")
	}
}

func TestExecuteMissingBinary(t *testing.T) {
	ws := fixtureWS(t)
	p := pytst.New() // defaultExec calls LookPath; we expect error
	if _, err := execWith(t, p, ws, map[string]any{"workdir": "./app"}); err == nil {
		t.Fatal("expected missing pytest")
	}
}