# Proposal: 027-blast-graph-walk

## Why

Blast v0 (`25`) reports Task IR touches/claims. The Graph already
stores `mentions` from prior runs (repo-search). Operators cannot see
that overlap without a second tool. The long `02` walk (Players,
Workflows, Symbols) stays out. This change is one inbound `mentions`
hop from `path` touches.

## What Changes

- Canonical `docs/27-blast-graph-walk-v0.md` (G-111..G-116 CONFIRMED)
- `blast.Walk` + Report field `affected` (**slice 15 — not this spec PR**)
- `BlastTask` reads `GetGraphSnapshot` after `Analyze`; degrades to `[]`
- CLI JSON gains `affected` (always present)
- G-104's Graph-walk exclusion is superseded **only** for this hop

## What Does Not Change

- Predicted claims / `risk` / overlay (`25`)
- Runner: still no auto-blast; no Execute gate
- QueryHits / ContextPack (`19`)
- TUI GRAPH (`26`): still no blast-from-node
- `shell.exec` argv; hello.json stays blast-empty (`affected: []`)
- Graph persistence / kinds (`18`)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `027-blast-graph-walk` |
| Doc canônico | [`docs/27-blast-graph-walk-v0.md`](../../../docs/27-blast-graph-walk-v0.md) |
| Gaps | G-111..G-116 **CONFIRMED** |
| Código | Ainda não — slice 15; **bloqueado** até este pacote + `04` |

## Approach

1. Seed unique `path` touches; 1 hop inbound `mentions`
2. Never fail the IR report on graph errors
3. Keep `schema_version` `0.1.0` (additive field)

## Impact

- `internal/core/blast`, `BlastTask` (slice 15)
- Docs `25` G-104 pointer, `02` Blast section, `18` “derivados do graph”
