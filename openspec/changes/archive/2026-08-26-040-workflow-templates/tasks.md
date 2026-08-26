# Tasks: 040-workflow-templates

## Docs (this change)

- [x] `docs/40-workflow-templates-v0.md` — G-194..G-200
- [x] Cross-refs em `04`, `10`, `08`, `09`, `AGENTS`, `README`,
      `docs/README`, `REVIEW`, `06`
- [x] OpenSpec `040-workflow-templates` (esta change)
- [x] `04-decisoes` G-40 vira recorte CONFIRMED (G-194..G-200)

## Slice 33 — Workflow Templates v0

- [x] Pacote `internal/core/templates` (Load, Compile, Lookup)
- [x] `api.Open` carrega templates; Graph `RefreshFromTemplates`
- [x] Intent `heuristic.template` antes de `matchShell`
- [x] CLI `runtgine template list|show|run`
- [x] Example `examples/templates/verify.json`
- [x] Unit tests (load, compile, unknown id, intent vs shell)
- [x] `go test ./...` / `go vet ./...` verdes
- [x] Arquivar OpenSpec `040` após código
