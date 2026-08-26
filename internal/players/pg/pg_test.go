package pg_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/pg"
)

func TestManifestNoQuery(t *testing.T) {
	m := pg.New().Manifest()
	if m.Name != "postgres" {
		t.Fatalf("name=%s", m.Name)
	}
	for _, c := range m.Capabilities {
		if c.Name == "pg.query" || c.Name == "pg.exec" {
			t.Fatal("write/query capability registered")
		}
		if !strings.Contains(string(c.InputSchema), `"additionalProperties":false`) {
			t.Fatal("schema must close extra fields")
		}
		if strings.Contains(string(c.InputSchema), `"password"`) || strings.Contains(string(c.InputSchema), `"sql"`) {
			t.Fatal("password/sql must not be in schema")
		}
	}
}

func TestRejectsFlagDB(t *testing.T) {
	if err := pg.ValidateStaticInput("/tmp", pg.CapPing, json.RawMessage(`{"dbname":"--command"}`)); err == nil {
		t.Fatal("expected invalid")
	}
}

func TestUnknownCapability(t *testing.T) {
	if err := pg.ValidateStaticInput("/tmp", "pg.query", json.RawMessage(`{"dbname":"app"}`)); err == nil {
		t.Fatal("expected unknown")
	}
}

func TestFakePing(t *testing.T) {
	p := pg.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, env, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return "1\n", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: pg.CapPing,
		Input:      json.RawMessage(`{"dbname":"app","user":"runtgine"}`),
		Workspace:  t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(saw, "--command") || !contains(saw, "SELECT 1") || !contains(saw, "--dbname") {
		t.Fatalf("args=%v", saw)
	}
	var got struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(out, &got); err != nil || !got.OK {
		t.Fatalf("out=%s", out)
	}
}

func TestFakeFail(t *testing.T) {
	p := pg.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, env, args []string) (string, string, int, error) {
		return "", "connection refused", 2, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: pg.CapPing,
		Input:      json.RawMessage(`{"dbname":"app"}`),
		Workspace:  t.TempDir(),
	})
	var ve result.Error
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodePlayerError {
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
