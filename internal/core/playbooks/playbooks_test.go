package playbooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndHits(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(`---
id: qa
title: QA playbook
capabilities:
  - test.go
---
Run go test before merge.
`)
	if err := os.WriteFile(filepath.Join(dir, "qa.md"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	books, err := Load(dir)
	if err != nil || len(books) != 1 || books[0].ID != "qa" {
		t.Fatalf("books=%v err=%v", books, err)
	}
	hits := Hits(books, "test.go", 2, 1500)
	if len(hits) != 1 || hits[0].ID != "qa" {
		t.Fatalf("hits=%v", hits)
	}
	if Hits(books, "git.status", 2, 1500) != nil && len(Hits(books, "git.status", 2, 1500)) != 0 {
		t.Fatal("git.status should not match qa playbook")
	}
}
