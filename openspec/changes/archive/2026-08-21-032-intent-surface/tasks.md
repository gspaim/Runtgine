# Tasks: 032-intent-surface

## Docs (this change)

- [x] `docs/32-intent-surface-v0.md` — canonical spec G-141..G-146
- [x] `docs/04-decisoes.md` — CONFIRMED section
- [x] `docs/14-tui-design.md` — INTENT tab
- [x] `docs/17-intent-engine-v0.md` — cross-ref; remove TUI exclusion
- [x] `docs/09-mvp.md`, `docs/10-gaps.md`, `docs/README.md`, `AGENTS.md`
- [x] Skill runtgine-tui-design — seven tabs
- [x] `openspec/changes/032-intent-surface/` package

## Slice 21 — TUI INTENT (done)

- [x] Extend TUI `CoreAPI` with `CompileIntent` / `SubmitIntent`
- [x] Add `tabIntent`; reorder tabs per G-142
- [x] INTENT view: input, preview, error, session history (cap 10)
- [x] Keymap: `Ctrl+p`, `Ctrl+Enter`, `Esc`; footer hints
- [x] Submit success → select run + tab LIVE
- [x] JSON mode toggle (Task IR direct submit)
- [x] Unit tests: tab cycle, compile, submit, resize, NO_COLOR
- [x] `go test ./...` green

## Fase 3 — Wails INTENT (future)

- [ ] Wails project + shadcn-svelte Constellation theme
- [ ] Intent view: equivalent to TUI G-143/G-144
- [ ] Bindings → Core; `source: "wails"`
- [ ] Submit → Live navigation
- [ ] Manual smoke on desktop target OS
