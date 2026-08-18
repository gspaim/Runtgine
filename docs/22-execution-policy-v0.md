# 22 — Execution Policy e HITL v0

Policy de execução por capability e pausa humana (`approval-required`)
antes do Player executar o step.

Inventário: [10-gaps.md](10-gaps.md) (G-81+; recorte de G-42).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Core + Validator + Runner estáveis (slices 1–9).
Consumidor seguinte: Docker Player ([23-docker-player-v0.md](23-docker-player-v0.md)).

**Status deste doc: CONFIRMED (v0).** G-81..G-86 autorizam o slice 10
de código. Resource Claims, Blast Radius e Human Player permanecem fora.

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-17-022-execution-policy/`](../openspec/changes/archive/2026-08-17-022-execution-policy/).
Deltas mergeados em `openspec/specs/execution-policy/`. Branch de implementação:
`feat/022-execution-policy`.

---

## 1. Problema

Sandbox por Player (Shell/Git/FS) limita *como* uma capability corre.
Não existe ainda *se* ela pode correr: allow / deny / pedir humano.

Sem isso, o próximo Player (Docker) ou `shell.exec` em workspace
sensível não tem porta de HITL. Approvals não são um Player: são o
Core recusando despachar até um Entry Point decidir.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Motor no Core (Runner + API) | Player `human` / `approval.*` |
| Verbo por **capability** exata | Policy por path, argv, ambiente ou blast radius |
| `allow` / `deny` / `approval-required` | Roles, multi-approver, assinatura |
| Pausa persistida no Run (SQLite) | Fila distribuída / webhook de aprovação |
| CLI + TUI Runs como Entry Point | Nova aba TUI; Board write-back de approve |

Regras:

1. Validator e Registry continuam soberanos. Policy **não** inventa
   capability nem ignora schema.
2. Deny é falha de admissão (compilador): a Task nem começa.
3. `approval-required` admite a Task, executa steps `allow` anteriores
   e **pausa antes** do `Execute` do step gated.
4. Humano aprova/rejeita via Core API, nunca falando com o Player.
5. Default global é `allow` — slices 1–9 não mudam de comportamento
   sem config ou `execution_policy` no Manifest.

---

## 3. Cortes confirmados (G-81+)

### G-81 — Papel

**Status: CONFIRMED**

- Execution Policy vive no Core (`internal/core/policy` ou equivalente).
- HITL é o mesmo motor + lifecycle; não é Player.
- Avaliação: depois do Validator/Registry/Plan, no Runner.
- Pacote não importa `entrypoint`.

### G-82 — Tabela de policy

**Status: CONFIRMED**

Verbos v0 (só estes):

| Verbo | Efeito |
|---|---|
| `allow` | Despacha o Player |
| `deny` | Rejeita a Task na admissão (`task.rejected`) |
| `approval-required` | Pausa o Run antes do Execute daquele step |

Precedência (alinha G-38):

```text
default global allow
  < Manifest.capability.execution_policy (se presente)
  < config.json execution_policy.capabilities
  < RUNTGINE_POLICY_DEFAULT (só o default global; sem mapa por cap via env)
```

`config.json` (opcional):

```json
{
  "execution_policy": {
    "default": "allow",
    "capabilities": {
      "shell.exec": "approval-required"
    }
  }
}
```

- Chave de capability **exata** (`domain.action`). Sem wildcards (`shell.*`) no v0.
- Verbo inválido ou capability desconhecida no mapa → erro ao carregar config
  (fail closed no boot), não ignore silencioso.
- Manifest: campo opcional `execution_policy` em cada capability
  (`allow` | `deny` | `approval-required`). Omitido = herda o default.
- `deny` em **qualquer** step do Plan → `task.rejected` com
  `policy.denied` **antes** de `InsertRun` / execução. Nenhum step corre.
- `approval-required` **não** rejeita na admissão.

### G-83 — Lifecycle HITL

**Status: CONFIRMED**

Estende [11-protocolo-v0.md](11-protocolo-v0.md) §8:

```text
accepted → planned → running → succeeded
                              ↘ failed
                              ↘ cancelled
                              ↘ waiting_approval → running   (approval granted)
                                                 ↘ failed     (approval denied)
                                                 ↘ cancelled
