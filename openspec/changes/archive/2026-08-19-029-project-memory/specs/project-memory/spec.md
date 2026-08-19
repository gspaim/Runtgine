# Delta for project-memory

## ADDED Requirements

### Requirement: Project Memory is a Core Provider, not a Player

The Core SHALL persist episodic Project Memory in the workspace SQLite
database via package `internal/core/memory`. The Registry SHALL NOT
register capabilities `memory.*`. Memory MUST NOT grant capabilities,
change Policy, or bypass the Validator.

#### Scenario: Unknown memory capability

- GIVEN a Task step `memory.query`
- WHEN the Validator runs
- THEN admission fails because the capability is unregistered

### Requirement: Episodes have explicit validity

An episode MUST have `kind` in `decision|failure|handoff|preference`
and `validity` in `active|superseded|archived`. `Query` MUST return
only `active` episodes. Supersession MUST be an explicit API call.

#### Scenario: Supersede hides the old episode

- GIVEN an active decision episode A
- WHEN `Supersede(A, B)` runs
- THEN A has `validity=superseded` and `successor_id=B.id`
- AND `Query` returns B and not A

### Requirement: ContextPack includes memory_hits

`AssembleContext` for LLM steps SHALL attach `memory_hits` with budget
`memory_max_hits` (default 8) and `memory_max_chars` (default 2000).
Provider errors SHALL yield `items: []` without failing the Run.
Intent heuristics `shell|pipeline` MUST NOT query Memory.

#### Scenario: Empty memory degrades

- GIVEN a workspace with no episodes
- WHEN AssembleContext runs for an LLM step
- THEN `memory_hits.items` is an empty array
- AND the Run continues
