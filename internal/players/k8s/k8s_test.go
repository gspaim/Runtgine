package k8s_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/k8s"
)

func TestManifestNoApply(t *testing.T) {
	m := k8s.New().Manifest()
	if m.Name != "k8s" {
		t.Fatalf("name=%s", m.Name)
	}
	found := 0
	for _, c := range m.Capabilities {
		if c.Name == "k8s.apply" {
			t.Fatal("k8s.apply must not be registered")
		}
		if c.Name == k8s.CapList || c.Name == k8s.CapGet {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("caps=%v", m.Capabilities)
	}
}

func TestRejectsFlagResource(t *testing.T) {
	err := k8s.ValidateStaticInput("/tmp", k8s.CapList, json.RawMessage(`{"resource":"--raw"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	var ve result.Error
	if !asErr(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("err=%v", err)
	}
}

func TestFakeList(t *testing.T) {
	p := k8s.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `{"kind":"PodList","items":[]}`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: k8s.CapList,
		Input:      json.RawMessage(`{"resource":"pods","namespace":"default"}`),
		Workspace:  t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(saw, "get") || !contains(saw, "pods") || !contains(saw, "-n") {
		t.Fatalf("args=%v", saw)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["resource"] != "pods" {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeFail(t *testing.T) {
	p := k8s.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		return "", "not found", 1, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: k8s.CapGet,
		Input:      json.RawMessage(`{"resource":"pods","name":"web"}`),
		Workspace:  t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve result.Error
	if !asErr(err, &ve) || ve.Code != result.CodePlayerError {
		t.Fatalf("err=%v", err)
	}
}

func asErr(err error, ve *result.Error) bool {
	e, ok := err.(result.Error)
	if ok {
		*ve = e
	}
	return ok
}

func contains(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
