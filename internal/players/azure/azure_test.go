package azure_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/azure"
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
	m := azure.New().Manifest()
	if m.Name != "azure" {
		t.Fatalf("name=%s", m.Name)
	}
	found := 0
	for _, c := range m.Capabilities {
		switch c.Name {
		case "azure.groups-create", "azure.role", "azure.storage-create":
			t.Fatalf("%s must not be registered", c.Name)
		case azure.CapIdentity, azure.CapSubscriptions, azure.CapGroups:
			found++
		}
	}
	if found != 3 {
		t.Fatalf("caps=%v", m.Capabilities)
	}
}

func TestUnknownCapability(t *testing.T) {
	err := azure.ValidateStaticInput("/tmp", "azure.groups-create", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeUnknownCapability {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestRejectsBadTimeout(t *testing.T) {
	err := azure.ValidateStaticInput("/tmp", azure.CapIdentity, json.RawMessage(`{"timeout_ms":-1}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestFakeIdentity(t *testing.T) {
	p := azure.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `{"id":"/subscriptions/s1","name":"demo","user":{"name":"ops@example.com"}}`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: azure.CapIdentity,
		Input:      json.RawMessage(`{}`),
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"account", "show", "-o", "json"}
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
	obj, ok := got["object"].(map[string]any)
	if !ok || obj["name"] != "demo" {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeGroups(t *testing.T) {
	p := azure.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `[{"name":"rg-demo"}]`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: azure.CapGroups,
		Input:      json.RawMessage(`{}`),
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"group", "list", "-o", "json"}
	if len(saw) != len(want) {
		t.Fatalf("args=%v", saw)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	groups, ok := got["object"].([]any)
	if !ok || len(groups) != 1 {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeBadJSON(t *testing.T) {
	p := azure.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		return "not json", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: azure.CapSubscriptions,
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
	p := azure.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		return "", "ERROR: Please run az login", 1, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: azure.CapIdentity,
		Input:      json.RawMessage(`{}`),
		Workspace:  "/tmp",
	})
	if ve := asResultErr(t, err); ve.Code != result.CodePlayerError {
		t.Fatalf("code=%s", ve.Code)
	}
}
