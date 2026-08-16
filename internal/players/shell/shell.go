package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
)

const CapExec = "shell.exec"

type Player struct{}

func New() *Player { return &Player{} }

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "shell",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{{
			Name: CapExec,
			InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["command"],
  "properties":{
    "command":{"type":"array","items":{"type":"string"},"minItems":1},
    "workdir":{"type":"string","default":"."},
    "env":{"type":"object","additionalProperties":{"type":"string"}},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
			OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["exit_code"],
  "properties":{
    "exit_code":{"type":"integer"},
    "stdout":{"type":"string"},
    "stderr":{"type":"string"}
  }
}`),
		}},
	}
}

type input struct {
	Command   []string          `json:"command"`
	Workdir   string            `json:"workdir"`
	Env       map[string]string `json:"env"`
	TimeoutMS int               `json:"timeout_ms"`
}

type output struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if req.Capability != CapExec {
		return nil, result.Validation(result.CodeUnknownCapability, "shell player cannot serve "+req.Capability, nil)
	}
	var in input
	if err := json.Unmarshal(req.Input, &in); err != nil {
		return nil, result.Validation(result.CodeInvalidInput, "invalid shell.exec input: "+err.Error(), nil)
	}
	if len(in.Command) == 0 || strings.TrimSpace(in.Command[0]) == "" {
		return nil, result.Validation(result.CodeInvalidInput, "command must be a non-empty argv array", nil)
	}
	// Sandbox: no implicit shell; argv only (G-18)
	if in.TimeoutMS <= 0 {
		in.TimeoutMS = 60000
	}
	if in.Workdir == "" {
		in.Workdir = "."
	}

	workAbs, err := resolveWorkdir(req.Workspace, in.Workdir)
	if err != nil {
		return nil, err
	}

	cctx, cancel := context.WithTimeout(ctx, time.Duration(in.TimeoutMS)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(cctx, in.Command[0], in.Command[1:]...)
	cmd.Dir = workAbs
	if len(in.Env) > 0 {
		// minimal env: only explicitly provided keys (plus empty base)
		env := make([]string, 0, len(in.Env))
		for k, v := range in.Env {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	out := output{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cctx.Err() == context.DeadlineExceeded {
		return nil, result.Runtime(result.CodeTimeout, "shell.exec timed out", true, map[string]any{
			"timeout_ms": in.TimeoutMS,
			"stderr":     out.Stderr,
		})
	}
	if cctx.Err() == context.Canceled {
		return nil, result.Runtime(result.CodeCancelled, "shell.exec cancelled", false, nil)
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			out.ExitCode = ee.ExitCode()
			b, _ := json.Marshal(out)
			// non-zero exit is a failed step but structured output returned via error wrapper
			return b, result.Runtime(result.CodePlayerError, fmt.Sprintf("command exited with code %d", out.ExitCode), false, map[string]any{
				"exit_code": out.ExitCode,
				"stderr":    out.Stderr,
			})
		}
		return nil, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
	out.ExitCode = 0
	return json.Marshal(out)
}

func resolveWorkdir(workspace, workdir string) (string, error) {
	ws, err := filepath.Abs(workspace)
	if err != nil {
		return "", result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	target := workdir
	if !filepath.IsAbs(target) {
		target = filepath.Join(ws, workdir)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "", result.Validation(result.CodeInvalidInput, "invalid workdir", nil)
	}
	rel, err := filepath.Rel(ws, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", result.Validation(result.CodeInvalidInput, "workdir must be inside workspace_root", map[string]any{
			"workspace": ws,
			"workdir":   target,
		})
	}
	return target, nil
}
