# Runtgine — Revisao Completa do Projeto

Documento unico consolidando todas as decisoes, conceitos, arquitetura
e estado atual do Runtgine. Para revisao e validacao.

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
JSON + JSON Schema (CONFIRMED), SQLite (CONFIRMED), slog (HYPOTHESIS).

Descartadas: Rust (I/O-bound), GPUI (exige Rust), Tauri (dois runtimes),
Electron (+100MB), Python (runtime), Node/TS (single-thread).

## 3. Fluxo Conceitual

Human Intent -> Intent Engine -> Task IR -> Validator ->
Execution Plan -> Event Bus -> Orchestrator ->
Capability Resolver -> Player Router -> Players -> Events -> Graph -> State

## 4. Modelo Conceitual

CONFIRMED: Task, Workflow, Execution Plan, Player, Capability,
Manifest, Event, Queue, Event Bus.

HYPOTHESIS: Intent Engine, Task IR, Task Validator, Runtime Graph,
Context Engine, Orchestrator, Player Router, Execution Policy,
Resource Claim, Blast Radius, Background Player.

Distinga: Task != Workflow != Execution Plan.
Event != Queue != Workflow. Player != Agent.
Runtgine != Chorus (complementares).

## 5. Arquitetura do Core

Tasks (Queue, Planner, Validator, Intent Engine)
Events (Event Bus, Event Store, Streams)
Players (Registry, Capabilities, Execution)
Policy Engine
Runtime Graph

Interfaces: CLI, TUI, Wails, API/Webhook — todas convergem
para o Public Protocol. Core e o produto. Interface e superficie.

## 6. Entry Points

CLI (MVP), TUI (MVP), Board/Github (MVP), Wails (Fase 3),
API (pos-MVP), Webhooks (futuro), Scheduler (futuro).
Todos convergem para o mesmo protocolo interno.
Entry Point != Player.

## 7. Players

Deterministic: Shell, Git, Filesystem, Docker, K8s, Terraform, Test.
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

Ciclo: Board -> Import -> Tech Review -> Spec Review ->
Repo Search -> Effort Estimation -> Difficulty ->
Task Decomposition -> Task Router -> Context Assembly -> Players

Inclui: Board Integration, Task model, pipeline de analise,
Repo Search, Effort Estimation, Decomposition, Router,
Context assembly, Player Registry, Event Bus, CLI.

Nao inclui: Shell Player, Workflow engine, Human-in-the-loop,
Policies, Plugin system, Wails, MCP, Event sourcing, API, NATS.

Ordem de implementacao: Board -> Registry -> Event Bus ->
Context -> LLM Player -> Pipeline -> Repo Search -> Effort ->
Decomposition -> Router -> CLI -> Mais Players.

## 10. Roadmap

Fase 0: Documentacao (atual)
Fase 1: MVP (Core Go, Event Bus, CLI, TUI)
Fase 2: Players (Shell, Git, LLM, Test)
Fase 3: Desktop (Wails)
Fase 4: Infra (SQLite, policies, blast radius)
Fase 5: Cloud (NATS, API, serverless)
Fase 6: Ecossistema (biblioteca de Players)

## 11. Status Geral

12 arquivos .md, ~2100 linhas.
Conceitos CONFIRMED: 15+. HYPOTHESIS: 11.
Stack toda CONFIRMED. Proximo passo: prototipacao.