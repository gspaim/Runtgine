# Proposal: 029-project-memory

## Why

ContextPack is intra-run. Graph Hits are structural. Between runs the
LLM still rediscovers decisions, failures, and handoffs. Project Memory
is episodic orientation for the same workspace — compiled observations,
not chat RAG. Sketch lives in `docs/16`; this change is the v0 cut.

## What Changes

- Canonical `docs/29-project-memory-v0.md` (G-123..G-128 CONFIRMED)
- Core package `internal/core/memory` (**slice 17 — not this spec PR**)
- Provider (not Player): Record / List / Query / Supersede / Archive
- ContextPack `memory_hits` + dedicated budget
- CLI `runtgine memory …`
- Optional capture `memory.capture=failures` (default off)

## What Does Not Change

- Event Store, Runtime Graph, Graph Hits semantics
- Task IR schema
- Players (no `memory.*`)
- G-44 MCP; G-45 HTTP API
- TUI tabs
- Validator / Policy / Claims / Blast authority

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `029-project-memory` |
| Doc canônico | [`docs/29-project-memory-v0.md`](../../../docs/29-project-memory-v0.md) |
| Gaps | G-123..G-128 **CONFIRMED** (recorte de G-46/G-47) |
| Código | Slice 17 feito |

## Approach

1. SQLite table in the existing Core DB
2. Lexical Query over `active` episodes only
3. AssembleContext (LLM path) attaches capped `memory_hits`
4. Provider errors degrade to empty items

## Impact

- New package `internal/core/memory`
- `internal/core/contextpack`, `internal/core/api`, CLI
- README Estágio: Slice 17 after code
