# Tasks: 031-mvp-1.0

Spec 031 (G-135..G-140). Código = slices 19–20. Implementado e arquivado após
merge da spec em `develop`.

## 0. Spec (PR #37)

- [x] 0.1 Reescrever `docs/09-mvp.md` (realizado vs 1.0 magro)
- [x] 0.2 `docs/31-context-engine-v0.md` (G-137..G-140)
- [x] 0.3 Emenda `docs/17-intent-engine-v0.md` (G-135..G-136)
- [x] 0.4 Promover `04` / `05` / `10` e espelhos
- [x] 0.5 OpenSpec proposal / design / deltas

## 1. Slice 19 — Intent heuristics

- [x] 1.1 `matchPlayer` antes de `matchShell`
- [x] 1.2 `go test` / `roda os testes` → `test.go` (não `shell.exec`)
- [x] 1.3 `git status|diff|log`, `docker ps`
- [x] 1.4 Métodos `heuristic.test|git|docker`; testes de Compile
- [x] 1.5 README Estágio: Slice 19 Feito (1.0 ainda aberto)

## 2. Slice 20 — Context Engine v0

- [x] 2.1 Semente `repo_hits` a partir de QueryHits path/symbol
- [x] 2.2 Não pisar repo-search; Graph vazio não falha o Run
- [x] 2.3 Intent LLM pack também semeia se vazio
- [x] 2.4 `go test ./...` e `go vet ./...`
- [x] 2.5 README Estágio: 1.0 magro Feito
- [x] 2.6 Archive this change into `openspec/specs/`
