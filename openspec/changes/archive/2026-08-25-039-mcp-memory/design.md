# Design: 039-mcp-memory

## Pacote

`internal/entrypoint/mcpserver` com `New(reader Reader, log
*slog.Logger)`, `Server()` (stdio) e `Handler() http.Handler`
(HTTP). Não é Player: não entra no Registry, não tem capabilities,
não participa de Pipeline. É um **entrypoint** (irmão de
`httpapi`, `tui`, `desktop`), como `serve`.

## Interface de leitura

Mesmo padrão do `Reader` de `38`: interface mínima em
`internal/core/memory`, já coberta pelo Provider existente.

```go
type MCPReader interface {
    Query(ctx context.Context, text string, limit int) ([]Episode, error)
    List(ctx context.Context, f Filter) ([]Episode, error)
}
```

Provider existente já implementa; a spec **expõe** sem mudar a
implementação. Sem migração de dados.

## Tools MCP (G-188)

| Tool | Input | Output (sucesso) |
|---|---|---|
| `memory.query` | `text` (string UTF-8, 1–512), `limit?` (1–20, default 8) | `hits[]` (id, kind, title, snippet ≤200 runes, score) |
| `memory.list` | `kind?` (enum 4), `limit?` (1–100, default 20) | `episodes[]` (id, kind, title, created_at) |

- Só episódios `validity=active` (default recall do G-125;
  `superseded`/`archived` nunca aparecem).
- Output é JSON estável; `snippet` usa o truncamento já existente.
- Input schema com `additionalProperties: false`.
- **Sem** tools de escrita (`record`/`supersede`/`archive`) — nem
  no Manifest.

## Fronteiras

1. **Read-only.** O server nunca grava. Escrita continua via CLI/
   API local + Lessons HITL (`33`). Coerente com "memória não é
   autoridade" (`29` §2): o agente externo lê guidance; não muda
   Policy/Validator.
2. **Não é alternativa ao MCP.** Visão `01` §54 preservada:
   Runtgine *expõe* memória via MCP; não compete com ele.
3. Sem rede externa. HTTP fica loopback-only (`CheckBind` do
   `httpapi`); stdio é pipe direto do cliente.
4. Sem Authority. Resultado das tools informa o agente externo;
   não altera Core algum.
5. Sem embeddings / RAG / Knowledge / cross-workspace.

## Transporte stdio (G-189)

Comando `runtgine mcp`:

- Abre Core via `openCore(*workspace, ...)` (mesmo path do serve).
- JSON-RPC 2.0 sobre stdin/stdout (MCP stdio); stderr só para
  warnings/diagnóstico.
- Processo filho do cliente MCP (Claude Desktop, IDEs).
- Sem auth (não há socket); isolamento = processo + workspace.

## Transporte HTTP (G-190)

Rota `/mcp` no `runtgine serve`:

- Registrada em `httpapi.Handler()` com a mesma middleware `auth()`
  (bearer token obrigatório).
- Binding continua loopback-only via `CheckBind` — nada novo.
- Mesmo lifecycle do server existente (nenhum processo extra).

## Segurança (G-191)

- Falha do Provider → tool retorna lista vazia bem-formada +
  warning slog; server nunca morre por erro de memória.
- Corpo de episódios já é sanitizado no Record (sem segredos/raw
  env, regra do schema G-124) — o server não adiciona exposição
  nova.
- Bearer token obrigatório no transporte HTTP (herdado do `34`);
  sem token → 401 antes de tocar o Provider.

## SDK vs JSON-RPC mínimo

Preferência: SDK oficial Go (`github.com/modelcontextprotocol/
go-sdk`) **se** puro-Go e sem deps pesadas (verificação no início
do slice). Caso contrário: JSON-RPC 2.0 mínimo implementado no
pacote (initialize, tools/list, tools/call, ping — superfície
pequena e estável para servidor read-only).

## Integração Core

1. `internal/core/memory` expõe interface de leitura p/ MCP
   (assinatura pública; Provider intacto)
2. `internal/entrypoint/mcpserver` com Server stdio + Handler HTTP
3. CLI: comando `mcp` em `root.go` (reutiliza `openCore`)
4. httpapi: rota `/mcp` na mesma auth middleware
5. Examples: config de cliente MCP (Claude Desktop JSON, IDE)
6. README: seção "MCP" em Começando

## Testes (slice 32)

- Stub Reader: query com hits → `hits[]` populado, tool ok
- Stub Reader: 0 hits → lista vazia, tool ok
- Stub Reader: erro → resultado vazio bem-formado + warning slog,
  server vivo
- Handshake stdio: initialize → resposta correta; tools/list → 2
  tools; tools/call desconhecida → erro JSON-RPC bem-formado
- HTTP: `/mcp` sem token → 401; com token → tools/list ok
- `text` vazio ou >512 → erro de validação da tool
- `kind` inválido em `memory.list` → erro de validação
- `go test ./...` / `go vet ./...` verdes

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Cliente MCP (Runtgine consulta providers externos) | Inverte o papel do G-44 ("transporte da Project Memory"); muda Context Engine (`31`); escopo de outro track |
| Tools de escrita via MCP | Escrita remota viola HITL de Lessons (`33`); memória não pode ser mutada por agente externo |
| Só stdio | Bloqueia uso programático local via serve; HTTP já existe com auth pronta |
| Só HTTP | Exigiria server rodando p/ uso individual; stdio é o padrão dos clientes MCP |
| Embeddings / busca semântica nas tools | Recorte `29` é lexical determinístico |
| Expor Event Store / Graph via MCP também | Outro domínio; G-44 fala de Project Memory |
| Memory Player via MCP | Player já existe in-process (slice 31); MCP é entrypoint, não Player |

## Risks

| Risco | Mitigação |
|---|---|
| Agente externo trata hits como ordem | Doc + description das tools deixam claro: guidance informativo, não autoridade |
| Exposição acidental de episódios arquivados/superseded | Filtro `active` fixo no server; testes cobrem |
| SDK MCP pesado / deps problemáticas | Fallback: JSON-RPC 2.0 mínimo (decisão documentada no slice) |
| Server HTTP expandido vira superfície nova | Rota sob mesma auth + loopback; nenhum handler antigo tocado |
| Falha do SQLite derruba o server | Degrada para vazio + warning (mesmo padrão de `38`) |

## Critérios de aceite

1. Tools de escrita ausentes → cliente MCP não as vê em
   `tools/list`.
2. `text` vazio → erro de validação da tool.
3. Stub Reader com hits → `hits[]` populado.
4. Stub Reader com erro → resultado vazio + warning, server vivo.
5. Só episódios `active` aparecem (superseded/archived nunca).
6. `/mcp` sem bearer token → 401; com token → funciona.
7. `runtgine mcp` responde initialize/tools/list/tools/call via
   stdio.
8. `go test ./...` e `go vet ./...` verdes.
9. OpenSpec `039` arquivado após código (slice 32).
