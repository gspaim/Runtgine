package aws_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/aws"
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
	m := aws.New().Manifest()
	if m.Name != "aws" {
		t.Fatalf("name=%s", m.Name)
	}
	found := 0
	for _, c := range m.Capabilities {
		switch c.Name {
		case "aws.s3-cp", "aws.s3-rm", "aws.s3-sync", "aws.ec2-run":
			t.Fatalf("%s must not be registered", c.Name)
		case aws.CapStsIdentity, aws.CapS3Buckets, aws.CapS3Objects:
			found++
		}
	}
	if found != 3 {
		t.Fatalf("caps=%v", m.Capabilities)
	}
}

func TestUnknownCapability(t *testing.T) {
	err := aws.ValidateStaticInput("/tmp", "aws.s3-cp", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeUnknownCapability {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestRejectsFlagBucket(t *testing.T) {
	err := aws.ValidateStaticInput("/tmp", aws.CapS3Objects, json.RawMessage(`{"bucket":"--endpoint-url"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestRejectsSpacePrefix(t *testing.T) {
	err := aws.ValidateStaticInput("/tmp", aws.CapS3Objects, json.RawMessage(`{"bucket":"data","prefix":"my logs"}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestRequiresBucket(t *testing.T) {
	err := aws.ValidateStaticInput("/tmp", aws.CapS3Objects, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestRejectsBadTimeout(t *testing.T) {
	err := aws.ValidateStaticInput("/tmp", aws.CapStsIdentity, json.RawMessage(`{"timeout_ms":999999}`))
	if err == nil {
		t.Fatal("expected invalid")
	}
	if ve := asResultErr(t, err); ve.Code != result.CodeInvalidInput {
		t.Fatalf("code=%s", ve.Code)
	}
}

func TestFakeStsIdentity(t *testing.T) {
	p := aws.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `{"UserId":"AIDEXAMPLE","Account":"123456789012","Arn":"arn:aws:iam::123456789012:user/demo"}`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: aws.CapStsIdentity,
		Input:      json.RawMessage(`{}`),
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"sts", "get-caller-identity", "--no-cli-pager", "--output", "json"}
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
	if !ok || obj["Account"] != "123456789012" {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeS3Objects(t *testing.T) {
	p := aws.New()
	var saw []string
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		saw = append([]string(nil), args...)
		return `{"Contents":[{"Key":"logs/a.log"}]}`, "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: aws.CapS3Objects,
		Input:      json.RawMessage(`{"bucket":"data","prefix":"logs","region":"us-east-1"}`),
		Workspace:  "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"s3api", "list-objects-v2", "--bucket", "data", "--prefix", "logs", "--region", "us-east-1", "--no-cli-pager", "--output", "json"}
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
	if got["truncated"] != false {
		t.Fatalf("out=%v", got)
	}
}

func TestFakeBadJSON(t *testing.T) {
	p := aws.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		return "not json", "", 0, nil
	})
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: aws.CapS3Buckets,
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
	p := aws.New()
	p.SetRunner(func(ctx context.Context, timeout time.Duration, args []string) (string, string, int, error) {
		return "", "Unable to locate credentials", 1, nil
	})
	_, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: aws.CapStsIdentity,
		Input:      json.RawMessage(`{}`),
		Workspace:  "/tmp",
	})
	if ve := asResultErr(t, err); ve.Code != result.CodePlayerError {
		t.Fatalf("code=%s", ve.Code)
	}
}
