# 01 — Visao consolidada

## O que e o Runtgine

Runtgine e um runtime/orquestrador de engenharia, focado em
receber intencoes/tarefas, transforma-las em execucao estruturada
e coordenar Players deterministicos, LLMs, servicos e humanos
atraves de eventos, filas e capabilities.

Nao e um sistema de agentes LLM. E um runtime que pode usar LLM
como uma das capacidades disponiveis.

## Runtgine e Chorus

Diferente do que foi documentado anteriormente, Runtgine e Chorus
NAO sao predecessor/sucessor. Sao complementares:

- Runtgine: runtime, execucao, Players, workflows, recursos
- Chorus: protocolo/comunicacao/orquestracao entre componentes

## Fluxo conceitual

Human Intent -> Intent Engine -> Task IR -> Validator ->
Execution Plan -> Event Bus -> Orchestrator -> Capability Resolver
-> Player Router -> Players -> Events -> Graph -> State

## Arquitetura

RUNTGINE CORE:
- Tasks: Queue, Planner, Validator, Intent Engine
- Events: Event Bus, Event Store, Streams
- Players: Registry, Capabilities, Execution
- Policy Engine
- Runtime Graph

Interfaces (superficies sobre o Core):
- CLI (automacao, scripting, CI/CD)
- TUI (primeira interface real do projeto)
- GPUI (futura interface desktop nativa)
- API, Webhooks, Slack, Scheduler (todos convergem para o Task Protocol)

Core e o produto. A interface e uma superficie sobre ele.
UI nunca chama Player diretamente.

## O que NAO e

Nao e chatbot, IDE, framework de agentes, wrapper de LLM,
alternativa ao MCP, workflow SaaS, dashboard web.
E um runtime que transforma intencao em execucao verificavel.

## Norte absoluto

Runtgine nao e um agente que executa coisas. E um runtime que
transforma intencao em execucao verificavel, usando determinismo
sempre que possivel e inteligencia quando necessario.