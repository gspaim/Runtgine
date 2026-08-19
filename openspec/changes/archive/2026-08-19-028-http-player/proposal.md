# Proposal: 028-http-player

## Why

Shell/Git/FS/Docker cover local execution. Fetching a remote JSON or
doc still goes through `shell.exec` + `curl`, which drops schemas,
size limits, and URL policy. This is the first read-only network
Player. It is not Runtgine's HTTP API (G-45) and not MCP (G-44).

## What Changes

- Canonical `docs/28-http-player-v0.md` (G-117..G-122 CONFIRMED)
- Player `http` in `internal/players/httpclient` (**slice 16 — not this spec PR**)
- Capabilities `http.get` / `http.head` (HTTPS only)
- Static URL/header validation; metadata/link-local denied
- Example `examples/http-get.json`

## What Does Not Change

- Shell / Git / FS / Docker / pipeline / LLM
- Task IR schema
- Claims / Blast tables (http is not a touch/claim)
- G-45 HTTP server; G-44 MCP; Project Memory
- POST, Authorization, `http://`, download-to-file
- TUI tabs

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `028-http-player` |
| Doc canônico | [`docs/28-http-player-v0.md`](../../../docs/28-http-player-v0.md) |
| Gaps | G-117..G-122 **CONFIRMED** (recorte de G-41) |
| Código | Ainda não — slice 16; **bloqueado** até este pacote + `04` |

## Approach

1. Manifest with two capabilities; `additionalProperties: false`
2. `http.Client` + injectable `RoundTripper` for offline tests
3. Validate scheme/headers at admission; filter resolved IPs at dial
4. Register in `api.Open`; Graph refresh picks up `http.*`

## Impact

- New package `internal/players/httpclient`
- `internal/core/api` register + runner static validation dispatch
- README Estágio: Slice 16 after code
