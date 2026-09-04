package aws

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
	CapStsIdentity = "aws.sts-identity"
	CapS3Buckets   = "aws.s3-buckets"
	CapS3Objects   = "aws.s3-objects"

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
	outSchema := json.RawMessage(`{
  "type":"object",
  "required":["object","truncated"],
  "properties":{
    "object":{},
    "truncated":{"type":"boolean"}
  }
}`)
	region := `"region":{"type":"string"}`
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "aws",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapStsIdentity,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    ` + region + `,
    ` + timeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: outSchema,
			},
			{
				Name: CapS3Buckets,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    ` + region + `,
    ` + timeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: outSchema,
			},
			{
				Name: CapS3Objects,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["bucket"],
  "properties":{
    "bucket":{"type":"string","minLength":1},
    "prefix":{"type":"string"},
    ` + region + `,
    ` + timeout + `
  },
  "additionalProperties":false
}`),
				OutputSchema: outSchema,
			},
		},
	}
}

type identityIn struct {
	Region    string `json:"region"`
	TimeoutMS int    `json:"timeout_ms"`
}

type bucketsIn struct {
	Region    string `json:"region"`
	TimeoutMS int    `json:"timeout_ms"`
}

type objectsIn struct {
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix"`
	Region    string `json:"region"`
	TimeoutMS int    `json:"timeout_ms"`
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapStsIdentity:
		var in identityIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS); err != nil {
			return err
		}
		return checkRegion(in.Region)
	case CapS3Buckets:
		var in bucketsIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS); err != nil {
			return err
		}
		return checkRegion(in.Region)
	case CapS3Objects:
		var in objectsIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return invalid(capability, err)
		}
		if err := checkTimeout(in.TimeoutMS); err != nil {
			return err
		}
		if err := safeRef("bucket", in.Bucket); err != nil {
			return err
		}
		if p := strings.TrimSpace(in.Prefix); p != "" {
			if err := safeRef("prefix", in.Prefix); err != nil {
				return err
			}
		}
		return checkRegion(in.Region)
	default:
		return result.Validation(result.CodeUnknownCapability, "aws player cannot validate "+capability, nil)
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	args, timeout, err := buildArgs(req.Capability, req.Input)
	if err != nil {
		return nil, err
	}
	args = append(args, "--no-cli-pager", "--output", "json")
	obj, truncated, err := p.runJSON(ctx, time.Duration(timeout)*time.Millisecond, args)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"object":    obj,
		"truncated": truncated,
	})
}

// buildArgs maps a validated capability input to closed argv;
// Execute appends the global --no-cli-pager --output json (spec 43).
func buildArgs(capability string, raw json.RawMessage) ([]string, int, error) {
	switch capability {
	case CapStsIdentity:
		var in identityIn
		_ = json.Unmarshal(raw, &in)
		args := []string{"sts", "get-caller-identity"}
		if r := strings.TrimSpace(in.Region); r != "" {
			args = append(args, "--region", r)
		}
		return args, orDefault(in.TimeoutMS), nil
	case CapS3Buckets:
		var in bucketsIn
		_ = json.Unmarshal(raw, &in)
		args := []string{"s3api", "list-buckets"}
		if r := strings.TrimSpace(in.Region); r != "" {
			args = append(args, "--region", r)
		}
		return args, orDefault(in.TimeoutMS), nil
	case CapS3Objects:
		var in objectsIn
		_ = json.Unmarshal(raw, &in)
		args := []string{"s3api", "list-objects-v2", "--bucket", in.Bucket}
		if pfx := strings.TrimSpace(in.Prefix); pfx != "" {
			args = append(args, "--prefix", pfx)
		}
		if r := strings.TrimSpace(in.Region); r != "" {
			args = append(args, "--region", r)
		}
		return args, orDefault(in.TimeoutMS), nil
	default:
		return nil, 0, result.Validation(result.CodeUnknownCapability, "unknown capability "+capability, nil)
	}
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
			msg = "aws exited with code " + strconv.Itoa(exit)
		}
		return nil, false, result.Runtime(result.CodePlayerError, "aws: "+msg, false, map[string]any{"exit_code": exit})
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
	if _, err := exec.LookPath("aws"); err != nil {
		return "", "", 0, result.Runtime(result.CodePlayerError, "aws binary not found in PATH", false, nil)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "aws", args...)
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
			return stdout.String(), stderr.String(), 0, result.Runtime(result.CodeTimeout, "aws: timeout", true, nil)
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

func checkRegion(region string) error {
	if strings.TrimSpace(region) == "" {
		return nil
	}
	return safeRef("region", region)
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

func orDefault(ms int) int {
	if ms <= 0 {
		return defaultTimeoutMS
	}
	return ms
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
