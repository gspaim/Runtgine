# ContextPack

Status: comportamento atual (pós-slice 7 / Graph Hits).

## Requirements

### Requirement: AssembleContext builds intra-run pack

The Core SHALL assemble a ContextPack before LLM Player `Complete` with
task view, step view, prior step outputs, repo hits, graph hits, and budget.

#### Scenario: Pack fields present
- GIVEN a run with at least one prior step output
- WHEN AssembleContext runs for the next LLM step
- THEN the pack includes `task`, `step`, `prior_outputs`, `repo_hits`, `graph_hits`, and `budget`
- AND `repo_hits` are derived only from `pipeline.repo-search` outputs in this run

### Requirement: Budget truncates deterministically

The Core SHALL truncate `prior_outputs` and cap `repo_hits` paths/symbols
using fixed defaults (`max_chars`, `max_files`) without non-determinism.

#### Scenario: Cap files
- GIVEN repo-search returned more paths than `budget.max_files`
- WHEN the pack is assembled
- THEN at most `max_files` paths appear in `repo_hits.paths`

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

When applying graph hit caps, the Core SHALL drop lowest-score `graph_hits`
first while leaving `repo_hits` and task/step identity intact.

#### Scenario: Drop lowest scores first
- GIVEN more graph hits than `graph_max_hits`
- WHEN the pack is assembled
- THEN only the highest-score items remain (ties broken by kind, id)
