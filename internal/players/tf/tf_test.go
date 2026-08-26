package tf_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/policy"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/tf"
)

func fixtureWS(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "main.tf"), []byte("terraform {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return ws
}

func TestManifestPlanRequiresApproval(t *testing.T) {
	m := tf.New().Manifest()
	if m.Name != "terraform" {
		t.Fatalf("name=%s", m.Name)
	}
	var plan *registry.Capability
	for i := range m.Capabilities {
		if m.Capabilities[i].Name == "tf.apply" {
			t.Fatal("tf.apply must not be registered")
		}
		if m.Capabilities[i].Name == tf.CapPlan {
			plan = &m.Capabilities[i]
		}
	}
	if plan == nil || plan.ExecutionPolicy != string(policy.ApprovalRequired) {
		t.Fatalf("plan policy=%v", plan)
	}
}

func TestMissingTF(t *testing.T) {
	ws := t.TempDir()
	err := tf.ValidateStaticInput(ws, tf.CapValidate, json.RawMessage(`{"workdir":"."}`))
	if err == nil {
		t.Fatal("expected missing tf")
	}
}

func TestWorkdirEscape(t *testing.T) {
	if err := tf.ValidateStaticInput("/tmp", tf.CapValidate, json.RawMessage(`{"workdir":"../x"}`)); err == nil {
		t.Fatal("expected escape")
	}
}

func TestFakeValidate(t *testing.T) {
	ws := fixtureWS(t)
	p := tf.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return "Success", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: tf.CapValidate,
		Input:      json.RawMessage(`{"workdir":"."}`),
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(saw) < 1 || saw[0] != "validate" {
		t.Fatalf("args=%v", saw)
	}
	var got struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(out, &got); err != nil || !got.OK {
		t.Fatalf("out=%s err=%v", out, err)
	}
}

func TestFakeFail(t *testing.T) {
	ws := fixtureWS(t)
	p := tf.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, dir string, args []string) (string, string, int, error) {
		return "", "error", 1, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: tf.CapValidate,
		Input:      json.RawMessage(`{}`),
		Workspace:  ws,
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
