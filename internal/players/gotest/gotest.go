package gotest

import (
	"bufio"
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
	CapGo = "test.go"

	defaultTimeoutMS = 120_000
	maxTimeoutMS     = 600_000
	defaultCount     = 1
	maxCount         = 10
	maxLogBytes      = 64 * 1024
)

// ExecFunc invokes `go` with argv (never a shell string). args[0] is "test".
type ExecFunc func(ctx context.Context, dir string, env, args []string) (stdout, stderr string, exitCode int, err error)

type Player struct {
	exec ExecFunc
}

func New() *Player {
	return &Player{exec: defaultExec}
}

func (p *Player) SetExec(fn ExecFunc) { p.exec = fn }

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "test",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapGo,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "packages":{"type":"array","items":{"type":"string","minLength":1}},
    "timeout_ms":{"type":"integer","minimum":1,"maximum":600000},
    "short":{"type":"boolean"},
    "count":{"type":"integer","minimum":1,"maximum":10},
    "run":{"type":"string"}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["ok","pass","fail","skip","elapsed_ms","exit_code","log"],
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
	Packages  []string `json:"packages"`
	TimeoutMS int      `json:"timeout_ms"`
	Short     bool     `json:"short"`
	Count     int      `json:"count"`
	Run       string   `json:"run"`
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

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	if capability != CapGo {
		return result.Validation(result.CodeUnknownCapability, "test player cannot validate "+capability, nil)
	}
	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid "+CapGo+" input: "+err.Error(), nil)
	}
	if in.TimeoutMS != 0 && (in.TimeoutMS < 1 || in.TimeoutMS > maxTimeoutMS) {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 600000", map[string]any{"timeout_ms": in.TimeoutMS})
	}
	if in.Count != 0 && (in.Count < 1 || in.Count > maxCount) {
		return result.Validation(result.CodeInvalidInput, "count must be between 1 and 10", map[string]any{"count": in.Count})
	}
	if err := validateRun(in.Run); err != nil {
		return err
	}
	pkgs := in.Packages
	if len(pkgs) == 0 {
		pkgs = []string{"./..."}
	}
	return validatePackages(workspace, pkgs)
}

func validateRun(run string) error {
	if run == "" {
		return nil
	}
	if strings.ContainsAny(run, "\x00\n\r") {
		return result.Validation(result.CodeInvalidInput, "run must not contain NUL or newlines", nil)
	}
	return nil
}

func validatePackages(workspace string, packages []string) error {
	ws, err := filepath.Abs(workspace)
	if err != nil {
		return result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	if resolved, err := filepath.EvalSymlinks(ws); err == nil {
		ws = resolved
	}
	for _, pkg := range packages {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			return result.Validation(result.CodeInvalidInput, "package must be non-empty", nil)
		}
		if strings.ContainsAny(pkg, "\x00\n\r") {
			return result.Validation(result.CodeInvalidInput, "package must not contain NUL or newlines", map[string]any{"package": pkg})
		}
		if strings.Contains(pkg, "://") {
			return result.Validation(result.CodeInvalidInput, "package must not be a URL", map[string]any{"package": pkg})
		}
		if filepath.IsAbs(pkg) {
			return result.Validation(result.CodeInvalidInput, "package must be relative to workspace", map[string]any{"package": pkg})
		}
		if strings.HasPrefix(pkg, "-") {
			return result.Validation(result.CodeInvalidInput, "package must not start with -", map[string]any{"package": pkg})
		}
		if !strings.HasPrefix(pkg, ".") {
			return result.Validation(result.CodeInvalidInput, "package must be a relative path starting with .", map[string]any{"package": pkg})
		}
		confine := strings.TrimSuffix(pkg, "/...")
		if confine == "" {
			confine = "."
		}
		if err := confinePath(ws, confine); err != nil {
			return err
		}
	}
	return nil
}

func confinePath(ws, rel string) error {
	target := filepath.Join(ws, rel)
	target, err := filepath.Abs(target)
	if err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid package path", nil)
	}
	resolved := target
	if r, err := filepath.EvalSymlinks(target); err == nil {
		resolved = r
	} else {
		resolved, err = evalExisting(target)
		if err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid package path: "+err.Error(), nil)
		}
	}
	outRel, err := filepath.Rel(ws, resolved)
	if err != nil || outRel == ".." || strings.HasPrefix(outRel, ".."+string(os.PathSeparator)) {
		return result.Validation(result.CodeInvalidInput, "package must be inside workspace_root", map[string]any{
			"workspace": ws,
			"package":   resolved,
		})
	}
	return nil
}

