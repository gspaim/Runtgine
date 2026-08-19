package gotest_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/gotest"
)

func execCap(t *testing.T, p *gotest.Player, input any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: gotest.CapGo,
		Input:      raw,
		Workspace:  t.TempDir(),
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return out
}

func execErr(t *testing.T, p *gotest.Player, workspace string, input any) error {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Execute(context.Background(), registry.ExecRequest{
		Capability: gotest.CapGo,
		Input:      raw,
		Workspace:  workspace,
	})
	return err
}

func jsonLine(action, test, output string) string {
	ev := map[string]any{"Action": action}
	if test != "" {
		ev["Test"] = test
	}
	if output != "" {
		ev["Output"] = output
	}
	b, _ := json.Marshal(ev)
	return string(b) + "\n"
}

func TestManifestCapability(t *testing.T) {
	m := gotest.New().Manifest()
	if m.Name != "test" || m.Kind != registry.KindDeterministic {
		t.Fatalf("manifest=%+v", m)
	}
	found := false
	for _, c := range m.Capabilities {
		if c.Name == gotest.CapGo {
			found = true
		}
		if c.Name == "test.python" {
			t.Fatal("test.python must not be registered")
		}
	}
	if !found {
		t.Fatal("missing test.go")
	}
}

func TestFakePass(t *testing.T) {
	p := gotest.New()
	var sawArgs []string
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		sawArgs = append([]string(nil), args...)
		stdout := jsonLine("output", "TestHello", "=== RUN   TestHello\n") +
			jsonLine("pass", "TestHello", "") +
			jsonLine("pass", "", "")
		return stdout, "", 0, nil
	})
	out := execCap(t, p, map[string]any{"packages": []string{"./internal/players/gotest"}})
	var got struct {
		OK       bool   `json:"ok"`
		Pass     int    `json:"pass"`
		Fail     int    `json:"fail"`
		Skip     int    `json:"skip"`
		ExitCode int    `json:"exit_code"`
		Log      string `json:"log"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if !got.OK || got.Pass < 1 || got.Fail != 0 || got.ExitCode != 0 {
		t.Fatalf("out=%+v", got)
	}
	if !strings.Contains(got.Log, "TestHello") {
		t.Fatalf("log=%q", got.Log)
	}
	joined := strings.Join(sawArgs, " ")
	if sawArgs[0] != "test" || !containsAll(sawArgs, "-json", "-mod=readonly", "-count", "-timeout") {
		t.Fatalf("argv=%v", sawArgs)
	}
	for _, denied := range []string{"-race", "-exec", "-fuzz", "-coverprofile"} {
		if strings.Contains(joined, denied) {
			t.Fatalf("denied flag in argv: %v", sawArgs)
		}
	}
}

func TestFakeFail(t *testing.T) {
	p := gotest.New()
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		stdout := jsonLine("output", "TestBoom", "--- FAIL: TestBoom\n") +
			jsonLine("fail", "TestBoom", "")
		return stdout, "", 1, nil
	})
	err := execErr(t, p, t.TempDir(), map[string]any{"packages": []string{"./"}})
	if err == nil {
		t.Fatal("expected player error")
	}
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodePlayerError {
		t.Fatalf("got %v", err)
	}
	fail, _ := ve.Details["fail"].(int)
	okVal, _ := ve.Details["ok"].(bool)
	if okVal || fail < 1 {
		t.Fatalf("details=%v", ve.Details)
	}
	log, _ := ve.Details["log"].(string)
	if !strings.Contains(log, "FAIL") {
		t.Fatalf("log=%q", log)
	}
}

func TestEscapePackageRejected(t *testing.T) {
	ws := t.TempDir()
	err := gotest.ValidateStaticInput(ws, gotest.CapGo, json.RawMessage(`{"packages":["../outside"]}`))
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("escape: %v", err)
	}
	abs := filepath.Join(ws, "pkg")
	err = gotest.ValidateStaticInput(ws, gotest.CapGo, json.RawMessage(`{"packages":[`+jsonString(abs)+`]}`))
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("abs: %v", err)
	}
	err = gotest.ValidateStaticInput(ws, gotest.CapGo, json.RawMessage(`{"packages":["https://example.com/mod"]}`))
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("url: %v", err)
	}
}

func TestRunAndCountAndShortArgv(t *testing.T) {
	p := gotest.New()
	var saw []string
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return jsonLine("pass", "TestShort", ""), "", 0, nil
	})
	_ = execCap(t, p, map[string]any{
		"packages":   []string{"./internal/players/gotest"},
		"short":      true,
		"count":      2,
		"run":        "TestShort",
		"timeout_ms": 5000,
	})
	if !containsAll(saw, "-short", "-count", "2", "-run", "TestShort", "-timeout", "5s") {
		t.Fatalf("argv=%v", saw)
	}
}

func TestEnvDoesNotInheritGOFLAGS(t *testing.T) {
	t.Setenv("GOFLAGS", "-race")
	t.Setenv("RUNTGINE_TOKEN", "secret")
	p := gotest.New()
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		for _, e := range env {
			if strings.HasPrefix(e, "GOFLAGS=") || strings.HasPrefix(e, "RUNTGINE_") {
				t.Fatalf("forbidden env %q", e)
			}
		}
		return jsonLine("pass", "TestEnv", ""), "", 0, nil
	})
	_ = execCap(t, p, map[string]any{"packages": []string{"./..."}})
}

func TestInvalidRunRejected(t *testing.T) {
	err := gotest.ValidateStaticInput(t.TempDir(), gotest.CapGo, json.RawMessage("{\"run\":\"a\\nb\"}"))
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("got %v", err)
	}
}

func TestTimeoutFromRunner(t *testing.T) {
	p := gotest.New()
	p.SetExec(func(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
		return "", "", -1, result.Runtime(result.CodeTimeout, "test.go timed out", true, nil)
	})
	err := execErr(t, p, t.TempDir(), map[string]any{"packages": []string{"./..."}})
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeTimeout {
		t.Fatalf("got %v", err)
	}
}

func containsAll(args []string, want ...string) bool {
	for _, w := range want {
		found := false
		for _, a := range args {
			if a == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
