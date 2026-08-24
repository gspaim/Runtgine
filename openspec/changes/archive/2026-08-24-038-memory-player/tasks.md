# Tasks: 038-memory-player

## Docs (this change)

- [ ] `docs/38-memory-player-v0.md` — G-180..G-186
- [ ] Cross-refs em `04`, `09`, `10`, `02`, `05`, `01`, `11`, `17`,
      `29`, `33`
- [ ] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [ ] OpenSpec `038-memory-player` (esta change)
- [ ] `04-decisoes` §OPEN QUESTION G-47 vira CONFIRMED + implementado

## Slice 31 — Player memory (corte v0)

- [ ] `internal/core/memory` expõe interface `Reader`
      (`Recall`, `Check`) — Provider já cobre; só
      assinatura pública
- [ ] Pacote `internal/players/memory` + Manifest
      (`memory.recall`, `memory.check`)
- [ ] `ValidateStaticInput` (ranges, pattern não vazio,
      `additionalProperties: false`)
- [ ] `Execute`: chama Reader; degrada em erro
- [ ] Registrar no `api.Open`; runner static dispatch
- [ ] Graph: `RefreshFromRegistry` cria `memory`,
      `memory.recall`, `memory.check`; edge `provides` para
      `internal/core/memory`
- [ ] Examples `examples/memory-recall.json`,
      `examples/memory-check.json`
- [ ] Unit tests com Reader stub (3 cenários por capability)
- [ ] `go test ./...` / `go vet ./...` verdes
- [ ] README Estágio: Slice 31; arquivar OpenSpec `038` após código
