# Design: 030-test-player

## Technical approach

### Package

`internal/players/gotest` with `New()`, `Manifest()`,
`ValidateStaticInput(workspace, input)`, `Execute`.

Player name in Manifest is `test`. Import alias in `api`/`runner`:
`gotestplayer` or `gotest` (stdlib `testing` must not collide).

Inject a `Runner` func (production: `os/exec` `go`). Tests never
invoke `go test ./...` of the Runtgine module.

### Argv

Always:

```text
go test -json -mod=readonly [-short] [-count N] [-timeout Dur] [-run RE] packages...
```

Default packages: `./...`. Timeout duration is derived from
`timeout_ms` (ceil to seconds, minimum 1s).

### Static validation

Reject before Execute:

- empty package after trim
- absolute path, `..` escape, URL-shaped package
- `count` out of 1..10
- `timeout_ms` out of range
- `run` containing NUL or newlines

### Failure model

Non-zero process exit → `result.Runtime(CodePlayerError, …)` with
details map holding the same JSON fields as success (`ok=false`,
counts, truncated log). The Runner already maps that to `step.failed`
/ `run.failed`.

### Blast / Claims

Do not add `test.go` to `claim.Required` or `blast.Touched`.
hello-style task → `risk: none`, empty touches.

### Tests

- Fake runner emitting `go test -json` pass event → `ok=true`
- Fake runner exit 1 + fail event → player error with `fail>=1`
- `../escape` package rejected
- `test.python` unregistered at Core admission
- no network; no nested full-repo `go test ./...`

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| wrap `shell.exec` | Same hole HTTP closed for curl |
| Generic `test.run` + runner enum | pytest/npm need their own sandbox; v0 is Go-only |
| Succeed with `ok=false` | Soft-fail hides verification; compiler philosophy |
| `-race` in v0 | CGO; flaky Cloud; extra flags |
| `-mod=mod` | Network download; not offline |

## Risks

| Risco | Mitigação |
|---|---|
| Nested `go test` in unit tests | Injectable runner |
| Module download | `-mod=readonly` |
| Huge JSON stream | Parse events; 64 KiB `log` cap |
| Workspace writes (`*.test` binaries) | workdir = workspace; no claim (like Shell) |

## Packages touched (slice 18, not this PR)

- `internal/players/gotest` (new)
- `internal/core/api`, runner static dispatch
- `examples/test-go.json`