func evalExisting(path string) (string, error) {
	if r, err := filepath.EvalSymlinks(path); err == nil {
		return r, nil
	}
	var missing []string
	p := path
	for {
		dir := filepath.Dir(p)
		base := filepath.Base(p)
		if dir == p {
			return "", fmt.Errorf("cannot resolve %q", path)
		}
		missing = append([]string{base}, missing...)
		if r, err := filepath.EvalSymlinks(dir); err == nil {
			return filepath.Join(append([]string{r}, missing...)...), nil
		}
		p = dir
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	var in input
	_ = json.Unmarshal(req.Input, &in)
	timeoutMS := in.TimeoutMS
	if timeoutMS == 0 {
		timeoutMS = defaultTimeoutMS
	}
	count := in.Count
	if count == 0 {
		count = defaultCount
	}
	pkgs := in.Packages
	if len(pkgs) == 0 {
		pkgs = []string{"./..."}
	}

	ws, err := shell.ResolveWorkdir(req.Workspace, ".")
	if err != nil {
		return nil, err
	}

	timeoutDur := time.Duration(timeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeoutDur)
	defer cancel()

	args := buildArgs(in.Short, count, timeoutMS, strings.TrimSpace(in.Run), pkgs)
	env := buildEnv()
	fn := p.exec
	if fn == nil {
		fn = defaultExec
	}

	start := time.Now()
	stdout, stderr, exitCode, runErr := fn(ctx, ws, env, args)
	elapsed := time.Since(start).Milliseconds()
	if runErr != nil {
		if re, ok := runErr.(result.Error); ok {
			return nil, re
		}
		return nil, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}

	summary := parseJSONStream(stdout)
	logText := summary.log
	if strings.TrimSpace(logText) == "" {
		logText = strings.TrimSpace(stderr)
	}
	if strings.TrimSpace(logText) == "" {
		logText = fmt.Sprintf("pass=%d fail=%d skip=%d exit=%d", summary.pass, summary.fail, summary.skip, exitCode)
	}
	logText = truncateUTF8(logText, maxLogBytes)

	out := output{
		OK:        exitCode == 0,
		Pass:      summary.pass,
		Fail:      summary.fail,
		Skip:      summary.skip,
		ElapsedMS: elapsed,
		ExitCode:  exitCode,
		Log:       logText,
	}
	if exitCode != 0 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", CapGo, exitCode), false, map[string]any{
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

func buildArgs(short bool, count, timeoutMS int, run string, packages []string) []string {
	sec := (timeoutMS + 999) / 1000
	if sec < 1 {
		sec = 1
	}
	args := []string{
		"test",
		"-json",
		"-mod=readonly",
		"-count", strconv.Itoa(count),
		"-timeout", strconv.Itoa(sec) + "s",
	}
	if short {
		args = append(args, "-short")
	}
	if run != "" {
		args = append(args, "-run", run)
	}
	return append(args, packages...)
}

type jsonSummary struct {
	pass, fail, skip int
	log              string
}

type testEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test"`
	Output string `json:"Output"`
}

func parseJSONStream(stdout string) jsonSummary {
	var sum jsonSummary
	var logBuf strings.Builder
	sc := bufio.NewScanner(strings.NewReader(stdout))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev testEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			logBuf.WriteString(line)
			logBuf.WriteByte('\n')
			continue
		}
		if ev.Output != "" {
			logBuf.WriteString(ev.Output)
		}
		if ev.Test == "" {
			continue
		}
		switch ev.Action {
		case "pass":
			sum.pass++
		case "fail":
			sum.fail++
		case "skip":
			sum.skip++
		}
	}
	sum.log = logBuf.String()
	return sum
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
	if _, err := exec.LookPath("go"); err != nil {
		return "", "", -1, result.Runtime(result.CodePlayerError, "go binary not found in PATH", false, nil)
	}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodeTimeout, "test.go timed out", true, map[string]any{
			"stderr": stderr.String(),
		})
	}
	if ctx.Err() == context.Canceled {
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodeCancelled, "test.go cancelled", false, nil)
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout.String(), stderr.String(), ee.ExitCode(), nil
		}
		return stdout.String(), stderr.String(), -1, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
	return stdout.String(), stderr.String(), 0, nil
}
