# Tasks: 021-filesystem-player

## 1. Player package

- [ ] 1.1 Scaffold `internal/players/filesystem` + Manifest (4 capabilities)
- [ ] 1.2 Implement workspace/path/symlink confinement helpers
- [ ] 1.3 Implement `fs.read` with UTF-8 and byte budget
- [ ] 1.4 Implement `fs.write` with atomic same-parent rename
- [ ] 1.5 Implement deterministic `fs.list` and `fs.stat`
- [ ] 1.6 Reject unsupported capabilities and invalid inputs

## 2. Tests

- [ ] 2.1 Read UTF-8 content and truncation
- [ ] 2.2 Write/create parents/atomic replacement
- [ ] 2.3 List ordering, recursion and entry limit
- [ ] 2.4 Stat file/directory/symlink
- [ ] 2.5 Reject path escape and external symlink
- [ ] 2.6 Reject invalid UTF-8, oversized writes and unsupported operations

## 3. Wire + example

- [ ] 3.1 Register Filesystem Player in `api.Open`
- [ ] 3.2 Add Runner static validation dispatch
- [ ] 3.3 Add `examples/fs-read.json`
- [ ] 3.4 Smoke `runtgine run examples/fs-read.json`

## 4. Docs / OpenSpec closeout

- [ ] 4.1 README Estágio: Slice 9 Feito; next = next Player or HITL spec
- [ ] 4.2 `docs/10-gaps.md` checklist slice 9
- [ ] 4.3 Archive `openspec/changes/021-filesystem-player` and merge delta
      into `openspec/specs/filesystem-player/spec.md`
