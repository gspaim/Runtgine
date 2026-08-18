# Design: 026-tui-graph

## Technical approach

### Tab index

Insert GRAPH before CONFIG so CONFIG stays last (read-only settings):

```go
tabRuns, tabLive, tabBoard, tabEvents, tabGraph, tabConfig
tabNames = []string{"RUNS", "LIVE", "BOARD", "EVENTS", "GRAPH", "CONFIG"}
```

`tabCount` becomes 6. Existing `tab`/`shift+tab` loops over `tabCount`.

### CoreAPI

```go
GetGraphSnapshot(ctx context.Context) (graph.Snapshot, error)
RefreshGraph(ctx context.Context) error
```

`api.Core` already has both. Slice 14 only extends the TUI interface +
fakeCore. Never import `internal/core/store` from the view.

### View cache

`Model` holds `graph.Snapshot` + `graphSelected int` + optional
`graphFilter` (reuse `filtering`/`filter` when `tab == tabGraph`, or a
dedicated field if EVENTS filter must stay independent — **prefer a
dedicated `graphFilter`** so switching tabs does not leak EVENTS `/`).

`r` on GRAPH: command that calls RefreshGraph then GetGraphSnapshot.
On other tabs `r` keeps today's full refresh (runs/events/config).

### Rendering

- Sort nodes: kind order `player, capability, task, run, path, symbol`,
  then `id`.
- Truncate `id` with the same rune-width helper RUNS uses.
- Detail: `json.MarshalIndent` of attrs; edge lines
  `fmt.Sprintf("%s %s:%s → %s:%s", e.Kind, e.From.Kind, e.From.ID, ...)`.
- Counts: walk snapshot once per View (cheap; snapshot is small in v0).

### Tests

- Tab cycle visits GRAPH.
- Fixture snapshot includes shell player + provides edge; detail shows it.
- Filter `capability` hides `player` rows.
- Width 70: View does not panic.
- `NO_COLOR` path still includes the word `GRAPH`.

No `computerUse` / interactive TTY required in CI (`14` already tests
via `tea.KeyMsg`).

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| GRAPH as LIVE overlay | LIVE is one Run; Graph is workspace |
| Embed `runtgine graph snapshot` JSON raw | Unreadable; counts+list is the v0 UX |
| QueryHits search box | That's `19`; keep GRAPH structural |
| Auto-refresh on every event | Extra SQLite; `r` is enough |
| Tecla `g` | Skill keymap is already crowded; tab is enough |

## Risks

| Risco | Mitigação |
|---|---|
| Snapshot grande | v0 kinds are few; truncate ids; no canvas |
| Filter clash with EVENTS | Dedicated `graphFilter` |
| Skill vs `14` drift | Update both in this spec PR |

## Packages touched (slice 14, not this PR)

- `internal/entrypoint/tui`
- tests `model_test.go` / fakeCore
