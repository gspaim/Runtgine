# Tasks: 039-mcp-memory

## Docs (this change)

- [x] `docs/39-mcp-memory-v0.md` — G-187..G-193
- [x] Cross-refs em `04`, `10`, `09`, `01`, `05`, `29`, `31`, `33`,
      `34`, `16`
- [x] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [x] OpenSpec `039-mcp-memory` (esta change)
- [x] `04-decisoes` §P3 G-44 vira recorte CONFIRMED
      (G-187..G-193)

## Slice 32 — MCP Memory Server (corte v0)

- [x] `internal/core/memory` expõe interface de leitura p/ MCP
      (`Query`, `List`) — Provider já cobre; só assinatura pública
- [x] Pacote `internal/entrypoint/mcpserver` (Server stdio +
      Handler HTTP; tools `memory.query`, `memory.list`; só
      `active`; input schemas fechados)
- [x] Decisão SDK oficial vs JSON-RPC mínimo registrada no design
      (JSON-RPC mínimo escolhido: superfície pequena, zero deps)
- [x] CLI: comando `runtgine mcp` (stdio; reutiliza `openCore`)
- [x] httpapi: rota `/mcp` com a mesma auth middleware
- [x] Unit tests com stub Reader (handshake, query, list, erro,
      401 via auth existente)
- [x] Examples de config de cliente MCP (`examples/mcp-claude-desktop.json`)
- [x] README seção MCP + Estágio: Slice 32
- [x] `go test ./...` / `go vet ./...` verdes
- [x] Arquivar OpenSpec `039` após código
