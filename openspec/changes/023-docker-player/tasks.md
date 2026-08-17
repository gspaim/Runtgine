# Tasks: 023-docker-player

**Blocked until** `022-execution-policy` is implemented and merged.

## 1. Player package

- [ ] 1.1 Scaffold `internal/players/docker` + Manifest (5 capabilities)
- [ ] 1.2 Command-runner interface (exec vs stub)
- [ ] 1.3 Implement `docker.ps` / `inspect` / `logs`
- [ ] 1.4 Implement `docker.run` argv (`--pull=never --network=none --rm`)
- [ ] 1.5 Implement `docker.build` with workspace confinement
- [ ] 1.6 Reject unsupported capabilities and flag injection

## 2. Policy + tests

- [ ] 2.1 Manifest `approval-required` on run/build
- [ ] 2.2 HITL test: no exec before grant; deny never execs
- [ ] 2.3 Confinement tests for context / workdir / mount_workspace
- [ ] 2.4 Suite passes without a Docker daemon

## 3. Wire + example

- [ ] 3.1 Register in `api.Open`
- [ ] 3.2 Runner static validation dispatch
- [ ] 3.3 Add `examples/docker-ps.json`

## 4. Docs / OpenSpec closeout

- [ ] 4.1 README Estágio: Slice 11 Feito
- [ ] 4.2 Archive this change into `openspec/specs/docker-player/`
