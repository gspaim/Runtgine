# Tasks: 026-tui-graph

Código = slice 14. Implementado e arquivado após merge da spec em `develop`.

## 0. Spec (PR #27)

- [x] 0.1 `docs/26-tui-graph-v0.md` (G-105..G-110)
- [x] 0.2 Amend `docs/14-tui-design.md` + TUI skill (sexta aba)
- [x] 0.3 OpenSpec proposal / design / deltas
- [x] 0.4 Promover `04` / `10` / `18` e espelhos

## 1. TUI GRAPH

- [x] 1.1 Extend `CoreAPI` + fakeCore: `GetGraphSnapshot`, `RefreshGraph`
- [x] 1.2 Sixth tab `GRAPH` between EVENTS and CONFIG
- [x] 1.3 List + counts + detail; kind/id sort
- [x] 1.4 Dedicated graph filter `/`
- [x] 1.5 `r` on GRAPH → RefreshGraph + snapshot
- [x] 1.6 Narrow width + `NO_COLOR` / ASCII

## 2. Tests + closeout

- [x] 2.1 Tab cycle includes GRAPH
- [x] 2.2 Shell player node + `provides` edge in detail
- [x] 2.3 Filter hides non-matching kinds
- [x] 2.4 `go test ./internal/entrypoint/tui/...`
- [x] 2.5 `go test ./...` and `go vet ./...`
- [x] 2.6 README Estágio: Slice 14 Feito; next = walk Blast←Graph / mais Players
- [x] 2.7 Archive this change into `openspec/specs/` (tui-graph + runtime-graph delta)
