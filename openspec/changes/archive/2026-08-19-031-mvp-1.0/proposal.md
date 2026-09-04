# Proposal: 031-mvp-1.0

## Why

`docs/09-mvp.md` still described the original minimum runtime (Core +
Shell + CLI/TUI + Board) and listed Intent, Policies, Claims, Memory
and extra Players as out of scope — while slices 1–18 already shipped
them. The product line is intent → **verifiable** execution. The
skinny 1.0 closes two holes without pulling G-45 / NATS / Wails / MCP:

1. NL still falls through to `shell.exec` (`go test`, `git status`)
2. LLM ContextPack `repo_hits` is empty unless `pipeline.repo-search`
   ran in the same Run

## What Changes

- Rewrite canonical `docs/09-mvp.md` (MVP realizado vs MVP 1.0 magro)
- Promote G-135..G-140 **CONFIRMED** in `04` / `10` / `05`
- Canonical Context Engine v0: `docs/31-context-engine-v0.md`
- Amend Intent Engine (`17`) with player heuristics
- **Code is not this PR** — slice 19 (Intent) then slice 20 (Context)

## What Does Not Change

- Task IR schema; Registry / Validator authority
- Existing Players, Policy, Claims, Blast, Memory, TUI
- G-45 HTTP server; G-44 MCP; G-36 NATS; Wails; Memory Player
- Player Router; Workflow Templates; embeddings; pytest/npm
- `git.add` / `commit` / `push` and `http.get` via NL (v0 heuristics
  are read-only git/docker + `test.go`)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `031-mvp-1.0` |
| Doc canônico | [`docs/09-mvp.md`](../../../docs/09-mvp.md) |
| Gaps | G-135..G-140 **CONFIRMED** |
| Código | Ainda não — slices 19–20; **bloqueado** até este pacote + `04` |

## Approach

1. Spec this PR: `09` + `31` + emenda `17` + espelhos
2. Slice 19: `matchPlayer` before `matchShell` (`go test` ≠ shell)
3. Slice 20: seed empty `repo_hits` from `QueryHits` path/symbol
4. Archive `031` after both code slices

## Impact

- Docs / OpenSpec only in this PR
- Later: `internal/core/intent`, `internal/core/contextpack` (+ runner)
