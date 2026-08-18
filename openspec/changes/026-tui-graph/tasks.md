# Tasks: 026-tui-graph

Código = slice 14. **Não implementar neste PR de spec.** Marcar só
depois de G-105..G-110 CONFIRMED em `04` (este change) e merge em
`develop`.

## 0. Spec (este PR)

- [x] 0.1 `docs/26-tui-graph-v0.md` (G-105..G-110)
- [x] 0.2 Amend `docs/14-tui-design.md` + TUI skill (sexta aba)
- [x] 0.3 OpenSpec proposal / design / deltas
- [x] 0.4 Promover `04` / `10` / `18` e espelhos

## 1. TUI GRAPH

- [ ] 1.1 Extend `CoreAPI` + fakeCore: `GetGraphSnapshot`, `RefreshGraph`
- [ ] 1.2 Sixth tab `GRAPH` between EVENTS and CONFIG
- [ ] 1.3 List + counts + detail; kind/id sort
- [ ] 1.4 Dedicated graph filter `/`
- [ ] 1.5 `r` on GRAPH → RefreshGraph + snapshot
- [ ] 1.6 Narrow width + `NO_COLOR` / ASCII

## 2. Tests + closeout

- [ ] 2.1 Tab cycle includes GRAPH
- [ ] 2.2 Shell player node + `provides` edge in detail
- [ ] 2.3 Filter hides non-matching kinds
- [ ] 2.4 `go test ./internal/entrypoint/tui/...`
- [ ] 2.5 `go test ./...` and `go vet ./...`
- [ ] 2.6 README Estágio: Slice 14 Feito; next = walk Blast←Graph / mais Players
- [ ] 2.7 Archive this change into `openspec/specs/` (tui-graph + runtime-graph delta)
