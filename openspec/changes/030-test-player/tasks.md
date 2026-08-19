# Tasks: 030-test-player

Código = slice 18. **Não implementar neste PR de spec.** Marcar só
depois de G-129..G-134 CONFIRMED em `04` (este change) e merge em
`develop`.

## 0. Spec (este PR)

- [x] 0.1 `docs/30-test-player-v0.md` (G-129..G-134)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / G-41 e espelhos

## 1. Player

- [ ] 1.1 Package `internal/players/gotest` + Manifest `test.go`
- [ ] 1.2 `ValidateStaticInput` (packages, ranges, `run`)
- [ ] 1.3 Injectable runner; `-json` parse; log truncate

## 2. Core

- [ ] 2.1 Register in `api.Open`; runner static dispatch
- [ ] 2.2 `examples/test-go.json`
- [ ] 2.3 Blast/Claims tables unchanged

## 3. Tests + closeout

- [ ] 3.1 Fake pass → `ok=true`
- [ ] 3.2 Fake fail → `runtime.player_error`
- [ ] 3.3 Escaping package rejected
- [ ] 3.4 `go test ./internal/players/gotest/...` offline
- [ ] 3.5 `go test ./...` and `go vet ./...`
- [ ] 3.6 README Estágio: Slice 18 Feito
- [ ] 3.7 Archive this change into `openspec/specs/test-player/`
