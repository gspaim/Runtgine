# Tasks: 031-mvp-1.0

Código = slices 19–20. **Não implementar Go neste PR de spec.**
Marcar 1.x / 2.x só depois de G-135..G-140 CONFIRMED em `04` (este
change) e merge em `develop`.

## 0. Spec (este PR)

- [x] 0.1 Reescrever `docs/09-mvp.md` (realizado vs 1.0 magro)
- [x] 0.2 `docs/31-context-engine-v0.md` (G-137..G-140)
- [x] 0.3 Emenda `docs/17-intent-engine-v0.md` (G-135..G-136)
- [x] 0.4 Promover `04` / `05` / `10` e espelhos
- [x] 0.5 OpenSpec proposal / design / deltas

## 1. Slice 19 — Intent heuristics

- [ ] 1.1 `matchPlayer` antes de `matchShell`
- [ ] 1.2 `go test` / `roda os testes` → `test.go` (não `shell.exec`)
- [ ] 1.3 `git status|diff|log`, `docker ps`
- [ ] 1.4 Métodos `heuristic.test|git|docker`; testes de Compile
- [ ] 1.5 README Estágio: Slice 19 Feito (1.0 ainda aberto)

## 2. Slice 20 — Context Engine v0

- [ ] 2.1 Semente `repo_hits` a partir de QueryHits path/symbol
- [ ] 2.2 Não pisar repo-search; Graph vazio não falha o Run
- [ ] 2.3 Intent LLM pack também semeia se vazio
- [ ] 2.4 `go test ./...` e `go vet ./...`
- [ ] 2.5 README Estágio: 1.0 magro Feito
- [ ] 2.6 Archive this change into `openspec/specs/`
