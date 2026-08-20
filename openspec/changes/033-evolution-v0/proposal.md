# Proposal: 033-evolution-v0

## Why

Runtgine has effort/difficulty heuristics, a single LLM backend default, basic
Task Router (G-26), and Project Memory v0 — but no multi-model routing, no
workspace playbooks/skills, and no assisted improvement loop from run failures.

Product direction (aligned in discussion): specialize via **Players +
playbooks**, route LLM by task signals, and improve continuously with **HITL** —
not an autonomous multi-agent framework.

## What Changes

- Canonical `docs/33-evolution-v0.md` (G-147..G-152 CONFIRMED)
- `docs/04-decisoes.md`, `docs/08-workflow-templates.md`, `docs/09-mvp.md`,
  `docs/10-gaps.md`, `docs/README.md`, `AGENTS.md`
- `docs/02-conceitos.md` — Player Router cross-ref
- OpenSpec package `033-evolution-v0`

## What Does Not Change

- Player ≠ Agent; Validator sovereign
- Project Memory semantics (`29`)
- Intent Surface (`32`), MVP 1.0 scope
- No agent registry, no chat RAG, no silent skill mutation

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `033-evolution-v0` |
| Doc canônico | [`docs/33-evolution-v0.md`](../../../docs/33-evolution-v0.md) |
| Gaps | G-147..G-152 **CONFIRMED** (spec) |
| Code | slices 22–24 — **not started** |

## Approach

Three implementation slices after Intent Surface (21):

1. **Slice 22** — Player Router + multi-provider config
2. **Slice 23** — Playbooks loader + `playbook_hits`
3. **Slice 24** — Lessons postmortem + HITL promotion

## Impact

- Future: `internal/core/router`, `internal/core/playbooks`, lessons job
- Config schema extension (`.runtgine/config.json`)
- ContextPack extension (`playbook_hits`)
- Docs only in this PR
