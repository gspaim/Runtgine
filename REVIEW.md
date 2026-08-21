# Runtgine — Revisao Completa do Projeto

Documento unico consolidando decisoes, conceitos, arquitetura
e estado atual do Runtgine. Deve espelhar `docs/`. Em conflito,
prevalece `docs/04-decisoes.md`.

---

## 1. Identidade e Visao

Runtgine e um runtime/orquestrador de engenharia que transforma
intencao em execucao verificavel, usando determinismo sempre que
possivel e inteligencia quando necessario.

Nao e um sistema de agentes LLM. E um runtime que pode usar LLM
como uma das capacidades disponiveis.

**Frase que resume:** Runtgine nao e um agente que executa coisas.
E um runtime que transforma intencao em execucao verificavel.

**Runtgine vs Chorus:** Runtgine = runtime/execucao.
Chorus = protocolo/comunicacao. Sao complementares, nao sucessao.

## 2. Stack Tecnologica

Go (CONFIRMED), Cobra (CONFIRMED), Bubble Tea (CONFIRMED),
Wails (CONFIRMED), Canal Go Event Bus (CONFIRMED),
JSON + JSON Schema (CONFIRMED), SQLite (CONFIRMED), slog (CONFIRMED).

REJECTED para o caminho atual: Rust (Core), GPUI, Tauri, Electron,
Python (runtime), Node/TS (single-thread).

## 3. Fluxo Conceitual

Human Intent -> Intent Engine -> Task IR -> Validator ->
Execution Plan -> Event Bus -> Orchestrator ->
Capability Resolver -> Player Router -> Players -> Events -> Graph -> State

No MVP realizado: Task IR via CLI/Board e `runtgine intent`.
O 1.0 magro (`09`) fecha heuristicas de Player + Context Engine v0.

## 4. Modelo Conceitual

CONFIRMED: Task, Workflow, Execution Plan, Player, Capability,
Manifest, Event, Queue, Event Bus, Entry Point != Player.

HYPOTHESIS: Player Router, Background Player, Workflow Template.

CONFIRMED (v0): Intent Engine, Task IR, Task Validator (subset), Runtime Graph,
Context Engine (semente `repo_hits`; `31`),
Execution Policy + HITL (`22`), Docker Player (`23`), Resource Claims (`24`),
Blast Radius (`25`).

Distinga: Task != Workflow != Execution Plan.
Event != Queue != Workflow. Player != Agent.
Entry Point != Player. Runtgine != Chorus.

## 5. Arquitetura do Core

Tasks (Queue, Planner, Validator, Intent Engine)
Events (Event Bus, Event Store, Streams)
Players (Registry, Capabilities, Execution)
Policy Engine
Resource Claims
Runtime Graph

Interfaces: CLI, TUI, Board, Wails, API/Webhook — todas convergem
para o Public Protocol. Core e o produto. Interface e superficie.

## 6. Entry Points

CLI (MVP), TUI (MVP), Board/Github (MVP), HTTP API (spec `34`;
slices 25–26), Wails (Fase 3), Scheduler (futuro).
Webhook inbound GitHub permanece fora (Board = polling).
Webhooks outbound de Run = G-156.
Todos convergem para o mesmo protocolo interno.
Entry Point != Player.

## 7. Players

Deterministic: Shell (MVP), Git, Filesystem, Docker (spec `23`), K8s, Terraform, Test.
AI: Claude, GPT, Gemini, local LLM. HITL: Core `ApproveRun` (nao e Player; spec `22`).
Service: PG, HTTP.

Visao: biblioteca grande de Players deterministicos.
Manifest: cada Player declara capabilities, entradas e saidas.

## 8. Principios de Design

1. Deterministic-first
2. Player e a abstracao central
3. Muitos Players deterministicos sao estrategicos
4. Event-driven e o coracao
5. Core e o produto. Interface e superficie.
6. Validacao antes da execucao (filosofia de compilador)
7. Entrada flexivel, protocolo unico
8. Runtime Graph = memoria estrutural
9. Contexto relevante, nao todo o projeto
10. LLM-agnostic

## 9. MVP (Fase 1)

Corte canônico: [docs/09-mvp.md](docs/09-mvp.md).

