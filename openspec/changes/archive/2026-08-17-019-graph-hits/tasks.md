# Tasks: 019-graph-hits

## 1. QueryHits

- [x] 1.1 Tipos `Query`, `Hit`, `Hits` em `internal/core/graph`
- [x] 1.2 Implementar ranking/dedup/limit + stopwords
- [x] 1.3 Degradação em erro de store (hits vazios + log)
- [x] 1.4 Testes unitários: seeds, mentions, keyword, dedup, limit, chars

## 2. ContextPack

- [x] 2.1 Campos `graph_hits` + `budget.graph_max_*` (defaults 20 / 4000)
- [x] 2.2 API Assemble / WithGraphHits respeitando hierarquia de truncamento
- [x] 2.3 Testes de marshal JSON e caps

## 3. Wire Runner + Intent

- [x] 3.1 Runner chama QueryHits antes de LLM `Complete`
- [x] 3.2 Intent `compileLLM` anexa graph_hits; heuristicas não chamam Graph
- [x] 3.3 Teste: heuristic.shell não consulta Graph
- [x] 3.4 Teste integração leve: após SyncFromRun com mentions, pack LLM
      recebe path hit quando seeds batem

## 4. Docs de estágio (no PR de código)

- [x] 4.1 README Estágio: Slice 7 Feito; atualizar Próximo
- [x] 4.2 `docs/10-gaps.md` checklist slice 7 feito
- [x] 4.3 Arquivar esta change → `openspec/changes/archive/` e merge
      deltas em `openspec/specs/`
