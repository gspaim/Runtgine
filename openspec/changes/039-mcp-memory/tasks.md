# Tasks: 039-mcp-memory

## Docs (this change)

- [ ] `docs/39-mcp-memory-v0.md` — G-187..G-193
- [ ] Cross-refs em `04`, `10`, `09`, `01`, `05`, `29`, `31`, `33`,
      `34`, `16`
- [ ] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [ ] OpenSpec `039-mcp-memory` (esta change)
- [ ] `04-decisoes` §P3 G-44 vira recorte CONFIRMED
      (G-187..G-193)

## Slice 32 — MCP Memory Server (corte v0)

- [ ] `internal/core/memory` expõe interface de leitura p/ MCP
      (`Query`, `List`) — Provider já cobre; só assinatura pública
- [ ] Pacote `internal/entrypoint/mcpserver` (Server stdio +
      Handler HTTP; tools `memory.query`, `memory.list`; só
      `active`; input schemas fechados)
- [ ] Decisão SDK oficial vs JSON-RPC mínimo registrada no design
- [ ] CLI: comando `runtgine mcp` (stdio; reutiliza `openCore`)
- [ ] httpapi: rota `/mcp` com a mesma auth middleware
- [ ] Unit tests com stub Reader (handshake, query, list, erro,
      401)
- [ ] Examples de config de cliente MCP (Claude Desktop, IDE)
- [ ] README seção MCP + Estágio: Slice 32
- [ ] `go test ./...` / `go vet ./...` verdes
- [ ] Arquivar OpenSpec `039` após código
