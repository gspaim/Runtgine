# Tasks: 020-git-player

## 1. Player package

- [x] 1.1 Scaffold `internal/players/git` + Manifest (5 capabilities)
- [x] 1.2 Workdir/path confinement helpers (reuse Shell patterns or shared util)
- [x] 1.3 Implement `git.status` / `git.diff` / `git.log`
- [x] 1.4 Implement `git.add` / `git.commit` (hooks off + identity -c)
- [x] 1.5 Reject unknown capability / bad paths / empty message

## 2. Tests

- [x] 2.1 Temp repo: status clean / dirty
- [x] 2.2 diff + log after commit
- [x] 2.3 path escape rejected
- [x] 2.4 commit without network/hooks

## 3. Wire + example

- [x] 3.1 `api.Open` registra Git Player
- [x] 3.2 `examples/git-status.json`
- [x] 3.3 Smoke: `runtgine run examples/git-status.json` no repo

## 4. Docs / OpenSpec closeout

- [x] 4.1 README Estágio: Slice 8 Feito; Próximo = próximo Player ou HITL spec
- [x] 4.2 `docs/10-gaps` checklist slice 8
- [x] 4.3 Arquivar `openspec/changes/020-git-player` → `archive/` e merge
      deltas em `openspec/specs/git-player/spec.md`
