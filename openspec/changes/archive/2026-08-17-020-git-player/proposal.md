# Proposal: 020-git-player

## Why

Só existe Player determinístico de I/O genérico (`shell.exec`). Operações
Git passam por argv sem contrato, dificultando validação, routing e
histórico no Runtime Graph. O protocolo já prevê `git.commit` como
exemplo de capability (`docs/11`).

## What Changes

- Novo Player `git` em `internal/players/git`
- Capabilities: `git.status`, `git.diff`, `git.log`, `git.add`, `git.commit`
- Sandbox mínima (workdir no workspace, argv-only, sem rede, hooks off
  no commit)
- Registro em `api.Open` + `examples/git-status.json`

## What Does Not Change

- Shell Player / pipeline / LLM
- Task IR schema
- HITL / Execution Policy (permanecem HYPOTHESIS / P3)
- `git.push` / `pull` / `clone` / rewrite history

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `020-git-player` |
| Doc canônico | [`docs/20-git-player-v0.md`](../../docs/20-git-player-v0.md) |
| Gaps | G-70..G-74 **CONFIRMED** (recorte de G-41) |
| Código | Ainda não — este pacote autoriza o slice 8 |

## Approach

1. Manifest + Execute espelhando padrões do Shell Player
2. Allowlist de subcommands/helpers; rejeitar escape de path
3. Testes com `git init` em tempdir (sem rede)
4. Wire no Core

## Impact

- Package novo: `internal/players/git`
- Graph ganha capabilities `git.*` no refresh
- Intent heuristics Git = opcional / fora do aceite mínimo
