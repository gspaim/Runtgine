# Proposal: 035-wails-v0

## Why

CLI, TUI and `runtgine serve` exist. Wails + Svelte has been CONFIRMED
since G-35 / `07`, and G-144 already defined INTENT desktop semantics,
but there is no desktop Entry Point. Operators who want a window still
have only the terminal TUI.

## What Changes

- Canonical `docs/35-wails-v0.md` (G-159..G-165 CONFIRMED)
- Pin Wails **v2** (stable); Wails v3 beta out of v0
- Cross-refs in `04`, `07`, `09`, `10`, `01`, `05`, `11`, `13`, `14`, `17`, `32`, `34`
- `docs/README.md`, `AGENTS.md`, `README.md` estágio, `REVIEW.md`
- OpenSpec package `035-wails-v0`
- Task IR `source.entry_point` enum gains `wails` (schema note; code in slice 27)

## What Does Not Change

- TUI / CLI / HTTP API behaviour
- Validator / Runner / Event Bus
- Core does not import `entrypoint`
- No Wails project scaffold in this PR (slices 27–28 later)
- MCP, NATS, extra Players

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `035-wails-v0` |
| Doc canônico | [`docs/35-wails-v0.md`](../../../docs/35-wails-v0.md) |
| Gaps | G-159..G-165 **CONFIRMED** (spec); fecha G-144 no desktop |
| Code | slices 27–28 — **not started** |

## Approach

Two implementation slices:

1. **Slice 27** — Wails v2 app + Core bindings + INTENT + LIVE
2. **Slice 28** — remaining Constellation views + Lessons HITL UI

## Impact

- Future: `internal/entrypoint/desktop`, CLI `runtgine desktop`
- Schema: `entry_point` includes `wails`
- Docs only in this PR
