# Delta for test-player

## ADDED Requirements

### Requirement: Test Player is a deterministic go test runner

The Registry SHALL expose Player `test` with capability `test.go`
only. The Player MUST invoke the `go` binary with an argv allowlist
(package `internal/players/gotest`) and MUST NOT invoke a shell
string. The process MUST include `-json` and `-mod=readonly`.

#### Scenario: Unknown runner rejected

- GIVEN a Task step `test.python`
- WHEN the Validator runs
- THEN admission fails because the capability is unregistered

### Requirement: Packages stay in the workspace

`packages` MUST be relative workspace paths (default `./...`).
Absolute paths, URL-shaped values, and `..` escapes MUST be rejected
at `ValidateStaticInput`.

#### Scenario: Escape rejected

- GIVEN `packages` `["../outside"]`
- WHEN `test.go` is validated
- THEN the error is `validation.invalid_input`

### Requirement: Failing tests fail the Run

A non-zero `go test` exit MUST surface as `runtime.player_error` with
structured details (`ok=false`, pass/fail/skip, truncated `log`).
The Player MUST NOT return step success with `ok=false`.

#### Scenario: Fake failing package

- GIVEN an injected runner that exits 1 with one fail event
- WHEN `test.go` executes
- THEN the error is `runtime.player_error`
- AND details include `fail >= 1`
