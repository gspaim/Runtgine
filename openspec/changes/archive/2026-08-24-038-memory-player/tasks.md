# Tasks: 038-memory-player

## Docs (this change)

- [x] `docs/38-memory-player-v0.md` — G-180..G-186
- [x] Cross-refs em `04`, `09`, `10`, `02`, `05`, `01`, `11`, `17`,
      `29`, `33`
- [x] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [x] OpenSpec `038-memory-player` (esta change)
- [x] `04-decisoes` §OPEN QUESTION G-47 vira CONFIRMED + implementado

## Slice 31 — Player memory (corte v0)

- [x] `internal/core/memory` expõe interface `Reader`
      (`Recall`, `Check`) — Provider já cobre; só
      assinatura pública
- [x] Pacote `internal/players/memory` + Manifest
      (`memory.recall`, `memory.check`)
- [x] `ValidateStaticInput` (ranges, pattern não vazio,
      `additionalProperties: false`)
- [x] `Execute`: chama Reader; degrada em erro
- [x] Registrar no `api.Open`; runner static dispatch
- [x] Graph: `RefreshFromRegistry` cria `memory`,
      `memory.recall`, `memory.check`; edge `provides` para
      `internal/core/memory`
- [x] Examples `examples/memory-recall.json`,
      `examples/memory-check.json`
- [x] Unit tests com Reader stub (3 cenários por capability)
- [x] `go test ./...` / `go vet ./...` verdes
- [x] README Estágio: Slice 31; arquivar OpenSpec `038` após código