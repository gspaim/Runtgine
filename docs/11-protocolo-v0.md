# 11 — Protocolo v0 (PROPOSTA)

Contratos minimos para o MVP Core.

**Status geral: P0 CONFIRMADO** (fechamento humano). Ver checklist no fim.

Inventario de gaps: [10-gaps.md](10-gaps.md).
Itens individuais marcados `CONFIRMED` nas secoes.

---

## 1. Encoding (G-14)

**Status: CONFIRMED** (fechamento humano)

- Contrato canonico: **JSON** + **JSON Schema** (draft 2020-12 ou draft-07).
- YAML na CLI e acucar: convertible → JSON antes da validacao (`runtgine run task.yaml`).
- Core e Event Bus so veem JSON.

---

## 2. Identificadores e versao

**Status: CONFIRMED** (fechamento humano)

| Campo | Formato |
|---|---|
| IDs (`task_id`, `run_id`, `event_id`) | UUID v7 string (time-ordered; RFC 9562) |
| `schema_version` | semver string no documento (`"0.1.0"`) |
| Nomes de capability | `domain.action` (ex.: `shell.exec`, `git.commit`) |

## 3. Capability naming (G-05)

**Status: CONFIRMED** (fechamento humano)

```text
capability = <domain> "." <action>[ "." <qualifier> ]
domain     = [a-z][a-z0-9-]*
action     = [a-z][a-z0-9-]*
```

MVP registry inicial:

| Capability | Player |
|---|---|
| `shell.exec` | Shell Player |

Regras:

- Runtime roteia por capability, nao por nome de Player.
- Capability desconhecida → Validator rejeita (erro de validacao, nao runtime).

---

## 4. Task IR v0 (G-01)

**Status: CONFIRMED** (fechamento humano)

Entrada estruturada do MVP (CLI/Board). Sem Intent Engine.

### Exemplo

```json
{
  "schema_version": "0.1.0",
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-08-16T02:00:00Z",
  "source": {
    "entry_point": "cli",
    "ref": "local"
  },
  "intent": {
    "summary": "Run unit tests",
    "notes": "optional free text; not executed directly"
  },
  "steps": [
    {
      "step_id": "s1",
      "capability": "shell.exec",
      "input": {
        "command": ["go", "test", "./..."],
        "workdir": ".",
        "timeout_ms": 120000
      },
      "depends_on": []
    }
  ],
  "metadata": {}
}
```

### Campos

| Campo | Obrigatorio | Notas |
|---|---|---|
| `schema_version` | sim | |
| `task_id` | sim | UUID v7; CLI pode gerar |
| `created_at` | sim | RFC3339 |
| `source.entry_point` | sim | `cli` \| `tui` \| `board` \| `api` \| `other` |
| `source.ref` | nao | id externo (card, file path) |
| `intent.summary` | sim | humano; nao e executavel |
| `intent.notes` | nao | |
| `steps[]` | sim | >= 1 no MVP |
| `steps[].step_id` | sim | unico na task |
| `steps[].capability` | sim | deve existir no Registry |
| `steps[].input` | sim | valida contra schema da capability |
| `steps[].depends_on` | nao | step_ids; default `[]` (MVP: so DAG linear/simples) |
| `metadata` | nao | mapa livre |

### Fora do v0

- Linguagem natural como unico input
- Workflow Template binding
- Policies por step (alem do sandbox global do Shell)
- Subtasks aninhadas profundas (Board P1 pode estender)

---

## 5. Execution Plan v0 (G-11)

**Status: CONFIRMED** (fechamento humano) — passthrough apos validacao

**Proposta:** no MVP, Plan e **quase passthrough** do Task IR apos validacao.

```json
{
  "schema_version": "0.1.0",
  "plan_id": "…",
  "task_id": "…",
  "run_id": "…",
  "steps": [
    {
      "step_id": "s1",
      "capability": "shell.exec",
      "player": "shell",
      "input": { },
      "depends_on": []
    }
  ]
}
```

O Runner resolve `capability` → `player` via Registry. Sem replanejamento dinamico no v0.

---

## 6. Player Manifest v0 (G-02)

**Status: CONFIRMED** (fechamento humano)

