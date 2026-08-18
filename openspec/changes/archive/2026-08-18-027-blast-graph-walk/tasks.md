# Tasks: 027-blast-graph-walk

Código = slice 15. Implementado e arquivado após merge da spec em `develop`.

## 0. Spec (PR #29)

- [x] 0.1 `docs/27-blast-graph-walk-v0.md` (G-111..G-116)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / `02` / `25` / `18` e espelhos

## 1. Walk engine

- [x] 1.1 `Affected` + `Walk(snapshot, touches)`
- [x] 1.2 Seed unique `path` touches; inbound `mentions` only
- [x] 1.3 Stable sort + dedupe; empty on missing nodes

## 2. API + CLI

- [x] 2.1 `BlastTask` fills `affected` after Analyze; snapshot error → `[]`
- [x] 2.2 CLI JSON always includes `affected`
- [x] 2.3 Runner / TUI GRAPH unchanged

## 3. Tests + closeout

- [x] 3.1 hello.json → `affected: []`
- [x] 3.2 path touch without graph node → `[]`
- [x] 3.3 path node + mentions run appears
- [x] 3.4 `risk` / overlay unchanged vs slice 13 fixtures
- [x] 3.5 `go test ./...` and `go vet ./...`
- [x] 3.6 README Estágio: Slice 15 Feito; next = mais Players / Memory
- [x] 3.7 Archive this change into `openspec/specs/` (blast-graph-walk + blast-radius delta)
