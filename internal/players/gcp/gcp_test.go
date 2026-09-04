package gcp_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/gcp"
)

func asResultErr(t *testing.T, err error) result.Error {
	t.Helper()
	e, ok := err.(result.Error)
	if !ok {
		t.Fatalf("err is not result.Error: %v", err)
	}
	return e
}

func TestManifestNoMutants(t *testing.T) {
	m := gcp.New().Manifest()
	if m.Name != "gcp" {
		t.Fatalf("name=%s", m.Name)
	}
	found := 0
	for _, c := range m.Capabilities {
		switch c.Name {
		case "gcp.projects-create", "gcp.iam", "gcp.deploy":
			t.Fatalf("%s must not be registered", c.Name)
		case gcp.CapIdentity, gcp.CapConfig, gcp.CapProjects:
			found++
		}
	}
	if found != 3 {
		t.Fatalf("caps=%v", m.Capabilities)
	}
}

func TestUnknownCapability(t *testing.T) {
	err := gcp.ValidateStaticInput("/tmp", "gcp.projects-create", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeUnknownCapability {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestRejectsBadTimeout(t *testing.T) {
	err := gcp.ValidateStaticInput("/tmp", gcp.CapIdentity, json.RawMessage(`{"timeout_ms":999999}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestFakeIdentity(t *testing.T) {
	p := gcp.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `[{"account":"ops@example.com","status":"ACTIVE"}]`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: gcp.CapIdentity,
		Input:      json.RawMessage(`{}`),
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"auth", "list", "--format=json"}
	if len(saw) != len(want) {
		t.Fatalf("args=%v", saw)
	}
	for i, v := range want {
		if saw[i] != v {
			t.Fatalf("args=%v", saw)
		}
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	accounts, ok := got["object"].([]any)
	if !ok || len(accounts) != 1 {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeProjects(t *testing.T) {
	p := gcp.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `[{"projectId":"demo"}]`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: gcp.CapProjects,
		Input:      json.RawMessage(`{"timeout_ms":60000}`),
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"projects", "list", "--format=json"}
	if len(saw) != len(want) {
		t.Fatalf("args=%v", saw)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["truncated"] != false {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeBadJSON(t *testing.T) {
	p := gcp.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		return "not json", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: gcp.CapConfig,
		Input:      json.RawMessage(`{}`),
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["object"] != "not json" || got["truncated"] != true {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeFail(t *testing.T) {
	p := gcp.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		return "", "You do not currently have an active account selected", 1, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: gcp.CapProjects,
		Input:      json.RawMessage(`{}`),
		Workspace:  "/tmp",
	})
	if ve := asResultErr(t, err); ve.Code != result.CodePlayerError {
		t.Fatalf("code=%s", ve.Code)
	}
}
