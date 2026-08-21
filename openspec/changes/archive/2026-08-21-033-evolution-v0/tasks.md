# Tasks: 033-evolution-v0

## Docs (this change)

- [x] `docs/33-evolution-v0.md` — G-147..G-152
- [x] Cross-refs in `04`, `08`, `09`, `10`, `02`, `README`, `AGENTS`
- [x] OpenSpec `033-evolution-v0`

## Slice 22 — Player Router (future)

- [ ] Config schema `llm_providers` + `llm_routing`
- [ ] `router.SelectLLM` with match precedence
- [ ] Wire LLM Player / Runner to router
- [ ] Migrate legacy `llm_backend` alias
- [ ] Tests + docs config example

## Slice 23 — Playbooks (future)

- [ ] Loader `.runtgine/playbooks/*.md`
- [ ] ContextPack `playbook_hits` + budget
- [ ] Example playbooks orchestrator/developer/qa
- [ ] Tests index + context assembly

## Slice 24 — Lessons (future)

- [ ] `lessons.capture` config
- [ ] Proposal store + analyzer on `run.failed`
- [ ] CLI `runtgine lessons list|approve|reject`
- [ ] Approve → memory.Record; optional patch file
- [ ] Tests HITL paths; no silent writes
