# workflow-templates

## ADDED Requirements

### Requirement: Templates are workspace JSON, not Players

The Core SHALL load Workflow Templates from
`<workspace>/.runtgine/templates/*.json` at boot (best-effort).
Templates MUST NOT be registered as Players or capabilities.
Unknown or invalid files MUST be skipped with a warning; Open
MUST still succeed.

#### Scenario: Invalid file does not fail Open

- **WHEN** the templates directory contains a non-JSON file and a
  valid template
- **THEN** Core opens
- **AND** `template list` includes only the valid template

### Requirement: Native loading closes G-40

Templates MUST be loaded only from the workspace directory. The
Core MUST NOT fetch templates from a git remote, HTTP URL, or
path outside the workspace.

#### Scenario: No remote fetch

- **WHEN** a template JSON names a URL or git ref
- **THEN** that field is rejected at load (`additionalProperties`
  / unknown field) or ignored because no such field exists in v0

### Requirement: Compile emits Task IR

`Compile` SHALL copy template steps into a Task IR v0 document
with `metadata.template` equal to the template id. Admission
MUST still run the Validator: an unknown capability is rejected.

#### Scenario: Unknown capability rejected at submit

- **WHEN** a template step names `k8s.apply`
- **AND** that capability is unregistered
- **THEN** `SubmitTask` rejects the Task (`validation.unknown_capability`
  or equivalent admission error)

### Requirement: Intent heuristic beats shell prefix

NL `run template <id>` SHALL compile with method
`heuristic.template` and MUST NOT become `shell.exec`. An unknown
id after a recognized prefix SHALL be `validation.invalid_input`.

#### Scenario: Preview uses template

- **WHEN** the operator submits NL `run template verify` and
  `verify` is loaded
- **THEN** Intent emits a multi-step Task with `metadata.template=verify`
- **AND** the method is `heuristic.template`

#### Scenario: Unknown id does not fall through to shell

- **WHEN** the operator submits NL `run template missing`
- **THEN** compile fails with `validation.invalid_input`
- **AND** the Task is not `shell.exec`

### Requirement: Graph records template nodes

Boot SHALL upsert a Graph node `kind=template` per loaded
template id (best-effort). Failure MUST NOT fail Open.

#### Scenario: Snapshot includes template

- **WHEN** a valid template `verify` is loaded
- **THEN** `GetGraphSnapshot` includes a node `{kind: template, id: verify}`
