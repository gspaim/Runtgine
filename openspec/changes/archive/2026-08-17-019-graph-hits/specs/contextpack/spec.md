# Delta for ContextPack

## ADDED Requirements

### Requirement: ContextPack includes graph_hits

The ContextPack SHALL include a `graph_hits` object with an `items` array of
structural hits `{kind, id, reason, score}` and budget fields
`graph_max_hits` (default 20) and `graph_max_chars` (default 4000).

#### Scenario: Empty graph degrades
- GIVEN a workspace graph with no matching nodes
- WHEN AssembleContext runs for an LLM step
- THEN `graph_hits.items` is empty
- AND the Run continues normally

#### Scenario: Hit kinds
- GIVEN QueryHits returns path and capability hits
- WHEN the pack is marshaled
- THEN each item has `kind` in `path|symbol|capability|run|task`
- AND `reason` is one of `seed|mentions|executed|instance_of|child_of|keyword`

### Requirement: Truncation hierarchy prefers repo_hits over graph_hits

When applying global char budget pressure, the Core SHALL truncate
`graph_hits` before removing `repo_hits` or task/step identity fields.

#### Scenario: Drop lowest scores first
- GIVEN more graph hits than `graph_max_hits`
- WHEN the pack is assembled
- THEN only the highest-score items remain (ties broken by kind, id)

## MODIFIED Requirements

### Requirement: Graph hits absent until Graph Hits slice

(Previously: no `graph_hits` field.)

The ContextPack SHALL include `graph_hits` after change `019-graph-hits` is
implemented. LLM Players MUST tolerate the field; deterministic Players MAY
ignore it.
