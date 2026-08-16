# 04 — Decisoes arquiteturais

## Autoridade

Em conflito entre documentos:
1. Este arquivo (`04-decisoes.md`)
2. Documentos oficiais em `docs/` (exceto `00-rascunho.md`)
3. `AGENTS.md` / `README.md` / `REVIEW.md`
4. `brainstorm.md` e `conversas-empryo.md` (historicos; nao sao autoridade)

Status: CONFIRMED | HYPOTHESIS | OPEN QUESTION | REJECTED

## Stack

| Tecnologia | Status | Notas |
|---|---|---|
| Go | CONFIRMED | Linguagem principal do Core |
| Cobra | CONFIRMED | CLI |
| Bubble Tea | CONFIRMED | TUI |
| Lip Gloss + Bubbles | CONFIRMED | Estilizacao e componentes TUI |
| Wails | CONFIRMED | Desktop (Go + Svelte/React) |
| Canal Go (Event Bus) | CONFIRMED | Pub/sub in-process |
| JSON + JSON Schema | CONFIRMED | Protocolo e contratos |
| SQLite (modernc.org/sqlite) | CONFIRMED | Persistencia local; pure Go |
| log/slog | CONFIRMED | Logger padrao — sessao de fechamento |
| NATS (futuro) | OPEN QUESTION | Event Bus distribuido |
| Rust (Core) | REJECTED | Adiado; stack atual e Go |
| GPUI | REJECTED | Exigiria Rust; preterido por Wails |
| Tauri | REJECTED | Preterido por Wails (dois runtimes) |
| Git | CONFIRMED | Version control |

## Arquitetura

| Decisao | Status | Notas |
|---|---|---|
| Deterministic-first | CONFIRMED | Preferir deterministico |
| Player abstraction | CONFIRMED | Abstracao central |
| Event-driven | CONFIRMED | Task -> Event -> Queue -> Player |
| Capability routing | CONFIRMED | Runtime pensa em capabilities |
| Core = produto | CONFIRMED | Interface e superficie |
| Core independente de UI | CONFIRMED | Core funciona sem TUI/CLI |
| LLM-agnostic | CONFIRMED | Players LLM sao um tipo entre outros |
| Task != Workflow != ExecPlan | CONFIRMED | Tres conceitos distintos |
| Event != Queue != Workflow | CONFIRMED | Tres conceitos distintos |
| Entry Point != Player | CONFIRMED | Entry Point traduz sinal externo |
| Intent Engine | HYPOTHESIS | Traduz intencao NL em Task IR |
| Task IR | CONFIRMED (v0) | Schema em 11-protocolo-v0; NL Intent Engine ainda HYPOTHESIS |
| Task Validator | CONFIRMED (v0) | Subset MVP: capabilities, inputs, schemas; ver 11 |
| Runtime Graph | HYPOTHESIS | Memoria estrutural |
| Context Engine | HYPOTHESIS | Monta contexto relevante |
| Player Router | HYPOTHESIS | Roteia por capability + custo |
| Execution Policy | HYPOTHESIS | Regras de seguranca |
| Resource Claim | HYPOTHESIS | Bloqueio concorrente |
| Blast Radius | HYPOTHESIS | Impact analysis |
| Many deterministic Players | CONFIRMED | Estrategico |
| Runtgine + Chorus | CONFIRMED | Complementares |
| Event Bus in-process (MVP) | CONFIRMED | Canais Go |
| Nativo (nao Electron) | CONFIRMED | Wails |
| Runner v0 | CONFIRMED | Orchestrator minimo do MVP |

## MVP (corte canônico)

Ver [09-mvp.md](09-mvp.md). Decisoes-chave:

| Decisao | Status | Notas |
|---|---|---|
| Shell Player no MVP | CONFIRMED | Prova deterministic-first |
| CLI + TUI minimas no MVP | CONFIRMED | Superficies, nao produto |
| Board no MVP | CONFIRMED | Primeiro Entry Point de produto |
| Entrada estruturada (Task IR v0) no MVP | CONFIRMED | Sem depender de Intent Engine NL |
| Intent Engine NL fora do MVP Core | CONFIRMED | Permanece HYPOTHESIS / P1 |
| Wails fora do MVP | CONFIRMED | Fase 3 |
| Escopo detalhado em 09-mvp.md | CONFIRMED | Prevalece sobre rascunhos |

## Modelo conceitual

