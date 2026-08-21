# Tasks: 033-evolution-v0

## Docs (this change)

- [x] `docs/33-evolution-v0.md` — G-147..G-152
- [x] Cross-refs in `04`, `08`, `09`, `10`, `02`, `README`, `AGENTS`
- [x] OpenSpec `033-evolution-v0`

## Slice 22 — Player Router (done)

- [x] Config schema `llm_providers` + `llm_routing`
- [x] `router.SelectLLM` with match precedence
- [x] Wire LLM Player / Runner to router
- [x] Migrate legacy `llm_backend` alias
- [x] Tests + docs config example

## Slice 23 — Playbooks (done)

- [x] Loader `.runtgine/playbooks/*.md`
- [x] ContextPack `playbook_hits` + budget
- [x] Example playbooks orchestrator/developer/qa
- [x] Tests index + context assembly

## Slice 24 — Lessons (done)

- [x] `lessons.capture` config
- [x] Proposal store + analyzer on `run.failed`
- [x] CLI `runtgine lessons list|approve|reject`
- [x] Approve → memory.Record; optional patch file
- [x] Tests HITL paths; no silent writes
