# Runtgine

> **Runtgine is a lightweight, protocol-first, event-driven universal runtime for executing and orchestrating heterogeneous Players — LLMs, tools, scripts, humans, services and workflows — with deterministic execution preferred whenever possible.**

> **Do not build an AI orchestration framework. Build an execution runtime in which AI is one possible capability.**

Runtgine é um **runtime universal de execução e orquestração**: leve, orientado a eventos e baseado em protocolo. Ele coordena unidades de trabalho heterogêneas — LLMs, ferramentas, scripts, humanos, serviços e workflows — sem depender de nenhuma delas. LLMs são apenas um dos possíveis participantes, nunca o centro do sistema.

## Status do projeto

**Fase 0 — Fundação (documentação).** Nenhum código ainda. As tecnologias estão em definição — ver [docs/04-decisoes.md](docs/04-decisoes.md) para o registro de decisões (CONFIRMED / HYPOTHESIS / OPEN QUESTION).

## O que é

- Um runtime de execução e orquestração baseado em protocolo
- Orientado a eventos, com estado derivado do fluxo de eventos
- **Deterministic-first**: usa IA apenas quando necessário
- Local-first, CLI/runtime-first, com UI desktop como camada de observação
- Extensível via Players que declaram capacidades em um manifest

## O que não é

- Não é um chatbot, IDE ou framework exclusivo de agentes
- Não é um wrapper de LLM nem uma alternativa ao MCP
- Não é um sistema onde todo trabalho passa por IA

## Estrutura do repositório

```text
AGENTS.md          — guia para LLMs e contribuidores trabalharem no projeto
brainstorm.md      — visão original consolidada (fonte bruta do projeto)
docs/              — documentação oficial
```

## Leitura recomendada

1. [docs/01-visao.md](docs/01-visao.md) — visão consolidada e norte do projeto
2. [docs/02-conceitos.md](docs/02-conceitos.md) — modelo conceitual (Player, Execution, Task, Event…)
3. [docs/05-prd.md](docs/05-prd.md) — PRD (problema, personas, requisitos, MVP)
4. [AGENTS.md](AGENTS.md) — como trabalhar no projeto
