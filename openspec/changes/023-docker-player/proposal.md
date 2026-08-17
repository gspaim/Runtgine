# Proposal: 023-docker-player

## Why

Não há capabilities `docker.*`. Container work cai em `shell.exec`, sem
schema, sem argv controlado e sem HITL. Docker é o primeiro Player cujo
Manifest declara `approval-required` de verdade.

## What Changes

- Player `docker` em `internal/players/docker`
- Capabilities: `docker.ps`, `docker.inspect`, `docker.logs`,
  `docker.run`, `docker.build`
- Sandbox argv: `--pull=never --network=none --rm` no run
- Manifest: `run` / `build` = `approval-required` (spec 22)
- Registro em `api.Open`, static validation, `examples/docker-ps.json`

## What Does Not Change

- Execution Policy engine (já é 022)
- Compose / K8s / push / pull / privileged
- Task IR schema
- TUI dedicada

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `023-docker-player` |
| Doc canônico | [`docs/23-docker-player-v0.md`](../../docs/23-docker-player-v0.md) |
| Gaps | G-87..G-92 **CONFIRMED** (recorte de G-41) |
| Código | Ainda não — slice 11; **bloqueado** até slice 10 (022) mergeado |

## Approach

1. Manifest com cinco capabilities e policy no run/build
2. `exec.Command("docker", ...)` allowlist de subcommands
3. Confinement de context/workdir como Filesystem
4. Testes sem daemon obrigatório (stub); HITL via Core de teste

## Impact

- Package novo: `internal/players/docker`
- Graph ganha `docker.*` no refresh
- Depende do estado `waiting_approval` da 022