```json
{
  "schema_version": "0.1.0",
  "name": "shell",
  "version": "0.1.0",
  "kind": "deterministic",
  "capabilities": [
    {
      "name": "shell.exec",
      "input_schema": {
        "type": "object",
        "required": ["command"],
        "properties": {
          "command": {
            "type": "array",
            "items": { "type": "string" },
            "minItems": 1
          },
          "workdir": { "type": "string", "default": "." },
          "env": {
            "type": "object",
            "additionalProperties": { "type": "string" }
          },
          "timeout_ms": {
            "type": "integer",
            "minimum": 1,
            "default": 60000
          }
        },
        "additionalProperties": false
      },
      "output_schema": {
        "type": "object",
        "required": ["exit_code"],
        "properties": {
          "exit_code": { "type": "integer" },
          "stdout": { "type": "string" },
          "stderr": { "type": "string" }
        }
      }
    }
  ]
}
```

`kind`: `deterministic` \| `ai` \| `human` \| `service` \| `workflow`.

---

## 7. Event envelope + tipos (G-03, G-04)

**Status: CONFIRMED** (fechamento humano)

```json
{
  "schema_version": "0.1.0",
  "event_id": "…",
  "type": "run.started",
  "ts": "2026-08-16T02:00:01Z",
  "run_id": "…",
  "task_id": "…",
  "step_id": null,
  "payload": {}
}
```

### Tipos minimos (MVP)

| type | Quando |
|---|---|
| `task.accepted` | Task IR passou parse |
| `task.rejected` | Validator falhou |
| `run.planned` | Plan gerado |
| `run.started` | Runner iniciou |
| `step.started` | Step despachado ao Player |
| `step.succeeded` | Player retornou ok |
| `step.failed` | Player falhou / exit != 0 (shell) |
| `run.succeeded` | Todos steps ok |
| `run.failed` | Run abortado por falha |
| `run.cancelled` | Cancelamento (se suportado; P2 pode adiar emissao) |

Payload de `task.rejected` / `step.failed` usa o Error model (§9).

---

## 8. Run lifecycle (G-09)

**Status: CONFIRMED** (fechamento humano)

```text
accepted → planned → running → succeeded
                              ↘ failed
                              ↘ cancelled   (opcional no MVP)
rejected   (terminal; sem run)
```

Estados da Task (visao Entry Point): `accepted` | `rejected` | `running` | `succeeded` | `failed` | `cancelled`.

Um `run_id` por tentativa de execucao de uma task aceita.

---

## 9. Result / Error (G-08)

**Status: CONFIRMED** (fechamento humano)

### Result (step)

```json
{
  "ok": true,
  "step_id": "s1",
  "capability": "shell.exec",
  "player": "shell",
  "output": {
    "exit_code": 0,
    "stdout": "…",
    "stderr": ""
  },
  "duration_ms": 1234
}
```

### Error

```json
{
  "code": "validation.unknown_capability",
  "message": "capability \"foo.bar\" is not registered",
  "retryable": false,
  "details": {}
}
```

Codigos iniciais:

| code | Fase |
|---|---|
| `validation.schema` | Validator |
| `validation.unknown_capability` | Validator |
| `validation.invalid_input` | Validator |
| `runtime.player_error` | Player |
| `runtime.timeout` | Runner/Player |
| `runtime.cancelled` | Runner |
| `runtime.internal` | Core |

---

## 10. Runner v0 (G-10) — Orchestrator minimo

**Status: CONFIRMED** (fechamento humano)

**Nome no MVP:** `Runner` (evitar confundir com Orchestrator completo HYPOTHESIS).

Responsabilidades v0:

1. Receber Task IR
2. Validar (Validator)
3. Montar Plan (capability → player)
4. Enfileirar steps respeitando `depends_on` (MVP: executar em ordem topologica; falha em um step falha o run)
5. Publicar eventos
6. Coletar Results

**Nao faz no v0:** replanejamento, policies avancadas, background players, blast radius.

Relacao: Orchestrator HYPOTHESIS futuro pode absorver Runner.

---

## 11. Queue v0 (G-12)

**Status: CONFIRMED** (fechamento humano) — variante B

- In-process, FIFO por steps prontos / runs.
- Sem prioridade no MVP.
- Sem persistencia da fila (ver §12).
- **Multiplos runs concorrentes** permitidos no MVP (limites de paralelismo
  configuraveis depois; default razoavel no processo).
- Steps de um mesmo run respeitam `depends_on` (ordem topologica);
  runs distintos nao se bloqueiam entre si no v0.

---

## 12. Persistencia (G-13)

