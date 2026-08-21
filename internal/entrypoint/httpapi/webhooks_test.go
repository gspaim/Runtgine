package httpapi_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/entrypoint/httpapi"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestWebhookPostsTerminalEvent(t *testing.T) {
	var mu sync.Mutex
	var got *http.Request
	var body string
	done := make(chan struct{})
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		got = r
		body = string(b)
		mu.Unlock()
		close(done)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header), Request: r}, nil
	})}
	d := httpapi.NewDispatcher([]config.Webhook{{
		ID: "ci-main", URL: "https://example.invalid/hooks/runtgine",
		Events: []string{event.TypeRunFailed},
	}}, "s3cret", client, slog.New(slog.NewTextHandler(io.Discard, nil)))
	d.FilterInvalid()
	if len(d.Hooks) != 1 {
		t.Fatalf("hooks=%d", len(d.Hooks))
	}
	ev, err := event.New(event.TypeRunFailed, "run-1", "task-1", nil, map[string]any{"error": "boom"})
	if err != nil {
		t.Fatal(err)
	}
	d.Handle(ev)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
	mu.Lock()
	defer mu.Unlock()
	if got.Method != http.MethodPost || got.URL.Path != "/hooks/runtgine" {
		t.Fatalf("req=%s %s", got.Method, got.URL)
	}
	mac := hmac.New(sha256.New, []byte("s3cret"))
	mac.Write([]byte(body))
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if got.Header.Get("X-Runtgine-Signature") != want {
		t.Fatalf("sig=%s want=%s", got.Header.Get("X-Runtgine-Signature"), want)
	}
	if !strings.Contains(body, event.TypeRunFailed) {
		t.Fatalf("body=%s", body)
	}
}

func TestWebhookHTTPSkipped(t *testing.T) {
	var hits atomic.Int32
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		hits.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header), Request: r}, nil
	})}
	d := httpapi.NewDispatcher([]config.Webhook{{
		ID: "bad", URL: "http://example.invalid/hooks", Events: []string{event.TypeRunFailed},
	}}, "", client, slog.New(slog.NewTextHandler(io.Discard, nil)))
	d.FilterInvalid()
	if len(d.Hooks) != 0 {
		t.Fatalf("expected skip, hooks=%d", len(d.Hooks))
	}
	ev, _ := event.New(event.TypeRunFailed, "run-1", "t", nil, nil)
	d.Handle(ev)
	time.Sleep(50 * time.Millisecond)
	if hits.Load() != 0 {
		t.Fatalf("hits=%d", hits.Load())
	}
}

func TestWebhook5xxDoesNotChangeRun(t *testing.T) {
	core := openAPI(t, "secret")
	var hits atomic.Int32
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		hits.Add(1)
		return &http.Response{StatusCode: 502, Body: io.NopCloser(strings.NewReader("no")), Header: make(http.Header), Request: r}, nil
	})}
	d := httpapi.NewDispatcher([]config.Webhook{{
		ID: "ci", URL: "https://example.invalid/hook", Events: []string{event.TypeRunFailed},
	}}, "", client, slog.New(slog.NewTextHandler(io.Discard, nil)))

	runID := submitFailingViaCore(t, core)
	snap, err := core.GetRun(context.Background(), runID)
	if err != nil {
		t.Fatal(err)
	}
	if store.Status(snap.Status) != store.StatusFailed {
		t.Fatalf("status=%s", snap.Status)
	}
	ev, _ := event.New(event.TypeRunFailed, runID, snap.TaskID, nil, map[string]any{"error": "x"})
	d.Handle(ev)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && hits.Load() < 2 {
		time.Sleep(10 * time.Millisecond)
	}
	if hits.Load() < 2 {
		t.Fatalf("expected retry, hits=%d", hits.Load())
	}
	again, err := core.GetRun(context.Background(), runID)
	if err != nil {
		t.Fatal(err)
	}
	if again.Status != snap.Status {
		t.Fatalf("run mutated: %s -> %s", snap.Status, again.Status)
	}
}

func submitFailingViaCore(t *testing.T, core *api.Core) string {
	t.Helper()
	h := httpapi.New(core, slog.Default()).Handler()
	body := []byte(`{
	  "schema_version": "0.1.0",
	  "intent": {"summary": "fail"},
	  "steps": [{"step_id":"s1","capability":"shell.exec","input":{"command":["false"],"workdir":".","timeout_ms":5000},"depends_on":[]}]
	}`)
	rec := doJSON(t, h, http.MethodPost, "/v0/tasks", "secret", body)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		RunID string `json:"run_id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		snap, err := core.GetRun(context.Background(), out.RunID)
		if err == nil && store.Status(snap.Status) == store.StatusFailed {
			return out.RunID
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timeout")
	return ""
}
