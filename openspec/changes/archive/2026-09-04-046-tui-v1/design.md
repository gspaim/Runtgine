# Design: 046-tui-v1

## Technical approach

### Stack — no new modules

`go.mod` already requires:

```text
charm.land/bubbletea/v2
charm.land/lipgloss/v2
charm.land/bubbles/v2
```

Slice 39 imports Bubbles subpackages that ship in that module
(`table`, `viewport`, `textarea`, `list`, `help`). Do **not** add
`huh`, Ratatui bindings, or a CSS-in-terminal kit.

`internal/entrypoint/tui/model.go` may split into focused files
(`intent.go`, `runs.go`, `live.go`, `help.go`) in the same package;
the `Model` remains the Elm root.

### CoreAPI

```go
QueryHits(ctx context.Context, q graph.Query) graph.Hits
BlastTask(ctx context.Context, t task.Task) (blast.Report, error)
```

`api.Core.BlastTask` already exists. `QueryHits` is a one-line wrapper
on `Core.Graph.QueryHits` (degrade to empty `Hits` if Graph is nil).
Fake Core in `model_test.go` returns fixtures; TUI still must not
import `store`.

INTENT `Ctrl+b`: `CompileIntent` (no submit) → `BlastTask`. Failure to
compile shows the existing preview error, not a blast panel.

LIVE `b`: unmarshal `snapshot.Task` → `BlastTask`. Missing/invalid
task → error line in the drawer.

### Hits extraction (LIVE)

Walk `snapshot.Events` newest-first; decode ContextPack JSON from
`step.started` / equivalent payloads already stored. Display up to the
same budgets as `19` (do not re-rank). If none, render `No hits.`

Do not call `QueryHits` on every tick. INTENT preview hits refresh
only on `Ctrl+p` (same cadence as Task IR preview).

### Help overlay

`?` toggles a fullscreen-ish `help.Model` (or a Lip Gloss panel using
the Bubbles help keymap). While open, other keys except `?`/`esc`/`q`
are swallowed. `q` still quits the app (existing behavior) — document
in the overlay.

### Layout bands (unchanged from `14`)

| Width | Shell |
|---|---|
| >= 120 | Tabs + side-by-side (list \| detail; INTENT textarea \| preview; LIVE trajectory \| hits/blast) |
| 80–119 | Main + drawer below |
| < 80 | Vertical stack; no extra columns |

`View` stays side-effect free. Resize via `tea.WindowSizeMsg` as today.

### GRAPH / BOARD / CONFIG

Chrome only: wrap existing data in `list`/`viewport`/`Panel`. No
semantic change. GRAPH filter `/` stays independent of EVENTS.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Ratatui / Textual | Quebra Go-unico; Wails já cobre GUI rica |
| Oitava aba BLAST ou HITS | `14` fixa sete abas; dados cabem em drawer |
| Hits na aba GRAPH | G-110; GRAPH é estrutural, Hits é ranking |
| Blast-from-GRAPH | G-115; walk já está no report CLI |
| `huh` forms | Dependência nova; textarea+confirm atuais bastam |
| Auto-blast a cada keystroke | Barulho; `Ctrl+b` é explícito |

## Risks

| Risco | Mitigação |
|---|---|
| `model.go` ainda maior | Split por aba no mesmo package |
| Charm v2 table API | Pin já no go.mod; testes de View sem TTY |
| ContextPack payload shape drift | Decode best-effort; empty hits, no panic |
| Key `b` vs filter | Só ativo em LIVE; `/` continua filtro |

## Packages touched (slice 39, not this PR)

- `internal/entrypoint/tui`
- `internal/core/api` (QueryHits wrapper, if not already on Core)
- tests `model_test.go` / fakeCore
