# Design: 024-resource-claims

## Technical approach

### Package

`internal/core/claim` with:

- `type Kind string` — `workspace`, `path`
- `type Resource struct { Kind Kind; Key string }`
- `Required(capability string, input map) (Resource, bool, error)`
  using the G-95 table
- `Acquire(runID, stepID, res) error` — `ErrConflict` with holder
- `ReleaseAll(runID) error`
- `SweepOrphans()` on `api.Open`

Overlap is exclusive. Normalize paths (slash, clean, reject `..`).
Empty / `.` → `workspace`. Path prefix is **segment**-aware
(`src` vs `src/a.go` conflict; `src` vs `src2` do not).

The Runner calls `Acquire` **after** policy resolve / HITL grant and
**before** `Player.Execute`. Deny never acquires. `waiting_approval`
does not hold a claim.

### Persistence

New SQLite table in the existing `.runtgine/runtgine.db`, e.g.
`resource_claims(run_id, kind, key, step_id, acquired_at, released_at)`.
Unique active (kind, key) where `released_at IS NULL`.

Boot: release rows whose run is not `running` and not
`waiting_approval`. A crashed `running` run follows the existing
failure path, then `ReleaseAll`.

Hold until Run **terminal**, not step end. Same-run re-acquire of the
same resource is idempotent.

### Events

Reuse envelope from `11`. New types:

- `claim.acquired` — payload `kind`, `key`, `step_id`, `capability`
- `claim.conflict` — payload `kind`, `key`, `holder_run_id`
- `claim.released` — payload `kind`, `key`

Error code: `claim.conflict`. No new Run status.

### Auto-claim table

| Capability | Resource |
|---|---|
| `fs.write` | `path` from input `path` |
| `git.add` / `git.commit` | `workspace` |
| `docker.build` | `path` from `context` (default `.` → workspace) |
| `docker.run` if `mount_workspace=true` | `workspace` |
| else (incl. `shell.exec`) | none |

### Surfaces

CLI: existing `run` / `status` print `claim.conflict` and holder.
No `runtgine claims`. TUI RUNS/LIVE already render `failed`; no new
tab or keys. Follow TUI skill: Core API only.

### Policy interaction

```text
deny            → task.rejected; no claim
approval-required → waiting_approval; no claim yet
grant           → Acquire then Execute
claim.conflict after grant → Run failed; do not re-open HITL
```

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| `shell.exec` auto-claim workspace | Mata concorrência G-30 / hello.json |
| Wait / `waiting_claim` | Não é HITL; holder pode ser longo; v0 fail-fast |
| Manifest `claims[]` | Superfície extra; tabela automática basta no v0 |
| Path-level for `git.add` | Vários paths + commit no mesmo Run = deadlock |
| Blast Radius no mesmo slice | Análise ≠ lock; padrão Graph vs Hits |
| Read locks | Complexidade; `fs.read` permanece livre |

## Risks

| Risco | Mitigação |
|---|---|
| `docker.build` context `.` bloqueia tudo | Documentado: `.` promove a workspace |
| Órfãos após crash | Sweep no `api.Open` + release no terminal |
| Testes 1–11 quebram | Só mutadores da tabela claimam |
| TUI scope | Sem aba; failed existente |

## Packages touched (slice 12, not this PR)

- `internal/core/claim` (novo)
- `internal/core/runner`, `store`, `api`, `event`
- `docs/11-protocolo-v0.md` (estendido neste PR de spec)
