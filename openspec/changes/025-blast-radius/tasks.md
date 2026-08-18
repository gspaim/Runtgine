# Tasks: 025-blast-radius

Código = slice 13. **Não implementar neste PR de spec.** Marcar só
depois de G-99..G-104 CONFIRMED em `04` (este change) e merge em
`develop`.

## 0. Spec (este PR)

- [x] 0.1 `docs/25-blast-radius-v0.md` (G-99..G-104)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / `02` / `11` e espelhos (mesmo PR)

## 1. Blast engine

- [ ] 1.1 Scaffold `internal/core/blast` (Report, Touch, Risk, Overlay)
- [ ] 1.2 `Touched` table (G-101); reuse `claim.NormalizePath`
- [ ] 1.3 Predicted claims via `claim.Required` (G-95 intact)
- [ ] 1.4 Overlay vs `ListActiveClaims` + `claim.Overlaps`

## 2. API + CLI

- [ ] 2.1 `api.BlastTask` validates; does not Submit / Acquire / Execute
- [ ] 2.2 CLI `runtgine blast <task.json|task.yaml>` prints JSON
- [ ] 2.3 Conflicts do not change process exit; validation errors do
- [ ] 2.4 Runner path unchanged (`hello.json` has no blast)

## 3. Tests + closeout

- [ ] 3.1 hello.json → `risk: none`, empty claims/touches
- [ ] 3.2 `fs.read` touch read; no predicted claim
- [ ] 3.3 `fs.write` predicted path; `git.add` touch paths + claim workspace
- [ ] 3.4 Overlay lists `holder_run_id` without Acquire
- [ ] 3.5 Escape path → validation error
- [ ] 3.6 `go test ./...` and `go vet ./...`
- [ ] 3.7 README Estágio: Slice 13 Feito; next = TUI GRAPH (após `14`) / Graph-walk blast
- [ ] 3.8 Archive this change into `openspec/specs/blast-radius/`
