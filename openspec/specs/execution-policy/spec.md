# Execution Policy

Status: comportamento atual (pós-slice 10 / Execution Policy + HITL v0).

## Requirements

### Requirement: Execution Policy is Core, not a Player

The Core SHALL evaluate an execution verb per registered capability
before Player execution. HITL SHALL be Core API (`ApproveRun`), not a
Human Player.

#### Scenario: Default allow

- GIVEN no `execution_policy` in config or Manifest
- WHEN a Task with `shell.exec` is submitted
- THEN the step executes as today (`allow`)

### Requirement: Three verbs only

The Core SHALL support exactly `allow`, `deny`, and `approval-required`
on exact capability names. Unknown verbs in config MUST fail closed at
load time.

#### Scenario: Deny at admission

- GIVEN `shell.exec` is `deny` in config
- WHEN a Task with that capability is submitted
- THEN the Core emits `task.rejected` with `policy.denied`
- AND the Player MUST NOT execute

### Requirement: Approval pauses the Run

When the effective verb is `approval-required`, the Runner SHALL persist
status `waiting_approval` and emit `run.waiting_approval` before
`Execute`. `ApproveRun(grant)` SHALL resume once; `ApproveRun(deny)`
SHALL fail the Run with `policy.approval_denied` without calling the
Player.

#### Scenario: Grant then execute

- GIVEN a step whose capability is `approval-required`
- WHEN the Run reaches that step
- THEN status is `waiting_approval`
- AND after `runtgine approve <run_id>` the Player executes

#### Scenario: Human deny

- GIVEN a Run in `waiting_approval`
- WHEN `ApproveRun(deny)` runs
- THEN the Run is `failed` with `policy.approval_denied`
- AND the Player does not execute

### Requirement: Precedence

Effective verb SHALL be: global default `allow`, then Manifest
`execution_policy` if set, then `config.json` capability map, then
`RUNTGINE_POLICY_DEFAULT` for the global default only.

#### Scenario: Config overrides Manifest

- GIVEN Manifest marks `docker.run` as `approval-required`
- AND config sets `docker.run` to `deny`
- WHEN a Task with `docker.run` is submitted
- THEN admission fails with `policy.denied`

### Requirement: CLI and TUI are Entry Points

The CLI SHALL expose `runtgine approve` and `runtgine deny`. The TUI
RUNS/LIVE tabs SHALL show `waiting_approval` and map `a`/`d` to
`ApproveRun` without adding a tab or calling a Player.
