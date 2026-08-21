# Tasks: 034-http-api

## Docs (this change)

- [x] `docs/34-http-api-v0.md` — G-153..G-158
- [x] Cross-refs in `04`, `09`, `10`, `05`, `01`, `02`, `11`, `12`, `28`
- [x] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [x] OpenSpec `034-http-api`

## Slice 25 — serve (future)

- [ ] Config `api.listen` / token env + boot guard (non-loopback)
- [ ] Package `internal/entrypoint/httpapi` handler (stdlib)
- [ ] CLI `runtgine serve`
- [ ] Routes G-155 + Error model mapping
- [ ] SSE for `GET /v0/runs/{id}/events`
- [ ] Tests httptest: healthz, 401, hello, validation, preview
- [ ] `go test ./...` / `go vet ./...` green

## Slice 26 — webhooks (future)

- [ ] Config `webhooks[]` + `RUNTGINE_WEBHOOK_SECRET`
- [ ] Dispatcher on terminal run events
- [ ] HTTPS + link-local deny; timeout/retry/warn
- [ ] Tests with fake RoundTripper; run state unchanged on 5xx
