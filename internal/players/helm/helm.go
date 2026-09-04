package helm

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
	CapLint     = "helm.lint"
	CapTemplate = "helm.template"
	CapList     = "helm.list"
	CapStatus   = "helm.status"

	chartMarker = "Chart.yaml"

	// lint/template render locally and may take longer than cluster reads.
	defaultLongTimeoutMS = 120_000
	maxLongTimeoutMS     = 600_000
	// list/status read from the cluster via the inherited kubeconfig.
	defaultShortTimeoutMS = 30_000
	maxShortTimeoutMS     = 120_000

	maxOutputBytes = 1 << 20
)

type runCmdFunc func(ctx context.Context, timeout time.Duration, dir string, args []string) (stdout, stderr string, exitCode int, err error)

type Player struct {
	runCmd runCmdFunc
}

func New() *Player { return &Player{runCmd: defaultRunCmd} }

func (p *Player) SetRunner(fn runCmdFunc) { p.runCmd = fn }

func (p *Player) Manifest() registry.Manifest {
	longTimeout := `"timeout_ms":{"type":"integer","minimum":1,"maximum":600000}`
	shortTimeout := `"timeout_ms":{"type":"integer","minimum":1,"maximum":120000}`
	runOut := json.RawMessage(`{
  "type":"object",
  "required":["ok","exit_code","elapsed_ms","log"],
  "properties":{
    "ok":{"type":"boolean"},
    "exit_code":{"type":"integer"},
    "elapsed_ms":{"type":"integer"},
    "log":{"type":"string"}
  }
}`)
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "helm",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapLint,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["chart"],
  "properties":{
    "chart":{"type":"string","minLength":1},
    ` + longTimeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: runOut,
			},
			{
				Name: CapTemplate,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["chart"],
  "properties":{
    "chart":{"type":"string","minLength":1},
    "release":{"type":"string"},
    "namespace":{"type":"string"},
    ` + longTimeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["output","truncated"],
  "properties":{
    "output":{"type":"string"},
    "truncated":{"type":"boolean"}
  }
}`),
			},
			{
				Name: CapList,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "namespace":{"type":"string"},
    ` + shortTimeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["releases","truncated"],
  "properties":{
    "releases":{},
    "truncated":{"type":"boolean"}
  }
}`),
			},
			{
				Name: CapStatus,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["release"],
  "properties":{
    "release":{"type":"string","minLength":1},
    "namespace":{"type":"string"},
    ` + shortTimeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: runOut,
			},
		},
	}
}

type lintIn struct {
	Chart     string `json:"chart"`
	TimeoutMS int    `json:"timeout_ms"`
}

type templateIn struct {
	Chart     string `json:"chart"`
	Release   string `json:"release"`
	Namespace string `json:"namespace"`
	TimeoutMS int    `json:"timeout_ms"`
}

type listIn struct {
	Namespace string `json:"namespace"`
	TimeoutMS int    `json:"timeout_ms"`
}

type statusIn struct {
	Release   string `json:"release"`
	Namespace string `json:"namespace"`
	TimeoutMS int    `json:"timeout_ms"`
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapLint:
		var in lintIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkLongTimeout(in.TimeoutMS); err != nil {
			return err
		}
		return checkChart(workspace, in.Chart)
	case CapTemplate:
		var in templateIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkLongTimeout(in.TimeoutMS); err != nil {
			return err
		}
		if err := checkChart(workspace, in.Chart); err != nil {
			return err
		}
		if strings.TrimSpace(in.Release) != "" {
			if err := safeRef("release", in.Release); err != nil {
				return err
			}
		}
		if strings.TrimSpace(in.Namespace) != "" {
			return safeRef("namespace", in.Namespace)
		}
		return nil
	case CapList:
		var in listIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkShortTimeout(in.TimeoutMS); err != nil {
			return err
		}
		if strings.TrimSpace(in.Namespace) != "" {
			return safeRef("namespace", in.Namespace)
		}
		return nil
	case CapStatus:
		var in statusIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkShortTimeout(in.TimeoutMS); err != nil {
			return err
		}
		if err := safeRef("release", in.Release); err != nil {
			return err
		}
		if strings.TrimSpace(in.Namespace) != "" {
			return safeRef("namespace", in.Namespace)
		}
		return nil
	default:
		return result.Validation(result.CodeUnknownCapability, "helm player cannot validate "+capability, nil)
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	switch req.Capability {
	case CapLint:
		return p.execRun(ctx, req.Workspace, CapLint, req.Input)
	case CapStatus:
		return p.execRun(ctx, req.Workspace, CapStatus, req.Input)
	case CapTemplate:
		return p.execTemplate(ctx, req)
	case CapList:
		return p.execList(ctx, req)
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "unknown capability "+req.Capability, nil)
	}
}

func (p *Player) execRun(ctx context.Context, workspace, capability string, raw json.RawMessage) (json.RawMessage, error) {
	var args []string
	var timeoutMS int
	switch capability {
	case CapLint:
		var in lintIn
		_ = json.Unmarshal(raw, &in)
		args = []string{"lint", in.Chart}
		timeoutMS = orDefault(in.TimeoutMS, defaultLongTimeoutMS)
	case CapStatus:
		var in statusIn
		_ = json.Unmarshal(raw, &in)
		args = []string{"status", in.Release}
		if ns := strings.TrimSpace(in.Namespace); ns != "" {
			args = append(args, "-n", ns)
		}
		timeoutMS = orDefault(in.TimeoutMS, defaultShortTimeoutMS)
	}
	fn := p.runCmd
	if fn == nil {
		fn = defaultRunCmd
	}
	start := time.Now()
	stdout, stderr, exit, runErr := fn(ctx, time.Duration(timeoutMS)*time.Millisecond, workspace, args)
	elapsed := time.Since(start).Milliseconds()
	if runErr != nil {
		if re, ok := runErr.(result.Error); ok {
			return nil, re
		}
		return nil, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
	out := map[string]any{
		"ok":         exit == 0,
		"exit_code":  exit,
		"elapsed_ms": elapsed,
		"log":        joinLog(stdout, stderr),
	}
	if exit != 0 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", capability, exit), false, out)
	}
	return json.Marshal(out)
}

