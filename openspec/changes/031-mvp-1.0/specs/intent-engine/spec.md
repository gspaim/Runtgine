# Delta for intent-engine

## MODIFIED Requirements

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
