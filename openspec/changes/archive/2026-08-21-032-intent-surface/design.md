# Design: 032-intent-surface

## Technical approach

### Tab index (TUI)

Insert INTENT as **first** tab (Entry Point before observation):

```go
tabIntent, tabRuns, tabLive, tabBoard, tabEvents, tabGraph, tabConfig
tabNames = []string{"INTENT", "RUNS", "LIVE", "BOARD", "EVENTS", "GRAPH", "CONFIG"}
```

`tabCount` becomes 7. Existing `tab`/`shift+tab` loops over `tabCount`.

### CoreAPI (TUI)

Extend TUI `CoreAPI` interface (slice 21):

```go
CompileIntent(ctx context.Context, text, source string) (task.TaskIR, string, error)
// returns TaskIR, method, error

SubmitIntent(ctx context.Context, text, source string) (string, error)
// returns run_id
```

`api.Core` already implements these. TUI passes `source: "tui"`.
Never import `internal/core/intent` from views — only through CoreAPI.

JSON mode: paste Task IR → Validator on submit path via `SubmitTask` (same as
`runtgine run -`), bypassing Intent Engine when input is valid JSON.

### Model state (TUI)

- `intentInput string` — multiline (Bubbles textarea)
- `intentMode` — `nl` | `json`
- `intentPreview *task.TaskIR` + `intentMethod string`
- `intentError string`
- `intentHistory []intentSubmission` — cap N=10 session-only (run_id, summary)

Commands:

- `Ctrl+p` → CompileIntent, show preview panel
- `Ctrl+Enter` → SubmitIntent; on success set `selectedRunID`, `tab = tabLive`
- `Esc` → clear input (with confirm if preview dirty)

### Wails (Fase 3)

- Sidebar item **Intent** (first or top of nav)
- Components: Input/Textarea, Card (preview JSON), Badge (method), Button submit
- Bindings: same Core methods, `source: "wails"`
- On success: router → Live view with `run_id`

Palette: Constellation tokens from `14` mapped to Tailwind/shadcn theme.

### Tests (TUI slice 21)

- Tab cycle includes INTENT at index 0
- FakeCore: CompileIntent returns fixture Task IR + method
- SubmitIntent returns run_id; model switches to LIVE
- Compile error surfaces message; no panic
- Width 70: View renders input + collapsed preview
- `NO_COLOR` path includes word `INTENT`

No interactive TTY required in CI (same pattern as GRAPH slice).

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Chat thread multi-turn | Contradiz visão (não chatbot); transcript RAG rejeitado |
| INTENT as modal over RUNS | Entry Point deve ser discoverable; aba dedicada |
| INTENT last tab | Entry Point should be first stop, not hidden after CONFIG |
| New `intent.compose` Player | Intent Engine já é compilador no Core |
| Replace CLI | Scripts/CI precisam de `runtgine intent` |

## Risks

| Risco | Mitigação |
|---|---|
| Confundir com chatbot | Copy UI: “Mission Brief”, preview Task IR obrigatório |
| Tab order breaks muscle memory | INTENT first is new default; document in footer |
| LLM path lento no submit | Show method + spinner; LIVE tab for progress |
| Wails delayed vs TUI | Same spec `32`; TUI slice 21 independent |

## Packages touched (slice 21, not this PR)

- `internal/entrypoint/tui`
- tests `model_test.go` / fakeCore
- `.cursor/skills/runtgine-tui-design/SKILL.md`
