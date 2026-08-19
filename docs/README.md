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
| 16 | 16-project-memory.md | Project Memory (esboco conceitual; corte v0 em `29`) |
| 17 | 17-intent-engine-v0.md | Intent Engine NL v0 (CONFIRMED) |
| 18 | 18-runtime-graph-v0.md | Runtime Graph v0 (**CONFIRMED**; G-60..G-65) |
| 19 | 19-graph-hits-v0.md | Graph Hits v0 (**CONFIRMED**; G-66..G-69) |
| 20 | 20-git-player-v0.md | Git Player v0 (**CONFIRMED**; G-70..G-74) |
| 21 | 21-filesystem-player-v0.md | Filesystem Player v0 (**CONFIRMED**; G-75..G-80) |
| 22 | 22-execution-policy-v0.md | Execution Policy + HITL v0 (**CONFIRMED**; G-81..G-86) |
| 23 | 23-docker-player-v0.md | Docker Player v0 (**CONFIRMED**; G-87..G-92) |
| 24 | 24-resource-claims-v0.md | Resource Claims v0 (**CONFIRMED**; G-93..G-98) |
| 25 | 25-blast-radius-v0.md | Blast Radius v0 (**CONFIRMED**; G-99..G-104) |
| 26 | 26-tui-graph-v0.md | TUI GRAPH v0 (**CONFIRMED**; G-105..G-110) |
| 27 | 27-blast-graph-walk-v0.md | Walk Blast←Graph v0 (**CONFIRMED**; G-111..G-116) |
| 28 | 28-http-player-v0.md | HTTP Player v0 (**CONFIRMED**; G-117..G-122) |
| 29 | 29-project-memory-v0.md | Project Memory v0 (**CONFIRMED**; G-123..G-128) |
| 30 | 30-test-player-v0.md | Test Player v0 (**CONFIRMED**; G-129..G-134) |
| — | [openspec/](../openspec/README.md) | Pacotes de mudança OpenSpec (`NNN-slug`) |

## Fontes historicas (raiz do repo)

| Arquivo | Papel |
|---|---|
| brainstorm.md | Visao original; desatualizado em stack (Rust/GPUI) |
| conversas-empryo.md | Consolidacao de discussoes; desatualizado em stack |
| REVIEW.md | Resumo executivo; deve espelhar docs/ |

Nao usar fontes historicas para decisoes de implementacao.

Antes de codar: P0 (`11`), P1 Board (`12`), P1b Intent (`17`) e P2 (`13`)
estao **CONFIRMADOS** (G-36 NATS = DEFERRED). Runtime Graph (`18`), Graph
Hits (`19`), Git Player (`20`) e Filesystem Player (`21`, G-75..G-80)
sao **CONFIRMED v0**. Execution Policy + HITL (`22`, G-81..G-86) e Docker
Player (`23`, G-87..G-92) estao **CONFIRMED v0** (codigo = slices 10 e 11).
Resource Claims (`24`, G-93..G-98) esta **CONFIRMED v0** (slice 12 feito).
Blast Radius (`25`, G-99..G-104) esta **CONFIRMED v0** (slice 13 feito).
TUI GRAPH (`26`, G-105..G-110) esta **CONFIRMED v0** (slice 14 feito).
Walk Blast←Graph (`27`, G-111..G-116) esta **CONFIRMED v0** (slice 15 feito).
HTTP Player (`28`, G-117..G-122) esta **CONFIRMED v0** (slice 16 feito).
Project Memory (`29`, G-123..G-128) esta **CONFIRMED v0** (slice 17 feito).
Test Player (`30`, G-129..G-134) esta **CONFIRMED v0** (slice 18 feito).
P3 restante: mais Players; G-45 API HTTP; G-44 MCP.
