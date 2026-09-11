package liveout

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gspaim/Runtgine/internal/core/event"
)

// StepResult is a human-readable step output for LIVE (TUI and desktop).
type StepResult struct {
	StepID string `json:"step_id"`
	OK     bool   `json:"ok"`
	Text   string `json:"text"`
}

func FromEvents(events []event.Event) []StepResult {
	var out []StepResult
	for _, e := range events {
		step := "-"
		if e.StepID != nil && *e.StepID != "" {
			step = *e.StepID
		}
		switch e.Type {
		case event.TypeStepSucceeded:
			text := FormatOutput(e.Payload["output"])
			if text == "" {
				text = "(no output)"
			}
			out = append(out, StepResult{StepID: step, OK: true, Text: text})
		case event.TypeStepFailed:
			text := payloadString(e.Payload, "error")
			if text == "" {
				text = FormatOutput(e.Payload["output"])
			}
			if text == "" {
				text = "step failed"
			}
			out = append(out, StepResult{StepID: step, OK: false, Text: text})
		}
	}
	return out
}

func FormatOutput(raw any) string {
	raw = unwrapJSON(raw)
	switch v := raw.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case map[string]any:
		stdout := strings.TrimSpace(anyString(v["stdout"]))
		stderr := strings.TrimSpace(anyString(v["stderr"]))
		var b strings.Builder
		if stdout != "" {
			b.WriteString(stdout)
		}
		if stderr != "" {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString("stderr:\n")
			b.WriteString(stderr)
		}
		if b.Len() > 0 {
			return b.String()
		}
		pretty, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(pretty)
	default:
		pretty, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return strings.TrimSpace(fmt.Sprint(v))
		}
		return string(pretty)
	}
}

func unwrapJSON(raw any) any {
	switch v := raw.(type) {
	case json.RawMessage:
		if len(v) == 0 {
			return nil
		}
		var parsed any
		if err := json.Unmarshal(v, &parsed); err != nil {
			return string(v)
		}
		return parsed
	case []byte:
		if len(v) == 0 {
			return nil
		}
		var parsed any
		if err := json.Unmarshal(v, &parsed); err != nil {
			return string(v)
		}
		return parsed
	case string:
		trim := strings.TrimSpace(v)
		if len(trim) > 0 && (trim[0] == '{' || trim[0] == '[') {
			var parsed any
			if json.Unmarshal([]byte(trim), &parsed) == nil {
				return parsed
			}
		}
		return v
	default:
		return raw
	}
}

func anyString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.RawMessage:
		return strings.Trim(string(t), `"`)
	default:
		return ""
	}
}

func payloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, _ := payload[key].(string)
	return value
}
