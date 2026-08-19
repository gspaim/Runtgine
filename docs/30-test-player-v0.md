# 30 — Test Player v0

Player determinístico para **rodar testes Go no workspace**: contrato
`test.go`, sem `shell.exec` / `go test` argv livre.

Inventário: [10-gaps.md](10-gaps.md) (G-129+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é o Test Player genérico (pytest/npm). Não é Coverage UI.
Não é MCP ([G-44](10-gaps.md)). Não é a API HTTP (G-45).

**Status deste doc: CONFIRMED (v0).** G-129..G-134 implementados no
slice 18. Outros runners e `-race` permanecem fora.

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-19-030-test-player/`](../openspec/changes/archive/2026-08-19-030-test-player/).
Deltas mergeados em `openspec/specs/test-player/`. Branch de implementação:
`feat/030-test-player`.

---

## 1. Problema

A frase do produto é “intenção → execução **verificável**”. Depois de
Shell, Git, FS, Docker e HTTP, a verificação ainda passa por
`shell.exec` + `go test …` — sem schema, sem allowlist de flags, sem
contagem estruturada de pass/fail.

`02-conceitos` já lista Test Player. O PRD tem persona QA. Falta o
Manifest e o sandbox.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Player `test` determinístico | Agent / LLM / Coverage server |
| Capability `test.go` | `test.python` / `test.npm` / `test.fuzz` |
| Invoca binário `go` com argv | Shell string / `sh -c` |
| Pacotes relativos ao workspace | `go test` remoto; download de módulos |
| Falha de teste → step fail | Soft-fail silencioso |

Regras:

1. Validator / Registry continuam soberanos.
2. Só `go test` + flags allowlist. Sem `-exec`, `-toolexec`, `-overlay`,
   `-gcflags`, `-ldflags`, `-race`, `-fuzz`, `-coverprofile`.
3. `-mod=readonly`: não baixa módulos na execução.
4. Pacotes: relativos (ex. `./...`, `./internal/core/memory`); após
   resolve, devem permanecer no workspace. Sem paths absolutos, URLs
   ou `..` que escape.
5. Policy default: **allow** (como `http.get` / `git.status`). Sem HITL
   no v0.
6. Blast / Claims: `test.go` **não** gera touch nem predicted claim
   (como `shell.exec` / `http.get`).
7. Pacote Go: `internal/players/gotest` (não colidir com `testing`).
   Nome do Player no Manifest: `test`.

---

## 3. Cortes confirmados (G-129+)

### G-129 — Papel e pacote

**Status: CONFIRMED**

- Player name: `test`
- Pacote: `internal/players/gotest`
- Kind: `deterministic`
- Registro em `api.Open` com os demais Players
- Recorte de G-41: só runner `go test` no workspace

### G-130 — Capabilities v0

**Status: CONFIRMED**

Uma capability:

| Capability | Entrada | Saída (sucesso) |
|---|---|---|
| `test.go` | `packages?` (array string, default `["./..."]`), `timeout_ms?` (default 120000, max 600000), `short?` (bool), `count?` (1–10, default 1), `run?` (regex `-run`, opcional) | `ok` (true), `pass`, `fail`, `skip`, `elapsed_ms`, `exit_code` (0), `log` (texto truncado) |

Schemas JSON no Manifest; `additionalProperties: false`.

`run`, se presente, é um único argumento `-run` (sem metacaracteres de
shell). Vazio = omitir a flag.

Internamente o Player usa `go test -json` para contar pass/fail/skip.
O campo `log` é um resumo UTF-8 truncado (não o stream JSON bruto).
Default `max_log_bytes` interno: 64 KiB (não é input do Task IR no v0).

### G-131 — Sandbox / argv

**Status: CONFIRMED**

| Regra | Corte v0 |
|---|---|
| Invocação | argv → `go test …`; nunca shell string |
| Binário | `go` no `PATH` (mesmo padrão permissivo+warn do Shell) |
| Workdir | workspace root |
| Timeout | `timeout_ms` no Core **e** `-timeout` no `go test` (duração equivalente) |
| Env | herança mínima igual Shell (sem tokens / `RUNTGINE_*`); `GOFLAGS` do input **proibido** |
| Módulos | sempre `-mod=readonly` |
| JSON | sempre `-json` (parse interno) |
| Flags allowlist | `-json`, `-mod=readonly`, `-count`, `-timeout`, `-short`, `-run` + pacotes |
| Flags negadas | `-exec`, `-toolexec`, `-overlay`, `-gcflags`, `-ldflags`, `-race`, `-fuzz`, `-coverprofile`, `-o`, `-c`, `-args` |

Executor injetável nos testes (como RoundTripper do HTTP Player):
`go test ./internal/players/gotest/...` **não** precisa invocar o
suite do repositório.

### G-132 — Falha vs sucesso

**Status: CONFIRMED**

- `exit_code == 0` → step `succeeded`; output JSON com `ok: true`.
- `exit_code != 0` (testes falharam, compile error, timeout do `go`)
  → `runtime.player_error`; details incluem `ok: false`, `pass` /
  `fail` / `skip`, `exit_code`, `log` truncado.
- Timeout do Runner → `runtime.timeout` (já existe); matar o processo
  `go`.

Não há modo “sempre succeeded com ok=false”. Teste vermelho falha o
Run — é a verificação.

### G-133 — Integração Core

**Status: CONFIRMED**

1. `api.Open` registra `gotest.New()`
2. `ValidateStaticInput` no admission (pacotes, `run`, ranges)
3. Runner despacha static validation como Git/FS/HTTP/Docker
4. Graph: `RefreshFromRegistry` cria nós `test` / `test.go`
5. Exemplo: `examples/test-go.json`
6. Intent heuristics `go test` → `test.go`: nice-to-have; **não**
   bloqueia o slice

### G-134 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- `test.python`, `test.npm`, `test.rust`, runners genéricos
- `-race`, fuzz, coverage files, `go test -c`
- `go run` / `go generate` / `go vet` como capabilities
- Download de módulos (`-mod=mod`); proxy GOPROXY custom no Task IR
- Claims / blast touches
- HITL neste Player; TUI dedicada
- MCP (G-44); API HTTP (G-45); Memory Player
- K8s / Terraform / PostgreSQL Players (continuam exemplos em `02`)

---

## 4. Critérios de aceite

1. Manifest registra `test.go`; `test.python` ausente → Validator rejeita.
2. Pacote `../outside` ou path absoluto → `validation.invalid_input`.
3. Flag equivalente a `-exec` no input → rejeitado (campo inexistente
   pelo schema `additionalProperties: false`).
4. Fake runner: `go test -json` com um `pass` → `ok=true`, `pass>=1`.
5. Fake runner: exit 1 + um `fail` → `runtime.player_error`, Run failed,
   details com `fail>=1`.
6. Unit tests **não** disparam `go test ./...` do Runtgine; executor
   injetado.
7. `go test ./internal/players/gotest/...` e `go test ./...` /
   `go vet ./...` verdes.
8. OpenSpec `030-test-player` arquivado após o **código** (slice 18).

---

## 5. Ordem do slice de código

Bloqueado até G-129..G-134 CONFIRMED — feito neste slice:

1. Pacote `internal/players/gotest` + Manifest + `ValidateStaticInput`
2. Executor injetável; parse `-json`; truncamento de `log`
3. Registrar no Core; exemplo `examples/test-go.json`
4. Testes fake pass/fail; README Estágio: Slice 18
5. Arquivar OpenSpec `030` após o código

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-129 Papel (`test` / `gotest`)
- [x] G-130 Capability `test.go`
- [x] G-131 Sandbox argv / `-mod=readonly`
- [x] G-132 Falha de teste falha o Run
- [x] G-133 Registry + Graph; sem claim/blast
- [x] G-134 Exclusões (outros runners, race, G-45, MCP)
