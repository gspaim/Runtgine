# Filesystem Player

Status: comportamento atual (pós-slice 9 / Filesystem Player v0).

## Requirements

### Requirement: Filesystem Player serves local fs capabilities

The Core SHALL register a deterministic Player named `filesystem` that
serves exactly `fs.read`, `fs.write`, `fs.list`, and `fs.stat`.

#### Scenario: Read workspace file

- GIVEN a UTF-8 file inside the workspace
- WHEN a Task step with capability `fs.read` runs
- THEN the step succeeds with `path`, `content`, `bytes`, and `truncated`

#### Scenario: Unsupported operation rejected

- GIVEN a Task step with capability `fs.delete`
- WHEN validation / registry resolution runs
- THEN the Task is rejected because the capability is not registered

### Requirement: Filesystem paths are confined to workspace

All filesystem paths SHALL resolve inside the workspace root after symlink
evaluation. Absolute paths and external symlinks MUST be rejected.

#### Scenario: Parent traversal rejected

- GIVEN `path: "../outside.txt"`
- WHEN `fs.read`, `fs.write`, `fs.list`, or `fs.stat` validates input
- THEN the operation returns a validation error before filesystem I/O

#### Scenario: External symlink rejected

- GIVEN a workspace symlink whose target is outside the workspace
- WHEN an operation would traverse that symlink
- THEN the operation returns a validation error

### Requirement: Filesystem operations enforce deterministic limits

`fs.read` SHALL default to a 1 MiB limit and never exceed 4 MiB.
`fs.write` SHALL reject content above 4 MiB. `fs.list` SHALL default to
200 entries and never return more than 1000 entries.

#### Scenario: Read truncation

- GIVEN a UTF-8 file larger than `max_bytes`
- WHEN `fs.read` runs
- THEN content is capped at the requested limit and `truncated` is true

#### Scenario: List truncation

- GIVEN a directory with more than `max_entries` entries
- WHEN `fs.list` runs
- THEN entries are lexicographically ordered and `truncated` is true

### Requirement: Filesystem writes are safe and atomic

`fs.write` SHALL accept UTF-8 content, reject symlink destinations, and
replace files using a temporary file in the same parent followed by rename.
It MUST NOT create parent directories unless `create_parents=true`.

#### Scenario: Atomic write

- GIVEN a writable path inside the workspace
- WHEN `fs.write` succeeds
- THEN the complete new content is present
- AND no partial destination is exposed

#### Scenario: Symlink destination rejected

- GIVEN `path` refers to an existing symlink
- WHEN `fs.write` runs
- THEN the operation fails without changing the symlink target

### Requirement: Filesystem metadata is observable

`fs.stat` SHALL return `type`, `size`, `mode`, and `modified_at` for a
workspace-local file or directory, and SHALL identify symlinks without
following external targets.

#### Scenario: Stat regular file

- GIVEN a regular file inside the workspace
- WHEN `fs.stat` runs
- THEN output contains `type=file` and the file size
