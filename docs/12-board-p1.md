# 12 — Board / pipeline vertical (P1)

Decisoes do cenario Board + pipeline de analise.
Complementa `09-mvp.md` e `10-gaps.md` (G-20+).
Protocolo Core: `11-protocolo-v0.md` (P0 CONFIRMADO).

Nao bloqueia o primeiro slice de codigo (CLI + Shell).
Necessario antes de implementar o Entry Point Board.

---

## G-20 — Card → Task IR

**Status: CONFIRMED**

- Board = Entry Point **adapter** (nao Player)
- Transporte MVP: **polling** (webhook pos-MVP)
- Auth: GitHub token via env/config (`GITHUB_TOKEN`)
- Mapeamento minimo do card:
  - titulo → `intent.summary`
  - body → `intent.notes`
  - id do card → `source.ref`
  - `source.entry_point` = `"board"`
  - `task_id` = UUID v7 gerado pelo adapter
- Steps iniciais: template fixo do pipeline **ou** um step placeholder
  ate Decomposition existir (detalhar em G-22)

Fluxo: poll card → montar Task IR v0 → `SubmitTask` → eventos/status.

---

## G-21 — Write-back no board

**Status: CONFIRMED**

Write-back minimo no MVP:
- Atualiza status/campo do card conforme lifecycle (`running` / `succeeded` / `failed`)
- Comenta no card (ou issue ligada) com `run_id` + resumo/erro
- **Nao** cria subtasks/cards filhos no board ate G-27
- Fonte de verdade dos steps: Core / SQLite

---

## G-22 — Contratos por etapa do pipeline

**Status: CONFIRMED**

Cada etapa = 1 step com capability `pipeline.*`. Steps lineares no Task IR
(`depends_on` encadeado). Sem Workflow engine.

| Etapa | Capability | Output minimo |
|---|---|---|
| Technical Review | `pipeline.tech-review` | `findings[]`, `risks[]` |
| Spec Review | `pipeline.spec-review` | `gaps[]`, `acceptance_hints[]` |
| Repo Search | `pipeline.repo-search` | `paths[]`, `symbols[]` |
| Effort Estimation | `pipeline.effort` | `effort` (S/M/L/XL), `rationale` |
| Difficulty | `pipeline.difficulty` | `difficulty` (1–5), `rationale` |
| Decomposition | `pipeline.decompose` | `subtasks[]` (`summary`, `capability` sugerida) |

Adapter (G-20) monta Task IR com esses steps (ou placeholder ate Players
existirem). Fronteira regras vs LLM: G-23.

---

## G-23 — Regras vs LLM por etapa

**Status: CONFIRMED**

| Capability | Implementacao MVP |
|---|---|
| `pipeline.repo-search` | Deterministico (walk/grep/`go list`/ripgrep) |
| `pipeline.effort` | Heuristica (+ LLM opcional se incerto) |
| `pipeline.difficulty` | Heuristica (effort + risks do tech-review) |
| `pipeline.tech-review` | LLM Player |
| `pipeline.spec-review` | LLM Player |
| `pipeline.decompose` | Regras + LLM para refinar `subtasks[]` |

---

## Em aberto (proximos)

- G-24 Context assembly basico
- G-25 LLM Player v0
- G-26 Task Router basico
- G-27 Modelo de subtasks
