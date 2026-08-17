# Tasks: 020-git-player

## 1. Player package

- [ ] 1.1 Scaffold `internal/players/git` + Manifest (5 capabilities)
- [ ] 1.2 Workdir/path confinement helpers (reuse Shell patterns or shared util)
- [ ] 1.3 Implement `git.status` / `git.diff` / `git.log`
- [ ] 1.4 Implement `git.add` / `git.commit` (hooks off + identity -c)
- [ ] 1.5 Reject unknown capability / bad paths / empty message

## 2. Tests

- [ ] 2.1 Temp repo: status clean / dirty
- [ ] 2.2 diff + log after commit
- [ ] 2.3 path escape rejected
- [ ] 2.4 commit without network/hooks

## 3. Wire + example

- [ ] 3.1 `api.Open` registra Git Player
- [ ] 3.2 `examples/git-status.json`
- [ ] 3.3 Smoke: `runtgine run examples/git-status.json` no repo

## 4. Docs / OpenSpec closeout

- [ ] 4.1 README Estágio: Slice 8 Feito; Próximo = próximo Player ou HITL spec
- [ ] 4.2 `docs/10-gaps` checklist slice 8
- [ ] 4.3 Arquivar `openspec/changes/020-git-player` → `archive/` e merge
      deltas em `openspec/specs/git-player/spec.md`
