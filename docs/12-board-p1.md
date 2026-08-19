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

## G-24 — Context assembly basico

**Status: CONFIRMED**

ContextPack v0 (G-24) + extensao Graph Hits (G-67, ver `19`):

- `task` — intent.summary/notes + task_id
- `step` — step_id + capability atual
- `prior_outputs` — outputs das etapas anteriores do mesmo run
- `repo_hits` — paths/symbols do `pipeline.repo-search` (capados)
- `graph_hits` — hits estruturais do Runtime Graph (**slice 7**; `19`)
- `budget` — max_chars / max_files + graph_max_hits / graph_max_chars

Regras: truncamento deterministico se exceder budget; montado pelo Core
(`AssembleContext`) antes do LLM Player; nao e Intent Engine.
`repo_hits` = intra-run; `graph_hits` = entre runs (estrutural).
`memory_hits` v0: spec `29` (G-123..G-128); codigo = slice 17.

---

## G-25 — LLM Player v0

**Status: CONFIRMED** (A+B)

- Interface interna unica: `Complete(ctx, ContextPack, OutputSchema) → JSON`
- **Dois backends no MVP:**
  1. OpenAI-compatible (base URL + API key via env)
  2. Anthropic SDK/API nativo (API key via env)
- Selecao do backend: config/default do runtime (nao hardcode no Core)
- Pode ser um Player `llm` com adapters plugaveis **ou** dois Players
  (`llm-openai-compat`, `llm-anthropic`) declarando as mesmas capabilities
  — preferencia: **um Player `llm` + adapters** (menos ruido no Registry)
- Capabilities: `pipeline.tech-review`, `pipeline.spec-review`, refine de
  `pipeline.decompose`
- Resposta JSON validada contra `output_schema`; retry 1x se parse falhar
- Sem streaming no MVP; timeout configuravel

---

## G-26 — Task Router basico

**Status: CONFIRMED**

Router minimo por regras (Player Router completo permanece HYPOTHESIS):

1. Capability exigida pelo step
2. Preferir `kind: deterministic` se houver candidato
3. Se so AI: backend default da config (`openai-compat` | `anthropic`)
4. Empate: primeiro registrado no Registry
5. Zero candidatos → erro de validacao/plan

Sem custo/latencia/policy rica no MVP.

---

## G-27 — Modelo de subtasks

**Status: CONFIRMED**

- Output de `pipeline.decompose`: `subtasks[]` com
  `{ subtask_id (UUID v7), summary, suggested_capability, notes }`
- Persistidos no SQLite ligados a `task_id` / `run_id` (fonte de verdade)
- Execucao MVP: **child runs** com `parent_run_id` (isolamento; casa com multi-run)
- Board: nao cria cards filhos (G-21); comentario pode citar contagem/resumo
- CLI `status` lista subtasks / child runs

---

## Status P1

**G-20..G-27 CONFIRMADOS.** Cenario Board/pipeline especificado o suficiente
para implementar apos o Core CLI+Shell.

Proximos gaps naturais: P2 de engenharia (`10-gaps.md` G-30+) ou comecar codigo.
