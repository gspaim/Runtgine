# Proposal: 032-intent-surface

## Why

Intent Engine v0 and `runtgine intent` exist on the CLI. TUI Constellation and
future Wails desktop are observation-first (RUNS, LIVE, EVENTS, …). Operators
have no visual Entry Point to submit NL intent without leaving the UI.

Product expectation: “where do I talk to Runtgine?” — today only the terminal.

## What Changes

- Canonical `docs/32-intent-surface-v0.md` (G-141..G-146 CONFIRMED)
- `docs/04-decisoes.md`: Intent Surface v0 section
- `docs/14-tui-design.md`: seventh tab **INTENT** (first in order)
- `docs/17-intent-engine-v0.md`: remove “TUI input de NL” from exclusions;
  point to `32`
- `docs/09-mvp.md`, `docs/10-gaps.md`, `docs/README.md`, `AGENTS.md`
- Skill `.cursor/skills/runtgine-tui-design/SKILL.md`: seven tabs + INTENT keys
- TUI `internal/entrypoint/tui` (**slice 21 — not this docs PR**)
- Wails INTENT view (**Fase 3 — separate implementation**)

## What Does Not Change

- Intent Engine compiler semantics (`17`)
- Validator / Runner / Event Bus
- Other TUI tabs semantics (LIVE, GRAPH, …)
- CLI `runtgine intent` (remains for scripts/CI)
- No chatbot thread, no `intent.*` Player, no transcript RAG

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `032-intent-surface` |
| Doc canônico | [`docs/32-intent-surface-v0.md`](../../../docs/32-intent-surface-v0.md) |
| Gaps | G-141..G-146 **CONFIRMED** |
| Código TUI | slice 21 — **blocked** until this package + `04` + `14` merged |
| Código Wails | Fase 3 — same product spec, later branch |

## Approach

1. INTENT tab: NL input + preview (`CompileIntent`) + submit (`SubmitIntent`)
2. After submit → select run + switch to LIVE
3. Wails mirrors the same flow with shadcn-svelte (Fase 3)
4. Explicit exclusions: not a chatbot, no Player calls from UI

## Impact

- `internal/entrypoint/tui` (slice 21)
- Future Wails frontend (Fase 3)
- Docs `14`, `17`, skill, `09`, `10`
