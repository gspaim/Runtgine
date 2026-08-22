# 01 — Visao consolidada

## O que e o Runtgine

Runtgine e um runtime/orquestrador de engenharia, focado em
receber intencoes/tarefas, transforma-las em execucao estruturada
e coordenar Players deterministicos, LLMs, servicos e humanos
atraves de eventos, filas e capabilities.

Nao e um sistema de agentes LLM. E um runtime que pode usar LLM
como uma das capacidades disponiveis.

## Runtgine e Chorus

Runtgine e Chorus NAO sao predecessor/sucessor. Sao complementares:

- Runtgine: runtime, execucao, Players, workflows, recursos
- Chorus: protocolo/comunicacao/orquestracao entre componentes

## Fluxo conceitual

Human Intent -> Intent Engine -> Task IR -> Validator ->
Execution Plan -> Event Bus -> Orchestrator -> Capability Resolver
-> Player Router -> Players -> Events -> Graph -> State

No MVP realizado, a entrada tipica e Task IR estruturado (CLI/Board)
ou `runtgine intent` (NL). O 1.0 magro (ver `09`) adiciona heuristicas
de Player e Context Engine v0; API HTTP v0 esta em `34` (slices 25–26
feitas). Desktop Wails v0 esta spec'd em `35` (slices 27–28 feitas).
NPM Player v0 esta spec'd em `36` (G-166..G-171; slice 29).

## Arquitetura

RUNTGINE CORE:
- Tasks: Queue, Planner, Validator, Intent Engine
- Events: Event Bus, Event Store, Streams
- Players: Registry, Capabilities, Execution
- Policy Engine
- Runtime Graph

Interfaces (superficies sobre o Core):
- CLI (automacao, scripting, CI/CD)
- TUI (primeira interface interativa do projeto)
- Wails (futura interface desktop; Go + frontend web)
- API, Webhooks, Slack, Scheduler (todos convergem para o Task Protocol)

Core e o produto. A interface e uma superficie sobre ele.
UI nunca chama Player diretamente.
Entry Point != Player.

## O que NAO e

Nao e chatbot, IDE, framework de agentes, wrapper de LLM,
alternativa ao MCP, workflow SaaS, dashboard web.
E um runtime que transforma intencao em execucao verificavel.

## Norte absoluto

Runtgine nao e um agente que executa coisas. E um runtime que
transforma intencao em execucao verificavel, usando determinismo
sempre que possivel e inteligencia quando necessario.
