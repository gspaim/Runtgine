# 25 — Blast Radius v0

Análise de impacto **determinística** de uma Task IR: o que o Run
*tocaria* e o que *claimaria*, **sem** executar Players.

Inventário: [10-gaps.md](10-gaps.md) (G-99+; resto de G-43).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Resource Claims v0 ([24-resource-claims-v0.md](24-resource-claims-v0.md),
slice 12). Claims respondem *quem segura*; Blast responde *o que seria
afetado*.
Consumidor seguinte: TUI GRAPH (exige [14-tui-design.md](14-tui-design.md)
+ skill) ou walk no Runtime Graph — **fora** deste corte.

**Status deste doc: CONFIRMED (v0).** G-99..G-104 implementados no
slice 13. Walk no Graph, gate de Execute, TUI GRAPH e Human Player
permanecem fora.

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-18-025-blast-radius/`](../openspec/changes/archive/2026-08-18-025-blast-radius/).
Deltas mergeados em `openspec/specs/blast-radius/`. Branch de implementação:
`feat/025-blast-radius`.

---

## 1. Problema

Policy decide *se* a capability pode correr. Claim decide *quem segura*
o recurso **no momento do Execute**. O operador não tem um relatório
estável, antes de submeter, de:

- quais paths / workspace a Task **toca** (inclui `fs.read`);
- quais recursos **seriam claimados** (tabela G-95);
- se isso **sobreporia** claims ativos agora.

O conceito em `02` (`Change → Graph → Affected …`) é a visão longa.
O v0 **não** caminha o Runtime Graph (mesmo recorte de `18`/`19`:
Blast derivado do Graph fica para spec futura). A fonte é a Task IR
já validada + a tabela de claims + um overlay read-only dos claims
ativos.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Motor no Core (`internal/core/blast`) | Player `blast.*` / Human Player |
| Relatório computado (on-demand) | Lock, fila, verbo de policy |
| Predicted claims = tabela G-95 | Campo Manifest / Task IR `blast[]` |
| Touches incluem leituras (`fs.read` / `list` / `stat`) | Inferir argv de `shell.exec` |
| Overlay contra claims ativos | `Acquire` / mutar `resource_claims` |
| CLI `runtgine blast` + `BlastTask` | Aba TUI GRAPH; auto-blast no Runner |

Regras:

1. Validator e Registry continuam soberanos. Blast **não** inventa
   capability nem ignora schema.
2. Ordem do **Runner não muda**: Validator → Policy → Claim → Execute.
   Blast **não** entra nesse pipeline no v0 (não atrasa `hello.json`).
3. `BlastTask` valida a Task (mesmo subset de admissão que `SubmitTask`
   usa **antes** de policy) e devolve o relatório. Não chama Player,
   não avalia policy, não adquire claim, não cria Run.
4. Predicted claims **reutilizam** `claim.Required` (G-95). Blast não
   redefine o lock.
5. `shell.exec` / pipeline / LLM: sem touch e sem predicted claim
   (hello.json → `risk: none`).
6. Pacote não importa `entrypoint`.

---

## 3. Cortes confirmados (G-99+)

### G-99 — Papel e pacote

**Status: CONFIRMED**

- Blast Radius vive no Core (`internal/core/blast` ou equivalente).
- Não é Player, não é Execution Policy, não é Resource Claim.
- Análise ≠ lock. `BlastTask` é read-only sobre o Store de claims.
- Recorte de **G-43**: Claims já estão em `24` (G-93..G-98). Este
  slice entrega **só** Blast v0 (Task IR → Impact Report).
- Walk `Change → Graph → symbols/workflows` permanece HYPOTHESIS
  (spec futura; não `025`).

### G-100 — Impact Report

**Status: CONFIRMED**

JSON estável (`schema_version` alinhado ao Task IR v0: `0.1.0`):

| Campo | Tipo | Notas |
|---|---|---|
| `schema_version` | string | `0.1.0` |
| `capabilities` | string[] | Capabilities dos steps, ordem de aparição, sem duplicar |
| `touches` | object[] | Um por step que toca path/workspace (ver G-101) |
| `predicted_claims` | object[] | União da tabela G-95; `kind` + `key` únicos |
| `risk` | enum | `none` \| `path` \| `workspace` — só a partir de `predicted_claims` |
| `conflicts` | object[] | Overlay vs claims **ativos**; vazio se nenhum overlap |
| `images` | string[] | `image` de `docker.run` / referência de `docker.build` se presente; informativo |

Touch:

```text
{ "kind": "workspace"|"path", "key": string, "capability": string,
  "step_id": string, "mode": "read"|"write" }
```

Predicted claim (espelha o Resource de `24`, mais proveniência):

```text
{ "kind": "workspace"|"path", "key": string, "capability": string,
  "step_id": string }
