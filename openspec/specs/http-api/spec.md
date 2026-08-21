# HTTP API

Status: comportamento atual (pós-slices 25–26 / `runtgine serve` + webhooks outbound).

## Requirements

### Requirement: HTTP Entry Point adapter

The system SHALL expose the existing Core API over HTTP as an Entry
Point adapter. The server MUST NOT call Players directly and MUST NOT
be the HTTP Player (`http.get` / `http.head`).

#### Scenario: Submit Task IR

- **WHEN** a client `POST /v0/tasks` with a valid Task IR JSON and a
  valid Bearer token
- **THEN** the Core accepts the Task via `SubmitTask`
- **AND** the response is `202` with `run_id`
- **AND** `source.entry_point` is `http` unless already set

#### Scenario: Validator remains sovereign

- **WHEN** the Task IR names a capability not in the Registry
- **THEN** the API returns `400` with a validation error
- **AND** no Run is inserted

### Requirement: Bind and token

The server SHALL default to loopback listen `127.0.0.1:7420` and SHALL
require a Bearer token on every route except `GET /v0/healthz`.

#### Scenario: Healthz without token

- **WHEN** a client `GET /v0/healthz`
- **THEN** the response is `200` without Authorization

#### Scenario: Protected route without token

- **WHEN** a client `POST /v0/tasks` without a valid Bearer token
- **THEN** the response is `401`

#### Scenario: Non-loopback requires token

- **WHEN** listen is not loopback and `RUNTGINE_API_TOKEN` is empty
- **THEN** `runtgine serve` refuses to start

### Requirement: Intent preview does not submit

`POST /v0/intent/preview` SHALL compile NL or validate Task IR without
creating a Run.

#### Scenario: Preview git status

- **WHEN** a client posts NL `git status` to `/v0/intent/preview`
- **THEN** the response is `200` with Task IR using `git.status`
- **AND** no Run exists for that request

### Requirement: Outbound terminal webhooks

The system SHALL optionally POST terminal Run events to configured
HTTPS URLs without failing the Run on delivery errors.

#### Scenario: Failed run notifies webhook

- **WHEN** `webhooks` includes `run.failed` and a Run fails
- **THEN** the dispatcher POSTs the Event envelope to the URL

#### Scenario: Delivery failure is best-effort

- **WHEN** the webhook destination returns 5xx or times out
- **THEN** the Run terminal state is unchanged
- **AND** the failure is logged

### Requirement: Not GitHub inbound and not HTTP Player

The HTTP API SHALL NOT replace Board polling and SHALL NOT add
`http.post` Player capabilities.

#### Scenario: Board transport unchanged

- **WHEN** the HTTP API is enabled
- **THEN** the Board adapter still polls GitHub Projects (G-20)
