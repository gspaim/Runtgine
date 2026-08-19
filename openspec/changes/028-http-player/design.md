# Design: 028-http-player

## Technical approach

### Package

`internal/players/httpclient` with `New()`, `Manifest()`,
`ValidateStaticInput(workspace, capability, input)`, `Execute`.

Player name in Manifest is `http`. Import alias in `api`/`runner`:
`httpplayer` (stdlib is `net/http`).

Inject `http.RoundTripper` (production: default transport with custom
`DialContext` for IP filter). Tests never dial the public internet.

### Capability contracts

| Capability | HTTP method | Body |
|---|---|---|
| `http.get` | GET | UTF-8 string, truncated |
| `http.head` | HEAD | omitted |

Redirects: `CheckRedirect` counts hops (max 5) and rejects non-https
Location.

### Static validation

Reject before Execute:

- missing/invalid URL
- scheme ≠ `https`
- userinfo present
- forbidden request header names (case-insensitive)
- `timeout_ms` / `max_bytes` out of range (if the schema does not already)

### Runtime IP filter

In `DialContext`, parse resolved IP. Deny:

- `169.254.0.0/16`
- IPv6 link-local `fe80::/10`
- IPv6 `fd00:ec2::/128` if used as metadata

Allow loopback and RFC1918.

### Blast / Claims

Do not add `http.get` to `claim.Required` or `blast.Touched`.
hello-style GET → `risk: none`, empty touches.

### Tests

- httptest-style RoundTripper map by URL
- binary body → `binary: true`
- Authorization header → validation error
- `http://` → validation error
- metadata IP → player error
- no network in `go test ./...`

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| wrap `curl` argv | Same as Shell; loses Go TLS/limits |
| Include POST in v0 | Mutation + secrets; needs HITL |
| Allow `http://` | Cleartext; credentials leak |
| Runtgine HTTP API (G-45) | Server, not a Player |
| Download to `path` | Mixes FS; compose two steps later |

## Risks

| Risco | Mitigação |
|---|---|
| SSRF to cloud metadata | IP filter after resolve + on redirect |
| Secrets in Task IR | Forbid Authorization/Cookie |
| Flaky CI | Fake transport only |

## Packages touched (slice 16, not this PR)

- `internal/players/httpclient` (new)
- `internal/core/api`, `internal/core/runner`
- `examples/http-get.json`
