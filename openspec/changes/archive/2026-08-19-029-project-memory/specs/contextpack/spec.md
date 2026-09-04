# Delta for contextpack

## MODIFIED Requirements

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
