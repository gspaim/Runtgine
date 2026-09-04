package httpclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/players/httpclient"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func textResp(req *http.Request, status int, body string, hdr http.Header) *http.Response {
	if hdr == nil {
		hdr = http.Header{}
	}
	return &http.Response{
		StatusCode: status,
		Header:     hdr,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func execCap(t *testing.T, p *httpclient.Player, cap string, input any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Execute(context.Background(), registry.ExecRequest{
		Capability: cap,
		Input:      raw,
		Workspace:  t.TempDir(),
	})
	if err != nil {
		t.Fatalf("%s: %v", cap, err)
	}
	return out
}

func execErr(t *testing.T, p *httpclient.Player, cap string, input any) error {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Execute(context.Background(), registry.ExecRequest{
		Capability: cap,
		Input:      raw,
		Workspace:  t.TempDir(),
	})
	return err
}

func TestManifestCapabilities(t *testing.T) {
	m := httpclient.New().Manifest()
	if m.Name != "http" || m.Kind != registry.KindDeterministic {
		t.Fatalf("manifest=%+v", m)
	}
	foundGet, foundHead := false, false
	for _, c := range m.Capabilities {
		if c.Name == httpclient.CapGet {
			foundGet = true
		}
		if c.Name == httpclient.CapHead {
			foundHead = true
		}
		if c.Name == "http.post" {
			t.Fatal("http.post must not be registered")
		}
	}
	if !foundGet || !foundHead {
		t.Fatal("missing get/head")
	}
}

func TestGetJSONOffline(t *testing.T) {
	p := httpclient.New()
	p.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "example.com" || req.Method != http.MethodGet {
			t.Fatalf("unexpected %s %s", req.Method, req.URL)
		}
		if req.Header.Get("User-Agent") != "runtgine-http/0.1" {
			t.Fatalf("ua=%q", req.Header.Get("User-Agent"))
		}
		hdr := http.Header{}
		hdr.Set("Content-Type", "application/json")
		hdr.Set("ETag", `"abc"`)
		hdr.Set("Set-Cookie", "secret=1")
		return textResp(req, 200, `{"ok":true}`, hdr), nil
	}))
	out := execCap(t, p, httpclient.CapGet, map[string]any{"url": "https://example.com/doc.json"})
	var g struct {
		Status    int               `json:"status"`
		URLFinal  string            `json:"url_final"`
		Headers   map[string]string `json:"headers"`
		Body      string            `json:"body"`
		Bytes     int               `json:"bytes"`
		Truncated bool              `json:"truncated"`
		Binary    bool              `json:"binary"`
	}
	if err := json.Unmarshal(out, &g); err != nil {
		t.Fatal(err)
	}
	if g.Status != 200 || g.Body != `{"ok":true}` || g.Binary || g.Truncated {
		t.Fatalf("get=%+v", g)
	}
	if g.Headers["content-type"] != "application/json" || g.Headers["etag"] != `"abc"` {
		t.Fatalf("headers=%v", g.Headers)
	}
	if _, ok := g.Headers["set-cookie"]; ok {
		t.Fatal("set-cookie must be omitted")
	}
}

func TestGetTruncationAndBinary(t *testing.T) {
	p := httpclient.New()
	p.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "bin") {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/octet-stream"}},
				Body:       io.NopCloser(bytes.NewReader([]byte{0xff, 0xfe, 0x00})),
				Request:    req,
			}, nil
		}
		return textResp(req, 200, "abcdef", nil), nil
	}))
	out := execCap(t, p, httpclient.CapGet, map[string]any{
		"url":       "https://example.com/long",
		"max_bytes": 3,
	})
	var g struct {
		Body      string `json:"body"`
		Bytes     int    `json:"bytes"`
		Truncated bool   `json:"truncated"`
		Binary    bool   `json:"binary"`
	}
	if err := json.Unmarshal(out, &g); err != nil {
		t.Fatal(err)
	}
	if g.Body != "abc" || !g.Truncated || g.Bytes != 3 || g.Binary {
		t.Fatalf("trunc=%+v", g)
	}

	out = execCap(t, p, httpclient.CapGet, map[string]any{"url": "https://example.com/bin"})
	if err := json.Unmarshal(out, &g); err != nil {
		t.Fatal(err)
	}
	if g.Body != "" || !g.Binary || g.Truncated {
		t.Fatalf("binary=%+v", g)
	}
}

