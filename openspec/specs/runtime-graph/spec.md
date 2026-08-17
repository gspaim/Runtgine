# Runtime Graph

Status: comportamento atual (Runtime Graph v0 estrutural, `docs/18`, G-60..G-65).

## Requirements

### Requirement: One graph per workspace in SQLite

The Core SHALL persist graph nodes and edges in the workspace SQLite database
(`.runtgine/runtgine.db`) with kinds defined in G-61/G-62.

#### Scenario: Boot refresh
- GIVEN registered Players
- WHEN `api.Open` completes
- THEN player and capability nodes exist with `provides` edges

### Requirement: Sync is best-effort

Graph sync failures MUST NOT fail the Run; the Core SHALL log and continue.

#### Scenario: Sync after terminal run
- GIVEN a run that reaches succeeded/failed/cancelled
- WHEN SyncFromRun runs
- THEN run/task/executed/instance_of edges are upserted when sync succeeds
- AND a sync error leaves the run terminal status unchanged

### Requirement: CLI snapshot is read-only

The CLI SHALL expose `runtgine graph snapshot` and `graph refresh` without a
TUI GRAPH tab in v0.

#### Scenario: Snapshot JSON
- GIVEN a non-empty graph
- WHEN `runtgine graph snapshot` runs
- THEN stable JSON of nodes and edges is printed without secrets

### Requirement: No ContextPack integration in structural v0

Structural Graph v0 SHALL NOT inject hits into ContextPack or Intent; that
behavior is owned by change `019-graph-hits`.
