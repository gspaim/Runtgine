// Package pytst hosts the Pytest Player (G-172..G-174). It exposes a
// single capability, pytest.run, that invokes the `pytest` binary
// directly with an argv allowlist, never a shell string. The
// Player fails the step when the runner returns a non-zero exit
// code (G-178), exactly like test.go / npm.test.
package pytst

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/shell"
)

const (
	CapRun = "pytest.run"

	defaultTimeoutMS = 120_000
	maxTimeoutMS     = 600_000
	defaultLimit     = 5
	maxLimit         = 20
	maxLogBytes      = 64 * 1024
	maxPackageRunes  = 256
)

// ExecFunc invokes `pytest` with argv (never a shell string). args
// start at the user-provided flags and packages.
type ExecFunc func(ctx context.Context, dir string, env, args []string) (stdout, stderr string, exitCode int, err error)

type Player struct {
	exec ExecFunc
}

func New() *Player { return &Player{exec: defaultExec} }

func (p *Player) SetExec(fn ExecFunc) { p.exec = fn }

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "pytest",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapRun,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "workdir":{"type":"string"},
    "packages":{"type":"array","items":{"type":"string","minLength":1}},
    "flags":{"type":"array","items":{"type":"string","minLength":1}},
    "timeout_ms":{"type":"integer","minimum":1,"maximum":600000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["ok","exit_code","elapsed_ms","log"],
  "properties":{
    "ok":{"type":"boolean"},
    "pass":{"type":"integer"},
    "fail":{"type":"integer"},
    "skip":{"type":"integer"},
    "elapsed_ms":{"type":"integer"},
    "exit_code":{"type":"integer"},
    "log":{"type":"string"}
  }
}`),
			},
		},
	}
}

type input struct {
	Workdir   string   `json:"workdir"`
	Packages  []string `json:"packages"`
	Flags     []string `json:"flags"`
	TimeoutMS int      `json:"timeout_ms"`
}

type output struct {
	OK        bool   `json:"ok"`
	Pass      int    `json:"pass"`
	Fail      int    `json:"fail"`
	Skip      int    `json:"skip"`
	ElapsedMS int64  `json:"elapsed_ms"`
	ExitCode  int    `json:"exit_code"`
	Log       string `json:"log"`
}

// ValidateStaticInput enforces ranges, marker files, and the flag
// allowlist before Execute.
func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	if capability != CapRun {
		return result.Validation(result.CodeUnknownCapability, "pytest player cannot validate "+capability, nil)
	}
	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid "+CapRun+" input: "+err.Error(), nil)
	}
	if in.TimeoutMS != 0 && (in.TimeoutMS < 1 || in.TimeoutMS > maxTimeoutMS) {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 600000", map[string]any{"timeout_ms": in.TimeoutMS})
	}
	for _, f := range in.Flags {
		if err := allowFlag(f); err != nil {
			return err
		}
	}
	if len(in.Packages) > 0 {
		for _, pkg := range in.Packages {
			if err := validatePackage(pkg); err != nil {
				return err
			}
		}
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
	if !hasPytestMarker(resolved) {
		return result.Validation(result.CodeInvalidInput, "workdir must contain pyproject.toml, pytest.ini, or tests/", map[string]any{"workdir": resolved})
	}
	return nil
}

// allowFlag matches the pytest flag allowlist (G-173). Pass-through
// for package-shaped values (-k REGEX, -m MARK).
func allowFlag(flag string) error {
	if flag == "" {
		return result.Validation(result.CodeInvalidInput, "flag must be non-empty", nil)
	}
	if strings.ContainsAny(flag, "\x00\n\r") {
		return result.Validation(result.CodeInvalidInput, "flag must not contain NUL or newlines", nil)
	}
	if utf8RuneCount(flag) > maxPackageRunes {
		return result.Validation(result.CodeInvalidInput, "flag exceeds 256 runes", nil)
	}
	switch {
	case flag == "-q",
		flag == "-x",
		flag == "--tb=short",
		flag == "--tb=line",
		flag == "--tb=no",
		flag == "--no-header",
		flag == "--color=no",
		strings.HasPrefix(flag, "-k="),
		strings.HasPrefix(flag, "-m="),
		strings.HasPrefix(flag, "--tb="):
		return nil
	}
	if strings.HasPrefix(flag, "-k") || strings.HasPrefix(flag, "-m") {
		// require a non-empty argument
		rest := strings.TrimPrefix(strings.TrimPrefix(flag, "-k"), "-m")
		if rest == "" {
			return result.Validation(result.CodeInvalidInput, "flag requires an argument", map[string]any{"flag": flag})
		}
		return nil
	}
	return result.Validation(result.CodeInvalidInput, "flag not in allowlist", map[string]any{"flag": flag})
}

func validatePackage(pkg string) error {
	if strings.TrimSpace(pkg) == "" {
		return result.Validation(result.CodeInvalidInput, "package must be non-empty", nil)
	}
	if strings.ContainsAny(pkg, "\x00\n\r") {
		return result.Validation(result.CodeInvalidInput, "package must not contain NUL or newlines", map[string]any{"package": pkg})
	}
	if strings.HasPrefix(pkg, "-") {
		return result.Validation(result.CodeInvalidInput, "package must not start with -", map[string]any{"package": pkg})
	}
	if utf8RuneCount(pkg) > maxPackageRunes {
		return result.Validation(result.CodeInvalidInput, "package exceeds 256 runes", map[string]any{"package": pkg})
	}
	return nil
}

func hasPytestMarker(dir string) bool {
	for _, name := range []string{"pyproject.toml", "pytest.ini", "tox.ini", "conftest.py"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	if info, err := os.Stat(filepath.Join(dir, "tests")); err == nil && info.IsDir() {
		return true
	}
	return false
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

	args := []string{}
	args = append(args, in.Flags...)
	if len(in.Packages) > 0 {
		args = append(args, in.Packages...)
	} else {
		args = append(args, ".")
	}
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

	pass, fail, skip := parsePytestSummary(stdout)
	out := output{
		OK:        exitCode == 0,
		Pass:      pass,
		Fail:      fail,
		Skip:      skip,
		ElapsedMS: elapsed,
		ExitCode:  exitCode,
		Log:       logText,
	}
	if exitCode != 0 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", CapRun, exitCode), false, map[string]any{
			"ok":         false,
			"pass":       out.Pass,
			"fail":       out.Fail,
			"skip":       out.Skip,
			"elapsed_ms": out.ElapsedMS,
			"exit_code":  out.ExitCode,
			"log":        out.Log,
		})
	}
	return json.Marshal(out)
}

// parsePytestSummary pulls the standard pytest summary line
// "=== X passed, Y failed, Z skipped in Ns ===" out of the output.
// Returns zeros when the line is absent.
func parsePytestSummary(stdout string) (pass, fail, skip int) {
	lines := strings.Split(stdout, "\n")
	var lastSummary string
	for _, ln := range lines {
		if strings.Contains(ln, "===") && strings.Contains(ln, "passed") {
			lastSummary = ln
		}
	}
	if lastSummary == "" {
		return 0, 0, 0
	}
	pass = parseIntBefore(lastSummary, "passed")
	fail = parseIntBefore(lastSummary, "failed")
	skip = parseIntBefore(lastSummary, "skipped")
	return pass, fail, skip
}

// parseIntBefore scans s for "<N> <word>" where word is the marker
// and returns N. Returns 0 when not found.
func parseIntBefore(s, marker string) int {
	idx := strings.Index(s, marker)
	if idx < 0 {
		return 0
	}
	left := strings.TrimSpace(s[:idx])
	// Find the last integer token on the left.
	fields := strings.Fields(left)
	for i := len(fields) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSuffix(fields[i], ","), ""))
		if err == nil {
			return n
		}
	}
	return 0
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

func utf8RuneCount(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func defaultExec(ctx context.Context, dir string, env, args []string) (string, string, int, error) {
	if _, err := exec.LookPath("pytest"); err != nil {
		return "", "", -1, result.Runtime(result.CodePlayerError, "pytest binary not found in PATH", false, nil)
	}
	cmd := exec.CommandContext(ctx, "pytest", args...)
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodeTimeout, "pytest.run timed out", true, map[string]any{
			"stderr": stderr.String(),
		})
	}
	if ctx.Err() == context.Canceled {
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodeCancelled, "pytest.run cancelled", false, nil)
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout.String(), stderr.String(), ee.ExitCode(), nil
		}
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
	return stdout.String(), stderr.String(), 0, nil
}