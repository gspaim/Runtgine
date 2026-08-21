# Design: 035-wails-v0

## Slice 27 — app + INTENT + LIVE

### Layout

```text
internal/entrypoint/desktop/
  app.go              # Wails v2 app; opens api.Core
  bindings.go         # methods the frontend calls
  frontend/           # Svelte 5 + Vite + shadcn-svelte
cmd stays ./cmd/runtgine — cobra command `desktop`
```

Core still does not import this package. `openCore` reused from CLI.

### Bindings

Thin wrappers. No business logic. Errors mapped to `{code, message}`
matching the Error model (`11` §9) so the UI can show Validator
failures the same way as the TUI.

`Subscribe`: register a Wails event name (`runtgine:event`) and
forward `event.Event` JSON. Frontend filters by `run_id` on LIVE.

### INTENT / LIVE

Port TUI semantics (`32`), not the Bubble Tea model. Preview must
call `CompileIntent` only. Submit uses `SubmitIntent` or `SubmitTask`.

### Tests

- Go: bindings with fake Core (preview does not InsertRun; submit
  returns run_id; unknown capability surfaces validation error).
- No `wails build` required in CI for slice 27 unit tests.
- Manual smoke: `runtgine desktop` listed in tasks, not gated on CI.

## Slice 28 — remaining views

Reuse the same bindings. GRAPH is read-only snapshot. CONFIG uses
`ConfigSnapshot` and must not print `api.token` / LLM keys.

BOARD: display-only of board-origin runs if polling is not started
from the desktop in v0 (operator still uses `runtgine board poll`).
Optional: a “poll once” button that calls the existing adapter —
**OPEN inside slice 28**; default is display-only to keep v0 small.

Lessons: list/approve/reject via existing Core methods.

### Tests

- Navigation smoke in frontend unit tests if cheap; otherwise Go
  tests that each binding is wired.
- `go test ./...` / `go vet ./...` still green without a display.
