# Proposal: 039-mcp-memory

## Why

G-44 (`docs/10-gaps.md`) define MCP como **"candidato a transporte
futuro da Project Memory (fora do v0 `29`)"**. O v0 da Project
Memory (`29`, slice 17) entregou o Provider in-process
(`internal/core/memory`) com API determinística (`Record/List/
Query/Supersede/Archive`, G-125) — e listou MCP explicitamente em
G-128 como exclusão "de v0, não permanente".

A lacuna hoje: a memória do workspace só é acessível in-process ou
via CLI local (`runtgine memory`). Agentes externos (Claude Desktop,
IDEs com MCP) não conseguem consultar o guidance compilado
(episódios `decision | failure | handoff | preference`), que é
exatamente o valor do doc `29` ("o que ainda vale como guidance").

Esta spec propõe fechar G-44 como **servidor MCP somente leitura**
sobre o Provider existente:

- Runtgine **expõe** memória via MCP; continua **não sendo
  alternativa ao MCP** (`01` §54, preservado).
- Leitura apenas (`Query` + `List`); escrita continua via CLI/API
  local + Lessons HITL (`33`).
- Mesmo padrão do entrypoint servidor existente: `runtgine serve`
  (`34`) já estabelece loopback-only + bearer token.

Não é cliente MCP, não é Player, não é RAG/embeddings/Knowledge,
não é cross-workspace sync.

## What Changes

- Canonical `docs/39-mcp-memory-v0.md` (G-187..G-193 CONFIRMED)
- Novo entrypoint servidor MCP read-only sobre
  `internal/core/memory` (corte v0)
  - Tools MCP: `memory.query` (busca lexical, só episódios
    `active`) e `memory.list` (listagem filtrada)
- Transporte stdio: comando `runtgine mcp`
- Transporte HTTP: endpoint `/mcp` no `runtgine serve`,
  reutilizando auth existente (loopback + bearer token)
- Falha do Provider degrada (resultado vazio), nunca derruba o
  server
- Sem escrita via MCP; sem capabilities novas no Registry; sem
  schema novo no store
- Examples de configuração para clientes MCP (Claude Desktop,
  IDEs)

## What Does Not Change

- Project Memory Provider (`29`, slice 17) — segue soberano;
  schema G-124 intacto
- Ranking lexical determinístico (G-125) — MCP só expõe `Query`/
  `List`; sem embeddings nem busca semântica
- ContextPack `memory_hits` (`31`) — AssembleContext segue
  consumindo Provider diretamente in-process
- Lessons HITL (`33`) — única via de supersession automática
- HTTP API existente (`34`) — `/mcp` é rota nova, handlers atuais
  intocados
- Players (todos, inclusive `memory.*` do slice 31)
- Task IR schema; Claims / Blast; Policy / Validator / Registry
- TUI tabs; Wails views
- NATS (G-36 DEFERRED); Knowledge base

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `039-mcp-memory` |
| Doc canônico | `docs/39-mcp-memory-v0.md` (a criar) |
| Gaps | G-187..G-193 **CONFIRMED (proposta)** (recorte de G-44) |
| Código | Slice 32 — **bloqueado** até esta spec + `04` |
| Depende | `029-project-memory`, `034-http-api` |

## Approach

1. Pacote `internal/entrypoint/mcpserver` envolvendo o Provider
   por uma interface de leitura mínima (mesmo padrão do `Reader`
   de `38`). Sem SQL direto.
2. Duas tools read-only; input schemas fechados
   (`additionalProperties: false`).
3. Transporte stdio (`runtgine mcp`): JSON-RPC 2.0 sobre stdio,
   processo filho do cliente MCP.
4. Transporte HTTP: handler `/mcp` registrado no `httpapi.Handler`
   com a mesma middleware `auth()`; loopback-only via `CheckBind`.
5. Falha do Provider → tool retorna resultado vazio bem-formado +
   warning slog; server nunca morre por erro de memória.
6. SDK oficial Go (`github.com/modelcontextprotocol/go-sdk`) se
   puro-Go e sem deps pesadas; caso contrário, JSON-RPC 2.0 mínimo
   implementado no pacote (decisão final no design.md).

## Impact

- New package `internal/entrypoint/mcpserver`
- `internal/entrypoint/cli/root.go`: comando `mcp`
- `internal/entrypoint/httpapi/server.go`: rota `/mcp` na mesma
  auth
- README Estágio: Slice 32 após código
- `04-decisoes` §P3 G-44 vira recorte CONFIRMED
  (G-187..G-193) + implementado
