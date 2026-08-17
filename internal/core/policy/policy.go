package policy

import (
	"fmt"
	"strings"

	"github.com/gspaim/Runtgine/internal/core/result"
)

type Verb string

const (
	Allow             Verb = "allow"
	Deny              Verb = "deny"
	ApprovalRequired  Verb = "approval-required"
)

func ParseVerb(s string) (Verb, error) {
	v := Verb(strings.TrimSpace(s))
	switch v {
	case "", Allow:
		if v == "" {
			return Allow, nil
		}
		return Allow, nil
	case Deny, ApprovalRequired:
		return v, nil
	default:
		return "", fmt.Errorf("unknown execution_policy verb %q (want allow|deny|approval-required)", s)
	}
}

// Resolve applies G-82 precedence: default < manifest < config override.
func Resolve(globalDefault, manifest, configOverride Verb) Verb {
	v := globalDefault
	if v == "" {
		v = Allow
	}
	if manifest != "" {
		v = manifest
	}
	if configOverride != "" {
		v = configOverride
	}
	return v
}

type Table struct {
	Default Verb
	Caps    map[string]Verb // config.json capability map
}

func (t Table) Verb(capability string, manifest Verb) Verb {
	def := t.Default
	if def == "" {
		def = Allow
	}
	override := Verb("")
	if t.Caps != nil {
		override = t.Caps[capability]
	}
	return Resolve(def, manifest, override)
}

func DeniedError(capability string) result.Error {
	return result.Validation(result.CodePolicyDenied,
		fmt.Sprintf("capability %q is denied by execution policy", capability),
		map[string]any{"capability": capability})
}

func NotWaitingError() result.Error {
	return result.Runtime(result.CodePolicyNotWaiting, "run is not waiting for approval", false, nil)
}

func ApprovalDeniedError(capability string) result.Error {
	return result.Runtime(result.CodePolicyApprovalDenied,
		fmt.Sprintf("approval denied for capability %q", capability), false,
		map[string]any{"capability": capability})
}
