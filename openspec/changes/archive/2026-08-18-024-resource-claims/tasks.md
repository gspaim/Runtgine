# Tasks: 024-resource-claims

Código = slice 12. Implementado e arquivado após merge em `develop`.

## 0. Spec (PR #23)

- [x] 0.1 `docs/24-resource-claims-v0.md` (G-93..G-98)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / `02` / `11` e espelhos (mesmo PR)

## 1. Claim engine

- [x] 1.1 Scaffold `internal/core/claim` (Kind, Resource, overlap, normalize)
- [x] 1.2 Auto-claim table (G-95) from capability + input
- [x] 1.3 SQLite `resource_claims` + unique active (kind, key)
- [x] 1.4 `SweepOrphans` on `api.Open`

## 2. Runner

- [x] 2.1 Order: Policy → Claim → Execute
- [x] 2.2 HITL: no claim while `waiting_approval`; acquire after grant
- [x] 2.3 Events `claim.acquired` / `claim.conflict` / `claim.released`
- [x] 2.4 `ReleaseAll` on Run terminal (`succeeded`/`failed`/`cancelled`)
- [x] 2.5 Conflict after grant does not re-open HITL

## 3. Surfaces

- [x] 3.1 CLI `run`/`status` show `claim.conflict` + `holder_run_id`
- [x] 3.2 TUI: no new tab/keys; `failed` already visible

## 4. Tests + closeout

- [x] 4.1 hello.json still concurrent (`shell.exec` does not claim)
- [x] 4.2 Two `fs.write` same path / prefix → second `claim.conflict`
- [x] 4.3 Disjoint paths may run concurrently
- [x] 4.4 `git.commit` vs `fs.write` (workspace vs path)
- [x] 4.5 Orphan sweep after terminal / boot
- [x] 4.6 `go test ./...` and `go vet ./...`
- [x] 4.7 README Estágio: Slice 12 Feito; next = Blast Radius (`025`) / TUI GRAPH após `14`
- [x] 4.8 Archive this change into `openspec/specs/resource-claims/`
