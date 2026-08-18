# Design: 025-blast-radius

## Technical approach

### Package

`internal/core/blast` with:

- `type Report struct` matching G-100 (`schema_version`,
  `capabilities`, `touches`, `predicted_claims`, `risk`, `conflicts`,
  `images`)
- `Touched(capability, input) ([]Touch, error)` — G-101 table
- `Risk(predicted []Resource) Risk` — `none` / `path` / `workspace`
- `Overlay(predicted []Resource, active []store.ResourceClaim) []Conflict`

Predicted claims MUST call `claim.Required` (do not fork G-95).
Path normalize MUST call `claim.NormalizePath`. Overlap MUST call
`claim.Overlaps`.

`BlastTask` in `internal/core/api`:

1. Validate Task IR (schema + registered capability + static input)
2. Build report from steps (topo order not required; appearance order)
3. If Store present: overlay active claims
4. Return report. Never `Submit`, never `Acquire`, never Player.

Invalid path / unknown capability → existing validation error codes.
Do not return a partial report.

### Persistence / events

None in v0. No table, no `blast.computed`. Overlay reads
`resource_claims` but does not write.

### CLI

`runtgine blast <file>` opens Core like `graph snapshot`, prints
`json.MarshalIndent` of the report. Non-zero exit only on validate /
open / I/O. Non-empty `conflicts` is still exit 0.

### Runner

Unchanged. `runtgine run` does not call Blast.

### Surfaces

No TUI change. Follow TUI skill: do not add GRAPH or a Blast tab
without `14`.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Walk Runtime Graph in v0 | `18`/`19` already exclude Blast-from-graph; Hits exist |
| Auto-blast on every `run` | Extra work on hello.json; analysis is on-demand |
| Gate Execute on `risk`/`conflicts` | Mixes analysis with Policy/Claim; operator still `run`s |
| `git.add` predicted claim per path | Contradicts G-95 (workspace); deadlock risk on Execute |
| Touch `shell.exec` argv | Undecidable; keeps hello concurrent |
| Persist reports | No consumer; SQLite already holds claims/events |
| TUI GRAPH this slice | `14` tabs are fixed; skill forbids multiplexer/new tabs |

## Risks

| Risco | Mitigação |
|---|---|
| Relatório “mentir” vs claim real | Predicted = `claim.Required` only |
| Overlay stale | Same SQLite as Execute; best-effort snapshot |
| Scope creep Graph | G-104 explicit; next spec after `14` if UI |

## Packages touched (slice 13, not this PR)

- `internal/core/blast` (novo)
- `internal/core/api`, `internal/entrypoint/cli`
- README CLI table
