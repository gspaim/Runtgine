# Delta for docker-player

## ADDED Requirements

### Requirement: Docker Player serves local docker capabilities

The Core SHALL register a deterministic Player named `docker` that serves
exactly `docker.ps`, `docker.inspect`, `docker.logs`, `docker.run`, and
`docker.build`.

#### Scenario: ps succeeds with stub or daemon

- GIVEN the Docker Player is registered
- WHEN a Task step with `docker.ps` runs
- THEN the step succeeds with JSON `containers` (possibly empty)

#### Scenario: Unknown docker capability rejected

- GIVEN a Task step with capability `docker.push`
- WHEN validation / registry resolution runs
- THEN the Task is rejected (capability not in Manifest)

### Requirement: run and build require HITL by default

The Manifest SHALL set `execution_policy` to `approval-required` for
`docker.run` and `docker.build`. The Player MUST NOT be invoked until
`ApproveRun(grant)`.

#### Scenario: run waits for approval

- GIVEN a Task with `docker.run`
- WHEN the Runner reaches that step
- THEN the Run is `waiting_approval`
- AND no `docker` argv has been started

### Requirement: run sandbox argv

`docker.run` SHALL pass `--pull=never`, `--network=none`, and `--rm`.
It MUST reject privileged, publish, host network, and arbitrary volumes.
`mount_workspace=true` MAY add a single read-only bind of the workspace
root.

#### Scenario: default run has no bind mount

- GIVEN `mount_workspace` omitted
- WHEN `docker.run` executes after grant
- THEN argv does not contain `-v` / `--volume`

### Requirement: build context confined to workspace

`docker.build` context and dockerfile paths SHALL resolve inside the
workspace after symlink evaluation, matching Filesystem confinement.

#### Scenario: escaping context rejected

- GIVEN `context: "../outside"`
- WHEN `docker.build` is validated
- THEN the Core returns a validation error before invoking Docker
