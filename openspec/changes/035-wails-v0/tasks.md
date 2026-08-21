# Tasks: 035-wails-v0

## Docs (this change)

- [x] `docs/35-wails-v0.md` — G-159..G-165
- [x] Cross-refs in `04`, `07`, `09`, `10`, `01`, `05`, `11`, `13`, `14`, `17`, `32`, `34`
- [x] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [x] OpenSpec `035-wails-v0`

## Slice 27 — app + INTENT/LIVE

- [x] Wails v3 scaffold under `internal/entrypoint/desktop`
- [x] Pin `github.com/wailsapp/wails/v3` beta tag in `go.mod`
- [x] CLI `runtgine desktop`
- [x] Wails v3 service over `api.Core` (Intent, Submit, GetRun, Subscribe, Cancel, Approve)
- [x] Task IR enum `wails` for `source.entry_point`
- [x] INTENT view: preview without Run; submit → LIVE
- [x] Service tests with fake Core; `go test ./...` without display
- [ ] Manual smoke of the window (not CI-gated)

## Slice 28 — remaining views (future)

- [ ] RUNS, BOARD, EVENTS, GRAPH, CONFIG
- [ ] Lessons HITL in the UI
- [ ] CONFIG hides secrets
- [ ] `go test ./...` / `go vet ./...` green