**Status: CONFIRMED** (fechamento humano) — variante B

SQLite cedo no MVP Core (nao esperar pos-Shell):

| Dado | MVP Core | Notas |
|---|---|---|
| Runs / estado | SQLite | Tabela de runs + status |
| Eventos | SQLite append-only | Nao e event sourcing completo; permite `status` apos restart |
| Fila em voo | Memoria | Reconstroi o que for necessario a partir do store se o processo cair |

Sem pretender Event Sourcing / replay completo no v0.
Driver: ver §15 (modernc proposto).

---

## 13. Core API — Entry Point → Core (G-07)

**Status: CONFIRMED** (fechamento humano)

**Mesmo protocolo interno; Entry Points sao adapters.**

```text
SubmitTask(TaskIR) -> (run_id | ValidationError)
GetRun(run_id) -> RunSnapshot
Subscribe(filter) -> <-chan Event   // TUI/CLI status
CancelRun(run_id) -> error          // pode ser stub no MVP
```

- Nao ha protocolo separado Board/CLI.
- Board traduz card → Task IR e chama `SubmitTask`.
- Entry Point != Player.

---

## 14. Shell Player + policy minima (G-06, G-18)

**Status: CONFIRMED** (fechamento humano)

Capability: `shell.exec` (schema no Manifest §6).

**Sandbox v0 (obrigatorio mesmo sem Execution Policy completa)**

| Regra | Default MVP |
|---|---|
| Shell | sem shell string; so `command` argv (sem `sh -c` implicito) |
| Workdir | deve estar dentro do workspace root configurado |
| Env | allowlist ou herda minimo; sem injecao livre de secrets do host alem do necessario |
| Timeout | obrigatorio (default 60s) |
| Rede | nao controlada no v0 (documentar risco); deny via OS fica P2 |
| Binarios | allowlist opcional (`go`, `git`, …) — default permissivo + warning no log |

Falha de sandbox → `validation.invalid_input` ou `runtime.player_error` com code dedicado futuro `policy.denied`.

---

## 15. Stack openers (G-15, G-16, G-37)

| Item | Status | Decisao |
|---|---|---|
| Logger | CONFIRMED | `log/slog` |
| SQLite driver | CONFIRMED | `modernc.org/sqlite` (pure Go, sem cgo) |
| Go version | CONFIRMED | 1.22+ |
| Module path | CONFIRMED | `github.com/gspaim/Runtgine` |

---

## 16. Repo layout v0 (G-17)

**Status: CONFIRMED** (fechamento humano)

```text
cmd/runtgine/          # CLI entry
internal/core/
  task/                # Task IR types + validate
  plan/                # Plan v0
  event/               # bus + envelope
  runner/              # Runner v0
  registry/            # Player registry
  result/              # Result/Error
internal/players/
  shell/
internal/entrypoint/
  cli/
  tui/                 # depois do CLI
pkg/protocol/          # tipos/schemas publicos estaveis (opcional cedo)
schemas/               # JSON Schema files
docs/
```

Core nao importa `entrypoint`. Players nao importam UI.

---

## 17. Validator v0

Checagens MVP:

1. JSON Schema do Task IR
2. `steps` nao vazio; `step_id` unicos; `depends_on` aciclico e referencias validas
3. Cada `capability` existe no Registry
4. `input` valida contra `input_schema` da capability
5. Regras de sandbox estaticas do Shell (workdir, argv)

---

## Checklist de confirmacao

Marcar em `04-decisoes.md` apos revisao humana:

- [x] Encoding JSON canonico + YAML so na borda CLI
- [x] IDs UUID v7 + capability `domain.action`
- [x] Task IR v0
- [x] Manifest v0
- [x] Event envelope + tipos minimos
- [x] Result/Error + lifecycle
- [x] Runner v0 (nome e escopo) + Plan passthrough
- [x] Queue in-memory FIFO (**multi-run** concorrente)
- [x] Persistencia: SQLite cedo (runs + events append-only)
- [x] Core API SubmitTask/GetRun/Subscribe
- [x] Shell sandbox v0
- [x] slog CONFIRMED
- [x] modernc + Go 1.25+ + module path (Go atualizado pelo Charm v2 no Slice 3)
- [x] Layout de pacotes

**P0 fechado.** Proximo: implementar Core na ordem de `09-mvp` / `AGENTS.md`.
Gaps P1 (Board/LLM) ainda abertos em `10-gaps.md`.
