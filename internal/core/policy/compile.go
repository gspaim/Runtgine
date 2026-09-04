package policy

import (
	"fmt"

	"github.com/gspaim/Runtgine/internal/core/registry"
)

type FileConfig struct {
	Default      string            `json:"default"`
	Capabilities map[string]string `json:"capabilities"`
}

func Compile(file FileConfig, envDefault string, reg *registry.Registry) (Table, error) {
	def := Allow
	if file.Default != "" {
		v, err := ParseVerb(file.Default)
		if err != nil {
			return Table{}, fmt.Errorf("execution_policy.default: %w", err)
		}
		def = v
	}
	if envDefault != "" {
		v, err := ParseVerb(envDefault)
		if err != nil {
			return Table{}, fmt.Errorf("RUNTGINE_POLICY_DEFAULT: %w", err)
		}
		def = v
	}
	tab := Table{Default: def, Caps: map[string]Verb{}}
	known := map[string]struct{}{}
	if reg != nil {
		for _, name := range reg.CapabilityNames() {
			known[name] = struct{}{}
		}
	}
	for cap, raw := range file.Capabilities {
		if _, ok := known[cap]; !ok {
			return Table{}, fmt.Errorf("execution_policy.capabilities: unknown capability %q", cap)
		}
		v, err := ParseVerb(raw)
		if err != nil {
			return Table{}, fmt.Errorf("execution_policy.capabilities[%q]: %w", cap, err)
		}
		tab.Caps[cap] = v
	}
	return tab, nil
}
