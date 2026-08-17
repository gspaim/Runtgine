# Git Player

Status: comportamento atual (pós-slice 8 / Git Player v0).

## Requirements

### Requirement: Git Player serves local git.* capabilities

The Core SHALL register a deterministic Player named `git` that serves
exactly `git.status`, `git.diff`, `git.log`, `git.add`, and `git.commit`.

#### Scenario: Status in a repo
- GIVEN a git workspace with at least one commit
- WHEN a Task step with capability `git.status` runs
- THEN the step succeeds with JSON including `branch` and `clean`

#### Scenario: Unknown git capability rejected
- GIVEN a Task step with capability `git.push`
- WHEN validation / registry resolution runs
- THEN the Task is rejected (capability not in Manifest)

### Requirement: Workspace path confinement

All `workdir` and `paths` inputs SHALL resolve inside the workspace root
after symlink evaluation.

#### Scenario: Escaping path rejected
- GIVEN `paths: ["../outside"]` on `git.add`
- WHEN the step is validated or executed
- THEN the Core returns a validation or player error
- AND no files outside the workspace are staged

### Requirement: Commit without hooks or network

`git.commit` SHALL run with hooks disabled and without network
capabilities; identity MAY be forced via `git -c` for determinism.

#### Scenario: Commit in temp repo
- GIVEN a temp git repo with a staged file
- WHEN `git.commit` runs with a non-empty message
- THEN a commit hash is returned
- AND no git hooks from the repo execute

### Requirement: No remote operations in v0

The Git Player Manifest MUST NOT declare `git.push`, `git.pull`,
`git.fetch`, or `git.clone`.
