# HTTP Player

Status: comportamento atual (pós-slice 16 / HTTP Player v0).

## Requirements

### Requirement: HTTP Player is a deterministic HTTPS client

The Registry SHALL expose Player `http` with capabilities `http.get`
and `http.head` only. The Player MUST use Go `net/http` (package
`internal/players/httpclient`) and MUST NOT invoke `curl` or Shell.

#### Scenario: Unknown method rejected

- GIVEN a Task step `http.post`
- WHEN the Validator runs
- THEN admission fails because the capability is unregistered

### Requirement: HTTPS and header allowlist

`url` MUST use scheme `https` without userinfo. Request `headers` MAY
only include `Accept`, `Accept-Language`, and `User-Agent`.
`Authorization`, `Cookie`, and `Host` MUST be rejected at
`ValidateStaticInput`.

#### Scenario: Cleartext URL

- GIVEN `url` `http://example.com/`
- WHEN `http.get` is validated
- THEN the error is `validation.invalid_input`

#### Scenario: Auth header

- GIVEN request header `Authorization`
- WHEN `http.get` is validated
- THEN the error is `validation.invalid_input`

### Requirement: Read-only body limits

`http.get` SHALL return UTF-8 `body` truncated at `max_bytes` (default
1 MiB, max 4 MiB). Non-UTF-8 bodies SHALL set `binary=true` and empty
`body`. `http.head` MUST NOT return `body`. Redirects are capped at 5
and MUST remain `https`. Link-local and cloud-metadata destinations
MUST fail without returning a body.

#### Scenario: Offline fake GET

- GIVEN a RoundTripper that returns 200 `{"ok":true}` as UTF-8 JSON
- WHEN `http.get` executes
- THEN `status` is 200
- AND `body` contains the JSON
- AND no public network dial occurs
