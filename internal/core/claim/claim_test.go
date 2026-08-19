package claim

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/result"
)

func TestNormalizePath(t *testing.T) {
	ws, err := NormalizePath("")
	if err != nil || ws != Workspace() {
		t.Fatalf("empty: %+v %v", ws, err)
	}
	dot, err := NormalizePath(".")
	if err != nil || dot.Kind != KindWorkspace {
		t.Fatalf("dot: %+v %v", dot, err)
	}
	p, err := NormalizePath("src/main.go")
	if err != nil || p.Kind != KindPath || p.Key != "src/main.go" {
		t.Fatalf("path: %+v %v", p, err)
	}
	cleaned, err := NormalizePath("src/./a.go")
	if err != nil || cleaned.Key != "src/a.go" {
		t.Fatalf("clean: %+v %v", cleaned, err)
	}
	if _, err := NormalizePath("../outside"); err == nil {
		t.Fatal("expected escape error")
	}
	if _, err := NormalizePath("/abs"); err == nil {
		t.Fatal("expected abs error")
	}
}

func TestOverlaps(t *testing.T) {
	src := Resource{Kind: KindPath, Key: "src"}
	main := Resource{Kind: KindPath, Key: "src/main.go"}
	src2 := Resource{Kind: KindPath, Key: "src2"}
	notes := Resource{Kind: KindPath, Key: "notes.md"}
	if !Overlaps(src, main) || !Overlaps(main, src) {
		t.Fatal("prefix should overlap")
	}
	if Overlaps(src, src2) {
		t.Fatal("src vs src2 must not overlap")
	}
	if Overlaps(notes, src) {
		t.Fatal("disjoint files")
	}
	if !Overlaps(Workspace(), notes) || !Overlaps(src, Workspace()) {
		t.Fatal("workspace overlaps any path")
	}
	if !Overlaps(Workspace(), Workspace()) {
		t.Fatal("workspace vs workspace")
	}
}

func TestRequiredTable(t *testing.T) {
	res, ok, err := Required("fs.write", json.RawMessage(`{"path":"notes.md","content":"x"}`))
	if err != nil || !ok || res.Key != "notes.md" {
		t.Fatalf("fs.write: %+v ok=%v err=%v", res, ok, err)
	}
	res, ok, err = Required("git.commit", json.RawMessage(`{"message":"m"}`))
	if err != nil || !ok || res.Kind != KindWorkspace {
		t.Fatalf("git.commit: %+v ok=%v err=%v", res, ok, err)
	}
	res, ok, err = Required("git.add", json.RawMessage(`{"paths":["a.txt"]}`))
	if err != nil || !ok || res.Kind != KindWorkspace {
		t.Fatalf("git.add: %+v", res)
	}
	res, ok, err = Required("docker.build", json.RawMessage(`{}`))
	if err != nil || !ok || res.Kind != KindWorkspace {
		t.Fatalf("docker.build default: %+v ok=%v err=%v", res, ok, err)
	}
	res, ok, err = Required("docker.build", json.RawMessage(`{"context":"svc"}`))
	if err != nil || !ok || res.Key != "svc" {
		t.Fatalf("docker.build ctx: %+v", res)
	}
	_, ok, err = Required("docker.run", json.RawMessage(`{"image":"alpine:3.19"}`))
	if err != nil || ok {
		t.Fatalf("docker.run no mount: ok=%v err=%v", ok, err)
	}
	res, ok, err = Required("docker.run", json.RawMessage(`{"image":"alpine:3.19","mount_workspace":true}`))
	if err != nil || !ok || res.Kind != KindWorkspace {
		t.Fatalf("docker.run mount: %+v ok=%v err=%v", res, ok, err)
	}
	_, ok, err = Required("shell.exec", json.RawMessage(`{"command":["echo","hi"]}`))
	if err != nil || ok {
		t.Fatalf("shell.exec must not claim: ok=%v err=%v", ok, err)
	}
	_, ok, err = Required("http.get", json.RawMessage(`{"url":"https://example.com/"}`))
	if err != nil || ok {
		t.Fatalf("http.get must not claim: ok=%v err=%v", ok, err)
	}
	_, ok, err = Required("test.go", json.RawMessage(`{"packages":["./..."]}`))
	if err != nil || ok {
		t.Fatalf("test.go must not claim: ok=%v err=%v", ok, err)
	}
	_, ok, err = Required("fs.read", json.RawMessage(`{"path":"notes.md"}`))
	if err != nil || ok {
		t.Fatalf("fs.read must not claim")
	}
}

func TestConflictErrorCode(t *testing.T) {
	err := ConflictError(Resource{Kind: KindPath, Key: "a.txt"}, "run-holder")
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeClaimConflict {
		t.Fatalf("%#v", err)
	}
	if ve.Details["holder_run_id"] != "run-holder" {
		t.Fatalf("details=%v", ve.Details)
	}
}
