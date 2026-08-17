# Proposal: 021-filesystem-player

## Why

O Runtgine tem Shell e Git Players, mas não possui capabilities
determinísticas para ler, escrever, listar e inspecionar arquivos. Usar
`shell.exec` para filesystem perde contratos de input/output e limites
de segurança.

## What Changes

- Novo Player `filesystem` em `internal/players/filesystem`
- Capabilities: `fs.read`, `fs.write`, `fs.list`, `fs.stat`
- Confinamento ao workspace com rejeição de symlink externo
- Limites de bytes/entries, UTF-8 no v0 e escrita atômica
- Registro em `api.Open`, static validation e `examples/fs-read.json`

## What Does Not Change

- Shell / Git / pipeline / LLM Players
- Task IR schema
- Delete, move, copy, chmod, symlink e execução
- HITL / Execution Policy / Claims / Blast Radius
- TUI e heurísticas específicas do Intent

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `021-filesystem-player` |
| Doc canônico | [`docs/21-filesystem-player-v0.md`](../../docs/21-filesystem-player-v0.md) |
| Gaps | G-75..G-80 **CONFIRMED** (recorte de G-41) |
| Código | Ainda não — este pacote autoriza o slice 9 |

## Approach

1. Manifest com quatro capabilities e schemas estritos
2. Resolver paths dentro do workspace, sem seguir symlink externo
3. Implementar operações com APIs Go, sem Shell Player
4. Cobrir limites, UTF-8, atomicidade e cancelamento em testes
5. Registrar no Core e atualizar Runtime Graph via Registry

## Impact

- Package novo: `internal/players/filesystem`
- Graph ganha capabilities `fs.*` no refresh
- Nenhuma mudança no Task IR ou no protocolo de eventos
