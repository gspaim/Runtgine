package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
)

const (
	CapIdentity      = "azure.identity"
	CapSubscriptions = "azure.subscriptions"
	CapGroups        = "azure.groups"

	defaultTimeoutMS = 30_000
	maxTimeoutMS     = 120_000
	maxJSONBytes     = 1 << 20
)

type runCmdFunc func(ctx context.Context, timeout time.Duration, args []string) (stdout, stderr string, exitCode int, err error)

type Player struct {
	runCmd runCmdFunc
}

func New() *Player { return &Player{runCmd: defaultRunCmd} }

func (p *Player) SetRunner(fn runCmdFunc) { p.runCmd = fn }

type input struct {
	TimeoutMS int `json:"timeout_ms"`
}

func (p *Player) Manifest() registry.Manifest {
	inSchema := json.RawMessage(`{
  "type":"object",
  "properties":{
    "timeout_ms":{"type":"integer","minimum":1,"maximum":120000}
  },
  "additionalProperties":false
}`)
	outSchema := json.RawMessage(`{
  "type":"object",
  "required":["object","truncated"],
  "properties":{
    "object":{},
    "truncated":{"type":"boolean"}
  }
}`)
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "azure",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{Name: CapIdentity, InputSchema: inSchema, OutputSchema: outSchema},
			{Name: CapSubscriptions, InputSchema: inSchema, OutputSchema: outSchema},
			{Name: CapGroups, InputSchema: inSchema, OutputSchema: outSchema},
		},
	}
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapIdentity, CapSubscriptions, CapGroups:
		var in input
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid "+capability+" input: "+err.Error(), nil)
		}
		return checkTimeout(in.TimeoutMS)
	default:
		return result.Validation(result.CodeUnknownCapability, "azure player cannot validate "+capability, nil)
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	var head []string
	switch req.Capability {
	case CapIdentity:
		head = []string{"account", "show"}
	case CapSubscriptions:
		head = []string{"account", "list"}
	case CapGroups:
		head = []string{"group", "list"}
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "unknown capability "+req.Capability, nil)
	}
	args := append(head, "-o", "json")
	obj, truncated, err := p.runJSON(ctx, timeoutMS(req.Input), args)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"object":    obj,
		"truncated": truncated,
	})
}

func (p *Player) runJSON(ctx context.Context, d time.Duration, args []string) (any, bool, error) {
	fn := p.runCmd
	if fn == nil {
		fn = defaultRunCmd
	}
	stdout, stderr, exit, err := fn(ctx, d, args)
	if err != nil {
		if re, ok := err.(result.Error); ok {
			return nil, false, re
		}
		return nil, false, result.Runtime(result.CodePlayerError, err.Error(), false, nil)
	}
	if exit != 0 {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = strings.TrimSpace(stdout)
		}
		if msg == "" {
			msg = "az exited with code " + strconv.Itoa(exit)
		}
		return nil, false, result.Runtime(result.CodePlayerError, "az: "+msg, false, map[string]any{"exit_code": exit})
	}
	truncated := false
	if len(stdout) > maxJSONBytes {
		stdout = stdout[:maxJSONBytes]
		truncated = true
	}
	stdout = strings.TrimSpace(stdout)
	var obj any
	if json.Unmarshal([]byte(stdout), &obj) != nil {
		obj = truncateRunes(stdout, 4096)
		truncated = true
	}
	return obj, truncated, nil
}

func defaultRunCmd(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
	if _, err := exec.LookPath("az"); err != nil {
		return "", "", 0, result.Runtime(result.CodePlayerError, "az binary not found in PATH", false, nil)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "az", args...)
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
			return stdout.String(), stderr.String(), 0, result.Runtime(result.CodeTimeout, "azure: timeout", true, nil)
		}
	}
	return stdout.String(), stderr.String(), exit, err
}

func checkTimeout(ms int) error {
	if ms == 0 {
		return nil
	}
	if ms < 1 || ms > maxTimeoutMS {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 120000", nil)
	}
	return nil
}

func timeoutMS(raw json.RawMessage) time.Duration {
	var in input
	_ = json.Unmarshal(raw, &in)
	if in.TimeoutMS <= 0 {
		return time.Duration(defaultTimeoutMS) * time.Millisecond
	}
	return time.Duration(in.TimeoutMS) * time.Millisecond
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
}
