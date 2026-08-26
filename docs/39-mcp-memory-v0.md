# 39 — MCP Memory Server v0

Entrypoint **servidor MCP read-only** sobre o Project Memory
Provider (`29`; slice 17 feito): tools `memory.query` e
`memory.list`. Não grava, não supersede, não arquiva. Falha do
Provider **degrada** (vazio + warning), nunca derruba o server.

Inventário: [10-gaps.md](10-gaps.md) (G-187+; recorte de G-44,
"candidato a transporte futuro da Project Memory" desde `29`).
Autoridade de status: [04-decisoes.md](04-decisoes.md). Não é
cliente MCP. Não é Player. Não é RAG. Não é Knowledge base. Não é
alternativa ao MCP (`01`: Runtgine *expõe* memória via MCP; não
compete com ele).

**Status deste doc: CONFIRMED v0 (slice 32 feito).**
G-187..G-193. Fecha G-44 como transporte de leitura da Project
Memory; a direção "servidor, não cliente" é decisão explícita
deste recorte (ver `16` §10, que exigia decidir sidecar vs store —
o store local venceu em `29`; o servidor MCP expõe esse store).

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-25-039-mcp-memory/`](../openspec/changes/archive/2026-08-25-039-mcp-memory/)
(arquivado após o slice 32).

---

## 1. Problema

A memória do workspace só é acessível in-process ou via CLI local
(`runtgine memory`). Agentes externos (Claude Desktop, IDEs com
MCP) não conseguem consultar o guidance compilado — episódios
`decision | failure | handoff | preference` ativos — que é
exatamente o valor do doc `29` ("o que ainda vale como guidance").

G-44 já reservava MCP como candidato a **transporte**; este doc
confirma a direção: **servidor** MCP somente leitura sobre o
Provider existente.

## 2. Fronteiras

| É | Não é |
|---|---|
| Servidor MCP read-only | Cliente MCP |
| Tools `memory.query` / `memory.list` | Escrita remota (record/supersede/archive) |
| Entrypoint (irmão de `serve`) | Player / capability no Registry |
| Provider soberano (Core) | Autoridade sobre Policy/Validator |
| stdio + HTTP loopback autenticado | Rede externa / cross-workspace |
| Falha degrada (vazio + warning) | Falha derruba o server |

Regras:

1. Server **nunca** grava, supersede ou arquiva. Escrita continua
   via CLI/API local + Lessons HITL (`33`).
2. Só episódios `validity=active` são retornados.
3. Ranking lexical determinístico do G-125 preservado — sem
   embeddings, sem busca semântica.
4. Falha do store → resultado vazio bem-formado + `slog.Warn`.
   O server nunca morre por erro de memória.
5. Pacote Go: `internal/entrypoint/mcpserver`. Não entra no
   Registry.
6. Hits são guidance informativo para o agente externo; não
   alteram Policy, Validator, Claims ou Blast.

## 3. Cortes confirmados (G-187+)

### G-187 — Papel e pacote

- Servidor MCP read-only sobre `internal/core/memory`
- Pacote: `internal/entrypoint/mcpserver` (entrypoint, não Player)
- Recorte de G-44: transporte de leitura da Project Memory
- Direção: **servidor**, não cliente (cliente = outro track)

### G-188 — Tools v0

| Tool | Input | Output |
|---|---|---|
| `memory.query` | `text` (1–512), `limit?` (1–20, default 8) | `hits[]` (`id`, `kind`, `title`, `snippet` ≤200 runes, `score`) |
| `memory.list` | `kind?` (enum 4 kinds), `limit?` (1–100, default 20) | `episodes[]` (`id`, `kind`, `title`, `created_at`) |

Schemas de input fechados (`additionalProperties: false`). Sem
tools de escrita nem no Manifest.

### G-189 — Transporte stdio

Comando `runtgine mcp`: JSON-RPC 2.0 sobre stdin/stdout (MCP
stdio), processo filho do cliente MCP, escopado ao workspace Core
(`openCore`). Stderr só para warnings/diagnóstico. Sem auth (não há
socket); isolamento = processo + workspace.

### G-190 — Transporte HTTP no serve

Rota `/mcp` registrada em `httpapi.Handler()` com a mesma
middleware `auth()` (bearer token obrigatório) e binding
loopback-only via `CheckBind` (`34`). Mesmo lifecycle; nenhum
processo extra; handlers existentes intocados.

### G-191 — Segurança e degradação

- Bearer token obrigatório no HTTP (sem token → 401 antes de
  tocar o Provider).
- Falha do Provider → tool responde lista vazia bem-formada +
  warning slog; server vivo para chamadas seguintes.
- Corpo de episódios já sanitizado no Record (sem segredos/raw
  env, regra do schema G-124); o server não adiciona exposição
  nova.
- SDK oficial Go (`github.com/modelcontextprotocol/go-sdk`) se
  puro-Go e sem deps pesadas; caso contrário JSON-RPC 2.0 mínimo
  implementado no pacote (decisão final registrada no slice).

### G-192 — Exclusões v0

- Tools de escrita via MCP (qualquer transporte)
- Cliente MCP (Runtgine consultar providers externos)
- Embeddings / RAG / vector search / Knowledge base
- Cross-workspace / sync remoto
- Event Store ou Runtime Graph expostos via MCP (outro domínio)
- Subscriptions/resources MCP avançados (superfície mínima:
  initialize, ping, tools/list, tools/call)

### G-193 — Critérios de interop e aceite

Handshake MCP correto nos dois transportes; `tools/list` mostra
exatamente as duas tools; chamada a tool desconhecida → erro
JSON-RPC bem-formado; `/mcp` sem token → 401; detalhes na §4.

## 4. Critérios de aceite

1. Tools de escrita ausentes → cliente MCP não as vê em
   `tools/list`.
2. `text` vazio ou >512 → erro de validação da tool.
3. Stub Reader com hits → `hits[]` populado, tool ok.
4. Stub Reader com erro → lista vazia + warning, server vivo.
5. Só episódios `active` aparecem (superseded/archived nunca).
6. `/mcp` sem bearer token → 401; com token → funciona.
7. `runtgine mcp` responde initialize/tools/list/tools/call via
   stdio.
8. `go test ./...` e `go vet ./...` verdes sem rede externa.
9. OpenSpec `039` arquivado após o **código** (slice 32).

## 5. Ordem do slice de código

1. Interface de leitura p/ MCP em `internal/core/memory`
   (`Query`, `List`) — Provider já cobre
2. Pacote `internal/entrypoint/mcpserver` (Server stdio + Handler
   HTTP; decision SDK vs mínimo registrada)
3. CLI: comando `mcp` em `root.go`
4. httpapi: rota `/mcp` na mesma auth middleware
5. Testes stub Reader (handshake, query, list, erro, 401)
6. Examples de config de cliente MCP + README seção MCP
7. `go test ./...` / `go vet ./...` verdes
8. Arquivar OpenSpec `039`

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-187 Papel (servidor MCP read-only;
      `internal/entrypoint/mcpserver`)
- [x] G-188 Tools (`memory.query`, `memory.list`)
- [x] G-189 Transporte stdio (`runtgine mcp`)
- [x] G-190 Transporte HTTP (`/mcp` no serve, mesma auth)
- [x] G-191 Segurança e degradação (bearer, loopback, falha
      degrada)
- [x] G-192 Exclusões v0 (escrita, cliente, embeddings,
      cross-workspace)
- [x] G-193 Interop + aceite
