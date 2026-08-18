# Blast Radius

Status: comportamento atual (pós-slice 13 / Blast Radius v0).

## Requirements

### Requirement: Blast Radius is Core analysis, not a Player

The Core SHALL compute an Impact Report from a validated Task IR
without calling a Player, evaluating Execution Policy, acquiring a
claim, or creating a Run. Blast SHALL NOT be a Player or a Run status.

#### Scenario: Shell is empty blast

- GIVEN `examples/hello.json` (`shell.exec` only)
- WHEN `BlastTask` runs
- THEN `risk` is `none`
- AND `predicted_claims` and `touches` are empty
- AND no Run is created

### Requirement: Predicted claims reuse Resource Claims

Predicted claims SHALL be derived only by `claim.Required` (G-95).
Blast MUST NOT introduce a second lock table or Manifest `blast[]`.

#### Scenario: git.add report vs lock

- GIVEN a Task step `git.add` with `paths: ["README"]`
- WHEN the report is computed
- THEN `touches` include `path` `README` with `mode` `write`
- AND `predicted_claims` contain a single `workspace` claim
- AND `risk` is `workspace`

### Requirement: Touches include reads

`fs.read`, `fs.list`, and `fs.stat` SHALL emit a `path` (or
`workspace` if `.`) touch with `mode` `read` and MUST NOT emit a
predicted claim.

#### Scenario: Read only

- GIVEN a Task whose only step is `fs.read` on `a.txt`
- WHEN `BlastTask` runs
- THEN `touches` contain `path` `a.txt` `mode` `read`
- AND `predicted_claims` is empty
- AND `risk` is `none`

### Requirement: Live overlay is read-only

When a Store is open, `BlastTask` SHALL list `conflicts` for predicted
claims that overlap active claims of another `run_id`. It MUST NOT
Acquire or Release.

#### Scenario: Conflict overlay

- GIVEN an active claim on `path` `notes.md`
- WHEN `BlastTask` runs for `fs.write` on `notes.md`
- THEN `conflicts` includes that `holder_run_id`
- AND no new row is inserted into `resource_claims`

### Requirement: CLI surface only

The CLI SHALL expose `runtgine blast`. The Runner MUST NOT auto-blast
on `runtgine run`. The TUI MUST NOT add a tab or key for Blast or GRAPH
in this change.

#### Scenario: run unchanged

- GIVEN `examples/hello.json`
- WHEN `runtgine run` executes
- THEN no blast report is required for `run.succeeded`
