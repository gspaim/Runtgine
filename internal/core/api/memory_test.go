package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/memory"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
)

func TestMemoryQueryUnknownCapabilityRejected(t *testing.T) {
	core := openTestCore(t)
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "query memory"},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "memory.query",
			Input:      json.RawMessage(`{}`),
		}},
	}
	_, err = core.SubmitTask(context.Background(), tk)
	if err == nil {
		t.Fatal("expected rejection")
	}
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeUnknownCapability {
		t.Fatalf("got %v", err)
	}
}

func TestMemoryCaptureOffDoesNotWriteOnFailedRun(t *testing.T) {
	core := openTestCore(t)
	if core.Runner.MemoryCapture != memory.CaptureOff {
		t.Fatalf("default capture=%s", core.Runner.MemoryCapture)
	}
	runID := submitFailingShell(t, core, "capture off should stay quiet")
	waitFailed(t, core, runID)

	rows, err := core.MemoryList(context.Background(), memory.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("unexpected episodes: %+v", rows)
	}
}

func TestMemoryCaptureFailuresRecordsOnFailedRun(t *testing.T) {
	core := openTestCore(t)
	core.Runner.MemoryCapture = memory.CaptureFailures
	runID := submitFailingShell(t, core, "sqlite memory capture")
	waitFailed(t, core, runID)

	rows, err := core.MemoryList(context.Background(), memory.Filter{Kind: memory.KindFailure})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("episodes=%+v", rows)
	}
	if rows[0].RunID != runID || rows[0].Validity != memory.ValidityActive {
		t.Fatalf("%+v", rows[0])
	}
	if rows[0].Title != "sqlite memory capture" {
		t.Fatalf("title=%s", rows[0].Title)
	}
}

func TestMemoryQueryErrorDoesNotFailRun(t *testing.T) {
	core := openTestCore(t)
	core.Runner.Memory = failingMemory{}
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: "memory sink fails"},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "shell.exec",
			Input:      json.RawMessage(`{"command":["echo","ok"],"workdir":".","timeout_ms":5000}`),
		}},
	}
	runID, err := core.SubmitTask(context.Background(), tk)
	if err != nil {
		t.Fatal(err)
	}
	waitTerminal(t, core, runID, store.StatusSucceeded)
}

func TestMemoryRecordQueryRoundTrip(t *testing.T) {
	core := openTestCore(t)
	ctx := context.Background()
	ep, err := core.MemoryRecord(ctx, memory.EpisodeInput{
		Kind:  memory.KindPreference,
		Title: "Prefer lexical query over embeddings",
		Body:  "v0 Project Memory",
	})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := core.MemoryQuery(ctx, "lexical embeddings memory", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].ID != ep.ID {
		t.Fatalf("hits=%+v", hits)
	}
}

func submitFailingShell(t *testing.T, core *api.Core, summary string) string {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	tk := task.Task{
		SchemaVersion: task.SchemaVersion,
		TaskID:        id.String(),
		Source:        task.Source{EntryPoint: "cli"},
		Intent:        task.Intent{Summary: summary},
		Steps: []task.Step{{
			StepID:     "s1",
			Capability: "shell.exec",
			Input:      json.RawMessage(`{"command":["false"],"workdir":".","timeout_ms":5000}`),
		}},
	}
	runID, err := core.SubmitTask(context.Background(), tk)
	if err != nil {
		t.Fatal(err)
	}
	return runID
}

func waitFailed(t *testing.T, core *api.Core, runID string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		snap, err := core.GetRun(context.Background(), runID)
		if err != nil {
			t.Fatal(err)
		}
		st := store.Status(snap.Status)
		if st == store.StatusFailed {
			return
		}
		if st == store.StatusSucceeded || st == store.StatusCancelled || st == store.StatusRejected {
			t.Fatalf("run %s: %s", snap.Status, snap.Error)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timeout")
}

type failingMemory struct{}

func (failingMemory) Query(context.Context, string, int) ([]memory.Hit, error) {
	return nil, errors.New("memory unavailable")
}

func (failingMemory) Record(context.Context, memory.EpisodeInput) (memory.Episode, error) {
	return memory.Episode{}, errors.New("memory unavailable")
}
