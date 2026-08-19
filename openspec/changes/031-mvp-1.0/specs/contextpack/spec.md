# Delta for contextpack

## ADDED Requirements

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
