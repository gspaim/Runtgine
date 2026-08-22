# NPM Player

Status: comportamento atual (pós-slice 29 / NPM Player v0).

## Requirements

### Requirement: NPM Player is a deterministic npm test runner

The Registry SHALL expose Player `npm` with capability `npm.test`
only. The Player MUST invoke the `npm` binary with an argv allowlist
(package `internal/players/npm`) and MUST NOT invoke a shell string.
The argv MUST be `npm test` or `npm --silent test`.

#### Scenario: Unknown capability rejected

- GIVEN a Task step `npm.install`
- WHEN the Validator runs
- THEN admission fails because the capability is unregistered

#### Scenario: Preview does not use shell

- GIVEN NL `npm test`
- WHEN `CompileIntent` runs
- THEN method is `heuristic.npm`
- AND the Task has one `npm.test` step
- AND it MUST NOT be `shell.exec`

### Requirement: Workdir stays in the workspace

`workdir` MUST be a relative workspace path (default `.`) that
contains `package.json`. Absolute paths, URL-shaped values, and `..`
escapes MUST be rejected at `ValidateStaticInput`.

#### Scenario: Escape rejected

- GIVEN `workdir` `../outside`
- WHEN `npm.test` is validated
- THEN the error is `validation.invalid_input`

#### Scenario: Missing package.json

- GIVEN a resolved workdir with no `package.json`
- WHEN `npm.test` is validated
- THEN the error is `validation.invalid_input`

### Requirement: Failing tests fail the Run

A non-zero `npm test` exit MUST surface as `runtime.player_error` with
structured details (`ok=false`, `exit_code`, truncated `log`).
The Player MUST NOT return step success with `ok=false`.
Unit tests MUST inject the runner and MUST NOT require Node or network.

#### Scenario: Fake failing package

- GIVEN an injected runner that exits 1
- WHEN `npm.test` executes
- THEN the error is `runtime.player_error`
- AND details include `log` and `exit_code`
