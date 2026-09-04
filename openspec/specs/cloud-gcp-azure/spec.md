# cloud-gcp-azure

## ADDED Requirements

### Requirement: GCP Player is read-only identity, config, projects

The Registry SHALL expose Player `gcp` with capabilities
`gcp.identity`, `gcp.config`, and `gcp.projects` only, invoking the
`gcloud` binary with argv containing `--format=json`. It MUST NOT
register any mutating capability.

#### Scenario: Mutant absent

- **WHEN** a Task step names `gcp.projects-create`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Format flag present

- **WHEN** `gcp.projects` runs
- **THEN** argv contains `--format=json`

### Requirement: Azure Player is read-only account and groups

The Registry SHALL expose Player `azure` with capabilities
`azure.identity`, `azure.subscriptions`, and `azure.groups` only,
invoking the `az` binary with argv ending in `-o json`. It MUST NOT
register any mutating capability.

#### Scenario: Mutant absent

- **WHEN** a Task step names `azure.groups-create`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Output flag present

- **WHEN** `azure.identity` runs
- **THEN** argv ends with `-o`, `json`

### Requirement: No input beyond timeout

Input for every capability MUST accept only `timeout_ms`
(`additionalProperties: false`). Credentials, project, and
subscription SHALL come from the inherited environment/config only.

#### Scenario: Extra project field rejected

- **WHEN** input includes `"project": "my-proj"`
- **THEN** JSON Schema / static validation rejects the input

### Requirement: Intent heuristics beat shell

NL `gcloud auth list`, `gcloud config list`,
`gcloud projects list`, `az account show`, `az account list`, and
`az group list` SHALL compile to the matching capability with
methods `heuristic.gcp` and `heuristic.az`.

#### Scenario: Azure identity preview

- **WHEN** the operator submits NL `az account show`
- **THEN** Intent emits `azure.identity`
- **AND** the Task is not `shell.exec`

#### Scenario: Mutant does not match

- **WHEN** the operator submits NL `gcloud projects create x`
- **THEN** no Player capability matches (falls through to shell/LLM)
