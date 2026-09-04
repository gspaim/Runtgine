# Design: 036-npm-player

## Slice 29 — Player (not this spec PR)

### Package

`internal/players/npm` with `New()`, `Manifest()`,
`ValidateStaticInput(workspace, capability, input)`, `Execute`.

Player name in Manifest is `npm`. Import in `api`/`runner`: the
package path is `internal/players/npm` (stdlib has no `npm`).

Inject `ExecFunc` (production: `exec.CommandContext("npm", args...)`
with `Dir` = resolved workdir). Tests never invoke a real npm.

### Capability contract

| Capability | Argv | Body |
|---|---|---|
| `npm.test` | `npm test` or `npm --silent test` | `ok`, `exit_code`, `elapsed_ms`, `log`, optional `script` |

`workdir` resolved like Git/Test (EvalSymlinks, stay in workspace).
Require `package.json` in that directory.

### Static validation

Reject before Execute:

- capability ≠ `npm.test`
- `workdir` absolute, URL-shaped, or `..` escape
- missing `package.json`
- `timeout_ms` out of range (if schema did not already)

### Failure

Non-zero exit → `runtime.player_error` (fail the Run). Still attach
`log` / `exit_code` on the error payload like `test.go`.

### Blast / Claims

Do not add `npm.test` to `claim.Required` or `blast.Touched`.

### Intent

In `internal/core/intent`, before `matchShell`: phrases `npm test`,
`npm run test`, `roda os testes npm`, `run npm tests` → one-step
Task `npm.test`, method `heuristic.npm`. Yarn/pnpm stay shell.

### Tests (slice 29)

- Fake exit 0 → `ok=true`
- Fake exit 1 → player error
- Missing package.json → validation
- Escape workdir → validation
- Intent `npm test` is not `shell.exec`
- `go test ./...` without Node on PATH required

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| pytest first | No Python tree in this repo; weaker dogfood |
| Extend Player `test` with `test.npm` | `gotest` package is Go-specific; G-41 pattern is one Player package per cut |
| `npm run` arbitrary scripts | Unbounded (start/build/rm); v0 is test only |
| `npm install` in v0 | Network + lock mutation; opposite of `-mod=readonly` |
| Parse TAP/JUnit | Extra surface; exit code is enough like a first cut |
| K8s / Terraform / PostgreSQL | PRD P3 infra; not the next G-41 test-runner cut |

## Risks

| Risco | Mitigação |
|---|---|
| `scripts.test` is arbitrary JS | Document; no install/npx/prefix; workspace workdir |
| Flaky CI needing Node | Fake ExecFunc only |
| Intent still routes `npm ci` to shell | Out of v0; only `npm test` heuristic |
