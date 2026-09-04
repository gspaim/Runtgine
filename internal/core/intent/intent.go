package intent

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode"

	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"github.com/gspaim/Runtgine/internal/core/graph"
	"github.com/gspaim/Runtgine/internal/core/memory"
	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/task"
	"github.com/gspaim/Runtgine/internal/core/templates"
	"github.com/gspaim/Runtgine/internal/players/llm"
)

const (
	MethodHeuristicShell    = "heuristic.shell"
	MethodHeuristicPipeline = "heuristic.pipeline"
	MethodHeuristicTest     = "heuristic.test"
	MethodHeuristicGit      = "heuristic.git"
	MethodHeuristicDocker   = "heuristic.docker"
	MethodHeuristicNPM      = "heuristic.npm"
	MethodHeuristicPytest   = "heuristic.pytest"
	MethodHeuristicYarn     = "heuristic.yarn"
	MethodHeuristicK8s      = "heuristic.k8s"
	MethodHeuristicTF       = "heuristic.tf"
	MethodHeuristicPG       = "heuristic.pg"
	MethodHeuristicHelm     = "heuristic.helm"
	MethodHeuristicTemplate = "heuristic.template"
	MethodLLM               = "llm"
)

// HitsQuerier is optional Graph access for the LLM compile path (G-69).
type HitsQuerier interface {
	QueryHits(ctx context.Context, q graph.Query) graph.Hits
}

// MemoryQuerier is optional Project Memory for the LLM compile path (G-126).
type MemoryQuerier interface {
	Query(ctx context.Context, text string, limit int) ([]memory.Hit, error)
}

// Engine compiles natural language into Task IR v0 (G-50..G-53).
// It is not a Player and never bypasses the Validator.
type Engine struct {
	Completer llm.Completer
	Graph     HitsQuerier   // optional; only used on LLM path
	Memory    MemoryQuerier // optional; only used on LLM path
	Templates []templates.Template
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

	if hit, ok := matchPlayer(text); ok {
		extra := map[string]any{}
		for k, v := range hit.extra {
			extra[k] = v
		}
		if hit.method == MethodHeuristicGit || hit.method == MethodHeuristicNPM || hit.method == MethodHeuristicPytest || hit.method == MethodHeuristicYarn || hit.method == MethodHeuristicTF {
			extra["workdir"] = "."
		}
		tk, err := playerTask(text, hit.capability, ep, req.Ref, extra)
		if err != nil {
			return CompileResult{}, err
		}
		return CompileResult{Task: tk, Method: hit.method}, nil
	}
	if tpl, ok, known := matchTemplate(text, e.Templates); known {
		if !ok {
			return CompileResult{}, result.Validation(result.CodeInvalidInput, "unknown template", map[string]any{"text": text})
		}
		tk, err := templates.Compile(tpl, ep, req.Ref, text)
		if err != nil {
			return CompileResult{}, err
		}
		return CompileResult{Task: tk, Method: MethodHeuristicTemplate}, nil
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
		Budget: contextpack.Budget{
			MaxChars:       contextpack.DefaultMaxChars,
			MaxFiles:       contextpack.DefaultMaxFiles,
			GraphMaxHits:   contextpack.DefaultGraphMaxHits,
			GraphMaxChars:  contextpack.DefaultGraphMaxChars,
			MemoryMaxHits:  contextpack.DefaultMemoryMaxHits,
			MemoryMaxChars: contextpack.DefaultMemoryMaxChars,
		},
		GraphHits:  contextpack.GraphHits{Items: []contextpack.GraphHit{}},
		MemoryHits: contextpack.MemoryHits{Items: []contextpack.MemoryHit{}},
	}
	if e.Graph != nil {
		hits := e.Graph.QueryHits(ctx, graph.Query{
			Text:     text,
			Limit:    pack.Budget.GraphMaxHits,
			MaxChars: pack.Budget.GraphMaxChars,
		})
		items := make([]contextpack.GraphHit, 0, len(hits.Items))
		for _, h := range hits.Items {
			items = append(items, contextpack.GraphHit{
				Kind: h.Kind, ID: h.ID, Reason: h.Reason, Score: h.Score,
			})
		}
		pack = contextpack.WithGraphHits(pack, items)
		pack = contextpack.WithSeededRepoHits(pack, pack.GraphHits.Items)
	}
	if e.Memory != nil {
		hits, err := e.Memory.Query(ctx, text, pack.Budget.MemoryMaxHits)
		if err != nil {
			pack = contextpack.WithMemoryHits(pack, nil)
		} else {
			items := make([]contextpack.MemoryHit, 0, len(hits))
			for _, h := range hits {
				items = append(items, contextpack.MemoryHit{
					ID:       h.ID,
					Kind:     h.Kind,
					Validity: h.Validity,
					Title:    h.Title,
					Snippet:  memory.Snippet(h.Body),
					Score:    h.Score,
				})
			}
			pack = contextpack.WithMemoryHits(pack, items)
		}
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

func playerTask(summary, capability, ep, ref string, input map[string]any) (task.Task, error) {
	id, err := task.NewID()
	if err != nil {
		return task.Task{}, err
	}
	if input == nil {
		input = map[string]any{}
	}
	raw, err := json.Marshal(input)
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
			Capability: capability,
			Input:      raw,
		}},
		Metadata: map[string]any{"intent_engine": "v0"},
	}, nil
}

