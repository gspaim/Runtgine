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
| 31 | 31-context-engine-v0.md | Context Engine v0 (**CONFIRMED**; G-137..G-139; slice 20 feito) |
| 32 | 32-intent-surface-v0.md | Intent Surface / aba INTENT (**CONFIRMED**; G-141..G-146; slice 21 TUI feito; Wails = spec `35`) |
| 33 | 33-evolution-v0.md | Evolution v0: Router, Playbooks, Lessons (**CONFIRMED**; G-147..G-152; slices 22–24 feitas) |
| 34 | 34-http-api-v0.md | HTTP API v0 / `runtgine serve` (**CONFIRMED**; G-153..G-158; slices 25–26 feitas) |
| 35 | 35-wails-v0.md | Desktop Wails v0 (**CONFIRMED**; G-159..G-165; slices 27–28 feitas) |
| 36 | 36-npm-player-v0.md | NPM Player v0 (**CONFIRMED**; G-166..G-171; slice 29 feito) |
| 37 | 37-pytest-yarn-players-v0.md | Pytest + Yarn Players v0 (**CONFIRMED**; G-172..G-179; slice 30) |
| 38 | 38-memory-player-v0.md | Memory Player v0 (**CONFIRMED**; G-180..G-186; slice 31; fecha G-47) |
| 39 | 39-mcp-memory-v0.md | MCP Memory Server v0 (**CONFIRMED**; G-187..G-193; recorte de G-44; slice 32 feito) |
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
MVP 1.0 magro (`09`, G-135..G-140) esta **CONFIRMED** (slices 19–20 feitos).
Intent Surface (`32`, G-141..G-146) esta **CONFIRMED** (slice 21 TUI feito; Wails = spec `35`).
Evolution v0 (`33`, G-147..G-152) esta **CONFIRMED** (slices 22–24 feitas).
HTTP API (`34`, G-153..G-158) esta **CONFIRMED** (slices 25–26 feitas).
Desktop Wails (`35`, G-159..G-165) esta **CONFIRMED** (slices 27–28 feitas).
NPM Player (`36`, G-166..G-171) esta **CONFIRMED** (slice 29 feito).
P3 restante: G-44 MCP; NATS; G-41 (pytest/yarn/infra).
