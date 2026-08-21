package api_test

import (
	"context"
	"testing"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/lessons"
	"github.com/gspaim/Runtgine/internal/core/memory"
)

func TestLessonsCaptureOffDoesNotPropose(t *testing.T) {
	core := openTestCore(t)
	runID := submitFailingShell(t, core, "lessons off")
	waitFailed(t, core, runID)
	rows, err := core.ListLessons(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("unexpected proposals: %+v", rows)
	}
}

func TestLessonsCaptureFailuresHITL(t *testing.T) {
	core := openTestCore(t)
	core.Runner.LessonsCapture = config.LessonsCaptureFailures
	runID := submitFailingShell(t, core, "lessons hitl")
	waitFailed(t, core, runID)

	pending, err := core.ListLessons(context.Background(), lessons.StatusPending)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].RunID != runID {
		t.Fatalf("pending=%+v", pending)
	}
	mem, err := core.MemoryList(context.Background(), memory.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(mem) != 0 {
		t.Fatalf("proposal must not write memory: %+v", mem)
	}

	ep, err := core.ApproveLesson(context.Background(), pending[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if ep.RunID != runID || ep.Kind != memory.KindFailure {
		t.Fatalf("episode=%+v", ep)
	}
	mem, err = core.MemoryList(context.Background(), memory.Filter{Kind: memory.KindFailure})
	if err != nil || len(mem) != 1 {
		t.Fatalf("memory=%v err=%v", mem, err)
	}

	run2 := submitFailingShell(t, core, "lessons reject")
	waitFailed(t, core, run2)
	pending, err = core.ListLessons(context.Background(), lessons.StatusPending)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending after second fail=%v err=%v", pending, err)
	}
	if err := core.RejectLesson(context.Background(), pending[0].ID); err != nil {
		t.Fatal(err)
	}
	mem, err = core.MemoryList(context.Background(), memory.Filter{})
	if err != nil || len(mem) != 1 {
		t.Fatalf("reject wrote memory: %v err=%v", mem, err)
	}
}
