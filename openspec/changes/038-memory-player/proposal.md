# Proposal: 038-memory-player

## Why

`docs/29-project-memory-v0.md` fechou G-47 **só pelo lado Provider**:
o Core consulta episódios via `memory_hits` no ContextPack. A spec
`29` explicitou que o **Memory Player** continua **fora** do v0
(G-128, "Sem Player"). A entrada de `04-decisoes` ainda marca como
**OPEN QUESTION**:

> "G-47 Provider local CONFIRMED v0 (`29`); Player OPEN QUESTION
> (fora do v0)"

Esta spec propõe fechar G-47 com **um Player de leitura** — o
contrapartide determinístico do Provider. Mesmo recorte de G-41
("biblioteca ampla de Players"), mas com fronteiras explícitas
para preservar o que `29` já decidiu:

- Memória **não** é autoridade de execução (regra 2 de `29` §2).
- "Evitar falha X" informa o LLM; **não** proíbe reexecutar.
- O Player **lê**; a escrita continua sendo via Provider (HITL
  via Lessons, spec `33`).

A lacuna real hoje: `heuristic.shell|pipeline` não consulta Memory
(G-69, preservado), mas qualquer Task LLM que precise de
"contexto de projeto consolidado" só consegue via ContextPack. Um
Player `memory.*` permite:

- `memory.recall` — buscar episódios `active` por texto.
- `memory.check` — checar se um padrão (comando, path, capability)
  tem episódio `failure` registrado.

Não é MCP (G-44), não é Knowledge base, não é RAG, não é indexação
de transcript, não é mutação.

## What Changes

- Canonical `docs/38-memory-player-v0.md` (G-180..G-186 CONFIRMED)
- Player `memory` em `internal/players/memory` (corte v0)
  - Capability `memory.recall` — query lexical de episódios
  - Capability `memory.check` — booleano "este padrão tem `failure`
    ativo?"
- Sandbox: Provider-only; **sem** escrita; **sem** rede
- Integração com o Provider existente (`internal/core/memory`); sem
  schema novo
- Falha do Provider degrada (retorna vazio), **não** falha o Run
- Intent heuristics: nenhuma (decisão explícita — ver §Alternatives)
- Examples `examples/memory-recall.json`, `examples/memory-check.json`

## What Does Not Change

- Project Memory Provider (`29`, slice 17) — segue soberano
- Event Store / Runtime Graph (`18`, `19`) — memória é outro eixo
- Lessons HITL (`33`); Playbooks; Router — Pipeline continua igual
- ContextPack `memory_hits` — `AssembleContext` segue consumindo
  Provider diretamente (não passa pelo Player)
- Task IR schema (`0.1.0`)
- Claims / Blast (Player de leitura = no touch / no claim)
- G-45 HTTP server; G-44 MCP; Knowledge base
- TUI tabs; Wails views
- npm / pytest / yarn Players

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `038-memory-player` |
| Doc canônico | `docs/38-memory-player-v0.md` (a criar) |
| Gaps | G-180..G-186 **CONFIRMED (proposta)** (recorte de G-47) |
| Código | Slice 31 — **bloqueado** até esta spec + `04` |
| Depende | `029-project-memory`, `033-evolution-v0` (Lessons HITL) |

## Approach

1. Player lê Provider via interface em `internal/core/memory`. Sem
   SQL direto; sem fallback.
2. Duas capabilities read-only; `additionalProperties: false`.
3. Saída em JSON estável (`hits[]` para recall; `has_failure` para
   check).
4. Falha do Provider → step `succeeded` com `hits: []` /
   `has_failure: false` (degrada, não falha).
5. Sem Intent heuristic — Decision Step no Pipeline usa Task IR
   explícita (Pipeline é soberano; Intent não infere leitura de
   memória).
6. Graph: `RefreshFromRegistry` cria nós `memory` / `memory.recall`
   / `memory.check`. `provides` aponta para o Provider.

## Impact

- New package `internal/players/memory`
- `internal/core/api` register + static validation dispatch
- `internal/core/memory` ganha interface pública `Reader` (sem
   mudar `Provider`)
- README Estágio: Slice 31 após código
- `04-decisoes` §OPEN QUESTION G-47 vira CONFIRMED + implementado