func (p *Player) execTemplate(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in templateIn
	_ = json.Unmarshal(req.Input, &in)
	args := []string{"template"}
	if rel := strings.TrimSpace(in.Release); rel != "" {
		args = append(args, rel)
	}
	args = append(args, in.Chart)
	if ns := strings.TrimSpace(in.Namespace); ns != "" {
		args = append(args, "-n", ns)
	}
	stdout, exit, err := p.run(ctx, orDefault(in.TimeoutMS, defaultLongTimeoutMS), req.Workspace, args)
	if err != nil {
		return nil, err
	}
	if exit != 0 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", req.Capability, exit), false, nil)
	}
	output, truncated := truncateBytes(stdout, maxOutputBytes)
	return json.Marshal(map[string]any{
		"output":    output,
		"truncated": truncated,
	})
}

func (p *Player) execList(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in listIn
	_ = json.Unmarshal(req.Input, &in)
	args := []string{"list"}
	if ns := strings.TrimSpace(in.Namespace); ns != "" {
		args = append(args, "-n", ns)
	}
	args = append(args, "-o", "json")
	stdout, exit, err := p.run(ctx, orDefault(in.TimeoutMS, defaultShortTimeoutMS), req.Workspace, args)
	if err != nil {
		return nil, err
	}
	if exit != 0 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", req.Capability, exit), false, nil)
	}
	stdout, truncated := truncateBytes(stdout, maxOutputBytes)
	stdout = strings.TrimSpace(stdout)
	var releases any
	if json.Unmarshal([]byte(stdout), &releases) != nil {
		releases = truncateRunes(stdout, 4096)
		truncated = true
	}
	return json.Marshal(map[string]any{
		"releases":  releases,
		"truncated": truncated,
	})
}

func (p *Player) run(ctx context.Context, timeoutMS int, dir string, args []string) (string, int, error) {
	fn := p.runCmd
	if fn == nil {
		fn = defaultRunCmd
	}
	stdout, _, exit, err := fn(ctx, time.Duration(timeoutMS)*time.Millisecond, dir, args)
	if err != nil {
		if re, ok := err.(result.Error); ok {
			return "", 0, re
		}
		return "", 0, result.Runtime(result.CodePlayerError, err.Error(), false, nil)
	}
	return stdout, exit, nil
}

func checkChart(workspace, chart string) error {
	chart = strings.TrimSpace(chart)
	if chart == "" {
		return result.Validation(result.CodeInvalidInput, "chart is required", nil)
	}
	if strings.ContainsAny(chart, "\x00\n\r") || strings.Contains(chart, "://") || filepath.IsAbs(chart) {
		return result.Validation(result.CodeInvalidInput, "chart must be relative to workspace", map[string]any{"chart": chart})
	}
	resolved, err := shell.ResolveWorkdir(workspace, chart)
	if err != nil {
		return err
	}
	if !hasChartMarker(resolved) {
		return result.Validation(result.CodeInvalidInput, "chart must contain Chart.yaml", map[string]any{"chart": resolved})
	}
	return nil
}

func hasChartMarker(dir string) bool {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if e.Name() == chartMarker {
			return true
		}
	}
	return false
}

func defaultRunCmd(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
	if _, err := exec.LookPath("helm"); err != nil {
		return "", "", 0, result.Runtime(result.CodePlayerError, "helm binary not found in PATH", false, nil)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "helm", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exit := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
			err = nil
		} else if cctx.Err() != nil {
			return stdout.String(), stderr.String(), 0, result.Runtime(result.CodeTimeout, "helm: timeout", true, nil)
		}
	}
	return stdout.String(), stderr.String(), exit, err
}

func safeRef(field, v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return result.Validation(result.CodeInvalidInput, field+" is required", nil)
	}
	if strings.ContainsAny(v, " \t\n") || strings.HasPrefix(v, "-") {
		return result.Validation(result.CodeInvalidInput, field+" must not contain spaces or start with -", nil)
	}
	return nil
}

func checkLongTimeout(ms int) error {
	if ms == 0 {
		return nil
	}
	if ms < 1 || ms > maxLongTimeoutMS {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 600000", nil)
	}
	return nil
}

func checkShortTimeout(ms int) error {
	if ms == 0 {
		return nil
	}
	if ms < 1 || ms > maxShortTimeoutMS {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 120000", nil)
	}
	return nil
}

func orDefault(ms, def int) int {
	if ms <= 0 {
		return def
	}
	return ms
}

func invalid(cap string, err error) error {
	return result.Validation(result.CodeInvalidInput, "invalid "+cap+" input: "+err.Error(), nil)
}

func joinLog(stdout, stderr string) string {
	logText := strings.TrimSpace(stdout)
	if s := strings.TrimSpace(stderr); s != "" {
		if logText != "" {
			logText += "\n"
		}
		logText += s
	}
	if utf8.RuneCountInString(logText) > maxOutputBytes {
		r := []rune(logText)
		logText = string(r[:maxOutputBytes])
	}
	return logText
}

func truncateBytes(s string, max int) (string, bool) {
	if len(s) <= max {
		return s, false
	}
	return s[:max], true
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
}
