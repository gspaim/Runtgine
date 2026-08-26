# infra-players

## ADDED Requirements

### Requirement: Kubernetes Player is read-only kubectl get

The Registry SHALL expose Player `k8s` with capabilities `k8s.list`
and `k8s.get` only. The Player MUST invoke `kubectl get` with argv
and MUST NOT register `k8s.apply`.

#### Scenario: Apply absent

- **WHEN** a Task step names `k8s.apply`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Flag-like resource rejected

- **WHEN** `resource` is `--raw`
- **THEN** the error is `validation.invalid_input`

### Requirement: Terraform Player validates and plans

Player `terraform` SHALL expose `tf.validate` (allow) and
`tf.plan` (approval-required). Workdir MUST contain `*.tf` or
`*.tf.json`. `tf.apply` MUST NOT be registered.

#### Scenario: Plan requires approval

- **WHEN** the operator inspects the Manifest for `tf.plan`
- **THEN** `execution_policy` is `approval-required`

#### Scenario: Missing tf files

- **WHEN** workdir has no `*.tf` / `*.tf.json`
- **THEN** the error is `validation.invalid_input`

### Requirement: Postgres Player only pings

Player `postgres` SHALL expose `pg.ping` only. Input MUST NOT
accept `sql` or `password`. A non-zero `psql` exit SHALL fail the
step (`runtime.player_error`).

#### Scenario: Extra SQL field rejected

- **WHEN** input includes `"sql": "DROP DATABASE x"`
- **THEN** JSON Schema / static validation rejects the input

### Requirement: Intent heuristics beat shell

NL `terraform validate`, `kubectl get pods`, and `pg ping` SHALL
compile to the matching capability with methods `heuristic.tf`,
`heuristic.k8s`, and `heuristic.pg`.

#### Scenario: Terraform preview

- **WHEN** the operator submits NL `terraform validate`
- **THEN** Intent emits `tf.validate`
- **AND** the Task is not `shell.exec`
