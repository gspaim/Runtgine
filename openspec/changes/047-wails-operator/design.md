# Design: 047-wails-operator

## Technical approach

### Stack — no new desktop toolkit

Keep `github.com/wailsapp/wails/v3` and the Svelte 5 frontend under
`internal/entrypoint/desktop/frontend`. Do **not** add a second UI
framework, Wails v2, or an HTTP client of `runtgine serve`.

### RESULT

Reuse `internal/entrypoint/liveout` (already used by the TUI). The
Svelte LIVE view already parses `step.succeeded` / `step.failed` in
`core.ts` (`#68`). Slice 40 must keep RESULT at the top of LIVE and
cover it in service tests via `liveout.FromEvents` on `GetRun` events.

Do not add a Core table for “result documents”.

### CoreAPI / Service

```go
QueryHits(ctx context.Context, q graph.Query) graph.Hits
BlastTask(ctx context.Context, t task.Task) (blast.Report, error)
```

`api.Core.BlastTask` and TUI `QueryHits` already exist. Desktop
`CoreAPI` + `Service` must expose them (G-162 listed Blast; the
frontend never called it). Fake Core in `service_test.go` returns
fixtures; desktop still must not import `store` or call a Player.

INTENT Blast: `CompileIntent` or JSON parse (no submit) → `BlastTask`.
Compile failure shows the existing INTENT error, not a fake report.

LIVE Blast: unmarshal `snapshot.Task` → `BlastTask`.

### Hits

LIVE: walk `snapshot.Events` (same as TUI `hitsFromEvents`). Do not
call `QueryHits` on every Wails event tick.

INTENT: `QueryHits` only on Preview (`CompileIntent` cadence).

### UI

Buttons on INTENT: Preview, Submit, Blast.
Buttons on LIVE: Blast, plus existing Cancel / Approve / Deny.
GRAPH: no Hits list, no Blast control.

Tokens stay Constellation (`14`). No eighth nav item.

### CI

`go test ./internal/entrypoint/desktop` without a display. Manual
smoke: `runtgine desktop` on a machine with GTK4 + WebKitGTK 6.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Substituir a TUI | G-164 / G-244; TUI permanece referência de terminal |
| Oitava view HITS/BLAST | Sete views fixas em `35` |
| Hits/Blast na view GRAPH | G-110 / G-115 |
| Cliente HTTP de `serve` | Superfície local = in-process |
| Mobile Wails | G-164 |
