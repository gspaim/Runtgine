package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMemoryCaptureDefaultOff(t *testing.T) {
	t.Setenv("RUNTGINE_MEMORY_CAPTURE", "")
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Memory.Capture != MemoryCaptureOff {
		t.Fatalf("capture=%s", cfg.Memory.Capture)
	}
}

func TestMemoryCaptureFromFileAndEnv(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, ".runtgine")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"memory":{"capture":"failures"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(ws)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Memory.Capture != MemoryCaptureFailures {
		t.Fatalf("file capture=%s", cfg.Memory.Capture)
	}

	t.Setenv("RUNTGINE_MEMORY_CAPTURE", "off")
	cfg, err = Load(ws)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Memory.Capture != MemoryCaptureOff {
		t.Fatalf("env capture=%s", cfg.Memory.Capture)
	}
}

func TestMemoryCaptureRejectsUnknown(t *testing.T) {
	t.Setenv("RUNTGINE_MEMORY_CAPTURE", "outcomes")
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatal("expected error")
	}
}
