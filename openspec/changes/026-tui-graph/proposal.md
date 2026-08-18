# Proposal: 026-tui-graph

## Why

Runtime Graph (018) and CLI `graph snapshot` exist. The TUI still has
only five tabs. LIVE shows one Run's `depends_on` trajectory, not the
workspace structural graph. `18` deferred the GRAPH tab until `14` +
skill. This change amends `14` and the TUI skill, then (slice 14)
adds the tab as a Core-API surface.

## What Changes

- Canonical `docs/26-tui-graph-v0.md` (G-105..G-110 CONFIRMED)
- `docs/14-tui-design.md`: sixth tab GRAPH (EVENTS … GRAPH … CONFIG)
- Skill `.cursor/skills/runtgine-tui-design/SKILL.md`: six required tabs
- TUI `internal/entrypoint/tui` (**slice 14 — not this spec PR**)
- `CoreAPI` grows `GetGraphSnapshot` + `RefreshGraph`

## What Does Not Change

- Graph storage / kinds / sync (`18`)
- QueryHits / ContextPack (`19`)
- Blast Radius CLI (`25`); no Blast-from-Graph
- Claims, Policy, Players
- No PTY, tuios, multiplexer, Wails
- LIVE tab semantics

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `026-tui-graph` |
| Doc canônico | [`docs/26-tui-graph-v0.md`](../../../docs/26-tui-graph-v0.md) |
| Gaps | G-105..G-110 **CONFIRMED** |
| Código | Ainda não — slice 14; **bloqueado** até este pacote + `04` + `14` |

## Approach

1. Read-only list+detail over `GetGraphSnapshot`
2. Same keymap as other tabs; `r` refreshes graph
3. Narrow terminals: list only
4. Fake Core in tests supplies a snapshot (TUI still non-interactive in CI)

## Impact

- `internal/entrypoint/tui` (slice 14)
- Docs `14`, skill, `18` “tab deferred” line
