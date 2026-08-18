# Delta for resource-claims

## ADDED Requirements

### Requirement: Resource Claim is Core, not a Player

The Core SHALL acquire exclusive resource claims after Execution Policy
resolution (and after HITL grant when applicable) and before
`Player.Execute`. Claims SHALL NOT be a Player or a new Run status.

#### Scenario: Shell does not claim

- GIVEN a Task whose only mutating step is `shell.exec`
- WHEN two such Tasks are submitted concurrently
- THEN both MAY succeed (no auto-claim)

### Requirement: Two kinds only

The Core SHALL support exactly kinds `workspace` and `path`. Empty or
`.` paths MUST be stored as `workspace`. Path overlap MUST be
segment-aware. There SHALL be no shared read-lock in v0.

#### Scenario: Prefix overlap

- GIVEN Run A holds `path` `src`
- WHEN Run B tries `path` `src/main.go`
- THEN Run B fails with `claim.conflict`
- AND Run A continues

#### Scenario: Workspace vs path

- GIVEN Run A holds `workspace`
- WHEN Run B tries `path` `a.txt`
- THEN Run B fails with `claim.conflict`

### Requirement: Auto-claim table

The Core SHALL derive claims from this table only: `fs.write` → `path`
from input `path`; `git.add` and `git.commit` → `workspace`;
`docker.build` → `path` from `context` (default `.`); `docker.run`
with `mount_workspace=true` → `workspace`. Other capabilities MUST NOT
auto-claim. Manifest/Task IR MUST NOT grow a `claims[]` field in v0.

#### Scenario: docker.run without mount

- GIVEN `docker.run` with `mount_workspace` false or omitted
- WHEN the step is admitted
- THEN no claim is acquired

### Requirement: Fail-fast conflict

On conflict the later Run SHALL fail with error code `claim.conflict`
and MUST NOT call the Player. The Core MUST NOT wait, queue, or enter
`waiting_claim`. Events `claim.acquired`, `claim.conflict`, and
`claim.released` SHALL use the existing envelope. Claims persist in
SQLite and MUST be released when the Run is terminal and swept on boot
if the Run is not `running` or `waiting_approval`.

#### Scenario: Two fs.write same path

- GIVEN Run A is executing `fs.write` on `notes.md`
- WHEN Run B submits `fs.write` on `notes.md`
- THEN Run B is `failed` with `claim.conflict`
- AND Run A is not interrupted

#### Scenario: HITL then claim

- GIVEN a step that is `approval-required` and auto-claims
- WHEN the Run is `waiting_approval`
- THEN no claim is held
- AND after `ApproveRun(grant)` the Core attempts acquire before Execute
- AND a conflict fails the Run without a second HITL prompt
