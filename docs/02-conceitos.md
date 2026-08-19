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
| Intent Engine | CONFIRMED (v0) | Traduz intencao em Task IR; ver `17` |
| Task IR | CONFIRMED (v0) | Schema em 11-protocolo-v0; NL via Intent Engine v0 |
| Task Validator | HYPOTHESIS | Valida antes de executar |
| Runtime Graph | CONFIRMED (v0) | Memoria estrutural; ver `18` |
| Context Engine | HYPOTHESIS | Monta contexto para cada Player |
| Orchestrator | HYPOTHESIS | Coordena fluxo de execucao |
| Execution Policy | CONFIRMED (v0) | allow/deny/approval-required; ver `22` |
| Resource Claim | CONFIRMED (v0) | Bloqueio concorrente; ver `24` |
| Blast Radius | CONFIRMED (v0) | Relatorio de impacto da Task IR; ver `25` |
| Blast Graph Walk | CONFIRMED (v0) | 1 hop mentions no Graph; ver `27` |
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

Status: CONFIRMED (v0) — ver `17-intent-engine-v0.md`

O usuario pode escrever: Pega a ultima versao da API, roda os
testes e coloca no staging. Em vez de executar diretamente,
o Intent Engine traduz essa intencao para Task IR.

Funciona como um compilador:

Human Intent -> Intent Engine -> Task IR -> Validator -> Execution Plan

No MVP inicial, a entrada tipica era Task IR v0 estruturado (JSON/YAML
via CLI/Board). Intent Engine NL v0 (`runtgine intent`) compila texto
para Task IR com heuristicas deterministicas e LLM opcional.

O Intent Engine e especializado em Runtgine Protocol,
Players, Capabilities e Task Schemas.
Ele NAO e autoridade — se inventar uma capability que nao existe,
o Registry/Validator rejeita.

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

Status: CONFIRMED (v0) — ver [18-runtime-graph-v0.md](18-runtime-graph-v0.md)

Representa relacoes entre Players, Capabilities, Tasks,
Resources, Repositories, Symbols, Runs e artefatos de path.

Enquanto o Event Bus responde O que esta acontecendo agora?,
o Runtime Graph responde O que existe e como as coisas se relacionam?.

Runtime Graph = memoria estrutural
Event Store = memoria temporal

O corte estrutural v0 (G-60..G-65) limita-se a nos/arestas minimos em SQLite
por workspace, sync best-effort apos runs, e CLI read-only — sem Workflow
Templates e sem tab TUI. Hits no ContextPack/Intent sao o slice Graph Hits
v0 ([19-graph-hits-v0.md](19-graph-hits-v0.md), G-66..G-69).

---

## Player

Status: CONFIRMED

Player nao e sinonimo de Agent. Player e qualquer entidade capaz
de fornecer capabilities.

Tipos: Deterministic, AI, Human, Service, Workflow Player.

Exemplos: Git Player, Filesystem Player, Shell Player, Docker Player,
K8s Player, Terraform Player, PostgreSQL Player, Test Player,
HTTP Player, Claude Player, GPT Player, Human Approval Player.

HTTP Player v0 (`28`, G-117..G-122): cliente HTTPS `http.get` /
`http.head`. Nao e a API HTTP do runtime (G-45).

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

Project Memory v0 (`29`, G-123..G-128) alimenta o ContextPack com
`memory_hits` operacionais (`active`). Nao substitui o Context Engine
completo (ainda HYPOTHESIS).

---

## Player Router

Status: HYPOTHESIS

Task -> Required Capability -> Player Candidates -> Router -> best Player

Criterios de escolha: capability, complexidade, custo, latencia, contexto, policy.

---

## Execution Policy

Status: CONFIRMED (v0) — ver [22-execution-policy-v0.md](22-execution-policy-v0.md).

Regras de seguranca/permissao por **capability** exata, no Core (nao e Player):

```text
allow | deny | approval-required
```

Default global: `allow`. Manifest pode declarar o verbo; `config.json`
sobrescreve. `deny` rejeita na admissao; `approval-required` pausa o Run
(`waiting_approval`) ate `ApproveRun`.

Fora do motor de policy: wildcards, Blast Radius como gate, Human Player.
Resource Claims v0: ver [24-resource-claims-v0.md](24-resource-claims-v0.md).
Blast Radius v0 (analise on-demand): ver [25-blast-radius-v0.md](25-blast-radius-v0.md).

---

## Resource Claim

Status: CONFIRMED (v0) — ver [24-resource-claims-v0.md](24-resource-claims-v0.md).

Bloqueio concorrente exclusivo no Core (nao e Player). Kinds v0:
`workspace` e `path`. O Runner deriva o claim de uma tabela automatica
(`fs.write`, `git.add`/`commit`, `docker.build`, `docker.run` com
mount). `shell.exec` nao claima. Conflito e fail-fast
(`claim.conflict`); hold ate o Run terminal.

Garante que dois Runs nao mutem o mesmo recurso simultaneamente.
Blast Radius v0: ver [25-blast-radius-v0.md](25-blast-radius-v0.md).

---

## Blast Radius

Status: CONFIRMED (v0) — ver [25-blast-radius-v0.md](25-blast-radius-v0.md).

Relatorio deterministico a partir da Task IR: o que a Task toca
(inclui leituras) e o que claimaria (tabela G-95), mais overlay dos
claims ativos. CLI `runtgine blast`; nao e gate de Execute e nao
entra no Runner.

Walk `Change -> Graph -> Affected` no v0 e **1 hop** inbound
`mentions` a partir de `touches` path — ver
[27-blast-graph-walk-v0.md](27-blast-graph-walk-v0.md). Players,
Workflows e Symbols no walk permanecem fora.

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