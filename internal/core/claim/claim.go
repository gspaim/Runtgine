package claim

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/gspaim/Runtgine/internal/core/result"
)

type Kind string

const (
	KindWorkspace Kind = "workspace"
	KindPath      Kind = "path"

	workspaceKey = "."
)

type Resource struct {
	Kind Kind
	Key  string
}

func NormalizePath(raw string) (Resource, error) {
	p := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if p == "" || p == "." {
		return Resource{Kind: KindWorkspace, Key: workspaceKey}, nil
	}
	if path.IsAbs(p) {
		return Resource{}, result.Validation(result.CodeInvalidInput,
			"claim path must be relative to the workspace", map[string]any{"path": raw})
	}
	cleaned := path.Clean(p)
	if cleaned == "." {
		return Resource{Kind: KindWorkspace, Key: workspaceKey}, nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return Resource{}, result.Validation(result.CodeInvalidInput,
			"claim path escapes the workspace", map[string]any{"path": raw})
	}
	return Resource{Kind: KindPath, Key: cleaned}, nil
}

func Workspace() Resource {
	return Resource{Kind: KindWorkspace, Key: workspaceKey}
}

// Overlaps reports exclusive v0 conflict between two resources (any holders).
func Overlaps(a, b Resource) bool {
	if a.Kind == KindWorkspace || b.Kind == KindWorkspace {
		return true
	}
	return pathOverlaps(a.Key, b.Key)
}

func pathOverlaps(a, b string) bool {
	if a == b {
		return true
	}
	as := strings.Split(a, "/")
	bs := strings.Split(b, "/")
	n := len(as)
	if len(bs) < n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

// Required derives the G-95 auto-claim. ok=false means no claim.
func Required(capability string, input json.RawMessage) (Resource, bool, error) {
	switch capability {
	case "fs.write":
		p, err := stringField(input, "path")
		if err != nil {
			return Resource{}, false, err
		}
		if strings.TrimSpace(p) == "" {
			return Resource{}, false, result.Validation(result.CodeInvalidInput,
				"fs.write requires path", nil)
		}
		res, err := NormalizePath(p)
		if err != nil {
			return Resource{}, false, err
		}
		return res, true, nil
	case "git.add", "git.commit":
		return Workspace(), true, nil
	case "docker.build":
		ctxPath, err := stringField(input, "context")
		if err != nil {
			return Resource{}, false, err
		}
		if strings.TrimSpace(ctxPath) == "" {
			ctxPath = "."
		}
		res, err := NormalizePath(ctxPath)
		if err != nil {
			return Resource{}, false, err
		}
		return res, true, nil
	case "docker.run":
		mount, err := boolField(input, "mount_workspace")
		if err != nil {
			return Resource{}, false, err
		}
		if !mount {
			return Resource{}, false, nil
		}
		return Workspace(), true, nil
	default:
		return Resource{}, false, nil
	}
}

func ConflictError(res Resource, holderRunID string) result.Error {
	return result.Runtime(result.CodeClaimConflict,
		fmt.Sprintf("resource %s:%s held by run %s", res.Kind, res.Key, holderRunID),
		false,
		map[string]any{
			"kind":          string(res.Kind),
			"key":           res.Key,
			"holder_run_id": holderRunID,
		})
}

func stringField(raw json.RawMessage, name string) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", result.Validation(result.CodeInvalidInput, "invalid input json: "+err.Error(), nil)
	}
	v, ok := obj[name]
	if !ok || v == nil {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", result.Validation(result.CodeInvalidInput,
			fmt.Sprintf("%s must be a string", name), nil)
	}
	return s, nil
}

func boolField(raw json.RawMessage, name string) (bool, error) {
	if len(raw) == 0 {
		return false, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return false, result.Validation(result.CodeInvalidInput, "invalid input json: "+err.Error(), nil)
	}
	v, ok := obj[name]
	if !ok || v == nil {
		return false, nil
	}
	b, ok := v.(bool)
	if !ok {
		return false, result.Validation(result.CodeInvalidInput,
			fmt.Sprintf("%s must be a boolean", name), nil)
	}
	return b, nil
}
