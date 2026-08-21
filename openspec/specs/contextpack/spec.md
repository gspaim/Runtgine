# ContextPack

Status: comportamento atual (pós-slice 20 / Context Engine v0).

## Requirements

### Requirement: AssembleContext builds intra-run pack

The Core SHALL assemble a ContextPack before LLM Player `Complete` with
task view, step view, prior step outputs, repo hits, graph hits, memory
hits, and budget.

#### Scenario: Pack fields present
- GIVEN a run with at least one prior step output
- WHEN AssembleContext runs for the next LLM step
- THEN the pack includes `task`, `step`, `prior_outputs`, `repo_hits`, `graph_hits`, `memory_hits`, and `budget`
- AND `repo_hits` come from `pipeline.repo-search` in this run when present

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

### Requirement: ContextPack includes memory_hits

The ContextPack SHALL include a `memory_hits` object with an `items`
array of episodic hits `{id, kind, validity, title, snippet, score}`
and budget fields `memory_max_hits` (default 8) and `memory_max_chars`
(default 2000). Items in the pack MUST be `validity=active` only.
`memory_hits` ranks below `graph_hits` in the truncation hierarchy.

#### Scenario: Empty memory degrades

- GIVEN QueryMemory returns an error or no rows
- WHEN AssembleContext runs for an LLM step
- THEN `memory_hits.items` is `[]`
- AND the Run continues

### Requirement: Empty repo_hits are seeded from Graph path/symbol hits

When AssembleContext (or the Intent LLM Completer pack) has empty
`repo_hits` after intra-run extraction, the Core SHALL copy
`QueryHits` items with `kind=path` into `repo_hits.paths` and
`kind=symbol` into `repo_hits.symbols`, capped by `budget.max_files`.
Existing repo-search hits MUST NOT be overwritten. Graph failure MUST
leave `repo_hits` empty and MUST NOT fail the Run.

#### Scenario: Seed from graph

- GIVEN no `pipeline.repo-search` output in this Run
- AND QueryHits returns a path hit `internal/core/intent/intent.go`
- WHEN AssembleContext runs for an LLM step
- THEN `repo_hits.paths` contains that path

#### Scenario: Preserve repo-search

- GIVEN `pipeline.repo-search` already filled `repo_hits`
- WHEN AssembleContext runs
- THEN those paths remain
- AND Graph path hits are not merged in

#### Scenario: Empty graph degrades

- GIVEN QueryHits returns no path/symbol items
- WHEN AssembleContext runs
- THEN `repo_hits.paths` is empty
- AND the Run continues

### Requirement: ContextPack includes playbook_hits

The ContextPack SHALL include a `playbook_hits` object with an `items`
array of `{id, title, snippet}` and budget fields `playbook_max_hits`
(default 2) and `playbook_max_chars` (default 1500). Hits come from
`.runtgine/playbooks/*.md` matched by declared capabilities.
`playbook_hits` ranks below `memory_hits` in the truncation hierarchy.

#### Scenario: Empty playbooks degrade

- GIVEN no playbooks are indexed
- WHEN AssembleContext runs
- THEN `playbook_hits.items` is `[]`
- AND the Run continues
