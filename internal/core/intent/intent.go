package intent

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode"

	"github.com/gspaim/Runtgine/internal/core/contextpack"
	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/task"
	"github.com/gspaim/Runtgine/internal/players/llm"
)

const (
	MethodHeuristicShell    = "heuristic.shell"
	MethodHeuristicPipeline = "heuristic.pipeline"
	MethodLLM               = "llm"
)

// Engine compiles natural language into Task IR v0 (G-50..G-53).
// It is not a Player and never bypasses the Validator.
type Engine struct {
	Completer llm.Completer
}

func New(c llm.Completer) *Engine {
	if c == nil {
		c = llm.HeuristicCompleter{}
	}
	return &Engine{Completer: c}
}

type Request struct {
	Text       string
	EntryPoint string
	Ref        string
}

type CompileResult struct {
	Task   task.Task
	Method string
}

func (e *Engine) Compile(ctx context.Context, req Request) (CompileResult, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return CompileResult{}, result.Validation(result.CodeInvalidInput, "intent text is required", nil)
	}
	ep := req.EntryPoint
	if ep == "" {
		ep = "cli"
	}

	if argv, ok := matchShell(text); ok {
		tk, err := shellTask(text, argv, ep, req.Ref)
		if err != nil {
			return CompileResult{}, err
		}
		return CompileResult{Task: tk, Method: MethodHeuristicShell}, nil
	}
	if looksLikePipeline(text) {
		tk, err := corepipe.NewTaskIR(summarize(text), text, ep, req.Ref)
		if err != nil {
			return CompileResult{}, err
		}
		return CompileResult{Task: tk, Method: MethodHeuristicPipeline}, nil
	}

	tk, err := e.compileLLM(ctx, text, ep, req.Ref)
	if err != nil {
		return CompileResult{}, err
	}
	return CompileResult{Task: tk, Method: MethodLLM}, nil
}

type llmOut struct {
	Summary      string   `json:"summary"`
	Notes        string   `json:"notes"`
	Route        string   `json:"route"` // shell | pipeline
	ShellCommand []string `json:"shell_command"`
}

func (e *Engine) compileLLM(ctx context.Context, text, ep, ref string) (task.Task, error) {
	schema := json.RawMessage(`{
  "type":"object",
  "required":["summary","route"],
  "properties":{
    "summary":{"type":"string"},
    "notes":{"type":"string"},
    "route":{"type":"string","enum":["shell","pipeline"]},
    "shell_command":{"type":"array","items":{"type":"string"},"minItems":1}
  },
  "additionalProperties":false
}`)
	pack := contextpack.Pack{
		Task: contextpack.TaskView{Summary: text, Notes: "intent.compile"},
		Step: contextpack.StepView{StepID: "intent", Capability: "intent.compile"},
	}

	var last error
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := e.Completer.Complete(ctx, pack, schema)
		if err != nil {
			last = err
			continue
		}
		raw = json.RawMessage(strings.TrimSpace(string(raw)))
		var out llmOut
		if err := json.Unmarshal(raw, &out); err != nil {
			last = result.Runtime(result.CodePlayerError, "intent LLM returned non-JSON object", true, nil)
			continue
		}
		summary := strings.TrimSpace(out.Summary)
		if summary == "" {
			summary = summarize(text)
		}
		notes := out.Notes
		if notes == "" {
			notes = text
		}
		switch strings.ToLower(strings.TrimSpace(out.Route)) {
		case "pipeline":
			return corepipe.NewTaskIR(summary, notes, ep, ref)
		case "shell":
			cmd := out.ShellCommand
			if len(cmd) == 0 || strings.TrimSpace(cmd[0]) == "" {
				cmd = []string{"echo", summary}
			}
			return shellTask(summary, cmd, ep, ref)
		default:
			last = result.Runtime(result.CodePlayerError, "intent LLM route must be shell|pipeline", true, nil)
		}
	}
	if last == nil {
		last = result.Runtime(result.CodePlayerError, "intent compile failed", true, nil)
	}
	return task.Task{}, last
}

func shellTask(summary string, argv []string, ep, ref string) (task.Task, error) {
	id, err := task.NewID()
	if err != nil {
		return task.Task{}, err
	}
	in, err := json.Marshal(map[string]any{
		"command":    argv,
		"workdir":    ".",
		"timeout_ms": 60000,
	})
	if err != nil {
		return task.Task{}, err
	}
	return task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id,
		CreatedAt:     time.Now().UTC(),
		Source:        task.Source{EntryPoint: ep, Ref: ref},
		Intent:        task.Intent{Summary: summarize(summary), Notes: summary},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "shell.exec",
			Input:      in,
		}},
		Metadata: map[string]any{"intent_engine": "v0"},
	}, nil
}

func matchShell(text string) ([]string, bool) {
	lower := strings.ToLower(text)
	prefixes := []string{"run ", "exec ", "execute ", "echo ", "go ", "make ", "npm ", "yarn ", "pnpm ", "cargo ", "./", "$ "}
	for _, p := range prefixes {
		if !strings.HasPrefix(lower, p) {
			continue
		}
		rest := strings.TrimSpace(text)
		switch {
		case strings.HasPrefix(lower, "run "):
			rest = strings.TrimSpace(text[4:])
		case strings.HasPrefix(lower, "exec "):
			rest = strings.TrimSpace(text[5:])
		case strings.HasPrefix(lower, "execute "):
			rest = strings.TrimSpace(text[8:])
		case strings.HasPrefix(lower, "$ "):
			rest = strings.TrimSpace(text[2:])
		}
		argv := splitARGV(rest)
		if len(argv) == 0 {
			return nil, false
		}
		return argv, true
	}
	argv := splitARGV(text)
	if len(argv) >= 1 && looksLikeArgvLine(text, argv[0]) {
		return argv, true
	}
	return nil, false
}

func looksLikeArgvLine(text, first string) bool {
	if strings.ContainsAny(text, "?!.;") {
		return false
	}
	if strings.Contains(text, " ") && len(strings.Fields(text)) > 8 {
		return false
	}
	switch strings.ToLower(first) {
	case "echo", "go", "make", "npm", "yarn", "pnpm", "cargo", "git", "ls", "pwd", "cat", "true", "false", "printf", "env", "date":
		return true
	}
	return strings.HasPrefix(first, "./") || strings.HasPrefix(first, "/")
}

func looksLikePipeline(text string) bool {
	lower := strings.ToLower(text)
	keys := []string{
		"review", "revisa", "revisar", "analisa", "analisar", "analyze", "analysis",
		"decompose", "decompor", "pipeline", "estima", "estimar", "effort",
		"difficulty", "dificuldade", "board", "spec", "arquitetura", "architecture",
	}
	for _, k := range keys {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

func splitARGV(s string) []string {
	var out []string
	var b strings.Builder
	inQuote := rune(0)
	for _, r := range s {
		switch {
		case inQuote != 0:
			if r == inQuote {
				inQuote = 0
			} else {
				b.WriteRune(r)
			}
		case r == '"' || r == '\'':
			inQuote = r
		case unicode.IsSpace(r):
			if b.Len() > 0 {
				out = append(out, b.String())
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

func summarize(text string) string {
	text = strings.TrimSpace(text)
	if len(text) <= 120 {
		return text
	}
	return text[:117] + "..."
}
