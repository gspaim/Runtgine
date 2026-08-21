# Design: 035-wails-v0

## Slice 27 — app + INTENT + LIVE

### Layout

```text
internal/entrypoint/desktop/
  app.go              # Wails v3 application; opens api.Core
  service.go          # application.Service methods the frontend calls
  frontend/           # Svelte 5 + Vite + shadcn-svelte (template svelte)
cmd stays ./cmd/runtgine — cobra command `desktop`
```

Module: `github.com/wailsapp/wails/v3`. CLI: `wails3`.
Pin the beta tag in `go.mod` at slice 27.

Core still does not import this package. `openCore` reused from CLI.

v3 uses **services** (static analysis → generated TypeScript under
`frontend/bindings/`), not the v2 `Bind` + `wails.Run` model.
One window only — v3 multi-window stays unused.

### Bindings / service

Thin wrappers. No business logic. Errors mapped to `{code, message}`
matching the Error model (`11` §9) so the UI can show Validator
failures the same way as the TUI.

`Subscribe`: emit a Wails v3 event name (`runtgine:event`) and
forward `event.Event` JSON. Frontend filters by `run_id` on LIVE.

### INTENT / LIVE

Port TUI semantics (`32`), not the Bubble Tea model. Preview must
call `CompileIntent` only. Submit uses `SubmitIntent` or `SubmitTask`.

### Tests

- Go: service methods with fake Core (preview does not InsertRun; submit
  returns run_id; unknown capability surfaces validation error).
- No `wails3 build` required in CI for slice 27 unit tests.
- Manual smoke: `runtgine desktop` listed in tasks, not gated on CI.

## Slice 28 — remaining views

Reuse the same service. GRAPH is read-only snapshot. CONFIG uses
`ConfigSnapshot` and must not print `api.token` / LLM keys.

BOARD: display-only of board-origin runs if polling is not started
from the desktop in v0 (operator still uses `runtgine board poll`).
Optional: a “poll once” button that calls the existing adapter —
**OPEN inside slice 28**; default is display-only to keep v0 small.

Lessons: list/approve/reject via existing Core methods.

### Tests

- Navigation smoke in frontend unit tests if cheap; otherwise Go
  tests that each service method is wired.
- `go test ./...` / `go vet ./...` still green without a display.
