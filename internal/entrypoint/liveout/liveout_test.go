package liveout

import (
	"strings"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/event"
)

func TestFromEventsStdout(t *testing.T) {
	step := "echo"
	got := FromEvents([]event.Event{{
		Type:   event.TypeStepSucceeded,
		StepID: &step,
		Payload: map[string]any{
			"output": map[string]any{"stdout": "hello-runtgine", "stderr": "", "exit_code": 0},
		},
	}})
	if len(got) != 1 || !got[0].OK || got[0].Text != "hello-runtgine" {
		t.Fatalf("got=%+v", got)
	}
}

func TestFromEventsFailed(t *testing.T) {
	step := "s1"
	got := FromEvents([]event.Event{{
		Type:    event.TypeStepFailed,
		StepID:  &step,
		Payload: map[string]any{"error": "boom"},
	}})
	if len(got) != 1 || got[0].OK || got[0].Text != "boom" {
		t.Fatalf("got=%+v", got)
	}
}

func TestFormatOutputPrettyJSON(t *testing.T) {
	text := FormatOutput(map[string]any{"branch": "develop", "dirty": false})
	if !strings.Contains(text, "develop") {
		t.Fatalf("text=%s", text)
	}
}
