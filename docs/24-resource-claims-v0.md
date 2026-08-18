# 24 — Resource Claims v0

Bloqueio concorrente de recursos no Core, para que dois Runs não
mutem o mesmo workspace/path ao mesmo tempo.

Inventário: [10-gaps.md](10-gaps.md) (G-93+; recorte de G-43).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Execution Policy + HITL v0 ([22-execution-policy-v0.md](22-execution-policy-v0.md),
slice 10) e Players mutadores Git/FS/Docker (slices 8–11).
Consumidor seguinte: Blast Radius (candidato `025`); não depende deste
slice para o Graph existir.

**Status deste doc: CONFIRMED (v0).** G-93..G-98 implementados no
slice 12. Blast Radius, wait/queue de claim e Human Player permanecem
fora.

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-18-024-resource-claims/`](../openspec/changes/archive/2026-08-18-024-resource-claims/).
Deltas mergeados em `openspec/specs/resource-claims/`. Branch de implementação:
`feat/024-resource-claims`.

---

## 1. Problema

G-30 já permite vários Runs no mesmo processo (`max_concurrent_runs`).
Git, Filesystem e Docker mutam o workspace. Execution Policy decide
*se* a capability pode correr; não decide *quem segura o recurso*.

Sem claim, dois `fs.write` no mesmo path (ou `git.commit` + `docker.build`)
correm em paralelo e corrompem o resultado. Claims não são Player: o
Core recusa o Execute quando o recurso já está tomado.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Motor no Core (`internal/core/claim`) | Player `claim.*` / Human Player |
| Lock exclusivo `workspace` \| `path` | db, env, deployment, rede |
| Tabela automática por capability | Campo Manifest `claims[]` |
| Fail-fast (`claim.conflict`) | Fila, wait, estado `waiting_claim` |
| Persistência SQLite no workspace | Lock distribuído / NATS |
| CLI erro claro; TUI RUNS/LIVE status | Aba nova; TUI GRAPH; `runtgine claims` |

Regras:

1. Validator e Registry continuam soberanos. Claim **não** inventa
   capability nem ignora schema.
2. Ordem: Validator → Plan → Policy → **Claim** → Execute.
3. `deny` nunca chega a claim. `approval-required` pausa **antes** do
   claim; `ApproveRun(grant)` tenta o claim; conflito pós-HITL falha o
   Run (não reabre approve).
4. Default: capabilities fora da tabela automática **não** claimam —
   `runtgine run examples/hello.json` (`shell.exec`) permanece concorrente.
5. Hold até o Run terminal (`succeeded` / `failed` / `cancelled`), não
   até o fim do step.

---

## 3. Cortes confirmados (G-93+)

### G-93 — Papel e pacote

**Status: CONFIRMED**

- Resource Claim vive no Core (`internal/core/claim` ou equivalente).
- Não é Player, não é Execution Policy, não é Blast Radius.
- Avaliação: no Runner, **depois** do verbo efetivo de policy e
  **antes** de `Player.Execute`.
- Pacote não importa `entrypoint`.
- Recorte de **G-43**: este slice entrega **só** Claims v0. Blast
  Radius permanece HYPOTHESIS.

### G-94 — Resource kinds v0

**Status: CONFIRMED**

Só estes kinds:

| kind | Chave | Notas |
|---|---|---|
| `workspace` | root canônico do workspace | Um por processo/workspace |
| `path` | path relativo limpo (`/` ; sem `..`) | Arquivo ou diretório no workspace |

Normalização:

- Path vazio, `.` ou equivalente → **promove** a `workspace`.
- Path deve permanecer dentro do workspace (mesma regra FS); escape →
  `validation.invalid_input` **antes** do claim (já é sandbox do Player).
- Comparação de `path` é por segmentos: `src` conflita com `src/a.go`;
  `src` **não** conflita com `src2`.

Conflito (exclusivo; sem read-lock compartilhado no v0):

- `workspace` conflita com qualquer outro claim no mesmo workspace
  (`workspace` ou `path`).
- `path` P conflita com `path` Q se P = Q ou um é prefixo segmentado
  do outro.
- Re-acquire do **mesmo** Run no mesmo recurso é idempotente.

Fora: `database`, `environment`, `deployment`, `repository` como kind
distinto (`git.*` usa `workspace`).

### G-95 — Tabela automática

**Status: CONFIRMED**

O Core deriva o claim do step; não há `claims[]` no Manifest nem no
Task IR no v0.

| Capability | Claim | Fonte da chave |
|---|---|---|
| `fs.write` | `path` | input `path` |
| `git.add` | `workspace` | workspace root (não usa `paths[]`) |
| `git.commit` | `workspace` | workspace root |
| `docker.build` | `path` (ou `workspace` se context `.`) | input `context` (default `.`) |
| `docker.run` com `mount_workspace=true` | `workspace` | workspace root |
| qualquer outra, incl. `shell.exec` | **nenhum** | — |

Sem auto-claim (exemplos): `fs.read` / `list` / `stat`, `git.status` /
`diff` / `log`, `docker.ps` / `inspect` / `logs`, `docker.run` com
`mount_workspace=false` (default), `shell.exec`, pipeline, LLM.

`git.add` é `workspace` de propósito: `paths[]` múltiplos + overlap
com `git.commit` no mesmo Run seria deadlock fácil no v0.

Capability fora da tabela não claima. Acrescentar linha exige nova
promoção em `04` (não Manifest ad hoc).

### G-96 — Lifecycle e persistência

**Status: CONFIRMED**

```text
policy allow (ou grant HITL)
  → acquire
      → ok: claim.acquired → Execute
      → conflito: claim.conflict → Run failed (Player não roda)
