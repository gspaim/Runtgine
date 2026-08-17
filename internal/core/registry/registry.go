package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Kind string

const (
	KindDeterministic Kind = "deterministic"
	KindAI            Kind = "ai"
	KindHuman         Kind = "human"
	KindService       Kind = "service"
	KindWorkflow      Kind = "workflow"
)

type Manifest struct {
	SchemaVersion string       `json:"schema_version"`
	Name          string       `json:"name"`
	Version       string       `json:"version"`
	Kind          Kind         `json:"kind"`
	Capabilities  []Capability `json:"capabilities"`
}

type Capability struct {
	Name         string          `json:"name"`
	InputSchema  json.RawMessage `json:"input_schema"`
	OutputSchema json.RawMessage `json:"output_schema"`
}

type ExecRequest struct {
	Capability   string
	Input        json.RawMessage
	Workspace    string
	RunID        string
	TaskID       string
	StepID       string
	Context      json.RawMessage
	PriorOutputs []store.StepOutput
}

// Player executes a capability.
type Player interface {
	Manifest() Manifest
	Execute(ctx context.Context, req ExecRequest) (json.RawMessage, error)
}

type Registry struct {
	mu        sync.RWMutex
	players   map[string]Player
	caps      map[string][]string
	meta      map[string]Capability
	kinds     map[string]Kind
	inputSch  map[string]*jsonschema.Schema
	compilerN int
}

func New() *Registry {
	return &Registry{
		players:  map[string]Player{},
		caps:     map[string][]string{},
		meta:     map[string]Capability{},
		kinds:    map[string]Kind{},
		inputSch: map[string]*jsonschema.Schema{},
	}
}

func (r *Registry) Register(p Player) error {
	m := p.Manifest()
	if m.Name == "" {
		return fmt.Errorf("player manifest name is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.players[m.Name]; ok {
		return fmt.Errorf("player %q already registered", m.Name)
	}
	compiled := map[string]*jsonschema.Schema{}
	for _, c := range m.Capabilities {
		sch, err := compileInputSchema(c.Name, c.InputSchema, &r.compilerN)
		if err != nil {
			return fmt.Errorf("player %q capability %q: %w", m.Name, c.Name, err)
		}
		compiled[c.Name] = sch
	}
	r.players[m.Name] = p
	r.kinds[m.Name] = m.Kind
	for _, c := range m.Capabilities {
		r.caps[c.Name] = append(r.caps[c.Name], m.Name)
		if _, ok := r.meta[c.Name]; !ok {
			r.meta[c.Name] = c
			r.inputSch[c.Name] = compiled[c.Name]
		}
	}
	return nil
}

func compileInputSchema(capName string, raw json.RawMessage, seq *int) (*jsonschema.Schema, error) {
	if len(raw) == 0 {
		raw = json.RawMessage(`{"type":"object"}`)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("input_schema: %w", err)
	}
	*seq++
	url := fmt.Sprintf("https://runtgine.dev/schemas/capability/%s/%d.json", capName, *seq)
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	c.AssertFormat()
	if err := c.AddResource(url, doc); err != nil {
		return nil, err
	}
	return c.Compile(url)
}

func (r *Registry) HasCapability(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.caps[name]) > 0
}

// ValidateInput checks raw step input against the capability input_schema.
func (r *Registry) ValidateInput(capability string, input json.RawMessage) error {
	r.mu.RLock()
	sch := r.inputSch[capability]
	r.mu.RUnlock()
	if sch == nil {
		return result.Validation(result.CodeUnknownCapability,
			fmt.Sprintf("capability %q is not registered", capability), nil)
	}
	if len(input) == 0 {
		return result.Validation(result.CodeInvalidInput, "input is required", nil)
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(input))
	if err != nil {
		return result.Validation(result.CodeInvalidInput, "invalid input json: "+err.Error(), nil)
	}
	if err := sch.Validate(inst); err != nil {
		msg := strings.ReplaceAll(err.Error(), "\n", "; ")
		if len(msg) > 512 {
			msg = msg[:512] + "…"
		}
		return result.Validation(result.CodeInvalidInput,
			fmt.Sprintf("input does not match schema for %q: %s", capability, msg), nil)
	}
	return nil
}

// Resolve picks a player for a capability (G-26 router rules).
func (r *Registry) Resolve(capability, llmBackendDefault string) (string, Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := r.caps[capability]
	if len(names) == 0 {
		return "", nil, result.Validation(result.CodeUnknownCapability,
			fmt.Sprintf("capability %q is not registered", capability), nil)
	}
	for _, n := range names {
		if r.kinds[n] == KindDeterministic {
			return n, r.players[n], nil
		}
	}
	if llmBackendDefault != "" {
		for _, n := range names {
			if n == "llm" || n == llmBackendDefault || n == "llm-"+llmBackendDefault {
				return n, r.players[n], nil
			}
		}
	}
	n := names[0]
	return n, r.players[n], nil
}

func (r *Registry) Get(name string) (Player, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.players[name]
	return p, ok
}

// Manifests returns registered player manifests, sorted by name.
func (r *Registry) Manifests() []Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Manifest, 0, len(r.players))
	for _, p := range r.players {
		out = append(out, p.Manifest())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
