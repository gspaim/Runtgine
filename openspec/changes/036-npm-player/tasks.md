# Tasks: 036-npm-player

## Docs (this change)

- [x] `docs/36-npm-player-v0.md` — G-166..G-171
- [x] Cross-refs in `04`, `09`, `10`, `02`, `05`, `01`, `11`, `17`, `30`
- [x] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [x] OpenSpec `036-npm-player`

## Slice 29 — Player (future)

- [ ] Package `internal/players/npm` + Manifest `npm.test`
- [ ] `ValidateStaticInput` (workdir, package.json, ranges)
- [ ] Injectable runner; log truncate; fail on non-zero exit
- [ ] Register in `api.Open`; runner static dispatch
- [ ] `examples/npm-test.json`
- [ ] Intent `heuristic.npm` (`npm test` beats shell prefix)
- [ ] Fake-exec tests; `go test ./...` / `go vet ./...` without Node
- [ ] README Estágio: Slice 29; archive this change
