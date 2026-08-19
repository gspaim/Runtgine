# Design: 029-project-memory

## Technical approach

### Package

`internal/core/memory` with a `Store` over the existing SQLite.

Not a Player. Wired in `api.Open` like `graph.Service`.

### Episode

```text
kind:      decision | failure | handoff | preference
validity:  active | superseded | archived
title:     1..200 runes
body:      0..4096 bytes UTF-8
```

`Supersede` is one transaction: old.validity=superseded,
old.successor_id=new.id, insert new active.

### Query

Tokenize query (lowercase, split on non-alphanumeric). Score = count of
tokens that appear in title+body. Filter validity=active. Sort score
desc, created_at desc, id asc. Limit default 8.

No FTS required in v0 if a full scan of a small table is enough;
FTS5 is optional later, not this slice.

### ContextPack

Add `MemoryHits` + budget fields. `Assemble` initializes
`memory_hits.items = []`. Runner LLM path: `Query(intent.summary)` then
`WithMemoryHits`. Shell/pipeline heuristics skip (G-69 pattern).

Truncation drops lowest score first; never drops task/step.

### Capture

Config `memory.capture`: `off` | `failures`. Default `off`.
On `run.failed` only if `failures`: Record best-effort; log and ignore
errors.

### Tests

- Record/List/Query/Supersede/Archive in temp DB
- Query hides superseded
- AssembleContext includes empty memory_hits
- Injected Query error → empty items, pack otherwise valid
- Intent heuristic path does not call memory
- capture=off does not write on failed run

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Wait for Fase A sidecar experiments | User asked to promote a Core cut now; local SQLite is the smallest honest v0 |
| MCP Provider (G-44) | Still P3; local store does not block later MCP |
| Memory Player `memory.*` | No Task-IR step needs it; Provider is enough |
| Embeddings / FTS | Overkill; lexical score matches Graph Hits style |
| Auto-capture all outcomes | Noisy; failures opt-in only |

## Risks

| Risco | Mitigação |
|---|---|
| Secrets in body | UTF-8 cap; no env/transcript dump; capture uses error message only |
| Memory as authority | Negative list in spec; no Policy hook |
| Context noise | Small budget (8 / 2000); active-only; below graph_hits |

## Packages touched (slice 17, not this PR)

- `internal/core/memory` (new)
- `internal/core/store` (table)
- `internal/core/contextpack`
- `internal/core/api`, CLI
- `internal/config` (`memory.capture`)
