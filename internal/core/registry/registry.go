package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
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
	mu      sync.RWMutex
	players map[string]Player
	caps    map[string][]string
	meta    map[string]Capability
	kinds   map[string]Kind
}

func New() *Registry {
	return &Registry{
		players: map[string]Player{},
		caps:    map[string][]string{},
		meta:    map[string]Capability{},
		kinds:   map[string]Kind{},
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
	r.players[m.Name] = p
	r.kinds[m.Name] = m.Kind
	for _, c := range m.Capabilities {
		r.caps[c.Name] = append(r.caps[c.Name], m.Name)
		if _, ok := r.meta[c.Name]; !ok {
			r.meta[c.Name] = c
		}
	}
	return nil
}

func (r *Registry) HasCapability(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.caps[name]) > 0
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
