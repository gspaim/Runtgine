# TUI GRAPH

Status: comportamento atual (pós-slice 14 / TUI GRAPH v0).

## Requirements

### Requirement: GRAPH is a TUI tab over Core snapshot

The TUI SHALL expose a sixth tab `GRAPH` that renders
`GetGraphSnapshot` without calling a Player or writing graph rows
except via `RefreshGraph`. GRAPH SHALL NOT be a Player or a new Core
package.

#### Scenario: Boot nodes visible

- GIVEN a Core that has refreshed the registry into the Graph
- WHEN the operator opens GRAPH
- THEN player and capability nodes are listed
- AND selecting `player` `shell` shows a `provides` edge to `shell.exec`

### Requirement: Six tabs in fixed order

The TUI SHALL use exactly this tab order:
`RUNS`, `LIVE`, `BOARD`, `EVENTS`, `GRAPH`, `CONFIG`.
Navigation SHALL remain `tab` / `shift+tab`. GRAPH MUST NOT add a
dedicated `g` key in v0.

#### Scenario: Cycle

- GIVEN the TUI is on EVENTS
- WHEN the operator presses `tab`
- THEN the active tab is GRAPH
- AND another `tab` selects CONFIG

### Requirement: List, counts, and detail

GRAPH SHALL show node/edge counts, a kind+id list sorted by G-61 kind
order then id, and a detail pane for the selected node (attrs +
incident edges). Width `< 80` MUST use a vertical list without a
horizontal edge diagram. There SHALL be no 2D canvas in v0.

#### Scenario: Narrow terminal

- GIVEN width 70
- WHEN GRAPH renders
- THEN the view does not panic
- AND node kind and id remain visible as text

### Requirement: Explicit refresh

`r` on GRAPH SHALL call `RefreshGraph` then `GetGraphSnapshot`.
Live event subscribe MUST NOT be required to update GRAPH in v0.

#### Scenario: Refresh key

- GIVEN GRAPH is the active tab
- WHEN the operator presses `r`
- THEN Core `RefreshGraph` is invoked
- AND the list is replaced by the new snapshot (or an error line)

### Requirement: Substring filter

`/` on GRAPH SHALL filter the node list with a case-insensitive
substring on `kind` and `id`, independent of the EVENTS filter.

#### Scenario: Filter capability

- GIVEN nodes of kinds `player` and `capability`
- WHEN the GRAPH filter is `capability`
- THEN player rows are hidden