```

Conflict:

```text
{ "kind": "workspace"|"path", "key": string, "holder_run_id": string }
```

`risk`:

- `none` — `predicted_claims` vazio (ex.: só `shell.exec` / `fs.read`);
- `path` — todos os predicted claims são `path`;
- `workspace` — pelo menos um predicted claim é `workspace`.

Leituras **não** sobem o risco. `fs.read` sozinho → `risk: none`.

Sem kinds novos (`image`, `database`, `symbol`). `images[]` é lista
plana, não recurso.

### G-101 — Tabelas de derivação

**Status: CONFIRMED**

**Predicted claims** = exatamente G-95 (`claim.Required`). Sem segunda
tabela de lock.

**Touches** (relatório; não lock):

| Capability | Touch | `mode` | Fonte |
|---|---|---|---|
| `fs.write` | `path` (ou `workspace` se `.`) | `write` | input `path` |
| `fs.read` / `fs.list` / `fs.stat` | `path` (ou `workspace` se `.`) | `read` | input `path` |
| `git.add` | um touch `path` **por** elemento de `paths[]` | `write` | `paths[]` |
| `git.commit` | `workspace` | `write` | — |
| `docker.build` | `path` do `context` (default `.` → workspace) | `write` | `context` |
| `docker.run` com `mount_workspace=true` | `workspace` | `write` | — |
| qualquer outra, incl. `shell.exec`, `git.status`/`diff`/`log`, `docker.ps`/`inspect`/`logs`, `docker.run` sem mount | **nenhum** | — | — |

Notas:

- `git.add` **toca** paths individuais no relatório, mas o **claim**
  previsto continua `workspace` (G-95). Não há deadlock: Blast não
  adquire.
- Normalização de path = a mesma de `24` (segmentos; `.` → workspace;
  rejeitar escape). Falha → `validation.invalid_input`, sem relatório.
- Capability fora da tabela de touch não aparece em `touches`.
- Acrescentar linha exige nova promoção em `04`.

### G-102 — Overlay de claims ativos

**Status: CONFIRMED**

Se o Core tem Store aberto, `BlastTask` consulta claims ativos
(`ListActiveClaims`) e preenche `conflicts[]` com predicted claims que
**Overlaps** (mesma regra G-94) contra outro `run_id`.

- Read-only. Nunca `Acquire` / `Release`.
- Sem Store / sem claims → `conflicts: []`.
- Não avalia policy: um Task que seria `deny` ainda recebe relatório
  (útil antes de `runtgine run`).
- HITL irrelevante: não há Run.

### G-103 — Superfície

**Status: CONFIRMED**

Core API:

```text
BlastTask(TaskIR) -> (ImpactReport | ValidationError)
```

CLI:

```text
runtgine blast <task.json|task.yaml>
```

Imprime o JSON do relatório (stdout). Abre o Core do workspace como
`graph snapshot` (para o overlay). Exit ≠ 0 só em validação / I/O;
`conflicts` não é erro de processo (o relatório já os lista).

- Sem `runtgine blast --apply`. Sem persistir o relatório.
- Sem evento `blast.computed` no v0 (não há Run).
- TUI: **sem aba nova**, sem tecla, sem GRAPH. O operador usa a CLI.
  EVENTS/RUNS não mudam.
- `runtgine run` **não** chama Blast automaticamente.

### G-104 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- Walk no Runtime Graph (symbols, workflows, `mentions`, QueryHits)
- Gate / bloquear Execute por `risk` ou `conflicts`
- Auto-blast no Runner; evento `blast.computed`
- Wait / queue / `waiting_claim` (continua fora, como em `24`)
- Campo Manifest / Task IR `blast[]` ou `claims[]`
- Inferir impacto de `shell.exec` argv
- Kinds além de `workspace` / `path`; kind `image`
- Persistência de histórico de blasts
- TUI GRAPH; painel Blast; Board write-back
- LLM / ranking / “risco” não derivado da tabela
- Policy por path; Human Player / `blast.*` capabilities
- Locks distribuídos; MCP; HTTP

---

## 4. Critérios de aceite

1. `runtgine blast examples/hello.json` imprime `risk: none`,
   `predicted_claims: []`, `touches: []`, `capabilities` contendo
   `shell.exec`. Não cria Run nem claim.
2. Task só com `fs.read` em `a.txt`: touch `path:a.txt` `mode:read`;
   `predicted_claims` vazio; `risk: none`.
3. Task `fs.write` `notes.md`: touch write + predicted claim `path`;
   `risk: path`.
4. Task `git.add` com `paths: ["README"]`: touches incluem `path:README`;
   predicted claim é `workspace`; `risk: workspace`.
5. Com um Run holder de `path:notes.md` (claim ativo), blast de outro
   `fs.write` no mesmo path lista `conflicts[].holder_run_id` sem
   falhar o processo e sem Acquire.
6. Path que escapa o workspace → erro de validação; sem relatório.
7. `runtgine run examples/hello.json` permanece inalterado (sem
   blast no Runner).
8. `go test ./internal/core/blast/...` cobre tabelas, risk, overlay e
   hello vazio.
9. `go test ./...` e `go vet ./...` verdes.
10. OpenSpec `025-blast-radius` arquivado após o merge do **código**
    (slice 13).

---

## 5. Ordem do slice de código

Slice 13 feito:

1. Pacote `internal/core/blast` (Report, Touched, risk, overlay)
2. Reuso de `claim.Required` + `claim.Overlaps` / `NormalizePath`
3. `api.BlastTask` (valida, não submete)
4. CLI `runtgine blast`
5. Testes das tabelas + overlay; README Estágio: Slice 13 Feito
6. OpenSpec `025` arquivado

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-99 Papel (Core, não Player; análise ≠ lock)
- [x] G-100 Impact Report + `risk`
- [x] G-101 Touches vs predicted claims (G-95 intacto)
- [x] G-102 Overlay read-only vs claims ativos
- [x] G-103 CLI `blast` + `BlastTask`; sem TUI GRAPH; sem auto no Runner
- [x] G-104 Exclusões (Graph walk, gate, shell argv, persistência)
