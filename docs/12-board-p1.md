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

## Em aberto (proximos)

- G-21 Write-back no board
- G-22 Contratos por etapa do pipeline
- G-23 Regras vs LLM por etapa
- G-24 Context assembly basico
- G-25 LLM Player v0
- G-26 Task Router basico
- G-27 Modelo de subtasks
