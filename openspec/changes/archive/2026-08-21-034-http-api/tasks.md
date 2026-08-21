# Tasks: 034-http-api

## Docs (this change)

- [x] `docs/34-http-api-v0.md` — G-153..G-158
- [x] Cross-refs in `04`, `09`, `10`, `05`, `01`, `02`, `11`, `12`, `28`
- [x] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [x] OpenSpec `034-http-api`

## Slice 25 — serve (done)

- [x] Config `api.listen` / token env + boot guard (non-loopback)
- [x] Package `internal/entrypoint/httpapi` handler (stdlib)
- [x] CLI `runtgine serve`
- [x] Routes G-155 + Error model mapping
- [x] SSE for `GET /v0/runs/{id}/events`
- [x] Tests httptest: healthz, 401, hello, validation, preview
- [x] `go test ./...` / `go vet ./...` green

## Slice 26 — webhooks (done)

- [x] Config `webhooks[]` + `RUNTGINE_WEBHOOK_SECRET`
- [x] Dispatcher on terminal run events
- [x] HTTPS + link-local deny; timeout/retry/warn
- [x] Tests with fake RoundTripper; run state unchanged on 5xx
