# Intent Engine

Status: comportamento atual (Intent Engine v0, `docs/17`, G-50..G-54).

## Requirements

### Requirement: Intent is not a Player

The Intent Engine SHALL compile natural language to Task IR v0 and MUST NOT
bypass Validator or Registry.

#### Scenario: Submit path
- GIVEN non-empty NL text
- WHEN `SubmitIntent` succeeds
- THEN a Task IR is validated via the same `SubmitTask` path as CLI JSON

### Requirement: Deterministic heuristics first

The Engine SHALL try shell then pipeline heuristics before the LLM Completer.

#### Scenario: Shell heuristic
- GIVEN text `echo hello-intent`
- WHEN `CompileIntent` runs
- THEN method is `heuristic.shell` and the Task has one `shell.exec` step

#### Scenario: Pipeline heuristic
- GIVEN text containing analysis keywords (e.g. `revisar`)
- WHEN `CompileIntent` runs
- THEN method is `heuristic.pipeline` or routes to the pipeline template

### Requirement: LLM path uses ContextPack without graph

Until `019-graph-hits`, the LLM Completer ContextPack SHALL NOT query the
Runtime Graph.

#### Scenario: No graph query on heuristics
- GIVEN text matching shell heuristic
- WHEN compile runs
- THEN the Runtime Graph is not consulted