Run terminal (succeeded | failed | cancelled)
  → release todos os claims do run_id → claim.released (por recurso)
```

- Acquire na admissão do **step** (não na admissão da Task inteira).
  Steps anteriores `allow` sem claim já podem ter corrido.
- Um Run pode acumular vários claims (ex.: `fs.write` A depois
  `fs.write` B); todos ficam até o terminal do Run.
- Persistência: tabela SQLite no mesmo `.runtgine/runtgine.db`
  (ex.: `resource_claims`: `run_id`, `kind`, `key`, `acquired_at`,
  `released_at`).
- Restart / `api.Open`: claims cujo Run não está `running` nem
  `waiting_approval` são released (órfãos). Run `running` abandonado
  pelo crash segue o caminho já existente de falha e então libera.
- `waiting_approval` **não** segura claim (HITL é antes do acquire).

Eventos novos:

| type | Quando |
|---|---|
| `claim.acquired` | Lock gravado; payload: `kind`, `key`, `step_id`, `capability` |
| `claim.conflict` | Recurso tomado por outro Run; payload: `kind`, `key`, `holder_run_id` |
| `claim.released` | Release no terminal; payload: `kind`, `key` |

Código de erro novo:

| code | Fase |
|---|---|
| `claim.conflict` | Runner, após policy, antes do Execute |

Não há estado de Run novo. Conflito → `failed` como qualquer step
falho de admissão tardia.

### G-97 — Conflito fail-fast e superfície

**Status: CONFIRMED**

- Conflito é **fail-fast** determinístico. Sem espera, sem fila, sem
  `waiting_claim`, sem timeout de lock.
- O Run que chega depois falha; o holder continua.
- CLI: `runtgine run` / `status` mostram `claim.conflict` e o
  `holder_run_id` na mensagem. Sem subcomando `runtgine claims`.
- TUI: RUNS/LIVE já mostram `failed`; o erro entra no mesmo canal de
  eventos. Sem aba nova, sem tecla nova, sem GRAPH.
- Cancelamento do holder (`CancelRun`) libera claims e permite retry
  do outro Task (nova submissão; o Run falho não retoma sozinho).

### G-98 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- Blast Radius (resto de G-43; spec futura)
- Wait / queue / estado `waiting_claim` / retry automático de claim
- Locks distribuídos (NATS / multi-processo)
- Wildcards de capability; claim por argv do `shell.exec`
- Campo Manifest / Task IR `claims[]`
- Read-lock compartilhado; claim por `fs.read`
- Kinds além de `workspace` e `path`
- Human Player / `claim.*` capabilities
- TUI GRAPH; aba Claims; CLI list/inspect de claims
- Policy por path (continua sandbox do Player)
- Inferir claims a partir do Runtime Graph

---

## 4. Critérios de aceite

1. `runtgine run examples/hello.json` continua `run.succeeded` com
   Runs concorrentes (`shell.exec` não claima).
2. Dois Runs `fs.write` no mesmo `path` (ou prefixo): o segundo falha
   com `claim.conflict` **antes** do segundo Execute; o primeiro
   completa.
3. `fs.write` em paths disjuntos (`a.txt` vs `b.txt`) pode concorrer.
4. `git.commit` (workspace) conflita com `fs.write` em qualquer path
   do mesmo workspace.
5. `docker.run` sem `mount_workspace` não claima; com
   `mount_workspace=true` claima `workspace`.
6. Step `approval-required` não segura claim enquanto
   `waiting_approval`; após grant, acquire ocorre; conflito falha o
   Run sem reabrir HITL.
7. Restart: claim de Run já `failed`/`succeeded`/`cancelled` não
   bloqueia um Run novo.
8. `go test ./internal/core/claim/...` (e Runner) cobrem acquire,
   overlap de path, workspace vs path, e órfãos no boot.
9. `go test ./...` e `go vet ./...` verdes.
10. OpenSpec `024-resource-claims` arquivado após o merge do **código**
    (slice 12).

---

## 5. Ordem do slice de código

Slice 12 feito:

1. Pacote `internal/core/claim` + migração SQLite
2. Tabela automática + overlap + boot sweep
3. Runner: Policy → Claim → Execute; eventos + `claim.conflict`
4. CLI mensagem; TUI sem aba nova (status `failed` existente)
5. Testes de corrida `fs.write`; README Estágio: Slice 12 Feito
6. OpenSpec `024` arquivado

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-93 Papel (Core, não Player)
- [x] G-94 Kinds `workspace` / `path` + overlap
- [x] G-95 Tabela automática (shell.exec fora)
- [x] G-96 Lifecycle + SQLite + eventos
- [x] G-97 Conflito fail-fast + superfície
- [x] G-98 Exclusões (Blast, wait, Manifest claims[], GRAPH)
