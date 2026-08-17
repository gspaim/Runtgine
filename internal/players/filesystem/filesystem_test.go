package filesystem_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/players/filesystem"
)

func execCap(t *testing.T, ws, cap string, input any) json.RawMessage {
	t.Helper()
	p := filesystem.New()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: cap,
		Input:      raw,
		Workspace:  ws,
	})
	if err != nil {
		t.Fatalf("%s: %v", cap, err)
	}
	return out
}

func execErr(t *testing.T, ws, cap string, input any) error {
	t.Helper()
	p := filesystem.New()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Execute(context.Background(), registry.ExecRequest{
		Capability: cap,
		Input:      raw,
		Workspace:  ws,
	})
	return err
}

func TestReadUTF8AndTruncation(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "note.txt"), []byte("hello-fs"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := execCap(t, ws, filesystem.CapRead, map[string]any{"path": "note.txt"})
	var r struct {
		Path      string `json:"path"`
		Content   string `json:"content"`
		Bytes     int    `json:"bytes"`
		Truncated bool   `json:"truncated"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatal(err)
	}
	if r.Path != "note.txt" || r.Content != "hello-fs" || r.Truncated || r.Bytes != 8 {
		t.Fatalf("read=%+v", r)
	}

	if err := os.WriteFile(filepath.Join(ws, "long.txt"), []byte("abcdef"), 0o644); err != nil {
		t.Fatal(err)
	}
	out = execCap(t, ws, filesystem.CapRead, map[string]any{"path": "long.txt", "max_bytes": 3})
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatal(err)
	}
	if r.Content != "abc" || !r.Truncated || r.Bytes != 3 {
		t.Fatalf("truncated=%+v", r)
	}
}

func TestWriteCreateParentsAndAtomic(t *testing.T) {
	ws := t.TempDir()
	out := execCap(t, ws, filesystem.CapWrite, map[string]any{
		"path": "a.txt", "content": "one",
	})
	var w struct {
		Path    string `json:"path"`
		Bytes   int    `json:"bytes"`
		Created bool   `json:"created"`
	}
	if err := json.Unmarshal(out, &w); err != nil {
		t.Fatal(err)
	}
	if !w.Created || w.Bytes != 3 {
		t.Fatalf("write=%+v", w)
	}
	body, _ := os.ReadFile(filepath.Join(ws, "a.txt"))
	if string(body) != "one" {
		t.Fatalf("body=%q", body)
	}

	out = execCap(t, ws, filesystem.CapWrite, map[string]any{
		"path": "a.txt", "content": "two",
	})
	if err := json.Unmarshal(out, &w); err != nil {
		t.Fatal(err)
	}
	if w.Created {
		t.Fatal("expected overwrite created=false")
	}
	body, _ = os.ReadFile(filepath.Join(ws, "a.txt"))
	if string(body) != "two" {
		t.Fatalf("overwrite=%q", body)
	}

	if err := execErr(t, ws, filesystem.CapWrite, map[string]any{
		"path": "nested/b.txt", "content": "x",
	}); err == nil {
		t.Fatal("expected missing parent rejection")
	}
	_ = execCap(t, ws, filesystem.CapWrite, map[string]any{
		"path": "nested/b.txt", "content": "ok", "create_parents": true,
	})
	body, _ = os.ReadFile(filepath.Join(ws, "nested", "b.txt"))
	if string(body) != "ok" {
		t.Fatalf("nested=%q", body)
	}
}

func TestListOrderingRecursionAndLimit(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "b.txt"), []byte("b"), 0o644)
	_ = os.WriteFile(filepath.Join(ws, "a.txt"), []byte("a"), 0o644)
	_ = os.Mkdir(filepath.Join(ws, "d"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "d", "c.txt"), []byte("c"), 0o644)

	out := execCap(t, ws, filesystem.CapList, map[string]any{"path": "."})
	var l struct {
		Entries   []struct{ Path, Type string } `json:"entries"`
		Truncated bool                          `json:"truncated"`
	}
	if err := json.Unmarshal(out, &l); err != nil {
		t.Fatal(err)
	}
	if l.Truncated || len(l.Entries) != 3 {
		t.Fatalf("list=%+v", l.Entries)
	}
	if l.Entries[0].Path != "a.txt" || l.Entries[1].Path != "b.txt" || l.Entries[2].Path != "d" {
		t.Fatalf("order=%+v", l.Entries)
	}

	out = execCap(t, ws, filesystem.CapList, map[string]any{"path": ".", "recursive": true})
	if err := json.Unmarshal(out, &l); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range l.Entries {
		if e.Path == "d/c.txt" {
			found = true
		}
	}
	if !found {
		t.Fatalf("recursive missing nested file: %+v", l.Entries)
	}

	out = execCap(t, ws, filesystem.CapList, map[string]any{"path": ".", "max_entries": 1})
	if err := json.Unmarshal(out, &l); err != nil {
		t.Fatal(err)
	}
	if !l.Truncated || len(l.Entries) != 1 || l.Entries[0].Path != "a.txt" {
		t.Fatalf("limit=%+v truncated=%v", l.Entries, l.Truncated)
	}
}

func TestStatFileDirSymlink(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "f.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(ws, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(ws, "f.txt"), filepath.Join(ws, "link.txt")); err != nil {
		t.Fatal(err)
	}

	out := execCap(t, ws, filesystem.CapStat, map[string]any{"path": "f.txt"})
	var s struct {
		Type string `json:"type"`
		Size int64  `json:"size"`
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatal(err)
	}
	if s.Type != "file" || s.Size != 2 || s.Mode == "" {
		t.Fatalf("file stat=%+v", s)
	}
	out = execCap(t, ws, filesystem.CapStat, map[string]any{"path": "sub"})
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatal(err)
	}
	if s.Type != "directory" {
		t.Fatalf("dir stat=%+v", s)
	}
	out = execCap(t, ws, filesystem.CapStat, map[string]any{"path": "link.txt"})
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatal(err)
	}
	if s.Type != "symlink" {
		t.Fatalf("symlink stat=%+v", s)
	}
}

func TestPathEscapeAndExternalSymlinkRejected(t *testing.T) {
	ws := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := execErr(t, ws, filesystem.CapRead, map[string]any{"path": "../outside.txt"}); err == nil {
		t.Fatal("expected parent escape rejection")
	}
	if err := execErr(t, ws, filesystem.CapRead, map[string]any{"path": "/etc/passwd"}); err == nil {
		t.Fatal("expected absolute path rejection")
	}
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(ws, "leak")); err != nil {
		t.Fatal(err)
	}
	if err := execErr(t, ws, filesystem.CapRead, map[string]any{"path": "leak"}); err == nil {
		t.Fatal("expected external symlink rejection")
	}
	if err := execErr(t, ws, filesystem.CapWrite, map[string]any{"path": "leak", "content": "x"}); err == nil {
		t.Fatal("expected write-to-symlink rejection")
	}

	if err := os.WriteFile(filepath.Join(ws, "ok.txt"), []byte("inside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(ws, "ok.txt"), filepath.Join(ws, "alias.txt")); err != nil {
		t.Fatal(err)
	}
	out := execCap(t, ws, filesystem.CapRead, map[string]any{"path": "alias.txt"})
	var r struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatal(err)
	}
	if r.Content != "inside" {
		t.Fatalf("internal symlink read=%q", r.Content)
	}
}

func TestInvalidUTF8OversizedWriteAndUnknownCap(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "bin.dat"), []byte{0xff, 0xfe, 0xfd}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := execErr(t, ws, filesystem.CapRead, map[string]any{"path": "bin.dat"}); err == nil {
		t.Fatal("expected invalid UTF-8 rejection")
	}
	big := strings.Repeat("a", 4<<20+1)
	if err := execErr(t, ws, filesystem.CapWrite, map[string]any{"path": "big.txt", "content": big}); err == nil {
		t.Fatal("expected oversized write rejection")
	}
	if err := execErr(t, ws, filesystem.CapWrite, map[string]any{"path": ".", "content": "x"}); err == nil {
		t.Fatal("expected root write rejection")
	}
	p := filesystem.New()
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: "fs.delete",
		Input:      json.RawMessage(`{"path":"x"}`),
		Workspace:  ws,
	})
	if err == nil {
		t.Fatal("expected unknown capability")
	}
}

func TestManifestCapabilities(t *testing.T) {
	m := filesystem.New().Manifest()
	if m.Name != "filesystem" {
		t.Fatalf("name=%s", m.Name)
	}
	want := map[string]bool{
		filesystem.CapRead: false, filesystem.CapWrite: false,
		filesystem.CapList: false, filesystem.CapStat: false,
	}
	for _, c := range m.Capabilities {
		if _, ok := want[c.Name]; !ok {
			t.Fatalf("unexpected %s", c.Name)
		}
		want[c.Name] = true
	}
	for name, ok := range want {
		if !ok {
			t.Fatalf("missing %s", name)
		}
	}
}

func TestValidateStaticInputEscape(t *testing.T) {
	ws := t.TempDir()
	err := filesystem.ValidateStaticInput(ws, filesystem.CapRead, json.RawMessage(`{"path":"../x"}`))
	if err == nil {
		t.Fatal("expected static rejection")
	}
}
