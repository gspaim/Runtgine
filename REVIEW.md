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

No MVP: entrada estruturada (Task IR v0) via CLI/Board; Intent Engine NL
v0 disponivel via `runtgine intent` (pos-Core).

## 4. Modelo Conceitual

CONFIRMED: Task, Workflow, Execution Plan, Player, Capability,
Manifest, Event, Queue, Event Bus, Entry Point != Player.

HYPOTHESIS: Context Engine, Player Router,
Execution Policy, Resource Claim, Blast Radius, Background Player,
Workflow Template.

CONFIRMED (v0): Intent Engine, Task IR, Task Validator (subset), Runtime Graph.

Distinga: Task != Workflow != Execution Plan.
Event != Queue != Workflow. Player != Agent.
Entry Point != Player. Runtgine != Chorus.

## 5. Arquitetura do Core

Tasks (Queue, Planner, Validator, Intent Engine)
Events (Event Bus, Event Store, Streams)
Players (Registry, Capabilities, Execution)
Policy Engine
Runtime Graph

Interfaces: CLI, TUI, Board, Wails, API/Webhook — todas convergem
para o Public Protocol. Core e o produto. Interface e superficie.

## 6. Entry Points

CLI (MVP), TUI (MVP), Board/Github (MVP), Wails (Fase 3),
API (pos-MVP), Webhooks (futuro), Scheduler (futuro).
Todos convergem para o mesmo protocolo interno.
Entry Point != Player.

## 7. Players

Deterministic: Shell (MVP), Git, Filesystem, Docker, K8s, Terraform, Test.
AI: Claude, GPT, Gemini, local LLM. Human: Approval. Service: PG, HTTP.

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
Human-in-the-loop, Policies, Plugin system, Wails, MCP, Event sourcing,
API, NATS. Intent Engine NL e Runtime Graph foram promovidos pos-Core
(`17` / `18`) e ja tem codigo; Graph Hits (`19`) e o proximo slice.

Ordem: Task IR -> Registry -> Event Bus -> Validator -> Shell ->
CLI -> TUI -> Board -> Context -> LLM pipeline -> Router.

## 10. Roadmap

Fase 0: Documentacao
Fase 1: MVP (Core Go, Event Bus, Shell, CLI, TUI, Board) — slices 1–4
Fase 2: Graph Hits v0 (slice 7) + mais Players / policies conforme promocao
em `04`; Project Memory so apos sair de HYPOTHESIS.
Fase 3: Desktop (Wails)
Fase 4: Infra (policies, blast radius; SQLite ja no MVP)
Fase 5: Cloud (NATS, API, serverless)
Fase 6: Ecossistema (biblioteca de Players)

## 11. Status Geral

Documentacao alinhada; stack CONFIRMED; MVP canônico em 09-mvp.md.
**P0 + P1 + P2 CONFIRMADOS** (`11`, `12`, `13`; NATS DEFERRED).
**Intent Engine v0 CONFIRMADO** (`17`, G-50..G-54).
Slices 1–6 implementados (Core → Intent → Runtime Graph estrutural).
**Runtime Graph v0: CONFIRMED** em `18` (G-60..G-65).
**Graph Hits v0: CONFIRMED** em `19` (G-66..G-69) — liberado codigo slice 7.
P3 (G-40+) futuro.
