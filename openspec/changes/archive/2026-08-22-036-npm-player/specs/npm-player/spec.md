# npm-player

## ADDED Requirements

### Requirement: NPM Player is a deterministic npm test runner

The Registry SHALL expose Player `npm` with capability `npm.test`
only. The Player MUST invoke the `npm` binary with an argv allowlist
(package `internal/players/npm`) and MUST NOT invoke a shell string.
The argv MUST be `npm test` or `npm --silent test`.

#### Scenario: Unknown capability rejected

- **WHEN** a Task step names `npm.install`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Preview does not use shell

- **WHEN** the operator submits NL `npm test`
- **THEN** Intent emits `npm.test` with method `heuristic.npm`
- **AND** the Task is not `shell.exec`

### Requirement: Workdir stays in the workspace

`workdir` MUST be a relative workspace path (default `.`) that
contains `package.json`. Absolute paths, URL-shaped values, and `..`
escapes MUST be rejected at `ValidateStaticInput`.

#### Scenario: Escape rejected

- **WHEN** `workdir` is `../outside`
- **THEN** the error is `validation.invalid_input`

#### Scenario: Missing package.json

- **WHEN** the resolved workdir has no `package.json`
- **THEN** the error is `validation.invalid_input`

### Requirement: Failing tests fail the Run

A non-zero `npm test` exit code SHALL fail the step and the Run
(`runtime.player_error`). Unit tests MUST inject the runner and MUST
NOT require Node or network.

#### Scenario: Fake fail

- **WHEN** the injected runner returns exit code 1
- **THEN** Execute returns `runtime.player_error`
- **AND** the payload still includes `log` and `exit_code`
