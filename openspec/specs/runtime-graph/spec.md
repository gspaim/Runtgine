# Runtime Graph

Status: comportamento atual (estrutural G-60..G-65 + QueryHits G-68).

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

### Requirement: QueryHits ranked structural search

The Graph Service SHALL expose `QueryHits(ctx, Query) -> Hits` that returns
deduplicated, score-ranked hits from existing nodes/edges without writing
new kinds. Failures degrade to empty Hits.

#### Scenario: Seed path hit
- GIVEN a `path` node already in the graph
- AND Query.SeedPaths contains that path
- WHEN QueryHits runs
- THEN an item with `kind=path`, `reason=seed`, and score ≥ 10 is returned

#### Scenario: Keyword match
- GIVEN a capability node id containing token `review` (len ≥ 3)
- AND Query.Text includes `review`
- WHEN QueryHits runs with Limit ≥ 1
- THEN a `keyword` reason hit for that capability MAY appear with score 2

#### Scenario: Best-effort errors
- GIVEN the store returns an error during QueryHits
- WHEN the caller is Runner or Intent
- THEN Hits are empty and execution continues