rejected   (terminal; sem run)  // inclui policy.denied
```

- Estado persistido: `waiting_approval` (SQLite, como os demais).
- Restart do processo: Run continua `waiting_approval`; **não** reexecuta
  o step até `ApproveRun`. Steps já `succeeded` não repetem.
- Um Run espera **no máximo um** step gated de cada vez (pipeline linear
  já existente). Próximo `approval-required` só depois de retomar.
- Payload mínimo do pedido: `run_id`, `step_id`, `capability`, `player`.
- `runtgine run --wait` permanece bloqueado durante `waiting_approval`
  até estado terminal (outro processo/TUI pode aprovar).

Eventos novos:

| type | Quando |
|---|---|
| `run.waiting_approval` | Runner pausou antes do Execute |
| `run.approval_granted` | `ApproveRun` grant |
| `run.approval_denied` | `ApproveRun` deny |

Códigos de erro novos:

| code | Fase |
|---|---|
| `policy.denied` | Admissão (verbo deny) |
| `policy.approval_denied` | Humano recusou; Run `failed` |

Cancelamento de um Run `waiting_approval` usa `CancelRun` já existente
→ `cancelled` / `runtime.cancelled`. Não exige segundo verbo de policy.

### G-84 — Core API e CLI

**Status: CONFIRMED**

```text
ApproveRun(run_id, decision) -> error
  decision: grant | deny
GetRun inclui: status, pending_approval? {step_id, capability, player}
```

- `ApproveRun` só é válido se status atual é `waiting_approval`.
- Grant: emite `run.approval_granted`, volta `running`, `Execute` o step.
- Deny: emite `run.approval_denied`, Run `failed` com `policy.approval_denied`;
  o Player **não** roda.
- Grant/deny em Run que não espera → erro (`runtime.internal` ou código
  dedicado `policy.not_waiting`; preferir `policy.not_waiting`).

CLI:

```text
runtgine approve <run_id>
runtgine deny <run_id>
```

`runtgine status <run_id>` já existe: deve mostrar `waiting_approval` e o
pending. Sem flag `--auto-approve` na CLI de produto (testes injetam
`ApproveRun` no Core).

### G-85 — TUI (aba RUNS / LIVE)

**Status: CONFIRMED**

Sem aba nova. Seguir [14-tui-design.md](14-tui-design.md) e a skill TUI.

- Status `waiting_approval` visível em RUNS e LIVE (texto + símbolo, não só cor).
- Cor: **Amber** (atenção humana), distinto de `running`.
- Com o Run selecionado em espera: `a` grant / `d` deny via Core API.
- Footer atualiza essas teclas só nesse estado.
- LIVE mostra o step gated como atual (amber) até grant/deny.
- Sem multiplexer, PTY ou tab GRAPH.

### G-86 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- Resource Claims (G-43)
- Blast Radius (G-43)
- Wildcards de capability
- Policy por input/path/argv (continua sandbox do Player)
- Human Player / capabilities `approval.*`
- Timeout automático de espera / auto-deny
- Multi-approver, RBAC, assinatura
- Write-back de aprovação no Board
- Webhook / API HTTP de approve
- Mudar o mapa de policy no meio de um Run
- Docker Player (spec `23`, slice 11)

---

## 4. Critérios de aceite

1. Sem `execution_policy` no config/Manifest, `runtgine run examples/hello.json`
   continua `run.succeeded` (default allow).
2. Config `shell.exec: deny` → Task com `shell.exec` é `task.rejected` /
   `policy.denied`; o binário não executa.
3. Config `shell.exec: approval-required` → Run entra `waiting_approval`;
   `runtgine approve` completa o step; `runtgine deny` falha sem Execute.
4. Restart com Run em `waiting_approval`: `GetRun` ainda mostra espera;
   approve posterior executa **uma** vez.
5. TUI: Run em espera visível; `a`/`d` chamam o Core.
6. `go test ./internal/core/policy/...` (e Runner) cobrem allow/deny/HITL.
7. `go test ./...` e `go vet ./...` verdes.
8. OpenSpec `022-execution-policy` arquivado após o merge do código.

---

## 5. Ordem do slice de código

1. G-81..G-86 CONFIRMED — este doc + OpenSpec
2. Motor de policy + persistência `waiting_approval` + eventos
3. `ApproveRun` + CLI `approve`/`deny` + `--wait` honesto
4. TUI RUNS/LIVE
5. README Estágio: Slice 10; próximo = Docker (`23`)

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-81 Papel (Core, não Player)
- [x] G-82 Tabela allow/deny/approval-required + precedência
- [x] G-83 Lifecycle + eventos + persistência
- [x] G-84 API `ApproveRun` + CLI
- [x] G-85 TUI RUNS/LIVE (sem aba nova)
- [x] G-86 Exclusões (Claims, wildcards, Human Player, Docker)
