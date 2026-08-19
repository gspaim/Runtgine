# Proposal: 030-test-player

## Why

Runtgine’s product line is intent → **verifiable** execution. After
Shell/Git/FS/Docker/HTTP, verification still goes through
`shell.exec` + free-form `go test`, dropping schemas, flag allowlists,
and structured pass/fail. `docs/02-conceitos.md` already names Test
Player. This is the next G-41 cut — not G-45, not MCP.

## What Changes

- Canonical `docs/30-test-player-v0.md` (G-129..G-134 CONFIRMED)
- Player `test` in `internal/players/gotest` (**slice 18 — not this spec PR**)
- Capability `test.go` (`go test` argv allowlist, `-mod=readonly`)
- Test failure fails the Run (`runtime.player_error`)
- Example `examples/test-go.json` (code slice)

## What Does Not Change

- Shell / Git / FS / Docker / HTTP / pipeline / LLM / Memory
- Task IR schema
- Claims / Blast tables (`test.go` is not a touch/claim)
- G-45 HTTP server; G-44 MCP; Memory Player
- pytest / npm / `-race` / fuzz / coverage files
- TUI tabs

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `030-test-player` |
| Doc canônico | [`docs/30-test-player-v0.md`](../../../docs/30-test-player-v0.md) |
| Gaps | G-129..G-134 **CONFIRMED** (recorte de G-41) |
| Código | Ainda não — slice 18; **bloqueado** até este pacote + `04` |

## Approach

1. Manifest with one capability; `additionalProperties: false`
2. Injectable `go test` runner for offline unit tests
3. Parse `go test -json` for pass/fail/skip; truncate `log`
4. Register in `api.Open`; Graph refresh picks up `test.go`

## Impact

- New package `internal/players/gotest`
- `internal/core/api` register + runner static validation dispatch
- README Estágio: Slice 18 after code
