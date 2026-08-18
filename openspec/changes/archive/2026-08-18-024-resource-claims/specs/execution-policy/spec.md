# Delta for execution-policy

## MODIFIED Requirements

### Requirement: Execution Policy is Core, not a Player

The Core SHALL evaluate an execution verb per registered capability
before Player execution. HITL SHALL be Core API (`ApproveRun`), not a
Human Player. When Resource Claims v0 is enabled, the Runner SHALL
apply the verb **before** claim acquire: `deny` never claims;
`approval-required` pauses without holding a claim; grant then acquire
then Execute.

#### Scenario: Default allow

- GIVEN no `execution_policy` in config or Manifest
- WHEN a Task with `shell.exec` is submitted
- THEN the step executes as today (`allow`)

#### Scenario: Deny never claims

- GIVEN `fs.write` is `deny` in config
- WHEN a Task with `fs.write` is submitted
- THEN admission fails with `policy.denied`
- AND no resource claim is acquired
