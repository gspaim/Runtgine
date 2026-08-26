// Package templates loads workspace Workflow Templates (G-194..G-200).
// Templates are JSON files that compile to Task IR; they are not Players.
package templates

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/task"
)

const SchemaVersion = "0.1.0"

const (
	maxSteps      = 20
	maxIDLen      = 64
	maxTitleRunes = 200
)

var idRE = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)

type Template struct {
	SchemaVersion string `json:"schema_version,omitempty"`
	ID            string `json:"id"`
	Title         string `json:"title"`
	Steps         []Step `json:"steps"`
}

type Step struct {
	StepID     string          `json:"step_id"`
	Capability string          `json:"capability"`
	Input      json.RawMessage `json:"input,omitempty"`
	DependsOn  []string        `json:"depends_on,omitempty"`
}

func Dir(workspace string) string {
	return filepath.Join(workspace, ".runtgine", "templates")
}

func Load(dir string) ([]Template, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	seen := map[string]string{}
	var out []Template
	var warnings []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			warnings = append(warnings, path+": "+err.Error())
			continue
		}
		tpl, err := parse(raw)
		if err != nil {
			warnings = append(warnings, path+": "+err.Error())
			continue
		}
		if prev, ok := seen[tpl.ID]; ok {
			warnings = append(warnings, path+": duplicate id "+tpl.ID+" (kept "+prev+")")
			continue
		}
		seen[tpl.ID] = e.Name()
		out = append(out, tpl)
	}
	return out, warnings, nil
}

func parse(raw []byte) (Template, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var tpl Template
	if err := dec.Decode(&tpl); err != nil {
		return Template{}, err
	}
	if err := validate(tpl); err != nil {
		return Template{}, err
	}
	return tpl, nil
}

func validate(tpl Template) error {
	if tpl.SchemaVersion != "" && tpl.SchemaVersion != SchemaVersion {
		return result.Validation(result.CodeInvalidInput, "template schema_version must be "+SchemaVersion, map[string]any{"schema_version": tpl.SchemaVersion})
	}
	if !idRE.MatchString(tpl.ID) {
		return result.Validation(result.CodeInvalidInput, "template id must match [a-zA-Z0-9._-]{1,64}", map[string]any{"id": tpl.ID})
	}
	title := strings.TrimSpace(tpl.Title)
	if title == "" || utf8.RuneCountInString(title) > maxTitleRunes {
		return result.Validation(result.CodeInvalidInput, "template title is required (max 200 runes)", nil)
	}
	if n := len(tpl.Steps); n < 1 || n > maxSteps {
		return result.Validation(result.CodeInvalidInput, "template must have 1–20 steps", map[string]any{"steps": n})
	}
	seen := map[string]bool{}
	for _, s := range tpl.Steps {
		if strings.TrimSpace(s.StepID) == "" {
			return result.Validation(result.CodeInvalidInput, "step_id is required", nil)
		}
		if seen[s.StepID] {
			return result.Validation(result.CodeInvalidInput, "duplicate step_id "+s.StepID, nil)
		}
		seen[s.StepID] = true
		if strings.TrimSpace(s.Capability) == "" || !strings.Contains(s.Capability, ".") {
			return result.Validation(result.CodeInvalidInput, "capability must be domain.action", map[string]any{"step_id": s.StepID})
		}
		if len(s.Input) > 0 && !json.Valid(s.Input) {
			return result.Validation(result.CodeInvalidInput, "step input is not JSON", map[string]any{"step_id": s.StepID})
		}
		if len(s.Input) > 0 {
			var obj map[string]any
			if err := json.Unmarshal(s.Input, &obj); err != nil {
				return result.Validation(result.CodeInvalidInput, "step input must be an object", map[string]any{"step_id": s.StepID})
			}
		}
	}
	for _, s := range tpl.Steps {
		for _, d := range s.DependsOn {
			if !seen[d] {
				return result.Validation(result.CodeInvalidInput, "depends_on unknown step "+d, map[string]any{"step_id": s.StepID})
			}
			if d == s.StepID {
				return result.Validation(result.CodeInvalidInput, "step depends on itself", map[string]any{"step_id": s.StepID})
			}
		}
	}
	if err := detectCycle(tpl.Steps); err != nil {
		return err
	}
	_ = maxIDLen
	return nil
}

func detectCycle(steps []Step) error {
	indeg := map[string]int{}
	adj := map[string][]string{}
	for _, s := range steps {
		if _, ok := indeg[s.StepID]; !ok {
			indeg[s.StepID] = 0
		}
		for _, d := range s.DependsOn {
			adj[d] = append(adj[d], s.StepID)
			indeg[s.StepID]++
		}
	}
	var q []string
	for id, n := range indeg {
		if n == 0 {
			q = append(q, id)
		}
	}
	seen := 0
	for len(q) > 0 {
		id := q[0]
		q = q[1:]
		seen++
		for _, nxt := range adj[id] {
			indeg[nxt]--
			if indeg[nxt] == 0 {
				q = append(q, nxt)
			}
		}
	}
	if seen != len(steps) {
		return result.Validation(result.CodeInvalidInput, "template steps contain a cycle", nil)
	}
	return nil
}

func Lookup(list []Template, id string) (Template, bool) {
	id = strings.TrimSpace(id)
	for _, t := range list {
		if t.ID == id {
			return t, true
		}
	}
	return Template{}, false
}

func Compile(tpl Template, entryPoint, ref, summary string) (task.Task, error) {
	if err := validate(tpl); err != nil {
		return task.Task{}, err
	}
	tid, err := task.NewID()
	if err != nil {
		return task.Task{}, err
	}
	if entryPoint == "" {
		entryPoint = "cli"
	}
	steps := make([]task.Step, 0, len(tpl.Steps))
	for _, s := range tpl.Steps {
		in := s.Input
		if len(bytes.TrimSpace(in)) == 0 {
			in = json.RawMessage(`{}`)
		}
		steps = append(steps, task.Step{
			StepID:     s.StepID,
			Capability: s.Capability,
			Input:      in,
			DependsOn:  append([]string{}, s.DependsOn...),
		})
	}
	sum := strings.TrimSpace(summary)
	if sum == "" {
		sum = tpl.Title
	}
	return task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        tid,
		CreatedAt:     time.Now().UTC(),
		Source:        task.Source{EntryPoint: entryPoint, Ref: ref},
		Intent:        task.Intent{Summary: tpl.Title, Notes: sum},
		Steps:         steps,
		Metadata:      map[string]any{"template": tpl.ID},
	}, nil
}
