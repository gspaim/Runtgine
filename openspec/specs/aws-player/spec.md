# aws-player

## ADDED Requirements

### Requirement: AWS Player is read-only identity and listing

The Registry SHALL expose Player `aws` with capabilities
`aws.sts-identity`, `aws.s3-buckets`, and `aws.s3-objects` only. The
Player MUST invoke the `aws` binary with argv ending in
`--output json` and MUST NOT register any mutating capability
(`s3 cp|mv|rm|sync|mb|rb`, `create-*`, `put-*`, `delete-*`).

#### Scenario: Mutant absent

- **WHEN** a Task step names `aws.s3-cp`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Flag-like bucket rejected

- **WHEN** `bucket` is `--endpoint-url`
- **THEN** the error is `validation.invalid_input`

### Requirement: Output is always JSON

Every capability SHALL append `--no-cli-pager --output json` and
return parsed JSON in `object` (raw string plus `truncated: true`
when parsing fails).

#### Scenario: Parse fallback

- **WHEN** the fake runner returns `not json` with exit 0
- **THEN** output contains the raw string in `object`
- **AND** `truncated` is `true`

### Requirement: Credentials never enter the Task IR

Input MUST NOT accept `access_key`, `secret`, `session`, `token`,
`profile`, `endpoint`, or `query` fields. Credentials SHALL come from
the inherited process environment only.

#### Scenario: Token field rejected

- **WHEN** input includes `"token": "AKIA..."`
- **THEN** JSON Schema / static validation rejects the input

### Requirement: Objects listing requires bucket

`aws.s3-objects` SHALL require `bucket` (`safeRef`) and MAY accept
`prefix` (`safeRef`).

#### Scenario: Missing bucket

- **WHEN** `aws.s3-objects` runs without `bucket`
- **THEN** the error is `validation.invalid_input`

### Requirement: Intent heuristics beat shell

NL `aws sts get-caller-identity`, `aws s3 ls`, and
`aws s3 ls s3://bucket/prefix` SHALL compile to the matching
capability with method `heuristic.aws`.

#### Scenario: S3 listing preview

- **WHEN** the operator submits NL `aws s3 ls`
- **THEN** Intent emits `aws.s3-buckets`
- **AND** the Task is not `shell.exec`

#### Scenario: URI parsed statically

- **WHEN** the operator submits NL `aws s3 ls s3://data/logs`
- **THEN** Intent emits `aws.s3-objects`
- **AND** input has `bucket` `data` and `prefix` `logs`

#### Scenario: Mutant does not match

- **WHEN** the operator submits NL `aws s3 rm s3://data/x`
- **THEN** no Player capability matches (falls through to shell/LLM)
