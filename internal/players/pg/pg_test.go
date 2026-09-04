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

func TestManifestCapabilities(t *testing.T) {
	m := pg.New().Manifest()
	if m.Name != "postgres" {
		t.Fatalf("name=%s", m.Name)
	}
	found := 0
	for _, c := range m.Capabilities {
		if c.Name == "pg.query" || c.Name == "pg.exec" {
			t.Fatal("write/query capability registered")
		}
		if !strings.Contains(string(c.InputSchema), `"additionalProperties":false`) {
			t.Fatal("schema must close extra fields")
		}
		if strings.Contains(string(c.InputSchema), `"password"`) {
			t.Fatal("password must not be in schema")
		}
		if strings.Contains(string(c.InputSchema), `"sql"`) && c.Name != pg.CapExplain {
			t.Fatal("sql is only allowed in pg.explain")
		}
		if c.Name == pg.CapPing || c.Name == pg.CapExplain {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("caps=%v", m.Capabilities)
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

func TestExplainRejectsMultiStatement(t *testing.T) {
	err := pg.ValidateStaticInput("/tmp", pg.CapExplain, json.RawMessage(`{"sql":"select 1; drop table users","dbname":"app"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
}

func TestExplainRejectsBackslash(t *testing.T) {
	err := pg.ValidateStaticInput("/tmp", pg.CapExplain, json.RawMessage(`{"sql":"\\! whoami","dbname":"app"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
}

func TestExplainRejectsNonSelect(t *testing.T) {
	err := pg.ValidateStaticInput("/tmp", pg.CapExplain, json.RawMessage(`{"sql":"explain analyze select 1","dbname":"app"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if err := pg.ValidateStaticInput("/tmp", pg.CapExplain, json.RawMessage(`{"sql":"insert into users values (1)","dbname":"app"}`)); err == nil {
		t.Fatal("expected invalid")
	}
}

func TestExplainAcceptsSelectAndCTE(t *testing.T) {
	if err := pg.ValidateStaticInput("/tmp", pg.CapExplain, json.RawMessage(`{"sql":"  select id from users where active = true","dbname":"app"}`)); err != nil {
		t.Fatal(err)
	}
	if err := pg.ValidateStaticInput("/tmp", pg.CapExplain, json.RawMessage(`{"sql":"with c as (select 1) select * from c","dbname":"app"}`)); err != nil {
		t.Fatal(err)
	}
}

func TestFakeExplain(t *testing.T) {
	p := pg.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, env, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `[{"Plan":{"Node Type":"Seq Scan"}}]`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: pg.CapExplain,
		Input:      json.RawMessage(`{"sql":"select id from users","dbname":"app"}`),
		Workspace:  t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	idx := indexOf(saw, "--command")
	if idx < 0 || idx+1 >= len(saw) || !strings.HasPrefix(saw[idx+1], "EXPLAIN (FORMAT JSON) select id from users") {
		t.Fatalf("args=%v", saw)
	}
	var got struct {
		Plan []map[string]any `json:"plan"`
	}
	if err := json.Unmarshal(out, &got); err != nil || len(got.Plan) != 1 {
		t.Fatalf("out=%s", out)
	}
}

func TestFakeExplainBadJSON(t *testing.T) {
	p := pg.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, env, args []string) (string, string, int, error) {
		return "not json", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: pg.CapExplain,
		Input:      json.RawMessage(`{"sql":"select 1","dbname":"app"}`),
		Workspace:  t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["plan"] != "not json" || got["truncated"] != true {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeExplainFail(t *testing.T) {
	p := pg.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, env, args []string) (string, string, int, error) {
		return "", `ERROR: relation "users" does not exist`, 1, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: pg.CapExplain,
		Input:      json.RawMessage(`{"sql":"select id from users","dbname":"app"}`),
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

func indexOf(ss []string, v string) int {
	for i, s := range ss {
		if s == v {
			return i
		}
	}
	return -1
}
