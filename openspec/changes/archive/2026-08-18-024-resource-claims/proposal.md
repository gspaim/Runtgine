# Proposal: 024-resource-claims

## Why

G-30 already allows concurrent Runs. Git, Filesystem and Docker mutate
the same workspace. Execution Policy (022) decides *whether* a
capability may run, not *who holds the resource*. Without claims, two
`fs.write` or `git.commit` + `docker.build` race.

## What Changes

- Core package `internal/core/claim` (slice 12 — **not this spec PR**)
- Exclusive kinds: `workspace` and `path`
- Auto-claim table: `fs.write`, `git.add`/`commit`, `docker.build`,
  `docker.run` with `mount_workspace=true`
- `shell.exec` does **not** auto-claim (keeps hello.json concurrent)
- Fail-fast `claim.conflict`; events `claim.acquired` /
  `claim.conflict` / `claim.released`
- Persist in the same SQLite DB; release on Run terminal and boot sweep

## What Does Not Change

- Execution Policy engine (already 022); order becomes
  Policy → Claim → Execute
- Blast Radius (rest of G-43)
- Wait / `waiting_claim` / distributed locks
- Manifest / Task IR `claims[]`
- TUI tabs (no GRAPH, no Claims tab)
- `shell.exec` sandbox

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `024-resource-claims` |
| Doc canônico | [`docs/24-resource-claims-v0.md`](../../../docs/24-resource-claims-v0.md) |
| Gaps | G-93..G-98 **CONFIRMED** (recorte de G-43) |
| Código | Ainda não — slice 12; **bloqueado** até este pacote + `04` |

## Approach

1. Hardcoded capability → resource table in Core (no Manifest field)
2. Exclusive overlap: workspace vs anything; path vs prefix segments
3. Acquire after policy allow / after HITL grant; hold until Run end
4. Conflict fails the later Run; holder unchanged

## Impact

- Package novo no slice 12: `internal/core/claim`
- Runner, store, event types
- Depende do verbo de 022 (deny nunca claima; HITL antes do acquire)
