# ContextPack

Status: comportamento atual (pós-slices 1–6). Extensão `graph_hits` em
`openspec/changes/019-graph-hits/`.

## Requirements

### Requirement: AssembleContext builds intra-run pack

The Core SHALL assemble a ContextPack before LLM Player `Complete` with
task view, step view, prior step outputs, repo hits, and budget.

#### Scenario: Pack fields present
- GIVEN a run with at least one prior step output
- WHEN AssembleContext runs for the next LLM step
- THEN the pack includes `task`, `step`, `prior_outputs`, `repo_hits`, and `budget`
- AND `repo_hits` are derived only from `pipeline.repo-search` outputs in this run

### Requirement: Budget truncates deterministically

The Core SHALL truncate `prior_outputs` and cap `repo_hits` paths/symbols
using fixed defaults (`max_chars`, `max_files`) without non-determinism.

#### Scenario: Cap files
- GIVEN repo-search returned more paths than `budget.max_files`
- WHEN the pack is assembled
- THEN at most `max_files` paths appear in `repo_hits.paths`

### Requirement: Graph hits absent until Graph Hits slice

The ContextPack SHALL NOT include `graph_hits` until change `019-graph-hits`
is implemented and archived.

#### Scenario: Current pack JSON
- GIVEN the shipped binary before slice 7
- WHEN a ContextPack is marshaled
- THEN no `graph_hits` field is required by LLM Players
