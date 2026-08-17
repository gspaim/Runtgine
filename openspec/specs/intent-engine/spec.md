# Intent Engine

Status: comportamento atual (Intent Engine v0 + Graph Hits G-69).

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

### Requirement: LLM compile path consumes QueryHits

When the Intent Engine uses the LLM Completer path, it SHALL call
`QueryHits` with the NL text and attach results as `graph_hits` on the
Completer ContextPack. Heuristic shell and pipeline paths MUST NOT call
QueryHits.

#### Scenario: LLM path queries graph
- GIVEN NL text that does not match shell or pipeline heuristics
- AND the Completer is invoked
- WHEN compile runs
- THEN QueryHits is called with Text set to the NL input

#### Scenario: Heuristic shell skips graph
- GIVEN text `echo hello-intent`
- WHEN CompileIntent runs
- THEN method is `heuristic.shell`
- AND QueryHits is not invoked

#### Scenario: Graph failure still compiles
- GIVEN QueryHits returns an empty result due to store error
- WHEN compile LLM runs
- THEN a Task IR is still produced when the Completer succeeds