| Conceito | Status | Notas |
|---|---|---|
| Task | CONFIRMED | Intencao/pedido do usuario |
| Workflow | CONFIRMED | Estrutura reutilizavel |
| Execution Plan | CONFIRMED | Plano para UMA execucao |
| Player | CONFIRMED | Entidade com capabilities |
| Event | CONFIRMED | Algo aconteceu |
| Queue | CONFIRMED | Trabalho aguardando |
| Intent Engine | HYPOTHESIS | Traduz intencao |
| Task IR | CONFIRMED (v0) | Schema em 11-protocolo-v0 |
| Task Validator | HYPOTHESIS | Valida antes de executar |
| Runtime Graph | HYPOTHESIS | Memoria estrutural |
| Context Engine | HYPOTHESIS | Monta contexto relevante |

## Decisoes CONFIRMED (visao geral)

- Go, Cobra, Bubble Tea, Wails, JSON/JSON Schema, SQLite
- Canal Go como Event Bus (MVP)
- Core independente de UI
- Deterministic-first
- Player abstraction + Capability routing
- Task != Workflow != Execution Plan
- Event != Queue != Workflow
- Entry Point != Player
- Runtgine + Chorus complementares
- Nativo (nao Electron); GPUI/Tauri/Rust-Core rejeitados para o caminho atual
- Biblioteca grande de Players deterministicos (visao)
- MVP: Core + Shell + CLI/TUI + Board (ver 09-mvp.md)

## Protocolo v0 — confirmado (sessao de fechamento)

Inventario: [10-gaps.md](10-gaps.md).
Texto completo: [11-protocolo-v0.md](11-protocolo-v0.md).

**Bloco P0 CONFIRMADO.** Liberado iniciar implementacao do Core.
Gaps P1 (Board/LLM) permanecem abertos.

| Proposta | Status | Notas |
|---|---|---|
| JSON canonico; YAML so na borda CLI | CONFIRMED | G-14 |
| Capability naming `domain.action` | CONFIRMED | G-05 |
| IDs UUID v7; schema_version semver | CONFIRMED | time-ordered; trocado de v4 na sessao |
| Task IR v0 schema | CONFIRMED | G-01 |
| Manifest v0 schema | CONFIRMED | G-02 |
| Event envelope + tipos minimos | CONFIRMED | G-03/G-04 |
| Result/Error + Run lifecycle | CONFIRMED | G-08/G-09 |
| Runner v0 (Orchestrator minimo) | CONFIRMED | G-10; Plan passthrough G-11 |
| Queue in-memory FIFO | CONFIRMED | G-12 multi-run |
| Persistencia MVP Core = SQLite cedo | CONFIRMED | G-13 variante B |
| Core API SubmitTask/GetRun/Subscribe | CONFIRMED | G-07 |
| Shell sandbox v0 (argv, workdir, timeout) | CONFIRMED | G-06/G-18 |
| log/slog | CONFIRMED | G-16 |
| SQLite via modernc.org/sqlite | CONFIRMED | G-15 |
| Go 1.22+; module github.com/gspaim/Runtgine | CONFIRMED | G-37 |
| Layout `cmd/` + `internal/core|players|entrypoint` | CONFIRMED | G-17 |

### Desvios em relacao a proposta inicial

- Queue: multi-run concorrente (nao 1 run so)
- Persistencia: SQLite cedo (nao so memoria)
- IDs: UUID **v7** (nao v4) — melhor localidade em SQLite / ordem temporal

### Resolvido nesta sessao

- Capability Resolver / Planner → absorvidos no Runner v0
- Event Store no MVP → events append-only em SQLite (nao event sourcing)
- Task IR / Validator basico / Runner v0 → CONFIRMED

## Board / pipeline (P1) — em fechamento

Ver [12-board-p1.md](12-board-p1.md).

| Item | Status | Notas |
|---|---|---|
| G-20 Card → Task IR (adapter + polling) | CONFIRMED | Mapeamento titulo/body/ref; token via env |
| G-21 Write-back no board | CONFIRMED | Status + comentario; sem subtasks no board |
| G-22 Contratos por etapa | CONFIRMED | capabilities `pipeline.*`; steps lineares |
| G-23 Regras vs LLM | CONFIRMED | repo-search/effort/difficulty det.; reviews LLM |
| G-24 Context assembly basico | CONFIRMED | ContextPack v0; AssembleContext no Core |
| G-25 LLM Player v0 | CONFIRMED | Interface unica; backends OpenAI-compat + Anthropic |
| G-26 Task Router basico | CONFIRMED | Regras: capability → deterministic → default AI |
| G-27 Subtasks | OPEN | |
