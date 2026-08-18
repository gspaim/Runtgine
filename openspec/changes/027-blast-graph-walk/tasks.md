# Tasks: 027-blast-graph-walk

Código = slice 15. **Não implementar neste PR de spec.** Marcar só
depois de G-111..G-116 CONFIRMED em `04` (este change) e merge em
`develop`.

## 0. Spec (este PR)

- [x] 0.1 `docs/27-blast-graph-walk-v0.md` (G-111..G-116)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / `02` / `25` / `18` e espelhos

## 1. Walk engine

- [ ] 1.1 `Affected` + `Walk(snapshot, touches)`
- [ ] 1.2 Seed unique `path` touches; inbound `mentions` only
- [ ] 1.3 Stable sort + dedupe; empty on missing nodes

## 2. API + CLI

- [ ] 2.1 `BlastTask` fills `affected` after Analyze; snapshot error → `[]`
- [ ] 2.2 CLI JSON always includes `affected`
- [ ] 2.3 Runner / TUI GRAPH unchanged

## 3. Tests + closeout

- [ ] 3.1 hello.json → `affected: []`
- [ ] 3.2 path touch without graph node → `[]`
- [ ] 3.3 path node + mentions run appears
- [ ] 3.4 `risk` / overlay unchanged vs slice 13 fixtures
- [ ] 3.5 `go test ./...` and `go vet ./...`
- [ ] 3.6 README Estágio: Slice 15 Feito; next = mais Players / Memory
- [ ] 3.7 Archive this change into `openspec/specs/` (blast-graph-walk + blast-radius delta)
