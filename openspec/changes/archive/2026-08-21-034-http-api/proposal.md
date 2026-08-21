# Proposal: 034-http-api

## Why

CI/CD (UC-02) can only enter Runtgine by invoking the CLI binary. The
Core already has SubmitTask / GetRun / Subscribe / Intent / HITL / Blast.
HTTP Player (`28`) is a **client** capability (`http.get`/`head`), not a
server. G-45 has been the standing P3 gap after MVP 1.0.

## What Changes

- Canonical `docs/34-http-api-v0.md` (G-153..G-158 CONFIRMED; closes G-45 v0)
- `docs/04-decisoes.md`, `docs/09-mvp.md`, `docs/10-gaps.md`, `docs/05-prd.md`
- Cross-refs: `01`, `02`, `11` layout, `12` (inbound still polling), `28`
- `docs/README.md`, `AGENTS.md`, `README.md` estágio, `REVIEW.md`
- OpenSpec package `034-http-api`

## What Does Not Change

- HTTP Player semantics (`28`)
- Board transport (G-20 polling)
- Validator / Runner / Event Bus
- NATS, MCP, Wails, Memory/Graph REST
- No code in this PR (slices 25–26 later)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `034-http-api` |
| Doc canônico | [`docs/34-http-api-v0.md`](../../../docs/34-http-api-v0.md) |
| Gaps | G-153..G-158 **CONFIRMED** (spec); G-45 recorte v0 |
| Code | slices 25–26 — **not started** |

## Approach

Two implementation slices (independent of 21–24):

1. **Slice 25** — `runtgine serve` + REST/SSE + bearer token
2. **Slice 26** — outbound HTTPS webhooks for terminal run events

## Impact

- Future: `internal/entrypoint/httpapi`, CLI command `serve`
- Config `api.listen` / token env `RUNTGINE_API_TOKEN`
- Docs only in this PR
