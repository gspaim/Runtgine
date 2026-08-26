# Proposal: 040-workflow-templates

## Why

G-40 (`docs/10-gaps.md` / `08`) permanece aberto: Workflow Templates
são HYPOTHESIS (nativo vs repo externo). Playbooks v0 (`33`, G-149,
slice 23) entregam markdown + `playbook_hits`, mas **não** compilam
para Task IR. Intent ainda cai em um step ou no pipeline Board.

Esta spec fecha G-40 como **carregamento nativo** no workspace:
JSON em `.runtgine/templates/*.json` → Compile → Task IR v0.
Validator / Registry continuam soberanos.

Não é Playbook, não é Player, não é auto-sizing TLC, não é
Verifier, não é marketplace git.

## What Changes

- Canonical `docs/40-workflow-templates-v0.md` (G-194..G-200 CONFIRMED)
- Pacote `internal/core/templates` (Load + Compile)
- CLI `runtgine template list|show|run`
- Intent `heuristic.template` (antes de `matchShell`)
- Graph kind aditivo `template` no boot
- Example `examples/templates/verify.json`

## What Does Not Change

- Task IR schema (`0.1.0`)
- Playbooks (`33`) — continuam markdown + hits
- Pipeline Board (`pipeline.*`)
- Players / Registry capabilities
- Claims / Blast / Policy / Validator contratos
- HTTP API, MCP, TUI tabs, Wails
- NATS (G-36 DEFERRED)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `040-workflow-templates` |
| Doc canônico | `docs/40-workflow-templates-v0.md` |
| Gaps | G-194..G-200 **CONFIRMED** (recorte de G-40) |
| Código | Slice 33 |
| Depende | `033-evolution-v0` (Playbooks distintos) |

## Approach

1. JSON nativo no workspace; load best-effort (skip + warn).
2. Compile só copia steps → Task IR; admissão valida capabilities.
3. Intent reconhece `run template <id>` **antes** do prefixo shell
   `run `.
4. Graph upsert `kind=template`; sem aresta nova.

## Impact

- New package `internal/core/templates`
- `internal/core/api`, `intent`, `graph`, CLI
- README Estágio: Slice 33 após código
- `04-decisoes` G-40 vira recorte CONFIRMED (G-194..G-200)
