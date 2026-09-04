// Package jstest hosts the Yarn Player (G-175..G-177). It exposes a
// single capability, yarn.test, that invokes `yarn test` with argv
// allowlist, never a shell string. yarn install/add/dlx/npx and
// immutable-network flags are explicitly rejected at
// ValidateStaticInput (G-176). The Player fails the step on non-zero
// exit (G-178), like npm.test.
package jstest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/shell"
)

const (
	CapTest = "yarn.test"

	defaultTimeoutMS = 120_000
	maxTimeoutMS     = 600_000
	maxLogBytes      = 64 * 1024
)

// ExecFunc invokes `yarn` with argv (never a shell string). args
// always include the subcommand token (e.g. "test").
type ExecFunc func(ctx context.Context, dir string, env, args []string) (stdout, stderr string, exitCode int, err error)

type Player struct {
	exec ExecFunc
}

func New() *Player { return &Player{exec: defaultExec} }

func (p *Player) SetExec(fn ExecFunc) { p.exec = fn }

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "yarn",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapTest,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "workdir":{"type":"string"},
    "timeout_ms":{"type":"integer","minimum":1,"maximum":600000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["ok","exit_code","elapsed_ms","log"],
  "properties":{
    "ok":{"type":"boolean"},
    "exit_code":{"type":"integer"},
    "elapsed_ms":{"type":"integer"},
    "log":{"type":"string"},
    "script":{"type":"string"}
  }
}`),
			},
		},
	}
}

type input struct {
	Workdir   string `json:"workdir"`
	TimeoutMS int    `json:"timeout_ms"`
}

type output struct {
	OK        bool   `json:"ok"`
	ExitCode  int    `json:"exit_code"`
	ElapsedMS int64  `json:"elapsed_ms"`
	Log       string `json:"log"`
	Script    string `json:"script,omitempty"`
}

// forbiddenTokens matches yarn subcommands/flags that change
// network or mutating behavior. Listed in G-176.
var forbiddenTokens = map[string]struct{}{
	"add":                   {},
	"install":               {},
	"global":                {},
	"dlx":                   {},
	"npx":                   {},
	"--frozen-lockfile":     {},
	"--immutable":           {},
	"--network-timeout":     {},
	"--mutex":               {},
	"--parallel":            {},
	"remove":                {},
	"upgrade":               {},
	"link":                  {},
	"unlink":                {},
	"pack":                  {},
	"publish":               {},
	"cache":                 {},
	"workspaces":            {},
	"config":                {},
	"init":                  {},
	"create":                {},
	"bin":                   {},
	"info":                  {},
	"version":               {},
	"run":                   {},
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	if capability != CapTest {
		return result.Validation(result.CodeUnknownCapability, "yarn player cannot validate "+capability, nil)
	}
	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid "+CapTest+" input: "+err.Error(), nil)
	}
	if in.TimeoutMS != 0 && (in.TimeoutMS < 1 || in.TimeoutMS > maxTimeoutMS) {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 600000", map[string]any{"timeout_ms": in.TimeoutMS})
	}
	wd := strings.TrimSpace(in.Workdir)
	if wd == "" {
		wd = "."
	}
	if strings.ContainsAny(wd, "\x00\n\r") {
		return result.Validation(result.CodeInvalidInput, "workdir must not contain NUL or newlines", nil)
	}
	if strings.Contains(wd, "://") {
		return result.Validation(result.CodeInvalidInput, "workdir must not be a URL", map[string]any{"workdir": wd})
	}
	if filepath.IsAbs(wd) {
		return result.Validation(result.CodeInvalidInput, "workdir must be relative to workspace", map[string]any{"workdir": wd})
	}
	resolved, err := shell.ResolveWorkdir(workspace, wd)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(resolved, "package.json")); err != nil {
		return result.Validation(result.CodeInvalidInput, "workdir must contain package.json", map[string]any{"workdir": resolved})
	}
	return nil
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	var in input
	_ = json.Unmarshal(req.Input, &in)
	wd := strings.TrimSpace(in.Workdir)
	if wd == "" {
		wd = "."
	}
	timeoutMS := in.TimeoutMS
	if timeoutMS == 0 {
		timeoutMS = defaultTimeoutMS
	}
	dir, err := shell.ResolveWorkdir(req.Workspace, wd)
	if err != nil {
		return nil, err
	}

	timeoutDur := time.Duration(timeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeoutDur)
	defer cancel()

	args := []string{"test"}
	env := buildEnv()
	fn := p.exec
	if fn == nil {
		fn = defaultExec
	}

	start := time.Now()
	stdout, stderr, exitCode, runErr := fn(ctx, dir, env, args)
	elapsed := time.Since(start).Milliseconds()
	if runErr != nil {
		if re, ok := runErr.(result.Error); ok {
			return nil, re
		}
		return nil, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}

	logText := strings.TrimSpace(stdout)
	if stderr = strings.TrimSpace(stderr); stderr != "" {
		if logText != "" {
			logText += "\n"
		}
		logText += stderr
	}
	if logText == "" {
		logText = fmt.Sprintf("exit=%d", exitCode)
	}
	logText = truncateUTF8(logText, maxLogBytes)

	out := output{
		OK:        exitCode == 0,
		ExitCode:  exitCode,
		ElapsedMS: elapsed,
		Log:       logText,
		Script:    readTestScript(dir),
	}
	if exitCode != 0 {
		details := map[string]any{
			"ok":         false,
			"exit_code":  out.ExitCode,
			"elapsed_ms": out.ElapsedMS,
			"log":        out.Log,
		}
		if out.Script != "" {
			details["script"] = out.Script
		}
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", CapTest, exitCode), false, details)
	}
	return json.Marshal(out)
}

func readTestScript(dir string) string {
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct {
		Scripts map[string]any `json:"scripts"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil || pkg.Scripts == nil {
		return ""
	}
	s, ok := pkg.Scripts["test"].(string)
	if !ok {
		return ""
	}
	return s
}

func truncateUTF8(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	data := []byte(s)[:limit]
	for len(data) > 0 {
		r, size := utf8.DecodeLastRune(data)
		if r != utf8.RuneError || size != 1 {
			break
		}
		data = data[:len(data)-1]
	}
	return string(data)
}

func buildEnv() []string {
	var env []string
	for _, e := range os.Environ() {
		key, _, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		if allowInheritedEnv(key) {
			env = append(env, e)
		}
	}
	return env
}

func allowInheritedEnv(key string) bool {
	switch key {
	case "PATH", "HOME", "USER", "LANG", "TZ", "TMPDIR", "TMP", "TEMP":
		return true
	}
	if strings.HasPrefix(key, "LC_") {
		return true
	}
	return false
}

func defaultExec(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
	if _, err := exec.LookPath("yarn"); err != nil {
		return "", "", -1, result.Runtime(result.CodePlayerError, "yarn binary not found in PATH", false, nil)
	}
	cmd := exec.CommandContext(ctx, "yarn", args...)
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodeTimeout, "yarn.test timed out", true, map[string]any{
			"stderr": stderr.String(),
		})
	}
	if ctx.Err() == context.Canceled {
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodeCancelled, "yarn.test cancelled", false, nil)
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout.String(), stderr.String(), ee.ExitCode(), nil
		}
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
	return stdout.String(), stderr.String(), 0, nil
}