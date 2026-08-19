# Intent Engine

Status: comportamento atual (Intent Engine v0 + player heuristics G-135 + Graph Hits G-69).

## Requirements

### Requirement: Intent is not a Player

The Intent Engine SHALL compile natural language to Task IR v0 and MUST NOT
bypass Validator or Registry.

#### Scenario: Submit path
- GIVEN non-empty NL text
- WHEN `SubmitIntent` succeeds
- THEN a Task IR is validated via the same `SubmitTask` path as CLI JSON

### Requirement: Deterministic heuristics first

The Engine SHALL try **player** heuristics, then shell, then pipeline,
before the LLM Completer. Player phrases in the MVP 1.0 table
(`test.go`, `git.status|diff|log`, `docker.ps`) MUST win over generic
shell prefixes (`go `, argv `git`).

#### Scenario: go test is not shell

- GIVEN text `go test`
- WHEN `CompileIntent` runs
- THEN method is `heuristic.test`
- AND the Task has one `test.go` step
- AND it MUST NOT be `shell.exec`

#### Scenario: git status

- GIVEN text `git status`
- WHEN `CompileIntent` runs
- THEN method is `heuristic.git`
- AND the capability is `git.status`

#### Scenario: Shell heuristic still works

- GIVEN text `echo hello-intent`
- WHEN `CompileIntent` runs
- THEN method is `heuristic.shell`
- AND the Task has one `shell.exec` step

#### Scenario: Pipeline heuristic
- GIVEN text containing analysis keywords (e.g. `revisar`)
- WHEN `CompileIntent` runs
- THEN method is `heuristic.pipeline` or routes to the pipeline template

### Requirement: LLM compile path consumes QueryHits

When the Intent Engine uses the LLM Completer path, it SHALL call
`QueryHits` with the NL text and attach results as `graph_hits` on the
Completer ContextPack. When `repo_hits` is empty it SHALL seed paths
and symbols from those hits. Heuristic shell, pipeline, and player
paths MUST NOT call QueryHits.

#### Scenario: LLM path queries graph
- GIVEN NL text that does not match shell, pipeline, or player heuristics
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
