package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/entrypoint/httpapi"
)

const helloJSON = `{
  "schema_version": "0.1.0",
  "source": {"entry_point": "http"},
  "intent": {"summary": "Echo hello from Shell Player"},
  "steps": [{
    "step_id": "s1",
    "capability": "shell.exec",
    "input": {"command":["echo","hello-runtgine"],"workdir":".","timeout_ms":5000},
    "depends_on": []
  }]
}`

func openAPI(t *testing.T, token string) *api.Core {
	t.Helper()
	ws := t.TempDir()
	cfg := config.Defaults()
	cfg.WorkspaceRoot = ws
	cfg.DBPath = filepath.Join(ws, ".runtgine", "runtgine.db")
	cfg.API.Token = token
	if err := config.EnsureRuntimeDir(cfg); err != nil {
		t.Fatal(err)
	}
	core, err := api.Open(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = core.Close() })
	return core
}

func doJSON(t *testing.T, h http.Handler, method, path, token string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthzNoAuth(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodGet, "/v0/healthz", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestProtectedRouteUnauthorized(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodGet, "/v0/runs", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), result.CodeUnauthorized) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestCheckBind(t *testing.T) {
	if err := httpapi.CheckBind("0.0.0.0:7420", ""); err == nil {
		t.Fatal("expected refuse")
	}
	if err := httpapi.CheckBind(":7420", ""); err == nil {
		t.Fatal("expected refuse empty host")
	}
	if err := httpapi.CheckBind("127.0.0.1:7420", ""); err != nil {
		t.Fatal(err)
	}
	if err := httpapi.CheckBind("0.0.0.0:7420", "tok"); err != nil {
		t.Fatal(err)
	}
}

func TestPostTaskHelloAndGetRun(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodPost, "/v0/tasks", "secret", []byte(helloJSON))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		RunID string `json:"run_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.RunID == "" {
		t.Fatalf("body=%s err=%v", rec.Body.String(), err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		got := doJSON(t, h, http.MethodGet, "/v0/runs/"+out.RunID, "secret", nil)
		if got.Code != http.StatusOK {
			t.Fatalf("get status=%d body=%s", got.Code, got.Body.String())
		}
		var snap api.RunSnapshot
		if err := json.Unmarshal(got.Body.Bytes(), &snap); err != nil {
			t.Fatal(err)
		}
		if store.Status(snap.Status) == store.StatusSucceeded {
			return
		}
		if store.Status(snap.Status) == store.StatusFailed {
			t.Fatalf("run failed: %s", snap.Error)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timeout waiting for succeeded")
}

func TestUnknownCapabilityNoRun(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	body := []byte(`{
	  "schema_version": "0.1.0",
	  "intent": {"summary": "nope"},
	  "steps": [{"step_id":"s1","capability":"nope.cap","input":{},"depends_on":[]}]
	}`)
	rec := doJSON(t, h, http.MethodPost, "/v0/tasks", "secret", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	list := doJSON(t, h, http.MethodGet, "/v0/runs", "secret", nil)
	if list.Body.String() != "[]\n" && list.Body.String() != "null\n" {
		var rows []api.RunSummary
		if err := json.Unmarshal(list.Body.Bytes(), &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 0 {
			t.Fatalf("runs created: %+v", rows)
		}
	}
}

func TestIntentPreviewNoRun(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodPost, "/v0/intent/preview", "secret", []byte(`{"text":"git status"}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "git.status") {
		t.Fatalf("body=%s", rec.Body.String())
	}
	list := doJSON(t, h, http.MethodGet, "/v0/runs", "secret", nil)
	var rows []api.RunSummary
	if err := json.Unmarshal(list.Body.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("preview created runs: %+v", rows)
	}
}

func TestGetRunNotFound(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodGet, "/v0/runs/00000000-0000-7000-0000-000000000001", "secret", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBodyTooLarge(t *testing.T) {
	core := openAPI(t, "secret")
	core.Cfg.API.MaxBodyBytes = 16
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodPost, "/v0/tasks", "secret", []byte(helloJSON))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBlastNoRun(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodPost, "/v0/blast", "secret", []byte(helloJSON))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	list := doJSON(t, h, http.MethodGet, "/v0/runs", "secret", nil)
	var rows []api.RunSummary
	if err := json.Unmarshal(list.Body.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("blast created runs: %+v", rows)
	}
}

func TestIntentSubmit(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodPost, "/v0/intent", "secret", []byte(`{"text":"echo hello-http"}`))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApproveNotWaitingConflict(t *testing.T) {
	core := openAPI(t, "secret")
	h := httpapi.New(core, slog.Default()).Handler()
	rec := doJSON(t, h, http.MethodPost, "/v0/tasks", "secret", []byte(helloJSON))
	var out struct {
		RunID string `json:"run_id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		snap, err := core.GetRun(context.Background(), out.RunID)
		if err == nil && (store.Status(snap.Status) == store.StatusSucceeded || store.Status(snap.Status) == store.StatusFailed) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	got := doJSON(t, h, http.MethodPost, "/v0/runs/"+out.RunID+"/approve", "secret", nil)
	if got.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", got.Code, got.Body.String())
	}
}
