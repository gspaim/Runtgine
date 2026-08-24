# 36 — NPM Player v0

Player determinístico para **rodar o script `test` do npm no
workspace**: contrato `npm.test`, sem `shell.exec` / `npm test` argv
livre.

Inventário: [10-gaps.md](10-gaps.md) (G-166+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é o Test Player Go (`30`, `test.go`). Não é pytest.
Não é `npm install` / `npx` / publish. Não é MCP ([G-44](10-gaps.md)).
Não é a API HTTP (G-45 / `34`).

**Status deste doc: CONFIRMED v0 (slice 29 feito).** G-166..G-171
autorizam o recorte. pytest, yarn/pnpm, `npm install` e Players de
infra (K8s / Terraform / PostgreSQL) permanecem outros recortes G-41.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-22-036-npm-player/`](../openspec/changes/archive/2026-08-22-036-npm-player/).

Por que **npm** e não pytest neste corte: o desktop Wails já tem
frontend `package.json` (`internal/entrypoint/desktop/frontend`). O
Player pode ser dogfoodado no próprio repo. pytest continua listado
em G-134 / `09` como recorte futuro.

---

## 1. Problema

`npm test` hoje cai no prefixo shell do Intent (`npm ` → `shell.exec`)
ou em argv livre. Perde `input_schema`, workdir no workspace, e a
regra “teste vermelho falha o Run”.

Depois de `test.go`, o próximo runner da exclusão G-134 que o
repositório consegue exercer é npm — não um Player de infra.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Player `npm` determinístico | Agent / LLM / Coverage server |
| Capability `npm.test` | `npm.install` / `npm.publish` / `npx` |
| Invoca binário `npm` com argv | Shell string / `sh -c` |
| Script `test` do `package.json` no workdir | `npm run <script>` arbitrário |
| Workdir dentro do workspace | Registry global (`-g`); `--prefix` que escape |
| Falha de teste → step fail | Soft-fail silencioso |

Regras:

1. Validator / Registry continuam soberanos.
2. Argv v0: `npm test` (equivalente a `npm run test`). Sem `install`,
   `ci`, `exec`, `npx`, `publish`, `link`, `-g`, `--prefix` no input.
3. `package.json` **obrigatório** no `workdir` (após resolve, dentro do
   workspace). Sem `package.json` → `validation.invalid_input`.
4. Sem rede no contrato: o Player **não** corre `npm install`. Se
   `node_modules` faltar, o script falha — isso é `runtime.player_error`
   (como `-mod=readonly` no `test.go`).
5. Policy default: **allow**. Sem HITL no v0.
6. Blast / Claims: `npm.test` **não** gera touch nem predicted claim
   (como `test.go` / `http.get`).
7. Pacote Go: `internal/players/npm`. Nome do Player no Manifest: `npm`.
8. O script `test` no `package.json` é código do **projeto**, não do
   Runtgine — pode ser tão perigoso quanto qualquer script. O v0
   reduz a superfície (sem install/npx/prefix) mas **não** parseia o
   script para deny-list. Documentar.

---

## 3. Cortes confirmados (G-166+)

### G-166 — Papel e pacote

**Status: CONFIRMED**

- Player name: `npm`
- Pacote: `internal/players/npm`
- Kind: `deterministic`
- Registro em `api.Open` com os demais Players
- Recorte de G-41: só runner `npm test` no workspace
- Distinto do Player `test` (`test.go` / `internal/players/gotest`)

### G-167 — Capabilities v0

**Status: CONFIRMED**

Uma capability:

| Capability | Entrada | Saída (sucesso) |
|---|---|---|
| `npm.test` | `workdir?`, `timeout_ms?`, `silent?` | `ok`, `exit_code`, `elapsed_ms`, `log`, `script?` |

| Campo | Default | Máximo / regra |
|---|---|---|
| `workdir` | `.` | relativo; após resolve, dentro do workspace; deve conter `package.json` |
| `timeout_ms` | 120000 | 600000 |
| `silent` | false | se true, argv inclui `--silent` **antes** de `test` |

`script` na saída: valor de `scripts.test` no `package.json` se for
string; omitido se ausente/não-string (npm ainda pode falhar no
execute).

`log`: stdout+stderr concatenados, truncados (teto 64 KiB, como
`test.go`). Sem parse TAP/JUnit no v0 — `ok` = `exit_code == 0`.

Schemas JSON no Manifest; `additionalProperties: false`.

### G-168 — Sandbox / argv

**Status: CONFIRMED**

| Regra | Corte v0 |
|---|---|
| Invocação | só argv → `npm` + flags allowlist + `test`; nunca shell string |
| Binário | `npm` no `PATH` (mesmo padrão permissivo+warn do Shell) |
| Flags allowlist | `--silent` (se `silent=true`). Nenhuma outra flag no input |
| Subcommands | só `test`. Sem `install`, `ci`, `exec`, `run` com nome ≠ implícito test, `npx`, `publish`, `pack`, `link`, `audit`, `outdated` |
| `--prefix` / `-g` / `--userconfig` | rejeitados (nem no input; argv montado pelo Player) |
| Workdir | dentro do workspace; `package.json` presente |
| Env | herança mínima igual Shell (sem tokens / `RUNTGINE_*` no env injetado pelo Player) |
| Rede | não expor capability de install; o script do projeto *pode* acessar rede — fora do contrato |

Falha de sandbox → `validation.invalid_input` (estático).

### G-169 — Falha vs sucesso

**Status: CONFIRMED**

- `exit_code == 0` → step succeeded; `ok=true`
- `exit_code != 0` (incluindo testes vermelhos) → `runtime.player_error`;
  o Run falha. Payload ainda inclui `log` / `exit_code` (espelha `test.go`)
- Timeout / binário ausente → `runtime.timeout` / `runtime.player_error`
- Testes unitários **injetam** `ExecFunc`; `go test ./...` **não** exige
  Node/npm nem rede

### G-170 — Integração Core

**Status: CONFIRMED**

1. `api.Open` registra `npm.New()`
2. `ValidateStaticInput` no admission (espelha Git/Test)
3. Graph: `RefreshFromRegistry` cria nós `npm` / `npm.test`
4. Blast / Claims: nenhuma linha nova nas tabelas G-95/G-101
5. Exemplo: `examples/npm-test.json` (workdir apontando a um fixture
   de teste do Player, **não** obrigar CI a correr o frontend Wails)
6. Intent: `npm test` / `npm run test` **ganha** do prefixo shell
   `npm ` (G-52). Method: `heuristic.npm`. Frases PT/EN de alta
   confiança: `npm test`, `roda os testes npm`, `run npm tests`.
   `yarn` / `pnpm` **não** entram neste recorte.

### G-171 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- `npm install` / `ci` / `update` / `publish` / `link` / `npx`
- `npm run <script>` arbitrário (`start`, `build`, `lint` como
  capabilities próprias)
- yarn / pnpm / bun Players
- pytest / `test.python` (outro recorte G-41)
- `-race` / fuzz / coverage files no `test.go`
- K8s / Terraform / PostgreSQL
- MCP (G-44); API HTTP do Runtgine (G-45)
- Parse TAP/JUnit; Coverage UI
- HITL neste Player; deny-list do conteúdo de `scripts.test`

---

## 4. Critérios de aceite (slice 29)

1. Manifest registra `npm.test`; `npm.install` ausente → Validator rejeita.
2. `workdir` que escape o workspace ou sem `package.json` →
   `validation.invalid_input`.
3. Fake exec exit 0 → `ok=true`; exit 1 → `runtime.player_error` e Run fail.
4. `runtgine intent "npm test"` (após heurística) → Task IR `npm.test`,
   method `heuristic.npm`, **não** `shell.exec`.
5. `go test ./internal/players/npm/...` verde **sem** Node.
6. `go test ./...` / `go vet ./...` verdes.
7. OpenSpec `036-npm-player` arquivado **após** o código (slice 29).

---

## 5. Ordem do slice de código

Slice **29** (não este PR de spec):

1. Pacote `internal/players/npm` + Manifest
2. `ValidateStaticInput` + `ExecFunc` injetável
3. Registrar no Core; exemplo `examples/npm-test.json`
4. Heurística Intent `heuristic.npm`
5. Testes fake; README Estágio: Slice 29
6. Arquivar OpenSpec `036` após o código

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-166 Papel (`npm` / `internal/players/npm`)
- [x] G-167 `npm.test` só
- [x] G-168 Sandbox argv (sem install/npx/prefix)
- [x] G-169 Teste vermelho falha o Run
- [x] G-170 Registry + Graph + heurística Intent
- [x] G-171 Exclusões (install, yarn/pnpm, pytest, G-44/G-45)
