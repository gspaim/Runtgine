package task

import (
	"fmt"
	"strings"
)

var validEntryPoints = map[string]bool{
	"cli": true, "tui": true, "board": true, "api": true, "http": true, "other": true,
}

// IdentityValidate checks schema_version and UUID v7 task_id after Parse.
func IdentityValidate(t Task) error {
	if t.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %q", SchemaVersion)
	}
	if t.TaskID == "" {
		return fmt.Errorf("task_id is required")
	}
	if !IsUUIDv7(t.TaskID) {
		return fmt.Errorf("task_id must be a UUID v7")
	}
	return nil
}

// StructuralValidate checks Task IR shape before registry/capability checks.
func StructuralValidate(t Task) error {
	if t.Intent.Summary == "" {
		return fmt.Errorf("intent.summary is required")
	}
	if !validEntryPoints[t.Source.EntryPoint] {
		return fmt.Errorf("source.entry_point must be one of cli|tui|board|api|http|other")
	}
	if len(t.Steps) == 0 {
		return fmt.Errorf("steps must contain at least one step")
	}

	seen := map[string]bool{}
	for _, s := range t.Steps {
		if s.StepID == "" {
			return fmt.Errorf("step_id is required")
		}
		if seen[s.StepID] {
			return fmt.Errorf("duplicate step_id %q", s.StepID)
		}
		seen[s.StepID] = true
		if s.Capability == "" {
			return fmt.Errorf("step %q: capability is required", s.StepID)
		}
		if !isCapabilityName(s.Capability) {
			return fmt.Errorf("step %q: invalid capability %q", s.StepID, s.Capability)
		}
		if len(s.Input) == 0 {
			return fmt.Errorf("step %q: input is required", s.StepID)
		}
	}

	// depends_on references + cycle check (Kahn)
	for _, s := range t.Steps {
		for _, dep := range s.DependsOn {
			if !seen[dep] {
				return fmt.Errorf("step %q: depends_on unknown step %q", s.StepID, dep)
			}
			if dep == s.StepID {
				return fmt.Errorf("step %q: depends on itself", s.StepID)
			}
		}
	}
	if err := detectCycle(t.Steps); err != nil {
		return err
	}
	return nil
}

func isCapabilityName(c string) bool {
	parts := strings.Split(c, ".")
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for i, r := range p {
			ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
			if i == 0 && (r < 'a' || r > 'z') {
				return false
			}
			if !ok {
				return false
			}
		}
	}
	return true
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
	visited := 0
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		visited++
		for _, n := range adj[cur] {
			indeg[n]--
			if indeg[n] == 0 {
				q = append(q, n)
			}
		}
	}
	if visited != len(steps) {
		return fmt.Errorf("steps contains a dependency cycle")
	}
	return nil
}

// TopoOrder returns step_ids in topological order.
func TopoOrder(steps []Step) ([]string, error) {
	if err := detectCycle(steps); err != nil {
		return nil, err
	}
	byID := map[string]Step{}
	indeg := map[string]int{}
	adj := map[string][]string{}
	for _, s := range steps {
		byID[s.StepID] = s
		indeg[s.StepID] = 0
	}
	for _, s := range steps {
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
	var out []string
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		out = append(out, cur)
		for _, n := range adj[cur] {
			indeg[n]--
			if indeg[n] == 0 {
				q = append(q, n)
			}
		}
	}
	return out, nil
}
