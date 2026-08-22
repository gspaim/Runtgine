# Proposal: 036-npm-player

## Why

G-41 (biblioteca de Players) continua em andamento. O Test Player v0
só cobre `test.go`. `npm test` ainda cai em `shell.exec` (prefixo
Intent `npm `). O desktop já tem `package.json` no frontend — npm é o
próximo recorte G-41 dogfoodável. Não é pytest, não é K8s, não é MCP.

## What Changes

- Canonical `docs/36-npm-player-v0.md` (G-166..G-171 CONFIRMED)
- Player `npm` in `internal/players/npm` (**slice 29 — not this spec PR**)
- Capability `npm.test` (`npm test` argv allowlist; `package.json` required)
- Test failure fails the Run (`runtime.player_error`)
- Intent heuristic `npm test` → `npm.test` (beats shell prefix)
- Example `examples/npm-test.json` (code slice)

## What Does Not Change

- Shell / Git / FS / Docker / HTTP / pipeline / LLM / Memory / `test.go`
- Task IR schema
- Claims / Blast tables (`npm.test` is not a touch/claim)
- G-45 HTTP server; G-44 MCP; Memory Player
- pytest / yarn / pnpm / `npm install` / `npx`
- TUI tabs; Wails views

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `036-npm-player` |
| Doc canônico | [`docs/36-npm-player-v0.md`](../../../docs/36-npm-player-v0.md) |
| Gaps | G-166..G-171 **CONFIRMED** (recorte de G-41) |
| Código | Ainda não — slice 29; **bloqueado** até este pacote + `04` |

## Approach

1. Manifest with one capability; `additionalProperties: false`
2. Injectable `npm test` runner for offline unit tests
3. `ok` from exit code; truncate `log`
4. Register in `api.Open`; Graph refresh picks up `npm.test`
5. Intent: `heuristic.npm` before `matchShell` `npm `

## Impact

- New package `internal/players/npm` (slice 29)
- `internal/core/api` register + runner static validation dispatch
- `internal/core/intent` heuristic
- README Estágio: Slice 29 after code
