# 37 — Pytest + Yarn Players v0

Dois Players determinísticos de teste no workspace:
`pytest.run` (`internal/players/pytst`) e `yarn.test`
(`internal/players/jstest`). Mesmo padrão de `test.go` (`30`) e
`npm.test` (`36`). Sem install, sem rede, sem shell string.

Inventário: [10-gaps.md](10-gaps.md) (G-172+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é tox / pytest-xdist / `pytest-cov`. Não é `yarn install` /
`yarn add` / `yarn dlx` / `npx` / `pnpm`. Não é MCP (G-44). Não é
Knowledge. Não é K8s/Terraform/PostgreSQL.

**Status deste doc: CONFIRMED v0 (slice 30 a fazer).** G-172..G-179.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-24-037-pytest-yarn-players/`](../openspec/changes/archive/2026-08-24-037-pytest-yarn-players/).
Spec atual: [`openspec/specs/pytest-yarn-players/`](../openspec/specs/pytest-yarn-players/).

---

## 1. Problema

G-41 (biblioteca ampla de Players determinísticos) continua em
andamento. Faltavam, depois de `npm.test` (slice 29), os dois
recortes dogfoodáveis óbvios: `pytest` e `yarn test`. Ambos caíam
em `shell.exec` via prefixo Intent `pytest ` / `yarn `, sem schema,
sem allowlist de flags.

## 2. Fronteiras

| É | Não é |
|---|---|
| `pytest.run` / `yarn.test` | tox / nox / coverage |
| Binário direto com argv allowlist | shell string |
| Marker obrigatório | Qualquer diretório sem config |
| Falha de teste falha o Run | Soft-fail silencioso |
| Read-only (não toca paths fora do workdir) | `pytest-cov`, xdist, plugins |

Regras:

1. Validator / Registry continuam soberanos.
2. Só argv allowlist por Player; sem `-exec`-equivalente.
3. Marker por capability: pytest = `pyproject.toml` \| `pytest.ini`
   \| `tests/`; yarn = `package.json`.
4. `-mod=readonly`-equivalente: nenhum download (pytest usa o env,
   yarn assume `node_modules/` resolvido).
5. Policy default: **allow** (como `test.go` / `npm.test`).
6. Blast / Claims: nenhum dos dois gera touch nem predicted claim
   (verificação é leitura).
7. Pacotes Go: `internal/players/pytst` e `internal/players/jstest`.

## 3. Cortes confirmados (G-172+)

### G-172 — Papel / pacote `pytest`

- Player name: `pytest`
- Pacote: `internal/players/pytst`
- Kind: `deterministic`

### G-173 — Capability `pytest.run`

| Capability | Argv | Body |
|---|---|---|
| `pytest.run` | `pytest` + flags allowlist + packages | `ok`, `pass`, `fail`, `skip`, `elapsed_ms`, `exit_code`, `log` |

Allowlist de flags: `-q`, `-x`, `-k`, `-m`, `--tb=short`,
`--no-header`, `--color=no`. Packages: paths relativos iniciando
em `.`. `timeout_ms` (default 120000, max 600000) igual `test.go`.

Negadas: `-n` (xdist), `--cov*`, `-W error`, `--basetemp` fora do
workspace, `-c` externo, `--rootdir` fora.

### G-174 — Workdir + marker

`workdir` resolvido (default `.`); deve estar dentro do workspace
e conter `pyproject.toml`, `pytest.ini` ou `tests/`. Sem marker →
`validation.invalid_input`.

### G-175 — Papel / pacote `yarn`

- Player name: `yarn`
- Pacote: `internal/players/jstest`
- Kind: `deterministic`

### G-176 — Capability `yarn.test`

| Capability | Argv | Body |
|---|---|---|
| `yarn.test` | `yarn test` | `ok`, `exit_code`, `elapsed_ms`, `log`, `script` opcional |

Negadas em `ValidateStaticInput`: `--frozen-lockfile`,
`--immutable`, `--network-timeout`, `--mutex`, `--parallel`, `add`,
`install`, `global`, `dlx`, `npx`. Schemas
`additionalProperties: false`.

### G-177 — Workdir + `package.json`

Mesma invariante do `npm`: `package.json` obrigatório no
workdir; paths absolutos/URL/`..` rejeitados.

### G-178 — Falha vs sucesso

Idêntico a `test.go` / `npm.test`:

- `exit_code == 0` → `ok: true`.
- `exit_code != 0` → `runtime.player_error`; payload inclui `log`
  truncado (64 KiB) + counts.
- Sem "ok: false succeeded".

### G-179 — Registry + Graph + Intent

1. `api.Open` registra `pytst.New()` e `jstest.New()`
2. `ValidateStaticInput` no admission (workdir, marker, ranges,
   flags negadas)
3. Runner despacha static validation como Git/FS/HTTP/Docker/npm
4. Graph: `RefreshFromRegistry` cria nós `pytest` / `pytest.run`
   e `yarn` / `yarn.test`
5. Intent: `heuristic.pytest` (antes de `matchShell`) e
   `heuristic.yarn`. Frases:
   - pytest: `pytest`, `pytest -q`, `pytest tests/foo.py`,
     `roda pytest`, `run pytest`
   - yarn: `yarn test`, `yarn run test`, `rodar yarn test`
6. `yarn install` / `yarn add` / `pip install` continuam
   `shell.exec` (fora)

## 4. Critérios de aceite

1. `pytest.python` ausente → Validator rejeita.
2. Workdir `../outside` ou absoluto → `validation.invalid_input`.
3. Workdir sem marker → `validation.invalid_input`.
4. `yarn install` no input → `validation.invalid_input`.
5. Fake runner pytest exit 0 → `ok: true`.
6. Fake runner pytest exit 1 → `runtime.player_error`.
7. Idem para yarn.
8. Unit tests **não** invocam `pytest` / `yarn` reais
   (executores injetados).
9. `go test ./...` e `go vet ./...` verdes sem Python/Yarn/Node
   no PATH.
10. OpenSpec `037` arquivado após o **código** (slice 30).

## 5. Ordem do slice de código

1. Pacote `internal/players/pytst` + Manifest + `ValidateStaticInput`
2. Executor injetável; `-q`/`-x`/`-k`/etc; truncate `log`
3. Pacote `internal/players/jstest` + Manifest +
   `ValidateStaticInput` (flags negadas)
4. Registrar no `api.Open`; exemplos `pytest-run.json`,
   `yarn-test.json`
5. Intent `heuristic.pytest` e `heuristic.yarn`
6. Testes fake-exec; `go test ./...` / `go vet ./...`
7. Arquivar OpenSpec `037`

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-172 Papel (`pytest` / `pytst`)
- [x] G-173 Capability `pytest.run`
- [x] G-174 Workdir + marker
- [x] G-175 Papel (`yarn` / `jstest`)
- [x] G-176 Capability `yarn.test`
- [x] G-177 Workdir + `package.json`
- [x] G-178 Falha vs sucesso
- [x] G-179 Registry + Graph + Intent