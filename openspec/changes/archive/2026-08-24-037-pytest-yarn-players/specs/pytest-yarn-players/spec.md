# pytest-yarn-players

## ADDED Requirements

### Requirement: Pytest Player is a deterministic pytest runner

The Registry SHALL expose Player `pytest` with capability
`pytest.run` only. The Player MUST invoke the `pytest` binary with
an argv allowlist (package `internal/players/pytst`) and MUST NOT
invoke a shell string. The argv MUST include `pytest` plus a
subset of `-q`, `-x`, `-k`, `-m`, `--tb=short`, `--no-header`,
`--color=no`, and package paths.

#### Scenario: Unknown capability rejected

- **WHEN** a Task step names `pytest.coverage`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Preview does not use shell

- **WHEN** the operator submits NL `pytest tests/foo.py`
- **THEN** Intent emits `pytest.run` with method `heuristic.pytest`
- **AND** the Task is not `shell.exec`

### Requirement: Workdir stays in the workspace

`workdir` MUST be a relative workspace path (default `.`) that
contains `pyproject.toml`, `pytest.ini`, or `tests/`. Absolute
paths, URL-shaped values, and `..` escapes MUST be rejected at
`ValidateStaticInput`.

#### Scenario: Escape rejected

- **WHEN** `workdir` is `../outside`
- **THEN** the error is `validation.invalid_input`

#### Scenario: Missing marker

- **WHEN** the resolved workdir has none of `pyproject.toml`,
  `pytest.ini`, `tests/`
- **THEN** the error is `validation.invalid_input`

### Requirement: Failing tests fail the Run

A non-zero `pytest` exit code SHALL fail the step and the Run
(`runtime.player_error`). Unit tests MUST inject the runner and
MUST NOT require Python or network.

#### Scenario: Fake fail

- **WHEN** the injected runner returns exit code 1
- **THEN** Execute returns `runtime.player_error`
- **AND** the payload still includes `log` and `exit_code`

### Requirement: Yarn Player is a deterministic yarn test runner

The Registry SHALL expose Player `yarn` with capability `yarn.test`
only. The Player MUST invoke the `yarn` binary with argv allowlist
(package `internal/players/jstest`) and MUST NOT invoke a shell
string. The argv MUST be `yarn test`; `--frozen-lockfile`,
`--immutable`, `--parallel`, `add`, `install`, `dlx`, `npx` MUST
be rejected at `ValidateStaticInput`.

#### Scenario: Install rejected

- **WHEN** the input requests `yarn install`
- **THEN** the error is `validation.invalid_input`

#### Scenario: Preview does not use shell

- **WHEN** the operator submits NL `yarn test`
- **THEN** Intent emits `yarn.test` with method `heuristic.yarn`
- **AND** the Task is not `shell.exec`

### Requirement: Yarn workdir requires package.json

`workdir` MUST be a relative workspace path that contains
`package.json`. Absolute paths, URL-shaped values, and `..`
escapes MUST be rejected at `ValidateStaticInput`.

#### Scenario: Missing package.json

- **WHEN** the resolved workdir has no `package.json`
- **THEN** the error is `validation.invalid_input`

### Requirement: Failing yarn tests fail the Run

A non-zero `yarn test` exit code SHALL fail the step and the Run
(`runtime.player_error`). Unit tests MUST inject the runner and
MUST NOT require Node, Yarn, or network.

#### Scenario: Fake fail

- **WHEN** the injected runner returns exit code 1
- **THEN** Execute returns `runtime.player_error`
- **AND** the payload still includes `log` and `exit_code`
