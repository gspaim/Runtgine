package k8s

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
	CapList = "k8s.list"
	CapGet  = "k8s.get"

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

func (p *Player) Manifest() registry.Manifest {
	timeout := `"timeout_ms":{"type":"integer","minimum":1,"maximum":120000}`
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "k8s",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapList,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["resource"],
  "properties":{
    "resource":{"type":"string","minLength":1},
    "namespace":{"type":"string"},
    ` + timeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["resource","object","truncated"],
  "properties":{
    "resource":{"type":"string"},
    "object":{},
    "truncated":{"type":"boolean"}
  }
}`),
			},
			{
				Name: CapGet,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["resource","name"],
  "properties":{
    "resource":{"type":"string","minLength":1},
    "name":{"type":"string","minLength":1},
    "namespace":{"type":"string"},
    ` + timeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["resource","name","object","truncated"],
  "properties":{
    "resource":{"type":"string"},
    "name":{"type":"string"},
    "object":{},
    "truncated":{"type":"boolean"}
  }
}`),
			},
		},
	}
}

type listIn struct {
	Resource  string `json:"resource"`
	Namespace string `json:"namespace"`
	TimeoutMS int    `json:"timeout_ms"`
}

type getIn struct {
	Resource  string `json:"resource"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	TimeoutMS int    `json:"timeout_ms"`
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapList:
		var in listIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS); err != nil {
			return err
		}
		if err := safeRef("resource", in.Resource); err != nil {
			return err
		}
		if strings.TrimSpace(in.Namespace) != "" {
			return safeRef("namespace", in.Namespace)
		}
		return nil
	case CapGet:
		var in getIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS); err != nil {
			return err
		}
		if err := safeRef("resource", in.Resource); err != nil {
			return err
		}
		if err := safeRef("name", in.Name); err != nil {
			return err
		}
		if strings.TrimSpace(in.Namespace) != "" {
			return safeRef("namespace", in.Namespace)
		}
		return nil
	default:
		return result.Validation(result.CodeUnknownCapability, "k8s player cannot validate "+capability, nil)
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	switch req.Capability {
	case CapList:
		return p.execList(ctx, req.Input)
	case CapGet:
		return p.execGet(ctx, req.Input)
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "unknown capability "+req.Capability, nil)
	}
}

func (p *Player) execList(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in listIn
	_ = json.Unmarshal(raw, &in)
	args := []string{"get", in.Resource}
	if ns := strings.TrimSpace(in.Namespace); ns != "" {
		args = append(args, "-n", ns)
	}
	args = append(args, "-o", "json")
	obj, truncated, err := p.runJSON(ctx, timeoutMS(in.TimeoutMS), args)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"resource":  in.Resource,
		"object":    obj,
		"truncated": truncated,
	})
}

func (p *Player) execGet(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in getIn
	_ = json.Unmarshal(raw, &in)
	args := []string{"get", in.Resource, in.Name}
	if ns := strings.TrimSpace(in.Namespace); ns != "" {
		args = append(args, "-n", ns)
	}
	args = append(args, "-o", "json")
	obj, truncated, err := p.runJSON(ctx, timeoutMS(in.TimeoutMS), args)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"resource":  in.Resource,
		"name":      in.Name,
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
			msg = "kubectl exited with code " + strconv.Itoa(exit)
		}
		return nil, false, result.Runtime(result.CodePlayerError, "kubectl: "+msg, false, map[string]any{"exit_code": exit})
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
	}
	return obj, truncated, nil
}

func defaultRunCmd(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return "", "", 0, result.Runtime(result.CodePlayerError, "kubectl binary not found in PATH", false, nil)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "kubectl", args...)
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
			return stdout.String(), stderr.String(), 0, result.Runtime(result.CodeTimeout, "k8s: timeout", true, nil)
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

func checkTimeout(ms int) error {
	if ms == 0 {
		return nil
	}
	if ms < 1 || ms > maxTimeoutMS {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 120000", nil)
	}
	return nil
}

func timeoutMS(ms int) time.Duration {
	if ms <= 0 {
		ms = defaultTimeoutMS
	}
	return time.Duration(ms) * time.Millisecond
}

func invalid(cap string, err error) error {
	return result.Validation(result.CodeInvalidInput, "invalid "+cap+" input: "+err.Error(), nil)
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
}
