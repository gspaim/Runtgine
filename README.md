# Runtgine

> **Runtgine is a lightweight, protocol-first, event-driven universal runtime for executing and orchestrating heterogeneous Players — LLMs, tools, scripts, humans, services and workflows — with deterministic execution preferred whenever possible.**

> **Do not build an AI orchestration framework. Build an execution runtime in which AI is one possible capability.**

Runtgine é um **runtime universal de execução e orquestração**: leve, orientado a eventos e baseado em protocolo. Ele coordena unidades de trabalho heterogêneas — LLMs, ferramentas, scripts, humanos, serviços e workflows — sem depender de nenhuma delas. LLMs são apenas um dos possíveis participantes, nunca o centro do sistema.

## Status do projeto

**Fase 1 (slice 1) — Core mínimo em Go.** Protocolo v0 / P1 / P2 documentados e confirmados.

```bash
go test ./...
go build -o bin/runtgine ./cmd/runtgine
./bin/runtgine run examples/hello.json
./bin/runtgine status <run_id>
```

Store local: `workspace/.runtgine/runtgine.db`

## O que é

- Um runtime de execução e orquestração baseado em protocolo
- Orientado a eventos, com estado derivado do fluxo de eventos
- **Deterministic-first**: usa IA apenas quando necessário
- Local-first, CLI/runtime-first, com UI desktop (Wails) como camada posterior
- Extensível via Players que declaram capacidades em um manifest

## O que não é

- Não é um chatbot, IDE ou framework exclusivo de agentes
- Não é um wrapper de LLM nem uma alternativa ao MCP
- Não é um sistema onde todo trabalho passa por IA

## Estrutura do repositório

```text
AGENTS.md            — guia para LLMs e contribuidores
cmd/runtgine/        — CLI
internal/core/       — runtime (task, event, runner, store, …)
internal/players/    — Shell (slice 1); LLM depois
internal/entrypoint/ — CLI (TUI depois)
examples/            — Task IR de exemplo
docs/                — documentação oficial
```

## Leitura recomendada

1. [docs/01-visao.md](docs/01-visao.md) — visão consolidada
2. [docs/02-conceitos.md](docs/02-conceitos.md) — modelo conceitual
3. [docs/03-principios.md](docs/03-principios.md) — princípios
4. [docs/04-decisoes.md](docs/04-decisoes.md) — decisões e status
5. [docs/09-mvp.md](docs/09-mvp.md) — corte do MVP
6. [docs/10-gaps.md](docs/10-gaps.md) — gaps antes do codigo
7. [docs/11-protocolo-v0.md](docs/11-protocolo-v0.md) — schemas PROPOSED
8. [AGENTS.md](AGENTS.md) — como trabalhar no projeto
