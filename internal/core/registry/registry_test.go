package registry_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
)

type fakePlayer struct {
	name string
	kind registry.Kind
	caps []registry.Capability
}

func (f *fakePlayer) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          f.name,
		Version:       "0.1.0",
		Kind:          f.kind,
		Capabilities:  f.caps,
	}
}

func (f *fakePlayer) Execute(_ context.Context, _ registry.ExecRequest) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}

func cap(name string, schema string) registry.Capability {
	return registry.Capability{Name: name, InputSchema: json.RawMessage(schema)}
}

func TestRegisterAndLookup(t *testing.T) {
	r := registry.New()
	p := &fakePlayer{name: "fake", kind: registry.KindDeterministic, caps: []registry.Capability{
		cap("fake.echo", `{"type":"object"}`),
	}}
	if err := r.Register(p); err != nil {
		t.Fatal(err)
	}
	if !r.HasCapability("fake.echo") {
		t.Fatal("capability must be registered")
	}
	if r.HasCapability("fake.missing") {
		t.Fatal("unknown capability must not resolve")
	}
	got, ok := r.Get("fake")
	if !ok || got != p {
		t.Fatalf("get=%v ok=%v", got, ok)
	}
	if _, ok := r.Get("nope"); ok {
		t.Fatal("unknown player must not resolve")
	}
}

func TestRegisterRejectsDuplicatesAndEmptyName(t *testing.T) {
	r := registry.New()
	p := &fakePlayer{name: "fake", caps: []registry.Capability{cap("fake.a", `{"type":"object"}`)}}
	if err := r.Register(p); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(&fakePlayer{name: "fake"}); err == nil {
		t.Fatal("duplicate player must fail")
	}
	if err := r.Register(&fakePlayer{name: ""}); err == nil {
		t.Fatal("empty name must fail")
	}
}

func TestRegisterRejectsInvalidSchema(t *testing.T) {
	r := registry.New()
	p := &fakePlayer{name: "bad", caps: []registry.Capability{cap("bad.a", `{not-json`)}}
	if err := r.Register(p); err == nil {
		t.Fatal("invalid input schema must fail registration")
	}
	// The failed player must not be registered.
	if _, ok := r.Get("bad"); ok {
		t.Fatal("failed registration must not keep the player")
	}
}

func TestCapabilityNamesSorted(t *testing.T) {
	r := registry.New()
	if err := r.Register(&fakePlayer{name: "z", caps: []registry.Capability{cap("z.b", `{"type":"object"}`), cap("z.a", `{"type":"object"}`)}}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(&fakePlayer{name: "a", caps: []registry.Capability{cap("a.c", `{"type":"object"}`)}}); err != nil {
		t.Fatal(err)
	}
	names := r.CapabilityNames()
	if len(names) != 3 || names[0] != "a.c" || names[1] != "z.a" || names[2] != "z.b" {
		t.Fatalf("names=%v", names)
	}
}

func TestValidateInput(t *testing.T) {
	r := registry.New()
	if err := r.Register(&fakePlayer{name: "fake", caps: []registry.Capability{
		cap("fake.echo", `{"type":"object","required":["text"],"additionalProperties":false,"properties":{"text":{"type":"string"}}}`),
	}}); err != nil {
		t.Fatal(err)
	}
	if err := r.ValidateInput("fake.echo", json.RawMessage(`{"text":"hi"}`)); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
	err := r.ValidateInput("fake.echo", json.RawMessage(`{"wrong":1}`))
	var ve result.Error
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("schema violation must be invalid_input, got %v", err)
	}
	err = r.ValidateInput("fake.echo", nil)
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("empty input must be invalid_input, got %v", err)
	}
	err = r.ValidateInput("fake.echo", json.RawMessage(`{broken`))
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("broken json must be invalid_input, got %v", err)
	}
	err = r.ValidateInput("nope.echo", json.RawMessage(`{}`))
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodeUnknownCapability {
		t.Fatalf("unknown capability, got %v", err)
	}
}

func TestResolvePrefersDeterministic(t *testing.T) {
	r := registry.New()
	// Service registered first; deterministic second — deterministic must win.
	if err := r.Register(&fakePlayer{name: "svc", kind: registry.KindService, caps: []registry.Capability{cap("dup.cap", `{"type":"object"}`)}}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(&fakePlayer{name: "det", kind: registry.KindDeterministic, caps: []registry.Capability{cap("dup.cap", `{"type":"object"}`)}}); err != nil {
		t.Fatal(err)
	}
	name, _, err := r.Resolve("dup.cap", "")
	if err != nil || name != "det" {
		t.Fatalf("resolve=%s err=%v", name, err)
	}
}

func TestResolveLLMBackendFallback(t *testing.T) {
	r := registry.New()
	if err := r.Register(&fakePlayer{name: "llm-openai", kind: registry.KindAI, caps: []registry.Capability{cap("llm.gen", `{"type":"object"}`)}}); err != nil {
		t.Fatal(err)
	}
	name, _, err := r.Resolve("llm.gen", "openai")
	if err != nil || name != "llm-openai" {
		t.Fatalf("resolve=%s err=%v", name, err)
	}
}

func TestResolveUnknownCapability(t *testing.T) {
	r := registry.New()
	_, _, err := r.Resolve("missing.cap", "")
	var ve result.Error
	if err == nil || !asErr(err, &ve) || ve.Code != result.CodeUnknownCapability {
		t.Fatalf("err=%v", err)
	}
}

func TestManifestPolicyAndManifests(t *testing.T) {
	r := registry.New()
	gate := cap("fake.gate", `{"type":"object"}`)
	gate.ExecutionPolicy = "approval-required"
	if err := r.Register(&fakePlayer{name: "fake", caps: []registry.Capability{cap("fake.echo", `{"type":"object"}`), gate}}); err != nil {
		t.Fatal(err)
	}
	if r.ManifestPolicy("fake.gate") != "approval-required" {
		t.Fatalf("policy=%q", r.ManifestPolicy("fake.gate"))
	}
	if r.ManifestPolicy("fake.echo") != "" {
		t.Fatalf("default policy must be empty, got %q", r.ManifestPolicy("fake.echo"))
	}
	manifests := r.Manifests()
	if len(manifests) != 1 || manifests[0].Name != "fake" || len(manifests[0].Capabilities) != 2 {
		t.Fatalf("manifests=%+v", manifests)
	}
}

func asErr(err error, ve *result.Error) bool {
	e, ok := err.(result.Error)
	if ok {
		*ve = e
	}
	return ok
}
