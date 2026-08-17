# Tasks: 019-graph-hits

## 1. QueryHits

- [ ] 1.1 Tipos `Query`, `Hit`, `Hits` em `internal/core/graph`
- [ ] 1.2 Implementar ranking/dedup/limit + stopwords
- [ ] 1.3 Degradação em erro de store (hits vazios + log)
- [ ] 1.4 Testes unitários: seeds, mentions, keyword, dedup, limit, chars

## 2. ContextPack

- [ ] 2.1 Campos `graph_hits` + `budget.graph_max_*` (defaults 20 / 4000)
- [ ] 2.2 API Assemble / WithGraphHits respeitando hierarquia de truncamento
- [ ] 2.3 Testes de marshal JSON e caps

## 3. Wire Runner + Intent

- [ ] 3.1 Runner chama QueryHits antes de LLM `Complete`
- [ ] 3.2 Intent `compileLLM` anexa graph_hits; heuristicas não chamam Graph
- [ ] 3.3 Teste: heuristic.shell não consulta Graph
- [ ] 3.4 Teste integração leve: após SyncFromRun com mentions, pack LLM
      recebe path hit quando seeds batem

## 4. Docs de estágio (no PR de código)

- [ ] 4.1 README Estágio: Slice 7 Feito; atualizar Próximo
- [ ] 4.2 `docs/10-gaps.md` checklist slice 7 feito
- [ ] 4.3 Arquivar esta change → `openspec/changes/archive/` e merge
      deltas em `openspec/specs/`
