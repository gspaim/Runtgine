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
	CapPing    = "pg.ping"
	CapExplain = "pg.explain"

	defaultTimeoutMS = 10_000
	maxTimeoutMS     = 60_000
	maxLogBytes      = 16 * 1024
	maxSQLBytes      = 10 * 1024
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
	pingSchema := json.RawMessage(`{
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
}`)
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
		Name:          "postgres",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name:         CapPing,
				InputSchema:  pingSchema,
				OutputSchema: runOut,
			},
			{
				Name: CapExplain,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["sql","dbname"],
  "properties":{
    "sql":{"type":"string","minLength":1,"maxLength":10240},
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
  "required":["plan","truncated"],
  "properties":{
    "plan":{},
    "truncated":{"type":"boolean"}
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

type explainIn struct {
	SQL       string `json:"sql"`
	DBName    string `json:"dbname"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	TimeoutMS int    `json:"timeout_ms"`
}

func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapPing:
		var in input
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid "+CapPing+" input: "+err.Error(), nil)
		}
		return validateConn(in.DBName, in.Host, in.Port, in.User, in.TimeoutMS)
	case CapExplain:
		var in explainIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid "+CapExplain+" input: "+err.Error(), nil)
		}
		if err := validateSQL(in.SQL); err != nil {
			return err
		}
		return validateConn(in.DBName, in.Host, in.Port, in.User, in.TimeoutMS)
	default:
		return result.Validation(result.CodeUnknownCapability, "postgres player cannot validate "+capability, nil)
	}
}

func validateConn(dbname, host string, port int, user string, timeoutMS int) error {
	if err := safeRef("dbname", dbname); err != nil {
		return err
	}
	if strings.TrimSpace(host) != "" {
		if err := safeRef("host", host); err != nil {
			return err
		}
	}
	if port != 0 && (port < 1 || port > 65535) {
		return result.Validation(result.CodeInvalidInput, "port must be 1–65535", nil)
	}
	if strings.TrimSpace(user) != "" {
		if err := safeRef("user", user); err != nil {
			return err
		}
	}
	if timeoutMS != 0 && (timeoutMS < 1 || timeoutMS > maxTimeoutMS) {
		return result.Validation(result.CodeInvalidInput, "timeout_ms must be between 1 and 60000", nil)
	}
	return nil
}

// validateSQL is the conservative allowlist of G-232: a single
// SELECT/WITH statement, no `;`, no `\` (psql meta-commands
// impossible), bounded size. False positives are acceptable; the
// known vectors are closed.
func validateSQL(sql string) error {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return result.Validation(result.CodeInvalidInput, "sql is required", nil)
	}
	if len(trimmed) > maxSQLBytes {
		return result.Validation(result.CodeInvalidInput, "sql exceeds 10240 bytes", nil)
	}
	first := strings.ToUpper(strings.Fields(trimmed)[0])
	if first != "SELECT" && first != "WITH" {
		return result.Validation(result.CodeInvalidInput, "sql must be a single SELECT or WITH statement", map[string]any{"first_word": first})
	}
	if strings.Contains(trimmed, ";") {
		return result.Validation(result.CodeInvalidInput, "sql must not contain ;", nil)
	}
	if strings.Contains(trimmed, "\\") {
		return result.Validation(result.CodeInvalidInput, "sql must not contain \\", nil)
	}
	return nil
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	switch req.Capability {
	case CapPing:
		return p.execPing(ctx, req.Input)
	case CapExplain:
		return p.execExplain(ctx, req.Input)
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "unknown capability "+req.Capability, nil)
	}
}

func (p *Player) execPing(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in input
	_ = json.Unmarshal(raw, &in)
	host, port, timeout := connParams(in.Host, in.Port, in.TimeoutMS)
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

func (p *Player) execExplain(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var in explainIn
	_ = json.Unmarshal(raw, &in)
	host, port, timeout := connParams(in.Host, in.Port, in.TimeoutMS)
	sql := strings.TrimSpace(in.SQL)
	args := []string{
		"--host", host,
		"--port", strconv.Itoa(port),
		"--dbname", in.DBName,
		"--no-psqlrc",
		"-t", "-A",
		"--pset", "pager=off",
		"--command", "EXPLAIN (FORMAT JSON) " + sql,
	}
	if u := strings.TrimSpace(in.User); u != "" {
		args = append(args, "--username", u)
	}
	fn := p.runCmd
	if fn == nil {
		fn = defaultRunCmd
	}
	stdout, stderr, exit, runErr := fn(ctx, time.Duration(timeout)*time.Millisecond, buildEnv(), args)
	if runErr != nil {
		if re, ok := runErr.(result.Error); ok {
			return nil, re
		}
		return nil, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
	if exit != 0 {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = strings.TrimSpace(stdout)
		}
		if msg == "" {
			msg = fmt.Sprintf("%s exited with code %d", CapExplain, exit)
		}
		return nil, result.Runtime(result.CodePlayerError, "psql: "+msg, false, map[string]any{"exit_code": exit})
	}
	truncated := false
	if len(stdout) > maxLogBytes {
		stdout = stdout[:maxLogBytes]
		truncated = true
	}
	stdout = strings.TrimSpace(stdout)
	var plan any
	if json.Unmarshal([]byte(stdout), &plan) != nil {
		plan = truncateRunes(stdout, 4096)
		truncated = true
	}
	return json.Marshal(map[string]any{
		"plan":      plan,
		"truncated": truncated,
	})
}

func connParams(host string, port, timeoutMS int) (string, int, int) {
	h := strings.TrimSpace(host)
	if h == "" {
		h = defaultHost
	}
	p := port
	if p == 0 {
		p = defaultPort
	}
	t := timeoutMS
	if t == 0 {
		t = defaultTimeoutMS
	}
	return h, p, t
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
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
