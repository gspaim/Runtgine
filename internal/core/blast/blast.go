package blast

import (
	"encoding/json"
	"fmt"

	"github.com/gspaim/Runtgine/internal/core/claim"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

const SchemaVersion = "0.1.0"

type Mode string

const (
	ModeRead  Mode = "read"
	ModeWrite Mode = "write"
)

type Risk string

const (
	RiskNone      Risk = "none"
	RiskPath      Risk = "path"
	RiskWorkspace Risk = "workspace"
)

type Touch struct {
	Kind       string `json:"kind"`
	Key        string `json:"key"`
	Capability string `json:"capability"`
	StepID     string `json:"step_id"`
	Mode       Mode   `json:"mode"`
}

type PredictedClaim struct {
	Kind       string `json:"kind"`
	Key        string `json:"key"`
	Capability string `json:"capability"`
	StepID     string `json:"step_id"`
}

type Conflict struct {
	Kind        string `json:"kind"`
	Key         string `json:"key"`
	HolderRunID string `json:"holder_run_id"`
}

type Report struct {
	SchemaVersion   string           `json:"schema_version"`
	Capabilities    []string         `json:"capabilities"`
	Touches         []Touch          `json:"touches"`
	PredictedClaims []PredictedClaim `json:"predicted_claims"`
	Risk            Risk             `json:"risk"`
	Conflicts       []Conflict       `json:"conflicts"`
	Images          []string         `json:"images"`
}

// Analyze builds a G-100 report from Task steps and optional active claims.
func Analyze(steps []task.Step, active []store.ResourceClaim) (Report, error) {
	rep := Report{
		SchemaVersion:   SchemaVersion,
		Capabilities:    []string{},
		Touches:         []Touch{},
		PredictedClaims: []PredictedClaim{},
		Risk:            RiskNone,
		Conflicts:       []Conflict{},
		Images:          []string{},
	}
	seenCap := map[string]bool{}
	seenClaim := map[string]bool{}
	seenImage := map[string]bool{}

	var predicted []claim.Resource
	for _, s := range steps {
		if s.Capability != "" && !seenCap[s.Capability] {
			seenCap[s.Capability] = true
			rep.Capabilities = append(rep.Capabilities, s.Capability)
		}
		touches, err := Touched(s.Capability, s.Input)
		if err != nil {
			return Report{}, err
		}
		for _, res := range touches {
			rep.Touches = append(rep.Touches, Touch{
				Kind:       string(res.Resource.Kind),
				Key:        res.Resource.Key,
				Capability: s.Capability,
				StepID:     s.StepID,
				Mode:       res.Mode,
			})
		}
		res, ok, err := claim.Required(s.Capability, s.Input)
		if err != nil {
			return Report{}, err
		}
		if ok {
			id := string(res.Kind) + "\x00" + res.Key
			if !seenClaim[id] {
				seenClaim[id] = true
				predicted = append(predicted, res)
				rep.PredictedClaims = append(rep.PredictedClaims, PredictedClaim{
					Kind:       string(res.Kind),
					Key:        res.Key,
					Capability: s.Capability,
					StepID:     s.StepID,
				})
			}
		}
		for _, img := range imagesOf(s.Capability, s.Input) {
			if img == "" || seenImage[img] {
				continue
			}
			seenImage[img] = true
			rep.Images = append(rep.Images, img)
		}
	}
	rep.Risk = RiskOf(predicted)
	rep.Conflicts = Overlay(predicted, active)
	return rep, nil
}

type touchHit struct {
	Resource claim.Resource
	Mode     Mode
}

// Touched derives G-101 report resources (not locks).
func Touched(capability string, input json.RawMessage) ([]touchHit, error) {
	switch capability {
	case "fs.write":
		res, err := pathResource(input, "path", true)
		if err != nil {
			return nil, err
		}
		return []touchHit{{Resource: res, Mode: ModeWrite}}, nil
	case "fs.read", "fs.stat":
		res, err := pathResource(input, "path", true)
		if err != nil {
			return nil, err
		}
		return []touchHit{{Resource: res, Mode: ModeRead}}, nil
	case "fs.list":
		res, err := pathResource(input, "path", false)
		if err != nil {
			return nil, err
		}
		return []touchHit{{Resource: res, Mode: ModeRead}}, nil
	case "git.add":
		paths, err := stringSliceField(input, "paths")
		if err != nil {
			return nil, err
		}
		out := make([]touchHit, 0, len(paths))
		for _, p := range paths {
			res, err := claim.NormalizePath(p)
			if err != nil {
				return nil, err
			}
			out = append(out, touchHit{Resource: res, Mode: ModeWrite})
		}
		return out, nil
	case "git.commit":
		return []touchHit{{Resource: claim.Workspace(), Mode: ModeWrite}}, nil
	case "docker.build":
		ctxPath, err := stringField(input, "context")
		if err != nil {
			return nil, err
		}
		if ctxPath == "" {
			ctxPath = "."
		}
		res, err := claim.NormalizePath(ctxPath)
		if err != nil {
			return nil, err
		}
		return []touchHit{{Resource: res, Mode: ModeWrite}}, nil
	case "docker.run":
		mount, err := boolField(input, "mount_workspace")
		if err != nil {
			return nil, err
		}
		if !mount {
			return nil, nil
		}
		return []touchHit{{Resource: claim.Workspace(), Mode: ModeWrite}}, nil
	default:
		return nil, nil
	}
}

func RiskOf(predicted []claim.Resource) Risk {
	if len(predicted) == 0 {
		return RiskNone
	}
	for _, r := range predicted {
		if r.Kind == claim.KindWorkspace {
			return RiskWorkspace
		}
	}
	return RiskPath
}

func Overlay(predicted []claim.Resource, active []store.ResourceClaim) []Conflict {
	out := []Conflict{}
	seen := map[string]bool{}
	for _, pred := range predicted {
		for _, row := range active {
			held := claim.Resource{Kind: claim.Kind(row.Kind), Key: row.Key}
			if !claim.Overlaps(pred, held) {
				continue
			}
			id := string(pred.Kind) + "\x00" + pred.Key + "\x00" + row.RunID
			if seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, Conflict{
				Kind:        string(pred.Kind),
				Key:         pred.Key,
				HolderRunID: row.RunID,
			})
		}
	}
	return out
}

func pathResource(input json.RawMessage, field string, required bool) (claim.Resource, error) {
	p, err := stringField(input, field)
	if err != nil {
		return claim.Resource{}, err
	}
	if required && p == "" {
		return claim.Resource{}, result.Validation(result.CodeInvalidInput,
			field+" is required", nil)
	}
	return claim.NormalizePath(p)
}

func imagesOf(capability string, input json.RawMessage) []string {
	switch capability {
	case "docker.run":
		img, err := stringField(input, "image")
		if err != nil || img == "" {
			return nil
		}
		return []string{img}
	case "docker.build":
		tag, err := stringField(input, "tag")
		if err != nil || tag == "" {
			return nil
		}
		return []string{tag}
	default:
		return nil
	}
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

func stringSliceField(raw json.RawMessage, name string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, result.Validation(result.CodeInvalidInput, "invalid input json: "+err.Error(), nil)
	}
	v, ok := obj[name]
	if !ok || v == nil {
		return nil, nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, result.Validation(result.CodeInvalidInput,
			fmt.Sprintf("%s must be an array of strings", name), nil)
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s, ok := item.(string)
		if !ok {
			return nil, result.Validation(result.CodeInvalidInput,
				fmt.Sprintf("%s must be an array of strings", name), nil)
		}
		out = append(out, s)
	}
	return out, nil
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