func TestHeadOmitsBody(t *testing.T) {
	p := httpclient.New()
	p.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodHead {
			t.Fatalf("method=%s", req.Method)
		}
		hdr := http.Header{}
		hdr.Set("Content-Type", "text/plain")
		hdr.Set("Content-Length", "12")
		return textResp(req, 200, "should-ignore", hdr), nil
	}))
	out := execCap(t, p, httpclient.CapHead, map[string]any{"url": "https://example.com/"})
	if strings.Contains(string(out), `"body"`) {
		t.Fatalf("head must not include body: %s", out)
	}
	var h struct {
		Status  int               `json:"status"`
		Headers map[string]string `json:"headers"`
	}
	if err := json.Unmarshal(out, &h); err != nil {
		t.Fatal(err)
	}
	if h.Status != 200 || h.Headers["content-type"] != "text/plain" {
		t.Fatalf("head=%+v", h)
	}
}

func TestRejectCleartextAndAuth(t *testing.T) {
	p := httpclient.New()
	p.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("must not dial")
		return nil, errors.New("dial")
	}))
	err := httpclient.ValidateStaticInput(t.TempDir(), httpclient.CapGet, json.RawMessage(`{"url":"http://example.com/"}`))
	assertInvalid(t, err)
	err = execErr(t, p, httpclient.CapGet, map[string]any{
		"url":     "https://example.com/",
		"headers": map[string]string{"Authorization": "Bearer x"},
	})
	assertInvalid(t, err)
	err = execErr(t, p, httpclient.CapGet, map[string]any{
		"url":     "https://example.com/",
		"headers": map[string]string{"Cookie": "a=b"},
	})
	assertInvalid(t, err)
	err = execErr(t, p, httpclient.CapGet, map[string]any{"url": "https://user:pass@example.com/"})
	assertInvalid(t, err)
}

func TestRejectMetadataLiteral(t *testing.T) {
	err := httpclient.ValidateStaticInput(t.TempDir(), httpclient.CapGet, json.RawMessage(`{"url":"https://169.254.169.254/latest"}`))
	assertInvalid(t, err)
	err = httpclient.ValidateStaticInput(t.TempDir(), httpclient.CapGet, json.RawMessage(`{"url":"https://metadata.google.internal/"}`))
	assertInvalid(t, err)
}

func TestRejectResolvedMetadataIP(t *testing.T) {
	p := httpclient.New()
	p.SetLookup(func(ctx context.Context, host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("169.254.169.254")}, nil
	})
	err := execErr(t, p, httpclient.CapGet, map[string]any{"url": "https://example.com/"})
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodePlayerError {
		t.Fatalf("want player_error got %v", err)
	}
	if strings.Contains(err.Error(), "{") || strings.Contains(strings.ToLower(err.Error()), "ami") {
		t.Fatalf("must not leak body: %v", err)
	}
}

func TestRejectHTTPRedirect(t *testing.T) {
	p := httpclient.New()
	p.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/from" {
			return &http.Response{
				StatusCode: 302,
				Header:     http.Header{"Location": []string{"http://evil.example/"}},
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		}
		t.Fatalf("followed %s", req.URL)
		return nil, errors.New("followed")
	}))
	err := execErr(t, p, httpclient.CapGet, map[string]any{"url": "https://example.com/from"})
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodePlayerError {
		t.Fatalf("want player_error got %v", err)
	}
}

func TestAllowlistedHeadersPass(t *testing.T) {
	p := httpclient.New()
	p.SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Accept") != "application/json" {
			t.Fatalf("accept=%q", req.Header.Get("Accept"))
		}
		if req.Header.Get("User-Agent") != "custom/1" {
			t.Fatalf("ua=%q", req.Header.Get("User-Agent"))
		}
		return textResp(req, 200, `{"ok":true}`, nil), nil
	}))
	_ = execCap(t, p, httpclient.CapGet, map[string]any{
		"url": "https://example.com/",
		"headers": map[string]string{
			"Accept":     "application/json",
			"User-Agent": "custom/1",
		},
	})
}

func TestUnknownCapabilityRejected(t *testing.T) {
	err := httpclient.ValidateStaticInput(t.TempDir(), "http.post", json.RawMessage(`{"url":"https://example.com/"}`))
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeUnknownCapability {
		t.Fatalf("got %v", err)
	}
}

func assertInvalid(t *testing.T, err error) {
	t.Helper()
	var ve result.Error
	if !errors.As(err, &ve) || ve.Code != result.CodeInvalidInput {
		t.Fatalf("want invalid_input got %v", err)
	}
}
