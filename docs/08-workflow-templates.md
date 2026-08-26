# 08 — Workflow Templates e Playbooks

Workflow Templates sao registros reutilizaveis no Runtime Graph
que definem um processo de execucao completo: fases, gates
deterministicos, Players necessarios e criterios de validacao.

Playbooks (skills) sao documentacao executavel que acompanha o
template. O Intent Engine consulta ambos para gerar Execution Plans.

**Corte implementavel v0 (Playbooks):** [33-evolution-v0.md](33-evolution-v0.md)
(G-149). **Corte implementavel v0 (Templates JSON):** [40-workflow-templates-v0.md](40-workflow-templates-v0.md)
(G-194..G-200; fecha G-40 como loading nativo). Auto-sizing / Verifier /
fases TLC permanecem fora do v0.
Lessons / auto-melhoria: G-150 em `33` (slice 24).

---

## TLC Spec-Driven (referencia)

Fonte: github.com/tech-leads-club/agent-skills
Skill: tlc-spec-driven v3.3.0

SDD e um workflow template de desenvolvimento com 4 fases
adaptativas e auto-sizing por complexidade.

### Fases

Specify (obrigatoria) -> Design (opcional) -> Tasks (opcional) -> Execute (obrigatoria)

### Artefatos

.specs/STATE.md, .specs/LESSONS.md, .specs/features/[feature]/
spec.md, design.md, tasks.md, validation.md

### Mapeamento TLC SDD -> Runtgine

Specify: Task LLM Player + Validator Gate
Design: Task LLM Player + Validator (opcional)
Tasks: Task Decomposition + Validator
Execute: N Tasks + Verifier
validate_spec.py: Deterministic Gate
validate_tasks.py: Deterministic Gate
check_commit.py: Deterministic Gate
validate_state.py: Deterministic Gate (Verifier)
Verifier sub-agent: Verifier Player
lessons.py: Lessons Engine
Sub-agent delegation: Background Players
STATE.md: Runtime Graph

### Auto-sizing

Small: <=3 arquivos, Spec inline, Skip Design, Skip Tasks, Inline Execute
Medium: <10 tasks, Spec breve, Design inline, Tasks inline, Verify
Large: multi-componente, Spec completo, Arquitetura, Tasks, Execute
Complex: dominio novo, Spec+discussao, Pesquisa+arquitetura, UAT

---

## Fluxo no Runtgine

1. Intent Engine reconhece SDD, consulta Runtime Graph
2. Graph retorna template: fases, gates, verifier
3. Intent Engine gera Execution Plan com auto-sizing
4. Orchestrator executa, emitindo eventos por etapa
5. Cada gate trava se falhar
6. Verifier roda automatico ao final (autor != verifier)

---

## Conceitos novos

| Conceito | Status | Descricao |
|---|---|---|
| Workflow Template | CONFIRMED v0 (`40`) | JSON nativo → Task IR; Graph kind `template` |
| Playbook / Skill | CONFIRMED v0 (`33`) | Documentacao executavel + `playbook_hits` |
| Phase | HYPOTHESIS | Etapa de um template (fora do v0 `40`) |
| Deterministic Gate | HYPOTHESIS | No v0 `40` um gate e um step deterministico |
| Verifier | HYPOTHESIS | Validacao final (autor != verifier) |
| Lessons Engine | CONFIRMED v0 (`33` G-150) | Postmortem + HITL |
| Auto-sizing | HYPOTHESIS | Profundidade por complexidade |

---

## Gates deterministicos

validate_spec.py: antes de confirmar spec (EARS, assumptions, IDs)
validate_tasks.py: antes de aprovar tasks (granularidade, deps, testes)
check_commit.py: em cada commit (Conventional Commits)
validate_state.py: antes de declarar done (validation.md PASS)

No Runtgine, executados pelo Task Validator.

---

## Questao G-40 — resolvida no v0

**Nativo no workspace** (`.runtgine/templates/*.json`). Repo externo /
marketplace fica fora do v0 (`40` G-196 / G-200). Playbooks
continuam markdown em `.runtgine/playbooks/` (`33`).