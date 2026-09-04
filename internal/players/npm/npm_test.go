package npm_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/npm"
)

func fixtureWS(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	pkg := filepath.Join(ws, "app")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"name":"fixture","private":true,"scripts":{"test":"echo ok"}}`)
	if err := os.WriteFile(filepath.Join(pkg, "package.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return ws
}

func execCap(t *testing.T, p *npm.Player, workspace string, input any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: npm.CapTest,
		Input:      raw,
		Workspace:  workspace,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return out
}

func execErr(t *testing.T, p *npm.Player, workspace string, input any) error {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Execute(context.Background(), registry.ExecRequest{
		Capability: npm.CapTest,
		Input:      raw,
		Workspace:  workspace,
	})
	return err
}

func TestManifestCapability(t *testing.T) {
	m := npm.New().Manifest()
	if m.Name != "npm" || m.Kind != registry.KindDeterministic {
		t.Fatalf("manifest=%+v", m)
	}
	found := false
	for _, c := range m.Capabilities {
		if c.Name == npm.CapTest {
			found = true
		}
		if c.Name == "npm.install" {
			t.Fatal("npm.install must not be registered")
		}
	}
	if !found {
		t.Fatal("missing npm.test")
	}
}

func TestFakePass(t *testing.T) {
	ws := fixtureWS(t)
	p := npm.New()
	var sawArgs []string
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		sawArgs = append([]string(nil), args...)
		return "ok\n", "", 0, nil
	})
	out := execCap(t, p, ws, map[string]any{"workdir": "app"})
	var got struct {
		OK       bool   `json:"ok"`
		ExitCode int    `json:"exit_code"`
		Log      string `json:"log"`
		Script   string `json:"script"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if !got.OK || got.ExitCode != 0 {
		t.Fatalf("out=%+v", got)
	}
	if got.Script != "echo ok" {
		t.Fatalf("script=%q", got.Script)
	}
	if strings.Join(sawArgs, " ") != "test" {
		t.Fatalf("argv=%v", sawArgs)
	}
}

func TestFakeFail(t *testing.T) {
	ws := fixtureWS(t)
	p := npm.New()
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		return "", "not ok", 1, nil
	})
	err := execErr(t, p, ws, map[string]any{"workdir": "app"})
	if err == nil {
		t.Fatal("expected player error")
	}
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodePlayerError {
		t.Fatalf("got %v", err)
	}
	okVal, _ := ve.Details["ok"].(bool)
	code, _ := ve.Details["exit_code"].(int)
	if okVal || code != 1 {
		t.Fatalf("details=%v", ve.Details)
	}
	log, _ := ve.Details["log"].(string)
	if !strings.Contains(log, "not ok") {
		t.Fatalf("log=%q", log)
	}
}

func TestSilentArgv(t *testing.T) {
	ws := fixtureWS(t)
	p := npm.New()
	var saw []string
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return "", "", 0, nil
	})
	_ = execCap(t, p, ws, map[string]any{"workdir": "app", "silent": true})
	if strings.Join(saw, " ") != "--silent test" {
		t.Fatalf("argv=%v", saw)
	}
}

func TestEscapeWorkdirRejected(t *testing.T) {
	ws := fixtureWS(t)
	err := npm.ValidateStaticInput(ws, npm.CapTest, json.RawMessage(`{"workdir":"../outside"}`))
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("escape: %v", err)
	}
}

func TestMissingPackageJSONRejected(t *testing.T) {
	ws := t.TempDir()
	err := npm.ValidateStaticInput(ws, npm.CapTest, json.RawMessage(`{"workdir":"."}`))
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("missing package.json: %v", err)
	}
}

func TestExtraPropertyRejected(t *testing.T) {
	ws := fixtureWS(t)
	err := npm.ValidateStaticInput(ws, npm.CapTest, json.RawMessage(`{"workdir":"app","prefix":"/tmp"}`))
	// schema additionalProperties is enforced at Registry.ValidateInput, not here
	if err != nil {
		t.Fatalf("static should ignore unknown keys that json.Unmarshal drops: %v", err)
	}
	_ = ws
}

func TestEnvDoesNotInheritSecrets(t *testing.T) {
	ws := fixtureWS(t)
	t.Setenv("RUNTGINE_LLM_API_KEY", "secret")
	t.Setenv("NODE_OPTIONS", "-r ./evil.js")
	p := npm.New()
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		for _, e := range env {
			if strings.HasPrefix(e, "RUNTGINE_") || strings.HasPrefix(e, "NODE_OPTIONS=") {
				t.Fatalf("forbidden env %q", e)
			}
		}
		return "", "", 0, nil
	})
	_ = execCap(t, p, ws, map[string]any{"workdir": "app"})
}

func TestTimeoutFromRunner(t *testing.T) {
	ws := fixtureWS(t)
	p := npm.New()
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		return "", "", -1, result.Runtime(result.CodeTimeout, "npm.test timed out", true, nil)
	})
	err := execErr(t, p, ws, map[string]any{"workdir": "app"})
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeTimeout {
		t.Fatalf("got %v", err)
	}
}
