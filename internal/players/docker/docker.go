package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gspaim/Runtgine/internal/core/policy"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/shell"
)

const (
	CapPS      = "docker.ps"
	CapInspect = "docker.inspect"
	CapLogs    = "docker.logs"
	CapRun     = "docker.run"
	CapBuild   = "docker.build"

	defaultReadTimeoutMS = 60_000
	defaultWriteTimeout  = 120_000
	maxTimeoutMS         = 600_000
	maxLogsTail          = 1000
	defaultLogsTail      = 100
	maxInspectChars      = 4 * 1024 * 1024
)

type runCmdFunc func(ctx context.Context, timeout time.Duration, args []string) (stdout, stderr string, err error)

type Player struct {
	runCmd runCmdFunc
}

func New() *Player {
	return &Player{runCmd: defaultRunCmd}
}

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "docker",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapPS,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "all":{"type":"boolean","default":false},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["containers"],
  "properties":{
    "containers":{"type":"array","items":{"type":"object"}}
  }
}`),
			},
			{
				Name: CapInspect,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["id"],
  "properties":{
    "id":{"type":"string","minLength":1},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["id","object"],
  "properties":{
    "id":{"type":"string"},
    "object":{}
  }
}`),
			},
			{
				Name: CapLogs,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["id"],
  "properties":{
    "id":{"type":"string","minLength":1},
    "tail":{"type":"integer","minimum":1,"maximum":1000,"default":100},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["id","logs","truncated"],
  "properties":{
    "id":{"type":"string"},
    "logs":{"type":"string"},
    "truncated":{"type":"boolean"}
  }
}`),
			},
			{
				Name:            CapRun,
				ExecutionPolicy: string(policy.ApprovalRequired),
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["image"],
  "properties":{
    "image":{"type":"string","minLength":1},
    "argv":{"type":"array","items":{"type":"string"}},
    "workdir":{"type":"string"},
    "mount_workspace":{"type":"boolean","default":false},
    "timeout_ms":{"type":"integer","minimum":1,"default":120000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["container_id","exit_code"],
  "properties":{
    "container_id":{"type":"string"},
    "exit_code":{"type":"integer"},
    "stdout":{"type":"string"},
    "stderr":{"type":"string"}
  }
}`),
			},
			{
				Name:            CapBuild,
				ExecutionPolicy: string(policy.ApprovalRequired),
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "context":{"type":"string","default":"."},
    "tag":{"type":"string"},
    "dockerfile":{"type":"string"},
    "timeout_ms":{"type":"integer","minimum":1,"default":120000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "image_id":{"type":"string"},
    "tag":{"type":"string"},
    "stderr":{"type":"string"}
  }
}`),
			},
		},
	}
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapPS:
		var in psIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		return checkTimeout(in.TimeoutMS, defaultReadTimeoutMS)
	case CapInspect:
		var in inspectIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS, defaultReadTimeoutMS); err != nil {
			return err
		}
		return safeRef("id", in.ID)
	case CapLogs:
		var in logsIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS, defaultReadTimeoutMS); err != nil {
			return err
		}
		if in.Tail > maxLogsTail {
			return result.Validation(result.CodeInvalidInput, "docker.logs tail exceeds 1000", nil)
		}
		return safeRef("id", in.ID)
	case CapRun:
		var in runIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS, defaultWriteTimeout); err != nil {
			return err
		}
		if err := safeRef("image", in.Image); err != nil {
			return err
		}
		for _, a := range in.Argv {
			if strings.TrimSpace(a) == "" {
				return result.Validation(result.CodeInvalidInput, "docker.run argv items must be non-empty", nil)
			}
		}
		if in.Workdir != "" {
			if _, err := shell.ResolveWorkdir(workspace, in.Workdir); err != nil {
				return err
			}
		}
		return nil
	case CapBuild:
		var in buildIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS, defaultWriteTimeout); err != nil {
			return err
		}
		ctxPath := in.Context
		if ctxPath == "" {
			ctxPath = "."
		}
		if _, err := shell.ResolveWorkdir(workspace, ctxPath); err != nil {
			return err
		}
		if in.Dockerfile != "" {
			joined := filepath.Join(ctxPath, in.Dockerfile)
			if _, err := shell.ResolveWorkdir(workspace, joined); err != nil {
				return err
			}
		}
		if in.Tag != "" {
			return safeRef("tag", in.Tag)
		}
		return nil
	default:
		return result.Validation(result.CodeUnknownCapability, "docker player cannot validate "+capability, nil)
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	switch req.Capability {
	case CapPS:
		return p.execPS(ctx, req.Input)
	case CapInspect:
		return p.execInspect(ctx, req.Input)
	case CapLogs:
		return p.execLogs(ctx, req.Input)
	case CapRun:
		return p.execRun(ctx, req)
	case CapBuild:
		return p.execBuild(ctx, req)
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "unsupported capability "+req.Capability, nil)
	}
}

type psIn struct {
	All       bool `json:"all"`
	TimeoutMS int  `json:"timeout_ms"`
}

type inspectIn struct {
	ID        string `json:"id"`
	TimeoutMS int    `json:"timeout_ms"`
}

type logsIn struct {
	ID        string `json:"id"`
	Tail      int    `json:"tail"`
	TimeoutMS int    `json:"timeout_ms"`
}

type runIn struct {
	Image          string   `json:"image"`
	Argv           []string `json:"argv"`
	Workdir        string   `json:"workdir"`
	MountWorkspace bool     `json:"mount_workspace"`
	TimeoutMS      int      `json:"timeout_ms"`
}

type buildIn struct {
	Context    string `json:"context"`
	Tag        string `json:"tag"`
	Dockerfile string `json:"dockerfile"`
	TimeoutMS  int    `json:"timeout_ms"`
}

func (p *Player) execPS(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in psIn
	_ = json.Unmarshal(raw, &in)
	args := []string{"ps", "--format", "json"}
	if in.All {
		args = append(args, "-a")
	}
	stdout, stderr, err := p.cmd(ctx, timeoutMS(in.TimeoutMS, defaultReadTimeoutMS), args)
	if err != nil {
		return nil, playerErr("docker.ps", err, stderr)
	}
	var containers []map[string]any
	dec := json.NewDecoder(strings.NewReader(stdout))
	for dec.More() {
		var row map[string]any
		if err := dec.Decode(&row); err != nil {
			break
		}
		containers = append(containers, map[string]any{
			"id":     firstString(row, "ID", "Id"),
			"image":  firstString(row, "Image"),
			"names":  firstString(row, "Names"),
			"status": firstString(row, "Status"),
		})
	}
	if containers == nil {
		containers = []map[string]any{}
	}
	return json.Marshal(map[string]any{"containers": containers})
}

func (p *Player) execInspect(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in inspectIn
	_ = json.Unmarshal(raw, &in)
	stdout, stderr, err := p.cmd(ctx, timeoutMS(in.TimeoutMS, defaultReadTimeoutMS), []string{"inspect", in.ID})
	if err != nil {
		return nil, playerErr("docker.inspect", err, stderr)
	}
	if len(stdout) > maxInspectChars {
		stdout = stdout[:maxInspectChars]
	}
	var parsed any
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		return nil, result.Runtime(result.CodePlayerError, "docker.inspect: invalid json", false, nil)
	}
	obj := parsed
	if arr, ok := parsed.([]any); ok && len(arr) > 0 {
		obj = arr[0]
	}
	return json.Marshal(map[string]any{"id": in.ID, "object": obj})
}

func (p *Player) execLogs(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in logsIn
	_ = json.Unmarshal(raw, &in)
	tail := in.Tail
	if tail <= 0 {
		tail = defaultLogsTail
	}
	stdout, stderr, err := p.cmd(ctx, timeoutMS(in.TimeoutMS, defaultReadTimeoutMS), []string{
		"logs", "--tail", fmt.Sprintf("%d", tail), in.ID,
	})
	if err != nil {
		return nil, playerErr("docker.logs", err, stderr)
	}
	truncated := false
	const maxLogs = 1 << 20
	if len(stdout) > maxLogs {
		stdout = stdout[:maxLogs]
		truncated = true
	}
	return json.Marshal(map[string]any{"id": in.ID, "logs": stdout, "truncated": truncated})
}

func (p *Player) execRun(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in runIn
	_ = json.Unmarshal(req.Input, &in)
	args := []string{"run", "--pull=never", "--network=none", "--rm"}
	if in.MountWorkspace {
		ws, err := filepath.Abs(req.Workspace)
		if err != nil {
			return nil, result.Runtime(result.CodeInternal, err.Error(), false, nil)
		}
		args = append(args, "-v", ws+":"+ws+":ro")
	}
	if in.Workdir != "" {
		wd, err := shell.ResolveWorkdir(req.Workspace, in.Workdir)
		if err != nil {
			return nil, err
		}
		args = append(args, "-w", wd)
	}
	args = append(args, in.Image)
	args = append(args, in.Argv...)
	stdout, stderr, err := p.cmd(ctx, timeoutMS(in.TimeoutMS, defaultWriteTimeout), args)
	exit := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			return nil, playerErr("docker.run", err, stderr)
		}
	}
	return json.Marshal(map[string]any{
		"container_id": "",
		"exit_code":    exit,
		"stdout":       stdout,
		"stderr":       stderr,
	})
}

func (p *Player) execBuild(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in buildIn
	_ = json.Unmarshal(req.Input, &in)
	ctxPath := in.Context
	if ctxPath == "" {
		ctxPath = "."
	}
	abs, err := shell.ResolveWorkdir(req.Workspace, ctxPath)
	if err != nil {
		return nil, err
	}
	args := []string{"build", "--pull=false"}
	if in.Tag != "" {
		args = append(args, "-t", in.Tag)
	}
	if in.Dockerfile != "" {
		args = append(args, "-f", filepath.Join(abs, in.Dockerfile))
	}
	args = append(args, abs)
	stdout, stderr, err := p.cmd(ctx, timeoutMS(in.TimeoutMS, defaultWriteTimeout), args)
	if err != nil {
		return nil, playerErr("docker.build", err, stderr)
	}
	out := map[string]any{"stderr": stderr, "image_id": strings.TrimSpace(stdout)}
	if in.Tag != "" {
		out["tag"] = in.Tag
	}
	return json.Marshal(out)
}

func (p *Player) cmd(ctx context.Context, d time.Duration, args []string) (string, string, error) {
	fn := p.runCmd
	if fn == nil {
		fn = defaultRunCmd
	}
	return fn(ctx, d, args)
}

// SetRunner replaces the docker CLI invoker (tests).
func (p *Player) SetRunner(fn func(ctx context.Context, timeout time.Duration, args []string) (stdout, stderr string, err error)) {
	p.runCmd = fn
}

func defaultRunCmd(ctx context.Context, timeout time.Duration, args []string) (string, string, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return "", "", result.Runtime(result.CodePlayerError, "docker binary not found in PATH", false, nil)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func checkTimeout(ms, def int) error {
	if ms == 0 {
		return nil
	}
	if ms > maxTimeoutMS {
		return result.Validation(result.CodeInvalidInput, "timeout_ms exceeds 10 minutes", nil)
	}
	if ms < 1 {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be >= 1", nil)
	}
	_ = def
	return nil
}

func timeoutMS(ms, def int) time.Duration {
	if ms <= 0 {
		ms = def
	}
	if ms > maxTimeoutMS {
		ms = maxTimeoutMS
	}
	return time.Duration(ms) * time.Millisecond
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

func invalid(cap string, err error) error {
	return result.Validation(result.CodeInvalidInput, "invalid "+cap+" input: "+err.Error(), nil)
}

func playerErr(cap string, err error, stderr string) error {
	msg := cap + ": " + err.Error()
	if strings.TrimSpace(stderr) != "" {
		msg += ": " + strings.TrimSpace(stderr)
		if len(msg) > 512 {
			msg = msg[:512] + "…"
		}
	}
	return result.Runtime(result.CodePlayerError, msg, false, nil)
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}
