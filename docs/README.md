# Documentacao do Runtgine

Ordem de leitura recomendada: 01 → 09, depois 10–11 (gaps e protocolo).
Autoridade de decisoes: [04-decisoes.md](04-decisoes.md).

| # | Documento | Conteudo |
|---|---|---|
| 00 | 00-rascunho.md | Discussoes em andamento (nao autoridade) |
| 01 | 01-visao.md | Visao, arquitetura, relacao com Chorus |
| 02 | 02-conceitos.md | Modelo conceitual completo |
| 03 | 03-principios.md | Principios de design |
| 04 | 04-decisoes.md | Decisoes com status (fonte de verdade) |
| 05 | 05-prd.md | PRD e requisitos |
| 06 | 06-glossario.md | Glossario de termos |
| 07 | 07-stack.md | Stack tecnologica e justificativas |
| 08 | 08-workflow-templates.md | Workflow Templates, TLC Spec-Driven |
| 09 | 09-mvp.md | Corte canônico do MVP |
| 10 | 10-gaps.md | Gaps e definicoes faltantes |
| 11 | 11-protocolo-v0.md | Protocolo v0 confirmado |
| 12 | 12-board-p1.md | Board / pipeline vertical |
| 13 | 13-p2.md | Engenharia P2 (G-30+) |
| 14 | 14-tui-design.md | TUI Constellation Mission Control |
| 15 | 15-git-workflow.md | Branches, RC e releases |
| 16 | 16-project-memory.md | Project Memory (esboco; HYPOTHESIS) |
| 17 | 17-intent-engine-v0.md | Intent Engine NL v0 (CONFIRMED) |
| 18 | 18-runtime-graph-v0.md | Runtime Graph v0 (**PROPOSED**; nao codificar ate `04`) |

## Fontes historicas (raiz do repo)

| Arquivo | Papel |
|---|---|
| brainstorm.md | Visao original; desatualizado em stack (Rust/GPUI) |
| conversas-empryo.md | Consolidacao de discussoes; desatualizado em stack |
| REVIEW.md | Resumo executivo; deve espelhar docs/ |

Nao usar fontes historicas para decisoes de implementacao.

Antes de codar: P0 (`11`), P1 Board (`12`), P1b Intent (`17`) e P2 (`13`)
estao **CONFIRMADOS** (G-36 NATS = DEFERRED). Runtime Graph (`18`, G-60+)
e **PROPOSED** — confirmar em `04` antes de implementar. P3 ainda futuro
em `10-gaps.md` (inclui Project Memory G-46/G-47 em `16` — HYPOTHESIS, sem codigo).
