# 05 — PRD: Product Requirements Document

## Problema

Ferramentas modernas executam agentes, CLIs, scripts e servicos,
mas cada componente funciona isoladamente. O Runtgine resolve
isso criando uma camada comum de execucao e orquestracao.

## Personas

- Dev: usa CLI/TUI para tarefas do dia a dia
- Tech Lead: gerencia board e workflows
- QA: executa pipelines de teste
- DevOps: automatiza deploy e infra
- LLM Player especializado: recebe contexto montado pelo Context Engine

## Casos de uso

UC-01: Usuario escreve intencao -> Intent Engine traduz ->
Validator valida -> Execution Plan -> Players executam

UC-02: CI/CD pipeline com Players deterministicos + LLM para diagnostico

UC-03: Task de board Kanban passa por decomposition e routing

## Requisitos

P0: Intent Engine basico, Task IR, Validator, Event Bus in-process,
Player Registry, Shell Player, CLI/TUI minima

P1: Runtime Graph, Context Engine, Player Router, SQLite

P2: Execution Policies, Resource Claims, Blast Radius, NATS, cloud

P3: Biblioteca grande de Players deterministicos (Git, Docker, K8s...)