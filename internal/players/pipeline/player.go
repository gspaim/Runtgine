package pipeline

import (
	"bufio"
	"context"
	"encoding/json"
	"io/fs"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/players/llm"
	"github.com/google/uuid"
)

// Player implements deterministic pipeline.* stages (G-23).
type Player struct {
	Refine llm.Completer // optional LLM refine for decompose
}

func New() *Player { return &Player{} }

func NewWithRefine(c llm.Completer) *Player { return &Player{Refine: c} }

func (p *Player) Manifest() registry.Manifest {
	obj := json.RawMessage(`{"type":"object","additionalProperties":false}`)
	caps := []registry.Capability{
		{Name: corepipe.CapRepoSearch, InputSchema: obj, OutputSchema: json.RawMessage(`{"type":"object","required":["paths","symbols"]}`)},
		{Name: corepipe.CapEffort, InputSchema: obj, OutputSchema: json.RawMessage(`{"type":"object","required":["effort","rationale"]}`)},
		{Name: corepipe.CapDifficulty, InputSchema: obj, OutputSchema: json.RawMessage(`{"type":"object","required":["difficulty","rationale"]}`)},
		{Name: corepipe.CapDecompose, InputSchema: obj, OutputSchema: json.RawMessage(`{"type":"object","required":["subtasks"]}`)},
	}
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "pipeline",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities:  caps,
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	_ = ctx
	switch req.Capability {
	case corepipe.CapRepoSearch:
		return searchRepo(req.Workspace)
	case corepipe.CapEffort:
		return estimateEffort(req.PriorOutputs)
	case corepipe.CapDifficulty:
		return classifyDifficulty(req.PriorOutputs)
	case corepipe.CapDecompose:
		raw, err := decompose(req.PriorOutputs)
		if err != nil || p.Refine == nil {
			return raw, err
		}
		return p.refineDecompose(ctx, req, raw)
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "pipeline player cannot serve "+req.Capability, nil)
	}
}

var skipDir = map[string]bool{
	".git": true, ".runtgine": true, "node_modules": true, "vendor": true,
	"bin": true, "dist": true, ".idea": true,
}

var goDecl = regexp.MustCompile(`^(func|type|var|const)\s+(\(?\w+\)?\s+)?(\w+)`)

func searchRepo(root string) (json.RawMessage, error) {
	var paths, symbols []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if skipDir[name] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		if strings.HasPrefix(rel, ".") && !strings.HasPrefix(rel, ".specs") {
			return nil
		}
		paths = append(paths, rel)
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			symbols = append(symbols, extractGoSymbols(path)...)
		}
		if len(paths) >= 200 {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil && err != fs.SkipAll {
		return nil, result.Runtime(result.CodePlayerError, err.Error(), false, nil)
	}
	return json.Marshal(map[string]any{"paths": paths, "symbols": cap(symbols, 80)})
}

func extractGoSymbols(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		m := goDecl.FindStringSubmatch(line)
		if len(m) >= 4 && m[3] != "" && m[1] != "var" && m[1] != "const" {
			out = append(out, m[1]+" "+m[3])
		}
		if len(out) >= 20 {
			break
		}
	}
	return out
}

func estimateEffort(priors []store.StepOutput) (json.RawMessage, error) {
	nFiles, nSym := repoCounts(priors)
	effort := "S"
	switch {
	case nFiles > 80 || nSym > 120:
		effort = "XL"
	case nFiles > 30 || nSym > 50:
		effort = "L"
	case nFiles > 10 || nSym > 15:
		effort = "M"
	}
	return json.Marshal(map[string]any{
		"effort":    effort,
		"rationale": fmt.Sprintf("heuristic from repo-search: files=%d symbols=%d", nFiles, nSym),
	})
}

func classifyDifficulty(priors []store.StepOutput) (json.RawMessage, error) {
	effort := priorString(priors, corepipe.CapEffort, "effort")
	risks := priorLen(priors, corepipe.CapTechReview, "risks")
	diff := 2
	switch effort {
	case "M":
		diff = 3
	case "L":
		diff = 4
	case "XL":
		diff = 5
	}
	if risks >= 3 && diff < 5 {
		diff++
	}
	return json.Marshal(map[string]any{
		"difficulty": diff,
		"rationale":  fmt.Sprintf("heuristic: effort=%s tech-review risks=%d", effort, risks),
	})
}

func decompose(priors []store.StepOutput) (json.RawMessage, error) {
	findings := priorStringSlice(priors, corepipe.CapTechReview, "findings")
	paths := priorStringSlice(priors, corepipe.CapRepoSearch, "paths")
	var subs []map[string]string
	for i, f := range findings {
		if i >= 5 {
			break
		}
		id, _ := uuid.NewV7()
		subs = append(subs, map[string]string{
			"subtask_id":           id.String(),
			"summary":              "Address finding: " + f,
			"suggested_capability": "shell.exec",
			"notes":                "from pipeline.tech-review",
		})
	}
	if len(subs) == 0 {
		for i, p := range paths {
			if i >= 3 {
				break
			}
			id, _ := uuid.NewV7()
			subs = append(subs, map[string]string{
				"subtask_id":           id.String(),
				"summary":              "Inspect " + p,
				"suggested_capability": "shell.exec",
				"notes":                "from pipeline.repo-search",
			})
		}
	}
	if len(subs) == 0 {
		id, _ := uuid.NewV7()
		subs = append(subs, map[string]string{
			"subtask_id":           id.String(),
			"summary":              "Follow up on task intent",
			"suggested_capability": "shell.exec",
			"notes":                "fallback decompose",
		})
	}
	return json.Marshal(map[string]any{"subtasks": subs, "refined": false})
}

func (p *Player) refineDecompose(ctx context.Context, req registry.ExecRequest, base json.RawMessage) (json.RawMessage, error) {
	var pack contextpack.Pack
	if len(req.Context) > 0 {
		_ = json.Unmarshal(req.Context, &pack)
	}
	out, err := p.Refine.Complete(ctx, pack, json.RawMessage(`{"type":"object","required":["subtasks"]}`))
	if err != nil || !json.Valid(out) {
		return base, nil
	}
	var parsed struct {
		Subtasks []map[string]any `json:"subtasks"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil || len(parsed.Subtasks) == 0 {
		return base, nil
	}
	return json.Marshal(map[string]any{"subtasks": parsed.Subtasks, "refined": true})
}

func repoCounts(priors []store.StepOutput) (int, int) {
	return priorLen(priors, corepipe.CapRepoSearch, "paths"), priorLen(priors, corepipe.CapRepoSearch, "symbols")
}

func priorLen(priors []store.StepOutput, cap, field string) int {
	return len(priorStringSlice(priors, cap, field))
}

func priorStringSlice(priors []store.StepOutput, cap, field string) []string {
	for _, o := range priors {
		if o.Capability != cap {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(o.Output, &m); err != nil {
			continue
		}
		raw, ok := m[field]
		if !ok {
			continue
		}
		arr, ok := raw.([]any)
		if !ok {
			continue
		}
		out := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func priorString(priors []store.StepOutput, cap, field string) string {
	for _, o := range priors {
		if o.Capability != cap {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(o.Output, &m); err != nil {
			continue
		}
		if s, ok := m[field].(string); ok {
			return s
		}
	}
	return ""
}

func cap(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return in[:n]
}
