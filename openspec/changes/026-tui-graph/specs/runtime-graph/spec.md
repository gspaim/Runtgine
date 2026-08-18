# Delta for runtime-graph

## MODIFIED Requirements

### Requirement: CLI snapshot is read-only

The CLI SHALL expose `runtgine graph snapshot` and `graph refresh`.
The TUI MAY expose the same snapshot on tab GRAPH (spec `26`) without
changing CLI behavior or Graph persistence.

#### Scenario: Snapshot JSON

- GIVEN a non-empty graph
- WHEN `runtgine graph snapshot` runs
- THEN stable JSON of nodes and edges is printed without secrets

#### Scenario: TUI GRAPH tab

- GIVEN the TUI is running against the same workspace
- WHEN the operator opens GRAPH
- THEN the view is derived from `GetGraphSnapshot`
- AND the CLI snapshot command still works
