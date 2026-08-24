# Design: 037-pytest-yarn-players

## Pacotes

| Player name | Pacote Go | Capability v0 | Binário |
|---|---|---|---|
| `pytest` | `internal/players/pytst` | `pytest.run` | `pytest` |
| `yarn` | `internal/players/jstest` | `yarn.test` | `yarn` |

Escolha dos pacotes segue o padrão existente (`gotest` para evitar
colisão com `testing`; `npm` na stdlib vazia; `pytst` para evitar
sombra de `pytest` no PyPI/import mental).

## Sandbox comum (recorte de G-41)

| Regra | pytest.run | yarn.test |
|---|---|---|
| Invocação | argv → `pytest …` | argv → `yarn test` |
| Workdir | resolvido com `EvalSymlinks`; fica no workspace | idem |
| Marker | `pyproject.toml` \| `pytest.ini` \| `tests/` | `package.json` |
| Timeout | `timeout_ms` (default 120000, max 600000) | idem |
| Env | herança mínima (sem `RUNTGINE_*` / tokens) | idem |
| Allowlist | `-q`, `-x`, `-k`, `-m`, `--tb=short`, `--no-header`, `--color=no`, packages | `yarn`, `test` (sem `--frozen-lockfile`, `--immutable`, `--no-progress`, sem flags) |
| Flags negadas | `-p no:cacheprovider` opcional OK; **não** `--basetemp` fora do workspace, **não** `-c path externo`, **não** `--rootdir` fora, **não** `-n` (xdist), **não** `--cov*`, **não** `-W error` | `--frozen-lockfile`, `--immutable`, `--network-timeout`, `--mutex`, `--parallel`, `add`, `install`, `global`, `dlx`, `npx` |
| Executor injetável | sim (fake `pytest`) | sim (fake `yarn`) |

`workdir` default = `.`. Rejeita path absoluto, URL, `..` que escape.

## Falha vs sucesso

- `exit_code == 0` → `ok: true` + counts (`pass`, `fail: 0`,
  `skip`, `elapsed_ms`).
- `exit_code != 0` → `runtime.player_error`; payload inclui
  `ok: false`, `exit_code`, `log` truncado (default 64 KiB).
- Sem "ok: false succeeded" — teste vermelho falha o Run.

## Blast / Claims

Nem `pytest.run` nem `yarn.test` entram em `claim.Required` nem em
`blast.Touched`. Mesma decisão de `test.go` / `npm.test`:
verificação é leitura, não mutação de path.

## Intent (recorte de G-51/G-52)

Em `internal/core/intent`, antes de `matchShell`:

| Frase | Capability | Método |
|---|---|---|
| `pytest`, `pytest -q`, `pytest tests/foo.py`, `roda pytest`, `run pytest` | `pytest.run` | `heuristic.pytest` |
| `yarn test`, `yarn run test`, `yarn tests`, `rodar yarn test` | `yarn.test` | `heuristic.yarn` |

Yarn workspaces, `yarn add`, `npx` continuam `shell.exec` (fora).

## Integração Core

1. `api.Open` registra `pytst.New()` e `jstest.New()`
2. `ValidateStaticInput` no admission (workdir, marker, ranges)
3. Runner despacha static validation como Git/FS/HTTP/Docker/npm
4. Graph: `RefreshFromRegistry` cria nós `pytest` / `pytest.run`
   e `yarn` / `yarn.test`
5. Exemplos `examples/pytest-run.json`, `examples/yarn-test.json`

## Testes (slice 30)

- Fake runner pytest: pass → `ok=true`, counts
- Fake runner pytest: fail → `runtime.player_error`, log truncado
- Workdir com `..` → validation `validation.invalid_input`
- Workdir sem marker → validation
- Yarn: fake exit 0/1; sem `package.json` rejeitado
- Intent `pytest tests/x.py` não vira `shell.exec`
- Intent `yarn install` continua `shell.exec`
- `go test ./...` / `go vet ./...` sem Python/Yarn no PATH

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| `test.python` + `test.js` (estender `test`) | Manifest do `gotest` é Go-specific; o padrão G-41 é um pacote por cut |
| Cobrir `tox` / `nox` / `pytest-xdist` | Unbounded flag surface; v0 fica no `pytest` raiz |
| `yarn dlx` / `yarn workspaces` | Network + mutação; recusa `yarn install` |
| `pnpm` / `bun test` | Mesmo padrão de Yarn; v0 não dogfooda; fica para G-41 futuro |
| K8s / Terraform / PostgreSQL | PRD P3 infra; não é o próximo cut de test-runner |
| Embute Python (uv / pip) no Player | Mesmo argumento de `npm install`: v0 é read-only |

## Risks

| Risco | Mitigação |
|---|---|
| `pytest` plugins injetam hooks | Allowlist de argv + sem `--rootdir` externo |
| `yarn` exige node_modules resolvido | `package.json` é o marker; install continua fora |
| CI sem Python ou Yarn no PATH | Executor injetado; smoke `go test ./...` cobre |
| Intent confunde `pip install` com `pytest` | `pip install` continua `shell.exec` (fora) |
| `pyproject.toml` é ambíguo (PEP 621 vs poetry) | Presença do marker é suficiente; não parseia TOML |

## Critérios de aceite

1. `pytest.python` ausente → Validator rejeita
   (somente `pytest.run` registrado).
2. Workdir `../outside` ou absoluto → `validation.invalid_input`.
3. Workdir sem marker → `validation.invalid_input`.
4. Fake runner pytest com `exit 0` → `ok: true`.
5. Fake runner pytest com `exit 1` → `runtime.player_error`.
6. Idem para `yarn test`.
7. Unit tests **não** invocam `pytest` nem `yarn` reais.
8. `go test ./...` e `go vet ./...` verdes.
9. OpenSpec `037` arquivado após o código (slice 30).
