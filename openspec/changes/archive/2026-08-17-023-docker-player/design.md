# Design: 023-docker-player

## Technical approach

### Package

`internal/players/docker` com `New()`, `Manifest()`,
`ValidateStaticInput`, `Execute`. Invoca o binário `docker` via argv
(como Git). Sem SDK.

### Capability contracts

| Capability | Behavior |
|---|---|
| `docker.ps` | `docker ps --format json` (ou parse estável); `all` → `-a` |
| `docker.inspect` | `docker inspect <id>` JSON truncado |
| `docker.logs` | `docker logs --tail N <id>` |
| `docker.run` | flags fixas + image + argv do input após grant |
| `docker.build` | `docker build --pull=false` no context do workspace |

Parsing de `docker ps`: preferir `--format json` (Docker 23+) com
fallback documentado se necessário para testes stub.

### Policy

Manifest sets `execution_policy` on `docker.run` and `docker.build`.
Unit/integration tests MUST assert `Execute` is not reached before
`ApproveRun(grant)`.

### Tests without daemon

Inject a command runner interface (`func(ctx, argv) (stdout, stderr, err)`).
Default production runner uses `exec.CommandContext`. Suite `Short` or
no-daemon CI never talks to a real engine.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Docker SDK | Dependência extra; Git já é binário |
| Incluir pull/push | Rede + credenciais; fora do v0 |
| Run sem HITL | Contraria a razão de existir a 022 |
| `--network=bridge` default | Superfície de rede implícita |

## Risks

| Risco | Mitigação |
|---|---|
| Daemon residual network on build | `--pull=false`; documentar residual |
| CI sem Docker | stub + skip |
| Image tag injection | rejeitar `id`/`image` com espaços e prefixo `-` |

## Packages touched

- `internal/players/docker` (novo)
- `internal/core/api` (register)
- `internal/core/runner` (static validation dispatch)
- `examples/docker-ps.json`
