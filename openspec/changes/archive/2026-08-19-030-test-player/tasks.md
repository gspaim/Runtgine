# Tasks: 030-test-player

Código = slice 18. Implementado e arquivado após merge da spec em `develop`.

## 0. Spec (PR #35)

- [x] 0.1 `docs/30-test-player-v0.md` (G-129..G-134)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / G-41 e espelhos

## 1. Player

- [x] 1.1 Package `internal/players/gotest` + Manifest `test.go`
- [x] 1.2 `ValidateStaticInput` (packages, ranges, `run`)
- [x] 1.3 Injectable runner; `-json` parse; log truncate

## 2. Core

- [x] 2.1 Register in `api.Open`; runner static dispatch
- [x] 2.2 `examples/test-go.json`
- [x] 2.3 Blast/Claims tables unchanged

## 3. Tests + closeout

- [x] 3.1 Fake pass → `ok=true`
- [x] 3.2 Fake fail → `runtime.player_error`
- [x] 3.3 Escaping package rejected
- [x] 3.4 `go test ./internal/players/gotest/...` offline
- [x] 3.5 `go test ./...` and `go vet ./...`
- [x] 3.6 README Estágio: Slice 18 Feito
- [x] 3.7 Archive this change into `openspec/specs/test-player/`
