# 05 — PRD: Product Requirements Document

## Problema

Ferramentas modernas executam agentes, CLIs, scripts e servicos,
mas cada componente funciona isoladamente. O Runtgine resolve
isso criando uma camada comum de execucao e orquestracao.

## Personas

- Dev: usa CLI/TUI para tarefas do dia a dia
- Tech Lead: gerencia board e workflows
- QA: executa pipelines de teste
- DevOps: automatiza deploy e infra (CLI; API HTTP = `34`, slices 25–26 feitas)
- LLM Player especializado: recebe contexto montado pelo Context Engine

## Casos de uso

UC-01: Usuario fornece Task IR (JSON/YAML) via CLI -> Validator valida
-> Execution Plan -> Players executam (Shell e/ou LLM e/ou Git/FS/…)

UC-02: CI/CD pipeline com Players deterministicos + LLM para diagnostico
— **pos-1.0**; HTTP API v0 em `34` (`runtgine serve`). Desktop = spec `35`.

UC-03: Task de board Kanban passa por decomposition e routing

UC-04: Usuario escreve intencao em linguagem natural -> Intent Engine
traduz -> Task IR -> Validator -> execucao (heuristicas Player no 1.0)

## Requisitos

P0 (MVP realizado — slices 1–18; ver `09-mvp.md`):
Task IR v0, Validator, Event Bus, Registry, Shell + Git/FS/Docker/HTTP/Test,
CLI/TUI, Board + pipeline, Intent NL, Graph, Memory, Policy/HITL, Claims, Blast

P1 (MVP 1.0 magro — G-135..G-140):
Heuristicas Intent → `test.go` / `git.status|diff|log` / `docker.ps`;
Context Engine v0 (semente `repo_hits` a partir do Graph)

P2: API HTTP (G-45 / spec `34`, feito); Desktop Wails (spec `35`); mais Players

P3: NATS (G-36 DEFERRED), MCP (G-44), Workflow Templates (G-40),
Memory Player (G-47), K8s/Terraform/PostgreSQL

## Escopo detalhado do MVP

Corte canônico em [09-mvp.md](09-mvp.md). Em conflito entre este PRD
e o MVP, prevalece `09-mvp.md` + `04-decisoes.md`.
