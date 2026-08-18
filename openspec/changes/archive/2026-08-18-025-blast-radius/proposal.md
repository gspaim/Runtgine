# Proposal: 025-blast-radius

## Why

Claims (024) lock *who holds* a workspace/path at Execute time.
Operators still cannot inspect *what a Task would touch or claim*
before `runtgine run`. The long-term idea (`02`: Change → Graph →
affected symbols) is a later spec. v0 is a deterministic Impact
Report from a validated Task IR plus a read-only overlay of active
claims.

## What Changes

- Canonical doc `docs/25-blast-radius-v0.md` (G-99..G-104 CONFIRMED)
- Core package `internal/core/blast` (**slice 13 — not this spec PR**)
- `BlastTask` API + CLI `runtgine blast`
- Touches include reads (`fs.read` / `list` / `stat`); predicted
  claims reuse G-95 (`claim.Required`)
- `risk`: `none` | `path` | `workspace` from predicted claims only
- `conflicts[]` against active claims; never Acquire

## What Does Not Change

- Runner order: Validator → Policy → Claim → Execute (no auto-blast)
- Resource Claim engine (024 / G-93..G-98)
- Execution Policy verbs
- Runtime Graph / QueryHits (no walk)
- TUI tabs (`14` still RUNS/LIVE/BOARD/EVENTS/CONFIG)
- `shell.exec` sandbox; hello.json stays concurrent and blast-empty

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `025-blast-radius` |
| Doc canônico | [`docs/25-blast-radius-v0.md`](../../../docs/25-blast-radius-v0.md) |
| Gaps | G-99..G-104 **CONFIRMED** (resto de G-43) |
| Código | Ainda não — slice 13; **bloqueado** até este pacote + `04` |

## Approach

1. Compute report from Task IR tables (touches vs G-95 claims)
2. Overlay `ListActiveClaims` + `Overlaps` (read-only)
3. CLI prints JSON; validation errors fail the process; conflicts do not
4. No persistence, no events, no Execute gate

## Impact

- Package novo no slice 13: `internal/core/blast`
- `internal/core/api`, CLI
- Depende de `claim.Required` / `Overlaps` (slice 12)
