# pg-explain

## ADDED Requirements

### Requirement: Explain is planner-only SQL

Player `postgres` SHALL expose `pg.explain` which runs
`EXPLAIN (FORMAT JSON) <statement>` via `psql --command` argv. The
Player MUST construct the `EXPLAIN (FORMAT JSON) ` prefix and MUST
NOT register `pg.query` or `pg.exec`.

#### Scenario: Query capability absent

- **WHEN** a Task step names `pg.query`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Plan output parsed

- **WHEN** the fake runner returns a JSON array with exit 0
- **THEN** output contains `plan` with the parsed array

### Requirement: SQL static allowlist

`sql` SHALL be rejected unless its first word (case-insensitive) is
`SELECT` or `WITH`, it contains no `;`, no `\`, and is at most 10
KiB.

#### Scenario: Multi-statement rejected

- **WHEN** `sql` is `select 1; drop table users`
- **THEN** the error is `validation.invalid_input`

#### Scenario: Backslash rejected

- **WHEN** `sql` contains a backslash
- **THEN** the error is `validation.invalid_input`

#### Scenario: Non-select rejected

- **WHEN** `sql` is `explain analyze select 1`
- **THEN** the error is `validation.invalid_input`

#### Scenario: Select accepted

- **WHEN** `sql` is `select id from users where active = true`
- **THEN** static validation passes

### Requirement: Credentials stay in the environment

Input MUST NOT accept `password` or connection string fields; the
command inherits only the pg allowlist environment from `41`
(`PGPASSWORD`, `PGSSLMODE`).

#### Scenario: Password field rejected

- **WHEN** input includes `"password": "hunter2"`
- **THEN** JSON Schema / static validation rejects the input

### Requirement: Intent explains selects

NL `explain select <query>` and `explain with <query>` SHALL compile
to `pg.explain` with method `heuristic.pg`.

#### Scenario: Explain preview

- **WHEN** the operator submits NL `explain select id from users`
- **THEN** Intent emits `pg.explain`
- **AND** input `sql` is `select id from users`
