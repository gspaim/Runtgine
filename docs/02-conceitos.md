# 02 — Modelo conceitual

## Visao geral

Runtgine transforma intencao em execucao verificavel.
O fluxo conceitual:

Human Intent -> Intent Engine -> Task IR -> Validator -> Execution Plan -> Event Bus -> Orchestrator -> Players -> Events -> Graph

---

## Status dos conceitos

| Conceito | Status | Notas |
|---|---|---|
| Task | CONFIRMED | Intencao/pedido do usuario |
| Workflow | CONFIRMED | Estrutura reutilizavel de execucao |
| Execution Plan | CONFIRMED | Plano especifico para UMA execucao |
| Player | CONFIRMED | Entidade executora com capabilities |
| Event | CONFIRMED | Algo aconteceu |
| Queue | CONFIRMED | Trabalho aguardando processamento |
| Event Bus | CONFIRMED | Transporte de eventos |
| Capability | CONFIRMED | O que um Player sabe fazer |
| Intent Engine | HYPOTHESIS | Traduz intencao em Task IR |
| Task IR | CONFIRMED (v0) | Schema em 11-protocolo-v0; Intent Engine NL ainda HYPOTHESIS |
| Task Validator | HYPOTHESIS | Valida antes de executar |
| Runtime Graph | HYPOTHESIS | Memoria estrutural do sistema |
| Context Engine | HYPOTHESIS | Monta contexto para cada Player |
| Orchestrator | HYPOTHESIS | Coordena fluxo de execucao |
| Execution Policy | HYPOTHESIS | Regras de seguranca/permissao |
| Resource Claim | HYPOTHESIS | Bloqueio concorrente de recurso |
| Blast Radius | HYPOTHESIS | Impacto de uma mudanca |
| Entry Point | CONFIRMED | Interface externa; nao e Player |

---

## Entry Point

Status: CONFIRMED

Entry Point traduz sinal do mundo externo para o protocolo interno
(CLI, TUI, Board, API, Wails, Webhooks…). Nao executa trabalho:
Player executa. UI/Board nunca chama Player diretamente.

---

## Task, Workflow, Execution Plan

Tres conceitos distintos, frequentemente confundidos:

### Task
A intencao/pedido do usuario. Ex: Faca deploy da API no staging.

### Workflow
Estrutura reutilizavel. Ex: test -> build -> deploy -> verify

### Execution Plan
O plano especifico criado para AQUELA execucao. Pode mudar
dependendo do contexto, Players disponiveis, recursos, politicas.

---

## Event, Queue, Workflow

Outra distincao importante:

| Conceito | Responde | Exemplo |
|---|---|---|
| Event | Algo aconteceu | deployment.failed |
| Queue | Existe trabalho aguardando | diagnosis.queue |
| Workflow | Estrutura de execucao | test -> build -> deploy |

---

## Intent Engine

Status: HYPOTHESIS (fora do MVP Core; ver `09-mvp.md`)

O usuario pode escrever: Pega a ultima versao da API, roda os
testes e coloca no staging. Em vez de executar diretamente,
o Intent Engine traduz essa intencao para Task IR.

Funciona como um compilador:

Human Intent -> Intent Engine -> Task IR -> Validator -> Execution Plan

No MVP, a entrada e Task IR v0 estruturado (JSON/YAML via CLI/Board).
Intent Engine de linguagem natural vem depois do Core estavel.

O Intent Engine e uma LLM especializada em Runtgine Protocol,
Players, Capabilities, Task Schemas, Policies e Runtime Graph.
Ela NAO e autoridade — se inventar uma capability que nao existe,
o Registry rejeita.

---

## Task Validator

Status: HYPOTHESIS (Validator basico incluso no MVP; ver `09-mvp.md`)

Antes de executar, o Validator verifica:
- capabilities existem
- inputs sao validos
- schemas corretos
- dependencias resolvidas
- resources existem
- permissions permitidas
- policies respeitadas
- execution graph valido

No MVP: subset — capabilities, inputs e schemas.
Filosofia: deslocar erros de runtime error para compile/validation error.

---

## Runtime Graph

Status: HYPOTHESIS

Representa relacoes entre Players, Capabilities, Tasks, Workflows,
Resources, Repositories, Symbols, Events, Runs, Artifacts e
Dependencies.

Enquanto o Event Bus responde O que esta acontecendo agora?,
o Runtime Graph responde O que existe e como as coisas se relacionam?.

Runtime Graph = memoria estrutural
Event Store = memoria temporal

---

## Player

Status: CONFIRMED

Player nao e sinonimo de Agent. Player e qualquer entidade capaz
de fornecer capabilities.

Tipos: Deterministic, AI, Human, Service, Workflow Player.

Exemplos: Git Player, Filesystem Player, Shell Player, Docker Player,
K8s Player, Terraform Player, PostgreSQL Player, Test Player,
HTTP Player, Claude Player, GPT Player, Human Approval Player.

Muitos Players deterministicos sao estrategicos — aumentam utilidade
sem IA, reduzem custo, aumentam confiabilidade. A visao e ter uma
biblioteca grande de Players deterministicos.

### Manifest
Cada Player possui um manifest com capabilities, entradas, saidas e
comportamento. O runtime pensa: preciso da capability X; qual Player
consegue fornece-la?

---

## Context Engine

Status: HYPOTHESIS

O LLM nao recebe todo o projeto. O Context Engine monta:
- Task
- Relevant Events
- Relevant Symbols
- Relevant Resources
- Previous Decisions
- Current State

Isso reduz tokens e melhora a qualidade da execucao.

---

## Player Router

Status: HYPOTHESIS

Task -> Required Capability -> Player Candidates -> Router -> best Player

Criterios de escolha: capability, complexidade, custo, latencia, contexto, policy.

---

## Execution Policy

Status: HYPOTHESIS

Regras de seguranca/permissao por Player ou acao:

filesystem: read
shell: deny
network: deny
production.deploy: approval-required

---

## Resource Claim

Status: HYPOTHESIS

Bloqueio concorrente de recursos. Player A claims resource X.
Recursos podem ser: file, repository, database, environment, deployment, workspace.

Garante que dois Players nao modifiquem o mesmo recurso simultaneamente.

---

## Blast Radius

Status: HYPOTHESIS

Antes de executar uma mudanca, analisa o impacto:

Change -> Graph -> Affected Players, Workflows, Resources, Symbols

Saida: Impact Analysis com workflows, recursos e risco afetados.

---

## Background Players

Status: HYPOTHESIS

Coordinator Player coordena sub-players via eventos:

Coordinator Player -> Research Player, Test Player, Review Player

Nao sao sub-agents — sao Players comuns coordenados por eventos.

---

## Orchestrator

Status: HYPOTHESIS

Coordena o fluxo de execucao. Escuta eventos do Event Bus,
resolv capabilities, roteia para Players, gerencia filas.