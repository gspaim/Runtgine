# helm-player

## ADDED Requirements

### Requirement: Helm Player is read-only lint, template, list, status

The Registry SHALL expose Player `helm` with capabilities `helm.lint`,
`helm.template`, `helm.list`, and `helm.status` only. The Player MUST
invoke the `helm` binary with argv and MUST NOT register
`helm.install` or `helm.upgrade`.

#### Scenario: Install absent

- **WHEN** a Task step names `helm.install`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Flag-like release rejected

- **WHEN** `release` is `--kubeconfig`
- **THEN** the error is `validation.invalid_input`

### Requirement: Chart resolves inside the workspace

`chart` SHALL be a workspace-relative path resolved with
`shell.ResolveWorkdir` and MUST contain a `Chart.yaml` marker.

#### Scenario: Chart without marker

- **WHEN** `chart` points to a directory without `Chart.yaml`
- **THEN** the error is `validation.invalid_input`

#### Scenario: Chart escaping the workspace

- **WHEN** `chart` resolves outside the workspace root
- **THEN** the error is `validation.invalid_input`

### Requirement: No values injection in v0

Input MUST NOT accept `values`, `set`, `repo`, or `kubeconfig`
fields, and argv MUST NOT include `--set*`, `--values`, `--repo`, or
`--kubeconfig`.

#### Scenario: Extra values field rejected

- **WHEN** input includes `"set": "image.tag=latest"`
- **THEN** JSON Schema / static validation rejects the input

### Requirement: List emits JSON

`helm.list` SHALL always append `-o json` and return parsed
`releases` (raw string + `truncated` when parsing fails).

#### Scenario: List output

- **WHEN** `helm.list` runs
- **THEN** argv ends with `-o json`
- **AND** output contains `releases`

### Requirement: Intent heuristics beat shell

NL `helm lint charts/demo`, `helm template charts/demo`,
`helm list`, and `helm status web` SHALL compile to the matching
capability with method `heuristic.helm`.

#### Scenario: Template preview

- **WHEN** the operator submits NL `helm template charts/demo`
- **THEN** Intent emits `helm.template`
- **AND** the Task is not `shell.exec`

#### Scenario: Install does not match

- **WHEN** the operator submits NL `helm install web charts/demo`
- **THEN** no Player capability matches (falls through to shell/LLM)
