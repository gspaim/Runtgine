package docker_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/policy"
	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/players/docker"
)

func stubPlayer(t *testing.T, stdout string, capture *[]string) *docker.Player {
	t.Helper()
	p := docker.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, error) {
		if capture != nil {
			*capture = append([]string{}, args...)
		}
		return stdout, "", nil
	})
	return p
}

func TestManifestRunBuildRequireApproval(t *testing.T) {
	m := docker.New().Manifest()
	foundRun, foundBuild := false, false
	for _, c := range m.Capabilities {
		if c.Name == docker.CapRun {
			foundRun = true
			if c.ExecutionPolicy != string(policy.ApprovalRequired) {
				t.Fatalf("run policy=%q", c.ExecutionPolicy)
			}
		}
		if c.Name == docker.CapBuild {
			foundBuild = true
			if c.ExecutionPolicy != string(policy.ApprovalRequired) {
				t.Fatalf("build policy=%q", c.ExecutionPolicy)
			}
		}
	}
	if !foundRun || !foundBuild {
		t.Fatal("missing run/build")
	}
}

func TestPSParsesJSONLines(t *testing.T) {
	p := stubPlayer(t, `{"ID":"abc","Image":"alpine","Names":"n1","Status":"Up"}`+"\n", nil)
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: docker.CapPS,
		Input:      json.RawMessage(`{"all":true}`),
		Workspace:  t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"id":"abc"`) {
		t.Fatalf("out=%s", out)
	}
}

func TestRunArgvSandbox(t *testing.T) {
	var got []string
	p := stubPlayer(t, "ok\n", &got)
	ws := t.TempDir()
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: docker.CapRun,
		Input:      json.RawMessage(`{"image":"alpine:3.19","argv":["echo","hi"]}`),
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "--pull=never") || !strings.Contains(joined, "--network=none") || !strings.Contains(joined, "--rm") {
		t.Fatalf("argv=%v", got)
	}
	if strings.Contains(joined, "-v") {
		t.Fatalf("unexpected mount: %v", got)
	}
}

func TestRunMountWorkspace(t *testing.T) {
	var got []string
	p := stubPlayer(t, "", &got)
	ws := t.TempDir()
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: docker.CapRun,
		Input:      json.RawMessage(`{"image":"alpine:3.19","mount_workspace":true}`),
		Workspace:  ws,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "-v") {
		t.Fatalf("expected mount: %v", got)
	}
}

func TestRejectFlagInjectionAndEscape(t *testing.T) {
	ws := t.TempDir()
	if err := docker.ValidateStaticInput(ws, docker.CapInspect, json.RawMessage(`{"id":"--privileged"}`)); err == nil {
		t.Fatal("expected reject --privileged as id")
	}
	if err := docker.ValidateStaticInput(ws, docker.CapBuild, json.RawMessage(`{"context":"../outside"}`)); err == nil {
		t.Fatal("expected reject escape")
	}
	if err := docker.ValidateStaticInput(ws, docker.CapRun, json.RawMessage(`{"image":"-e"}`)); err == nil {
		t.Fatal("expected reject image flag")
	}
}

func TestUnknownCapabilityRejected(t *testing.T) {
	p := docker.New()
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: "docker.push",
		Input:      json.RawMessage(`{}`),
		Workspace:  t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected unknown")
	}
}
