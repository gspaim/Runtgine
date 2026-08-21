package lessons_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/lessons"
	"github.com/gspaim/Runtgine/internal/core/memory"
	"github.com/gspaim/Runtgine/internal/core/store"
)

func TestApproveWritesMemoryRejectDoesNot(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "runtgine.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mem := memory.New(st, log)
	svc := lessons.New(st, mem, log)
	ctx := context.Background()

	if err := svc.OnRunFailed(ctx, "run-1", "task-1", "boom", "exit 1", []event.Event{{Type: event.TypeRunFailed}}); err != nil {
		t.Fatal(err)
	}
	pending, err := svc.List(ctx, lessons.StatusPending)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending=%v err=%v", pending, err)
	}

	ep, err := svc.Approve(ctx, pending[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if ep.Kind != memory.KindFailure || ep.RunID != "run-1" {
		t.Fatalf("episode=%+v", ep)
	}
	rows, err := mem.List(ctx, memory.Filter{Kind: memory.KindFailure})
	if err != nil || len(rows) != 1 {
		t.Fatalf("memory=%v err=%v", rows, err)
	}

	if err := svc.OnRunFailed(ctx, "run-2", "task-2", "again", "exit 2", nil); err != nil {
		t.Fatal(err)
	}
	pending, err = svc.List(ctx, lessons.StatusPending)
	if err != nil || len(pending) != 1 {
		t.Fatalf("second pending=%v err=%v", pending, err)
	}
	if err := svc.Reject(ctx, pending[0].ID); err != nil {
		t.Fatal(err)
	}
	rows, err = mem.List(ctx, memory.Filter{})
	if err != nil || len(rows) != 1 {
		t.Fatalf("reject must not write memory: %v err=%v", rows, err)
	}
}
