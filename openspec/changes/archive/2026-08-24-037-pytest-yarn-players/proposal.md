# Proposal: 037-pytest-yarn-players

## Why

G-41 (biblioteca ampla de Players determinísticos) continua em
andamento (`04-decisoes` §Arquitetura). Depois de `test.go` (slice 18,
spec `30`) e `npm.test` (slice 29, spec `36`), restam dois recortes
**dogfoodáveis e de baixo risco**:

- `pytest` — runner de testes Python no workspace (equivalente direto
  do `test.go`).
- `yarn test` — equivalente Yarn do `npm.test`; mesmo recorte de
  pacote, sem `yarn install` / `npx` / workspaces.

Ambos cabem num único package (`internal/players/jstest` e
`internal/players/pytst`) com a mesma forma do `npm` e do `gotest`:
binário direto, argv allowlist, sem rede, sem install. Promover
fecha 6 gaps P3 de uma vez (G-41 restante dogfood).

Não é pytest parametrizado, não é tox, não é coverage, não é
`yarn install`, não é `npx`, não é MCP (G-44), não é K8s/Terraform.

## What Changes

- Canonical `docs/37-pytest-yarn-players-v0.md` (G-172..G-179
  CONFIRMED)
- Player `pytest` em `internal/players/pytst` (corte v0)
  - Capability `pytest.run` (workspace; `pytest …`)
- Player `yarn` em `internal/players/jstest` (corte v0)
  - Capability `yarn.test` (`yarn test`)
- Sandbox argv (allowlist; sem install/npx/prefix/--frozen-lockfile)
- `package.json` para yarn; `pyproject.toml` *ou* `pytest.ini`
  *ou* `tests/` para pytest
- Falha de teste falha o Run (`runtime.player_error`)
- Intent heuristics `pytest …` → `pytest.run`; `yarn test` →
  `yarn.test` (beats shell prefix)
- Examples `examples/pytest-run.json`, `examples/yarn-test.json`
- Read-only `go test ./...` (executores injetados)

## What Does Not Change

- Shell / Git / FS / Docker / HTTP / `test.go` / `npm.test`
- Task IR schema (`0.1.0`)
- Claims / Blast tables (test runners = no touch / no claim)
- G-45 HTTP server; G-44 MCP; Memory Player
- pytest parametrizado / tox / coverage / pytest-xdist
- `yarn install` / `yarn add` / `yarn workspaces` / `npx`
- TUI tabs; Wails views; Board adapter
- NPM Player v0 (spec `36`)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `037-pytest-yarn-players` |
| Doc canônico | `docs/37-pytest-yarn-players-v0.md` (a criar) |
| Gaps | G-172..G-179 **CONFIRMED (proposta)** (recorte de G-41) |
| Código | Slice 30 — **bloqueado** até esta spec + `04` |
| Depende | `030-test-player`, `036-npm-player` (forma de Manifest) |

## Approach

1. `internal/players/pytst` espelha `internal/players/gotest`
   (exec injetado, `-q` allowlist, timeout, log truncado).
2. `internal/players/jstest` espelha `internal/players/npm`
   (argv fixo, `package.json` obrigatório, `node_modules/`
   não é claim).
3. Validação estática: `workdir` resolve + fica no workspace;
   marker file presente.
4. Fail = `runtime.player_error` com `exit_code` + `log`.
5. Intent: heurística `pytest …` / `yarn test` antes de
   `matchShell`; métodos `heuristic.pytest` e `heuristic.yarn`.
6. Não mexe em Claims (`24`) nem Blast (`25`).

## Impact

- New packages `internal/players/pytst`, `internal/players/jstest`
- `internal/core/api` register + runner static validation dispatch
- `internal/core/intent` 2 novas heurísticas
- README Estágio: Slice 30 após código
