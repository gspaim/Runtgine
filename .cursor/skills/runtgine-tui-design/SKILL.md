---
name: runtgine-tui-design
description: Design or implement the Runtgine Bubble Tea TUI using the confirmed Constellation Mission Control visual system. Use for tabs, layouts, Lip Gloss styles, components, interaction, responsive behavior, screenshots, and TUI reviews.
---

# Runtgine TUI Design

Use this skill whenever a task creates, changes, or reviews the Runtgine TUI.

## Required context

Read before implementation:

1. `docs/14-tui-design.md`
2. `docs/04-decisoes.md`
3. `docs/11-protocolo-v0.md`
4. Relevant Core APIs in `internal/core/api`

## Non-negotiable architecture

- TUI is an Entry Point/surface over the Core.
- Use Core APIs (`GetRun`, `Subscribe`, `CancelRun`, submission APIs).
- Never call a Player directly.
- Never bypass Validator, Runner, Event Bus, or SQLite.
- Do not turn the TUI into a terminal multiplexer.
- Do not add tuios or embedded PTYs without a new confirmed decision.

## Visual direction

System name: **Constellation Mission Control**.

Space theme is visual only. Keep technical domain words:

- Task
- Run
- Step
- Event
- Player
- Board

Do not replace them with fictional names in commands, schemas, or Core APIs.

## Stack

- Bubble Tea
- Lip Gloss
- Bubbles

Prefer reusable components and centralized theme tokens. Do not scatter raw
color values across views.

## Tokens

```go
Space     = "#070B14"
Panel     = "#101A2F"
Starlight = "#8FB8FF"
Violet    = "#9D8CFF"
Telemetry = "#73E2C2"
Amber     = "#FFB86B"
Anomaly   = "#FF6B8A"
Muted     = "#64748B"
```

Provide ANSI-256/ASCII fallbacks and honor `NO_COLOR`.

## Required tabs

1. `INTENT`
2. `RUNS`
3. `LIVE`
4. `BOARD`
5. `EVENTS`
6. `GRAPH`
7. `CONFIG`

Use one full-screen application. No overlapping shell windows.

INTENT is the visual Entry Point (Mission Brief). Spec: `docs/32-intent-surface-v0.md`.
Not a chatbot — compile NL/JSON to Task IR and submit via Core APIs.

### INTENT

- Multiline NL input (or JSON Task IR mode toggle).
- Preview panel: Task IR + `method` via `CompileIntent` (`Ctrl+p`).
- Submit via `SubmitIntent` (`Ctrl+Enter`); on success select run and switch to LIVE.
- Session history cap (~10 entries: run_id + summary); no long transcript persistence.
- Never call Players directly.

### RUNS

Show run ID, intent, status, and elapsed time. Selection uses starlight/violet.
Status includes text/symbol in addition to color.

### LIVE

Show the selected run, step dependency trajectory, progress, current Player,
capability, ContextPack size, and latest telemetry. Use a vertical step list
when the terminal is too narrow for a constellation trajectory.

### BOARD

Show `INTAKE`, `IN FLIGHT`, `LANDED`. Cards display GitHub reference, title,
run ID, and pipeline state. Do not create subtask cards in the MVP.

### EVENTS

Show UTC, event type, run ID, step, Player, filter input, and JSON payload.

### GRAPH

Read-only workspace Runtime Graph (`GetGraphSnapshot`). Show node/edge
counts, a `kind`+`id` list, and selected-node attrs plus incident edges.
`r` calls `RefreshGraph` then reloads the snapshot. `/` filters kind/id
(independent of EVENTS). Width `< 80`: vertical list only, no 2D canvas.
Do not mix Blast or QueryHits into this tab.

### CONFIG

Read-only in v0. Mask secrets. Show precedence:
`defaults < config.json < env < CLI flags`.

## Interaction

Default keymap:

| Key | Action |
|---|---|
| `tab` / `shift+tab` | next/previous tab |
| arrows or `j`/`k` | navigate |
| `enter` | inspect |
| `c` | cancel selected active run (confirm first) |
| `a` | approve selected run in `waiting_approval` (Core `ApproveRun` grant) |
| `d` | deny selected run in `waiting_approval` (Core `ApproveRun` deny) |
| `/` | filter |
| `r` | refresh |
| `Ctrl+p` | preview intent (INTENT tab only) |
| `Ctrl+Enter` | submit intent (INTENT tab only) |
| `q` | quit |

Keep key hints visible in the footer.

## Responsive behavior

- `>=120`: side-by-side panels.
- `80–119`: one main panel plus detail/drawer.
- `<80`: compact vertical mode; no horizontal constellation graph.

All views must survive resize events.

## Accessibility

- Never encode state by color alone.
- Support `NO_COLOR`.
- Provide ASCII-safe glyphs.
- Keep focus visible.
- Animations must be optional.
- Ensure readable contrast.

## Implementation workflow

1. Identify which Core API and event data the screen uses.
2. Define/update theme tokens and reusable components.
3. Implement model/update/view without side effects in `View`.
4. Route commands through Core APIs.
5. Test state transitions, key handling, resize behavior, and no-color mode.
6. Add snapshot/golden tests only where stable; prefer model/update unit tests.
7. Run `go test ./...`.

## Review checklist

- [ ] No direct Player calls
- [ ] No duplicated runtime state as source of truth
- [ ] All seven tabs remain coherent
- [ ] Status has text/symbol, not color only
- [ ] Narrow terminal fallback works
- [ ] `NO_COLOR` works
- [ ] Secrets are masked
- [ ] Board behavior matches `docs/12-board-p1.md`
- [ ] Visual tokens match `docs/14-tui-design.md`
- [ ] No tuios/PTY scope creep
