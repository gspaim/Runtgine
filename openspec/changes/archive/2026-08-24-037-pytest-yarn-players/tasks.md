# Tasks: 037-pytest-yarn-players

## Docs (this change)

- [ ] `docs/37-pytest-yarn-players-v0.md` — G-172..G-179
- [ ] Cross-refs em `04`, `09`, `10`, `02`, `05`, `01`, `11`, `17`,
      `30`, `36`
- [ ] `docs/README.md`, `AGENTS.md`, `README.md`, `REVIEW.md`
- [ ] OpenSpec `037-pytest-yarn-players` (esta change)

## Slice 30 — Player pytest (corte v0)

- [ ] Pacote `internal/players/pytst` + Manifest `pytest.run`
- [ ] `ValidateStaticInput` (workdir, marker, ranges)
- [ ] Executor injetável; truncate `log`; fail on non-zero exit
- [ ] Registrar no `api.Open`; runner static dispatch
- [ ] `examples/pytest-run.json`
- [ ] Intent `heuristic.pytest` (pytest prefix beats shell)
- [ ] Fake-exec tests; `go test ./...` sem Python no PATH
- [ ] README Estágio: Slice 30; arquivar OpenSpec `037` após código

## Slice 30 — Player yarn (corte v0)

- [ ] Pacote `internal/players/jstest` + Manifest `yarn.test`
- [ ] `ValidateStaticInput` (workdir, `package.json`, install/flags negadas)
- [ ] Executor injetável; truncate `log`; fail on non-zero exit
- [ ] Registrar no `api.Open`; runner static dispatch
- [ ] `examples/yarn-test.json`
- [ ] Intent `heuristic.yarn` (yarn test beats shell)
- [ ] Fake-exec tests; `go test ./...` sem Yarn/Node no PATH
- [ ] `go vet ./...` verde