Inclui: Task IR v0, Validator basico, Event Bus, Player Registry,
Shell Player, CLI, TUI minima, Board Integration, pipeline de analise
basico, Context assembly, LLM Player(s), Decomposition, Router.

Nao inclui (corte MVP original em `09`): Workflow engine completo,
Plugin system, Wails, MCP, Event sourcing, API, NATS. Intent Engine NL,
Runtime Graph, Graph Hits, Git/FS Players foram promovidos pos-Core.
Execution Policy + HITL (`22`), Docker Player (`23`) e Resource Claims
(`24`, G-93..G-98) estao CONFIRMED v0 (slices 10–12). Blast Radius
(`25`, G-99..G-104) esta CONFIRMED v0 (slice 13 feito).

Ordem: Task IR -> Registry -> Event Bus -> Validator -> Shell ->
CLI -> TUI -> Board -> Context -> LLM pipeline -> Router.

## 10. Roadmap

Fase 0: Documentacao
Fase 1: MVP (Core Go, Event Bus, Shell, CLI, TUI, Board) — slices 1–4
Fase 2: Graph Hits + Git/FS Players feitos; Execution Policy + Docker
specs em `22`/`23` (slices 10–11); Project Memory v0 em `29` (slice 17 feito).
Fase 3: Desktop (Wails)
Fase 4: Infra (Claims v0 slice 12; Blast Radius v0 slice 13 feitos;
TUI GRAPH v0 spec `26` / G-105..G-110 — slice 14 feito;
Walk Blast←Graph v0 spec `27` / G-111..G-116 — slice 15 feito;
HTTP Player v0 spec `28` / G-117..G-122 — slice 16 feito;
Project Memory v0 spec `29` / G-123..G-128 — slice 17 feito;
Test Player v0 spec `30` / G-129..G-134 — slice 18 feito;
Intent heuristics slice 19 / Context Engine slice 20 — 1.0 magro feito)
Fase 5: Cloud — HTTP API v0 spec `34` (slices 25–26); NATS DEFERRED; serverless depois
Fase 6: Ecossistema (biblioteca de Players)

## 11. Status Geral

Documentacao alinhada; stack CONFIRMED; MVP canônico em 09-mvp.md.
**P0 + P1 + P2 CONFIRMADOS** (`11`, `12`, `13`; NATS DEFERRED).
**Intent Engine v0 CONFIRMADO** (`17`, G-50..G-54).
Slices 1–9 implementados (Core → Intent → Graph → Git → Filesystem).
**Runtime Graph v0: CONFIRMED** em `18` (G-60..G-65).
**Git Player v0: CONFIRMED + implementado** (`20`, G-70..G-74; slice 8).
**Filesystem Player v0: CONFIRMED + implementado** (`21`, G-75..G-80; slice 9).
**Execution Policy + HITL v0: CONFIRMED** (`22`, G-81..G-86; slice 10).
**Docker Player v0: CONFIRMED + implementado** (`23`, G-87..G-92; slice 11).
**Resource Claims v0: CONFIRMED + implementado** (`24`, G-93..G-98; slice 12).
**Blast Radius v0: CONFIRMED + implementado** (`25`, G-99..G-104; slice 13).
**TUI GRAPH v0: CONFIRMED + implementado** (`26`, G-105..G-110; slice 14).
**Walk Blast←Graph v0: CONFIRMED + implementado** (`27`, G-111..G-116; slice 15).
**HTTP Player v0: CONFIRMED + implementado** (`28`, G-117..G-122; slice 16).
**Project Memory v0: CONFIRMED + implementado** (`29`, G-123..G-128; slice 17).
**Test Player v0: CONFIRMED + implementado** (`30`, G-129..G-134; slice 18).
**MVP 1.0 magro: CONFIRMED + implementado** (`09`/`31`, G-135..G-140; slices 19–20).
**Intent Surface v0: CONFIRMED + TUI implementada** (`32`, G-141..G-146; slice 21; Wails Fase 3).
**Evolution v0: CONFIRMED + implementado** (`33`, G-147..G-152; slices 22–24).
**HTTP API v0: CONFIRMED + implementado** (`34`, G-153..G-158; slices 25–26).
P3 restante: G-44 MCP; NATS; Wails; G-41 mais Players.
