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

UC-01: Usuario fornece Task IR (JSON/YAML) via CLI -> Validator valida
-> Execution Plan -> Players executam (Shell e/ou LLM)

UC-02: CI/CD pipeline com Players deterministicos + LLM para diagnostico

UC-03: Task de board Kanban passa por decomposition e routing

UC-04 (pos-MVP Core): Usuario escreve intencao em linguagem natural ->
Intent Engine traduz -> Task IR -> Validator -> execucao

## Requisitos

P0 (MVP — ver `09-mvp.md`):
Task IR v0, Validator basico, Event Bus in-process, Player Registry,
Shell Player, CLI/TUI minima, Board Integration + pipeline vertical basico

P1: Intent Engine (NL), Runtime Graph, Context Engine completo,
Player Router completo
(SQLite para runs/events ja no MVP Core — ver G-13 CONFIRMED)

P2: Execution Policies (`22`), Resource Claims (`24`), Blast Radius (`25` spec), NATS, cloud

P3: Biblioteca grande de Players deterministicos (Git, Docker, K8s...)

## Escopo detalhado do MVP

Corte canônico em [09-mvp.md](09-mvp.md). Em conflito entre este PRD
e o MVP, prevalece `09-mvp.md` + `04-decisoes.md`.
