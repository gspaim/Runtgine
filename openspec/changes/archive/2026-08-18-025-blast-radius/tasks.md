# Tasks: 025-blast-radius

Código = slice 13. Implementado e arquivado após merge em `develop`.

## 0. Spec (PR #25)

- [x] 0.1 `docs/25-blast-radius-v0.md` (G-99..G-104)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / `02` / `11` e espelhos (mesmo PR)

## 1. Blast engine

- [x] 1.1 Scaffold `internal/core/blast` (Report, Touch, Risk, Overlay)
- [x] 1.2 `Touched` table (G-101); reuse `claim.NormalizePath`
- [x] 1.3 Predicted claims via `claim.Required` (G-95 intact)
- [x] 1.4 Overlay vs `ListActiveClaims` + `claim.Overlaps`

## 2. API + CLI

- [x] 2.1 `api.BlastTask` validates; does not Submit / Acquire / Execute
- [x] 2.2 CLI `runtgine blast <task.json|task.yaml>` prints JSON
- [x] 2.3 Conflicts do not change process exit; validation errors do
- [x] 2.4 Runner path unchanged (`hello.json` has no blast)

## 3. Tests + closeout

- [x] 3.1 hello.json → `risk: none`, empty claims/touches
- [x] 3.2 `fs.read` touch read; no predicted claim
- [x] 3.3 `fs.write` predicted path; `git.add` touch paths + claim workspace
- [x] 3.4 Overlay lists `holder_run_id` without Acquire
- [x] 3.5 Escape path → validation error
- [x] 3.6 `go test ./...` and `go vet ./...`
- [x] 3.7 README Estágio: Slice 13 Feito; next = TUI GRAPH (após `14`) / Graph-walk blast
- [x] 3.8 Archive this change into `openspec/specs/blast-radius/`