type playerHit struct {
	capability string
	method     string
	extra      map[string]any
}

func matchPlayer(text string) (playerHit, bool) {
	n := normalizeNL(text)
	if hit, ok := matchK8s(n); ok {
		return hit, true
	}
	if hit, ok := matchHelm(n); ok {
		return hit, true
	}
	switch {
	case hasPhrase(n, "terraform validate"):
		return playerHit{capability: "tf.validate", method: MethodHeuristicTF}, true
	case hasPhrase(n, "terraform plan"):
		return playerHit{capability: "tf.plan", method: MethodHeuristicTF}, true
	case hasPhrase(n, "pg ping"), hasPhrase(n, "postgres ping"), hasPhrase(n, "psql ping"):
		return playerHit{capability: "pg.ping", method: MethodHeuristicPG, extra: map[string]any{"dbname": "postgres"}}, true
	case hasPhrase(n, "npm test"), hasPhrase(n, "npm run test"), hasPhrase(n, "roda os testes npm"), hasPhrase(n, "run npm tests"):
		return playerHit{capability: "npm.test", method: MethodHeuristicNPM}, true
	case hasPhrase(n, "pytest"), hasPhrase(n, "roda pytest"), hasPhrase(n, "run pytest"), hasPhrase(n, "rodar pytest"):
		return playerHit{capability: "pytest.run", method: MethodHeuristicPytest}, true
	case hasPhrase(n, "yarn test"), hasPhrase(n, "yarn run test"), hasPhrase(n, "rodar yarn test"):
		return playerHit{capability: "yarn.test", method: MethodHeuristicYarn}, true
	case hasPhrase(n, "go test"), hasPhrase(n, "roda os testes"), hasPhrase(n, "rodar testes"), hasPhrase(n, "run tests"):
		return playerHit{capability: "test.go", method: MethodHeuristicTest}, true
	case hasPhrase(n, "git status"):
		return playerHit{capability: "git.status", method: MethodHeuristicGit}, true
	case hasPhrase(n, "git diff"):
		return playerHit{capability: "git.diff", method: MethodHeuristicGit}, true
	case hasPhrase(n, "git log"):
		return playerHit{capability: "git.log", method: MethodHeuristicGit}, true
	case hasPhrase(n, "docker ps"):
		return playerHit{capability: "docker.ps", method: MethodHeuristicDocker}, true
	}
	return playerHit{}, false
}

func matchK8s(n string) (playerHit, bool) {
	for _, p := range []string{"kubectl get ", "k8s get "} {
		if !strings.HasPrefix(n, p) {
			continue
		}
		rest := strings.Fields(n[len(p):])
		if len(rest) == 0 || strings.HasPrefix(rest[0], "-") {
			return playerHit{}, false
		}
		extra := map[string]any{"resource": rest[0]}
		if len(rest) >= 2 && !strings.HasPrefix(rest[1], "-") {
			extra["name"] = rest[1]
			return playerHit{capability: "k8s.get", method: MethodHeuristicK8s, extra: extra}, true
		}
		return playerHit{capability: "k8s.list", method: MethodHeuristicK8s, extra: extra}, true
	}
	return playerHit{}, false
}

// matchHelm matches high-confidence read-only helm invocations (spec 42).
// install/upgrade/rollback/uninstall never match here.
func matchHelm(n string) (playerHit, bool) {
	singleArg := func(prefix string) (string, bool) {
		if !strings.HasPrefix(n, prefix) {
			return "", false
		}
		fields := strings.Fields(n[len(prefix):])
		if len(fields) != 1 || strings.HasPrefix(fields[0], "-") {
			return "", false
		}
		return fields[0], true
	}
	if chart, ok := singleArg("helm lint "); ok {
		return playerHit{capability: "helm.lint", method: MethodHeuristicHelm, extra: map[string]any{"chart": chart}}, true
	}
	if chart, ok := singleArg("helm template "); ok {
		return playerHit{capability: "helm.template", method: MethodHeuristicHelm, extra: map[string]any{"chart": chart}}, true
	}
	if release, ok := singleArg("helm status "); ok {
		return playerHit{capability: "helm.status", method: MethodHeuristicHelm, extra: map[string]any{"release": release}}, true
	}
	if strings.HasPrefix(n, "helm list") && strings.TrimSpace(n[len("helm list"):]) == "" {
		return playerHit{capability: "helm.list", method: MethodHeuristicHelm}, true
	}
	return playerHit{}, false
}

func matchTemplate(text string, list []templates.Template) (templates.Template, bool, bool) {
	n := normalizeNL(text)
	prefixes := []string{
		"run template ",
		"roda o template ",
		"rodar template ",
		"roda template ",
		"template ",
	}
	for _, p := range prefixes {
		if !strings.HasPrefix(n, p) {
			continue
		}
		id := strings.TrimSpace(n[len(p):])
		if id == "" {
			return templates.Template{}, false, true
		}
		if tpl, ok := templates.Lookup(list, id); ok {
			return tpl, true, true
		}
		return templates.Template{}, false, true
	}
	return templates.Template{}, false, false
}

func normalizeNL(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func hasPhrase(normalized, phrase string) bool {
	return strings.Contains(" "+normalized+" ", " "+phrase+" ")
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
