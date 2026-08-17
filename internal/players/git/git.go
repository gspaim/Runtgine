package git

import (
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

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/shell"
)

const (
	CapStatus = "git.status"
	CapDiff   = "git.diff"
	CapLog    = "git.log"
	CapAdd    = "git.add"
	CapCommit = "git.commit"

	defaultTimeoutMS = 60000
	defaultLogMax    = 10
	maxLogMax        = 50
	maxDiffChars     = 200_000
)

type Player struct{}

func New() *Player { return &Player{} }

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "git",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapStatus,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "workdir":{"type":"string","default":"."},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["branch","porcelain","clean"],
  "properties":{
    "branch":{"type":"string"},
    "porcelain":{"type":"array","items":{"type":"string"}},
    "clean":{"type":"boolean"}
  }
}`),
			},
			{
				Name: CapDiff,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "workdir":{"type":"string","default":"."},
    "staged":{"type":"boolean","default":false},
    "paths":{"type":"array","items":{"type":"string"}},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["diff"],
  "properties":{
    "diff":{"type":"string"}
  }
}`),
			},
			{
				Name: CapLog,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "workdir":{"type":"string","default":"."},
    "max":{"type":"integer","minimum":1,"maximum":50,"default":10},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["entries"],
  "properties":{
    "entries":{"type":"array","items":{
      "type":"object",
      "required":["hash","subject","author","date"],
      "properties":{
        "hash":{"type":"string"},
        "subject":{"type":"string"},
        "author":{"type":"string"},
        "date":{"type":"string"}
      }
    }}
  }
}`),
			},
			{
				Name: CapAdd,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["paths"],
  "properties":{
    "paths":{"type":"array","items":{"type":"string"},"minItems":1},
    "workdir":{"type":"string","default":"."},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["added","exit_code"],
  "properties":{
    "added":{"type":"array","items":{"type":"string"}},
    "exit_code":{"type":"integer"}
  }
}`),
			},
			{
				Name: CapCommit,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["message"],
  "properties":{
    "message":{"type":"string","minLength":1},
    "workdir":{"type":"string","default":"."},
    "allow_empty":{"type":"boolean","default":false},
    "timeout_ms":{"type":"integer","minimum":1,"default":60000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["commit","exit_code"],
  "properties":{
    "commit":{"type":"string"},
    "exit_code":{"type":"integer"},
    "stderr":{"type":"string"}
  }
}`),
			},
		},
	}
}

type commonIn struct {
	Workdir   string `json:"workdir"`
	TimeoutMS int    `json:"timeout_ms"`
}

type statusOut struct {
	Branch    string   `json:"branch"`
	Porcelain []string `json:"porcelain"`
	Clean     bool     `json:"clean"`
}

type diffIn struct {
	commonIn
	Staged bool     `json:"staged"`
	Paths  []string `json:"paths"`
}

type diffOut struct {
	Diff string `json:"diff"`
}

type logIn struct {
	commonIn
	Max int `json:"max"`
}

type logEntry struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
	Author  string `json:"author"`
	Date    string `json:"date"`
}

type logOut struct {
	Entries []logEntry `json:"entries"`
}

type addIn struct {
	commonIn
	Paths []string `json:"paths"`
}

type addOut struct {
	Added    []string `json:"added"`
	ExitCode int      `json:"exit_code"`
}

type commitIn struct {
	commonIn
	Message    string `json:"message"`
	AllowEmpty bool   `json:"allow_empty"`
}

type commitOut struct {
	Commit   string `json:"commit"`
	ExitCode int    `json:"exit_code"`
	Stderr   string `json:"stderr,omitempty"`
}

// ValidateStaticInput applies sandbox checks for git.* capabilities (G-72).
func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapStatus, CapLog:
		var in commonIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid "+capability+" input: "+err.Error(), nil)
		}
		wd := in.Workdir
		if wd == "" {
			wd = "."
		}
		_, err := shell.ResolveWorkdir(workspace, wd)
		return err
	case CapDiff:
		var in diffIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid git.diff input: "+err.Error(), nil)
		}
		return validatePaths(workspace, in.Workdir, in.Paths)
	case CapAdd:
		var in addIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid git.add input: "+err.Error(), nil)
		}
		if len(in.Paths) == 0 {
			return result.Validation(result.CodeInvalidInput, "git.add paths must be non-empty", nil)
		}
		return validatePaths(workspace, in.Workdir, in.Paths)
	case CapCommit:
		var in commitIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid git.commit input: "+err.Error(), nil)
		}
		if strings.TrimSpace(in.Message) == "" {
			return result.Validation(result.CodeInvalidInput, "git.commit message must be non-empty", nil)
		}
		wd := in.Workdir
		if wd == "" {
			wd = "."
		}
		_, err := shell.ResolveWorkdir(workspace, wd)
		return err
	default:
		return result.Validation(result.CodeUnknownCapability, "git player cannot validate "+capability, nil)
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	switch req.Capability {
	case CapStatus:
		return p.execStatus(ctx, req)
	case CapDiff:
		return p.execDiff(ctx, req)
	case CapLog:
		return p.execLog(ctx, req)
	case CapAdd:
		return p.execAdd(ctx, req)
	case CapCommit:
		return p.execCommit(ctx, req)
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "git player cannot serve "+req.Capability, nil)
	}
}

func (p *Player) execStatus(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in commonIn
	_ = json.Unmarshal(req.Input, &in)
	workAbs, timeout, err := prep(req.Workspace, in.Workdir, in.TimeoutMS)
	if err != nil {
		return nil, err
	}
	branch, stderr, code, err := runGit(ctx, workAbs, timeout, nil, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, mapGitErr("git.status", err, code, stderr)
	}
	porc, stderr, code, err := runGit(ctx, workAbs, timeout, nil, "status", "--porcelain=v1")
	if err != nil {
		return nil, mapGitErr("git.status", err, code, stderr)
	}
	lines := splitNonEmpty(porc)
	out := statusOut{
		Branch:    strings.TrimSpace(branch),
		Porcelain: lines,
		Clean:     len(lines) == 0,
	}
	return json.Marshal(out)
}

func (p *Player) execDiff(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in diffIn
	_ = json.Unmarshal(req.Input, &in)
	workAbs, timeout, err := prep(req.Workspace, in.Workdir, in.TimeoutMS)
	if err != nil {
		return nil, err
	}
	args := []string{"diff"}
	if in.Staged {
		args = append(args, "--staged")
	}
	if len(in.Paths) > 0 {
		args = append(args, "--")
		args = append(args, in.Paths...)
	}
	diff, stderr, code, err := runGit(ctx, workAbs, timeout, nil, args...)
	if err != nil {
		return nil, mapGitErr("git.diff", err, code, stderr)
	}
	if len(diff) > maxDiffChars {
		diff = diff[:maxDiffChars]
	}
	return json.Marshal(diffOut{Diff: diff})
}

func (p *Player) execLog(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in logIn
	_ = json.Unmarshal(req.Input, &in)
	workAbs, timeout, err := prep(req.Workspace, in.Workdir, in.TimeoutMS)
	if err != nil {
		return nil, err
	}
	n := in.Max
	if n <= 0 {
		n = defaultLogMax
	}
	if n > maxLogMax {
		n = maxLogMax
	}
	raw, stderr, code, err := runGit(ctx, workAbs, timeout, nil,
		"log", "-n", strconv.Itoa(n), "--format=%H%x00%s%x00%an%x00%aI")
	if err != nil {
		return nil, mapGitErr("git.log", err, code, stderr)
	}
	out := logOut{Entries: []logEntry{}}
	for _, line := range strings.Split(strings.TrimSuffix(raw, "\n"), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\x00")
		if len(parts) < 4 {
			continue
		}
		out.Entries = append(out.Entries, logEntry{
			Hash: parts[0], Subject: parts[1], Author: parts[2], Date: parts[3],
		})
	}
	return json.Marshal(out)
}

func (p *Player) execAdd(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in addIn
	_ = json.Unmarshal(req.Input, &in)
	workAbs, timeout, err := prep(req.Workspace, in.Workdir, in.TimeoutMS)
	if err != nil {
		return nil, err
	}
	args := append([]string{"add", "--"}, in.Paths...)
	_, stderr, code, err := runGit(ctx, workAbs, timeout, nil, args...)
	if err != nil {
		return nil, mapGitErr("git.add", err, code, stderr)
	}
	return json.Marshal(addOut{Added: in.Paths, ExitCode: 0})
}

func (p *Player) execCommit(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in commitIn
	_ = json.Unmarshal(req.Input, &in)
	workAbs, timeout, err := prep(req.Workspace, in.Workdir, in.TimeoutMS)
	if err != nil {
		return nil, err
	}
	cfg := []string{
		"-c", "core.hooksPath=/dev/null",
		"-c", "user.name=runtgine",
		"-c", "user.email=runtgine@local",
	}
	args := append(append([]string{}, cfg...), "commit", "-m", in.Message)
	if in.AllowEmpty {
		args = append(args, "--allow-empty")
	}
	_, stderr, code, err := runGit(ctx, workAbs, timeout, nil, args...)
	if err != nil {
		return nil, mapGitErr("git.commit", err, code, stderr)
	}
	hash, stderr2, code2, err := runGit(ctx, workAbs, timeout, nil, "rev-parse", "HEAD")
	if err != nil {
		return nil, mapGitErr("git.commit", err, code2, stderr2)
	}
	return json.Marshal(commitOut{
		Commit:   strings.TrimSpace(hash),
		ExitCode: 0,
		Stderr:   stderr,
	})
}

func prep(workspace, workdir string, timeoutMS int) (workAbs string, timeout time.Duration, err error) {
	if workdir == "" {
		workdir = "."
	}
	workAbs, err = shell.ResolveWorkdir(workspace, workdir)
	if err != nil {
		return "", 0, err
	}
	if timeoutMS <= 0 {
		timeoutMS = defaultTimeoutMS
	}
	return workAbs, time.Duration(timeoutMS) * time.Millisecond, nil
}

func validatePaths(workspace, workdir string, paths []string) error {
	if workdir == "" {
		workdir = "."
	}
	workAbs, err := shell.ResolveWorkdir(workspace, workdir)
	if err != nil {
		return err
	}
	ws, err := filepath.Abs(workspace)
	if err != nil {
		return result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	if resolved, err := filepath.EvalSymlinks(ws); err == nil {
		ws = resolved
	}
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			return result.Validation(result.CodeInvalidInput, "path must be non-empty", nil)
		}
		if filepath.IsAbs(p) {
			return result.Validation(result.CodeInvalidInput, "path must be relative to workdir", map[string]any{"path": p})
		}
		target := filepath.Join(workAbs, p)
		target, err = filepath.Abs(target)
		if err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid path", nil)
		}
		resolved := target
		if r, err := filepath.EvalSymlinks(target); err == nil {
			resolved = r
		} else {
			// leaf may not exist yet (git add new file) — resolve parents
			resolved, err = evalExisting(target)
			if err != nil {
				return result.Validation(result.CodeInvalidInput, "invalid path: "+err.Error(), nil)
			}
		}
		rel, err := filepath.Rel(ws, resolved)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return result.Validation(result.CodeInvalidInput, "path must be inside workspace_root", map[string]any{
				"workspace": ws,
				"path":      resolved,
			})
		}
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

func runGit(ctx context.Context, workAbs string, timeout time.Duration, extraEnv []string, args ...string) (stdout string, stderr string, exitCode int, err error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "git", args...)
	cmd.Dir = workAbs
	cmd.Env = buildEnv(extraEnv)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if cctx.Err() == context.DeadlineExceeded {
		return stdout, stderr, -1, result.Runtime(result.CodeTimeout, "git timed out", true, map[string]any{
			"stderr": stderr,
		})
	}
	if cctx.Err() == context.Canceled {
		return stdout, stderr, -1, result.Runtime(result.CodeCancelled, "git cancelled", false, nil)
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout, stderr, ee.ExitCode(), runErr
		}
		return stdout, stderr, -1, result.Runtime(result.CodePlayerError, runErr.Error(), false, nil)
	}
	return stdout, stderr, 0, nil
}

func mapGitErr(capName string, err error, code int, stderr string) error {
	if re, ok := err.(result.Error); ok {
		return re
	}
	if code >= 0 {
		return result.Runtime(result.CodePlayerError, fmt.Sprintf("%s exited with code %d", capName, code), false, map[string]any{
			"exit_code": code,
			"stderr":    stderr,
		})
	}
	return err
}

func buildEnv(extra []string) []string {
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
	env = append(env, extra...)
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
	// Never inherit GIT_DIR / GIT_WORK_TREE / tokens
	return false
}

func splitNonEmpty(s string) []string {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}
