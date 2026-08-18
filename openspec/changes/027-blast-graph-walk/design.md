# Design: 027-blast-graph-walk

## Technical approach

### Types

In `internal/core/blast`:

```go
type Affected struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Reason string `json:"reason"` // seed | mentions
	Via    string `json:"via,omitempty"`
}

func Walk(snap graph.Snapshot, touches []Touch) []Affected
```

`Report.Affected []Affected` always serialized (empty slice, not null).

`Analyze` stays graph-free. `BlastTask`:

1. Validate + `Analyze` (unchanged)
2. `snap, err := Graph.Snapshot` — on error, log, `affected=[]`, return report
3. `rep.Affected = Walk(snap, rep.Touches)`

Do not call `RefreshGraph`. Do not import `entrypoint`.

### Walk

- Collect unique path keys from touches (`kind == "path"`)
- Index snapshot nodes by `(kind,id)` and edges by `to`
- For each path key in first-seen order: if node exists, append seed
- For mentions edges `from → path`: append from (run|task only)
- Sort: kind rank path < task < run, then id; dedupe key
  `(kind,id,reason,via)`

Workspace touches ignored. Unknown edge kinds ignored.

### CLI

Same `runtgine blast`. Tests that decode `risk` keep working.
Hello fixture asserts `affected` is empty (not omitted).

### TUI

No change. GRAPH does not call `BlastTask`.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| QueryHits instead of hop | Hits are ranked search (`19`); blast needs explicit mentions |
| Walk `executed` on capabilities | Noisy (`shell.exec` on every hello run) |
| Multi-hop / symbols | `02` long vision; unbounded; G-116 |
| `--graph` flag default off | Extra UX; walk is cheap and degrades |
| Blast from GRAPH selection | Opposite direction; `26` G-110 |
| Bump `schema_version` | Additive field; keep `0.1.0` |

## Risks

| Risco | Mitigação |
|---|---|
| Path id mismatch claim vs graph | Same `NormalizePath` / relative ids; exact match only |
| Snapshot error hides IR | Walk failure never fails BlastTask |
| GRAPH vs walk confusion | Spec + G-115: CLI only |

## Packages touched (slice 15, not this PR)

- `internal/core/blast` (`Walk`, `Report.Affected`)
- `internal/core/api` (`BlastTask`)
- tests `blast_test.go` / `api/blast_test.go`
