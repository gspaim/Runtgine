# Tasks: 029-project-memory

Código = slice 17. Implementado e arquivado após merge da spec em `develop`.

## 0. Spec (PR #33)

- [x] 0.1 `docs/29-project-memory-v0.md` (G-123..G-128)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / G-46 e espelhos

## 1. Store

- [x] 1.1 Table + `Record` / `List` / `Archive`
- [x] 1.2 `Query` lexical, active-only
- [x] 1.3 `Supersede` transacional

## 2. Core + CLI

- [x] 2.1 Wire in `api.Open`; Core methods
- [x] 2.2 CLI `runtgine memory list|query|record|supersede|archive`
- [x] 2.3 ContextPack `memory_hits` + budget; LLM AssembleContext
- [x] 2.4 `memory.capture` off/failures; never fail the Run

## 3. Tests + closeout

- [x] 3.1 Query hides superseded/archived
- [x] 3.2 Empty hits + injected error degrade
- [x] 3.3 Intent heuristic does not query memory
- [x] 3.4 `go test ./internal/core/memory/...` offline
- [x] 3.5 `go test ./...` and `go vet ./...`
- [x] 3.6 README Estágio: Slice 17 Feito
- [x] 3.7 Archive this change into `openspec/specs/project-memory/`
