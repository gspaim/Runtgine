package pg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
)

const (
	CapPing = "pg.ping"

	defaultTimeoutMS = 10_000
	maxTimeoutMS     = 60_000
	maxLogBytes      = 16 * 1024
	defaultHost      = "127.0.0.1"
	defaultPort      = 5432
)

type runCmdFunc func(ctx context.Context, timeout time.Duration, env, args []string) (stdout, stderr string, exitCode int, err error)

type Player struct {
	runCmd runCmdFunc
}

func New() *Player { return &Player{runCmd: defaultRunCmd} }

func (p *Player) SetRunner(fn runCmdFunc) { p.runCmd = fn }

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "postgres",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapPing,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["dbname"],
  "properties":{
    "dbname":{"type":"string","minLength":1},
    "host":{"type":"string"},
    "port":{"type":"integer","minimum":1,"maximum":65535},
    "user":{"type":"string"},
    "timeout_ms":{"type":"integer","minimum":1,"maximum":60000}
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
    "log":{"type":"string"}
  }
}`),
			},
		},
	}
}

type input struct {
	DBName    string `json:"dbname"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	TimeoutMS int    `json:"timeout_ms"`
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	if capability != CapPing {
		return result.Validation(result.CodeUnknownCapability, "postgres player cannot validate "+capability, nil)
	}
	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid "+CapPing+" input: "+err.Error(), nil)
	}
	if err := safeRef("dbname", in.DBName); err != nil {
		return err
	}
	if strings.TrimSpace(in.Host) != "" {
		if err := safeRef("host", in.Host); err != nil {
			return err
		}
	}
	if in.Port != 0 && (in.Port < 1 || in.Port > 65535) {
		return result.Validation(result.CodeInvalidInput, "port must be 1–65535", nil)
	}
	if strings.TrimSpace(in.User) != "" {
		if err := safeRef("user", in.User); err != nil {
			return err
		}
	}
	if in.TimeoutMS != 0 && (in.TimeoutMS < 1 || in.TimeoutMS > maxTimeoutMS) {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 60000", nil)
	}
	return nil
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	var in input
	_ = json.Unmarshal(req.Input, &in)
	host := strings.TrimSpace(in.Host)
	if host == "" {
		host = defaultHost
	}
	port := in.Port
	if port == 0 {
		port = defaultPort
	}
	timeout := in.TimeoutMS
	if timeout == 0 {
		timeout = defaultTimeoutMS
	}
	args := []string{
		"--host", host,
		"--port", strconv.Itoa(port),
		"--dbname", in.DBName,
		"--no-psqlrc",
		"-t", "-A",
		"--pset", "pager=off",
		"--command", "SELECT 1",
	}
	if u := strings.TrimSpace(in.User); u != "" {
		args = append(args, "--username", u)
	}
	fn := p.runCmd
	if fn == nil {
		fn = defaultRunCmd
	}
	start := time.Now()
	stdout, stderr, exit, runErr := fn(ctx, time.Duration(timeout)*time.Millisecond, buildEnv(), args)
	elapsed := time.Since(start).Milliseconds()
	if runErr != nil {
		if re, ok := runErr.(result.Error); ok {
			return nil, re
		}
		return nil, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
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
	out := map[string]any{
		"ok":         exit == 0,
		"exit_code":  exit,
		"elapsed_ms": elapsed,
		"log":        logText,
	}
	if exit != 0 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", CapPing, exit), false, out)
	}
	return json.Marshal(out)
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
	case "PATH", "HOME", "USER", "LANG", "TZ", "TMPDIR", "TMP", "TEMP", "PGPASSWORD", "PGSSLMODE":
		return true
	}
	return strings.HasPrefix(key, "LC_")
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

func defaultRunCmd(ctx context.Context, timeout time.Duration, env, args []string) (string, string, int, error) {
	if _, err := exec.LookPath("psql"); err != nil {
		return "", "", 0, result.Runtime(result.CodePlayerError, "psql binary not found in PATH", false, nil)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "psql", args...)
	cmd.Env = env
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
			return stdout.String(), stderr.String(), 0, result.Runtime(result.CodeTimeout, "pg: timeout", true, nil)
		}
	}
	return stdout.String(), stderr.String(), exit, err
}
