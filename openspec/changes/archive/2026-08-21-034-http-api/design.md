# Design: 034-http-api

## Slice 25 — serve + REST/SSE

### Package

`internal/entrypoint/httpapi` — stdlib `net/http` only. No chi/gin.

```go
func New(core *api.Core, cfg config.Config, log *slog.Logger) http.Handler
func ListenAndServe(ctx context.Context, handler http.Handler, listen string) error
```

CLI: `runtgine serve [--listen] [--workspace]`. Opens Core via existing
`openCore`, blocks until SIGINT/SIGTERM, then `core.Close()`.

Core does **not** import this package.

### Auth middleware

- Skip `GET /v0/healthz`
- Require `Authorization: Bearer` matching `cfg.APIToken`
- Constant-time compare; `401` JSON Error model

Boot guard: if listen host is not loopback and token is empty, return
error before `Listen`.

### Handlers

Map G-155 1:1 onto `api.Core` methods. Decode JSON with
`http.MaxBytesReader`. `SubmitTask` / `BlastTask` / intent JSON mode
run existing `task.ValidateDocument` + `Parse` (same as CLI).

SSE: `Subscribe` + filter `run_id`; write `data: <event json>\n\n`;
flush; exit when run reaches terminal state (poll `GetRun` or inspect
event type).

### Tests

- `httptest.NewRecorder` / `httptest.NewServer` against handler
- 401 without token; 200 healthz without token
- hello Task IR → 202; invented capability → 400
- intent preview does not `InsertRun`
- non-loopback + empty token → constructor/listen error

## Slice 26 — outbound webhooks

### Config

`Config.Webhooks []Webhook` — `id`, `url`, `events[]`. Secret only from
`RUNTGINE_WEBHOOK_SECRET` (single shared secret v0).

### Dispatcher

Subscribe to Event Bus (or hook terminal transitions in Runner
observer). For matching event types, POST envelope JSON.

- Inject `http.RoundTripper` for tests
- URL allowlist: https + deny link-local/metadata (share helper with
  `internal/players/httpclient` dest checks, or duplicate the small
  function — do not import player from entrypoint; extract to
  `internal/core/netpolicy` **only if** duplication hurts; v0 may copy
  the deny list)
- Timeout 5s, one retry, warn on failure

### Tests

- Fake transport records POST URL/body/signature
- Run failure delivers `run.failed`
- 500 from destination leaves run state unchanged
- `http://` URL rejected at load or skipped
