package tf

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

	"github.com/gspaim/Runtgine/internal/core/policy"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/shell"
)

const (
	CapValidate = "tf.validate"
	CapPlan     = "tf.plan"

	defaultTimeoutMS = 120_000
	maxTimeoutMS     = 600_000
	maxLogBytes      = 64 * 1024
)

type runCmdFunc func(ctx context.Context, timeout time.Duration, dir string, args []string) (stdout, stderr string, exitCode int, err error)

type Player struct {
	runCmd runCmdFunc
}

func New() *Player { return &Player{runCmd: defaultRunCmd} }

func (p *Player) SetRunner(fn runCmdFunc) { p.runCmd = fn }

func (p *Player) Manifest() registry.Manifest {
	inSchema := json.RawMessage(`{
  "type":"object",
  "properties":{
    "workdir":{"type":"string"},
    "timeout_ms":{"type":"integer","minimum":1,"maximum":600000}
  },
  "additionalProperties":false
}`)
	outSchema := json.RawMessage(`{
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
		Name:          "terraform",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{Name: CapValidate, InputSchema: inSchema, OutputSchema: outSchema},
			{
				Name:            CapPlan,
				ExecutionPolicy: string(policy.ApprovalRequired),
				InputSchema:     inSchema,
				OutputSchema:    outSchema,
			},
		},
	}
}

type input struct {
	Workdir   string `json:"workdir"`
	TimeoutMS int    `json:"timeout_ms"`
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	if capability != CapValidate && capability != CapPlan {
		return result.Validation(result.CodeUnknownCapability, "terraform player cannot validate "+capability, nil)
	}
	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid "+capability+" input: "+err.Error(), nil)
	}
	if in.TimeoutMS != 0 && (in.TimeoutMS < 1 || in.TimeoutMS > maxTimeoutMS) {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 600000", nil)
	}
	wd := strings.TrimSpace(in.Workdir)
	if wd == "" {
		wd = "."
	}
	if strings.ContainsAny(wd, "\x00\n\r") || strings.Contains(wd, "://") || filepath.IsAbs(wd) {
		return result.Validation(result.CodeInvalidInput, "workdir must be relative to workspace", map[string]any{"workdir": wd})
	}
	resolved, err := shell.ResolveWorkdir(workspace, wd)
	if err != nil {
		return err
	}
	if !hasTF(resolved) {
		return result.Validation(result.CodeInvalidInput, "workdir must contain *.tf or *.tf.json", map[string]any{"workdir": resolved})
	}
	return nil
}

func hasTF(dir string) bool {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		n := strings.ToLower(e.Name())
		if strings.HasSuffix(n, ".tf") || strings.HasSuffix(n, ".tf.json") {
			return true
		}
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
	dir, err := shell.ResolveWorkdir(req.Workspace, wd)
	if err != nil {
		return nil, err
	}
	args := []string{"validate", "-no-color"}
	if req.Capability == CapPlan {
		args = []string{"plan", "-lock=false", "-input=false", "-no-color"}
	}
	timeout := in.TimeoutMS
	if timeout == 0 {
		timeout = defaultTimeoutMS
	}
	fn := p.runCmd
	if fn == nil {
		fn = defaultRunCmd
	}
	start := time.Now()
	stdout, stderr, exit, runErr := fn(ctx, time.Duration(timeout)*time.Millisecond, dir, args)
	elapsed := time.Since(start).Milliseconds()
	if runErr != nil {
		if re, ok := runErr.(result.Error); ok {
			return nil, re
		}
		return nil, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
	logText := joinLog(stdout, stderr)
	out := map[string]any{
		"ok":         exit == 0,
		"exit_code":  exit,
		"elapsed_ms": elapsed,
		"log":        logText,
	}
	if exit != 0 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", req.Capability, exit), false, out)
	}
	return json.Marshal(out)
}

func joinLog(stdout, stderr string) string {
	logText := strings.TrimSpace(stdout)
	if s := strings.TrimSpace(stderr); s != "" {
		if logText != "" {
			logText += "\n"
		}
		logText += s
	}
	if utf8.RuneCountInString(logText) > maxLogBytes {
		r := []rune(logText)
		logText = string(r[:maxLogBytes])
	}
	return logText
}

func defaultRunCmd(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
	if _, err := exec.LookPath("terraform"); err != nil {
		return "", "", 0, result.Runtime(result.CodePlayerError, "terraform binary not found in PATH", false, nil)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "terraform", args...)
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
			return stdout.String(), stderr.String(), 0, result.Runtime(result.CodeTimeout, "tf: timeout", true, nil)
		}
	}
	return stdout.String(), stderr.String(), exit, err
}
