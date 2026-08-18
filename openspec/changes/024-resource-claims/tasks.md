# Tasks: 024-resource-claims

Código = slice 12. **Não implementar neste PR de spec.** Marcar só
depois de G-93..G-98 CONFIRMED em `04` (este change) e merge em
`develop`.

## 0. Spec (este PR)

- [x] 0.1 `docs/24-resource-claims-v0.md` (G-93..G-98)
- [x] 0.2 OpenSpec proposal / design / deltas
- [ ] 0.3 Promover `04` / `10` / `02` / `11` e espelhos (mesmo PR)

## 1. Claim engine

- [ ] 1.1 Scaffold `internal/core/claim` (Kind, Resource, overlap, normalize)
- [ ] 1.2 Auto-claim table (G-95) from capability + input
- [ ] 1.3 SQLite `resource_claims` + unique active (kind, key)
- [ ] 1.4 `SweepOrphans` on `api.Open`

## 2. Runner

- [ ] 2.1 Order: Policy → Claim → Execute
- [ ] 2.2 HITL: no claim while `waiting_approval`; acquire after grant
- [ ] 2.3 Events `claim.acquired` / `claim.conflict` / `claim.released`
- [ ] 2.4 `ReleaseAll` on Run terminal (`succeeded`/`failed`/`cancelled`)
- [ ] 2.5 Conflict after grant does not re-open HITL

## 3. Surfaces

- [ ] 3.1 CLI `run`/`status` show `claim.conflict` + `holder_run_id`
- [ ] 3.2 TUI: no new tab/keys; `failed` already visible

## 4. Tests + closeout

- [ ] 4.1 hello.json still concurrent (`shell.exec` does not claim)
- [ ] 4.2 Two `fs.write` same path / prefix → second `claim.conflict`
- [ ] 4.3 Disjoint paths may run concurrently
- [ ] 4.4 `git.commit` vs `fs.write` (workspace vs path)
- [ ] 4.5 Orphan sweep after terminal / boot
- [ ] 4.6 `go test ./...` and `go vet ./...`
- [ ] 4.7 README Estágio: Slice 12 Feito; next = Blast Radius (`025`) / TUI GRAPH após `14`
- [ ] 4.8 Archive this change into `openspec/specs/resource-claims/`
