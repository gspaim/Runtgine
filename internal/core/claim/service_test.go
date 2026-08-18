package claim

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
)

func TestServiceAcquireConflictAndRelease(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc := New(st)
	ctx := context.Background()
	notes := Resource{Kind: KindPath, Key: "notes.md"}
	if err := svc.Acquire(ctx, "run-a", "s1", notes); err != nil {
		t.Fatal(err)
	}
	if err := svc.Acquire(ctx, "run-a", "s1", notes); err != nil {
		t.Fatal("same-run re-acquire should be idempotent")
	}
	child := Resource{Kind: KindPath, Key: "notes.md/x"}
	if err := svc.Acquire(ctx, "run-a", "s2", child); err != nil {
		t.Fatal("self prefix overlap is allowed")
	}
	err = svc.Acquire(ctx, "run-b", "s1", Resource{Kind: KindPath, Key: "notes.md"})
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeClaimConflict {
		t.Fatalf("expected conflict, got %#v", err)
	}
	if ve.Details["holder_run_id"] != "run-a" {
		t.Fatalf("holder=%v", ve.Details)
	}
	if _, err := svc.ReleaseAll(ctx, "run-a"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Acquire(ctx, "run-b", "s1", notes); err != nil {
		t.Fatal(err)
	}
}

func TestServicePrefixOtherRun(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc := New(st)
	ctx := context.Background()
	if err := svc.Acquire(ctx, "run-a", "s1", Resource{Kind: KindPath, Key: "src"}); err != nil {
		t.Fatal(err)
	}
	err = svc.Acquire(ctx, "run-b", "s1", Resource{Kind: KindPath, Key: "src/main.go"})
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeClaimConflict {
		t.Fatalf("expected prefix conflict, got %#v", err)
	}
}

func TestServiceWorkspaceVsPath(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc := New(st)
	ctx := context.Background()
	if err := svc.Acquire(ctx, "run-a", "s1", Workspace()); err != nil {
		t.Fatal(err)
	}
	err = svc.Acquire(ctx, "run-b", "s1", Resource{Kind: KindPath, Key: "a.txt"})
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeClaimConflict {
		t.Fatalf("expected workspace vs path conflict, got %#v", err)
	}
}

func TestSweepOrphans(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	if err := st.InsertRun(ctx, store.Run{RunID: "dead", TaskID: "t1", Status: store.StatusFailed}, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	svc := New(st)
	if err := svc.Acquire(ctx, "dead", "s1", Resource{Kind: KindPath, Key: "x"}); err != nil {
		t.Fatal(err)
	}
	active, err := st.ListActiveClaims(ctx)
	if err != nil || len(active) != 1 {
		t.Fatalf("active=%v err=%v", active, err)
	}
	if err := svc.SweepOrphans(ctx); err != nil {
		t.Fatal(err)
	}
	active, err = st.ListActiveClaims(ctx)
	if err != nil || len(active) != 0 {
		t.Fatalf("after sweep active=%v err=%v", active, err)
	}
}
