# Tasks: 021-filesystem-player

## 1. Player package

- [x] 1.1 Scaffold `internal/players/filesystem` + Manifest (4 capabilities)
- [x] 1.2 Implement workspace/path/symlink confinement helpers
- [x] 1.3 Implement `fs.read` with UTF-8 and byte budget
- [x] 1.4 Implement `fs.write` with atomic same-parent rename
- [x] 1.5 Implement deterministic `fs.list` and `fs.stat`
- [x] 1.6 Reject unsupported capabilities and invalid inputs

## 2. Tests

- [x] 2.1 Read UTF-8 content and truncation
- [x] 2.2 Write/create parents/atomic replacement
- [x] 2.3 List ordering, recursion and entry limit
- [x] 2.4 Stat file/directory/symlink
- [x] 2.5 Reject path escape and external symlink
- [x] 2.6 Reject invalid UTF-8, oversized writes and unsupported operations

## 3. Wire + example

- [x] 3.1 Register Filesystem Player in `api.Open`
- [x] 3.2 Add Runner static validation dispatch
- [x] 3.3 Add `examples/fs-read.json`
- [x] 3.4 Smoke `runtgine run examples/fs-read.json`

## 4. Docs / OpenSpec closeout

- [x] 4.1 README Estágio: Slice 9 Feito; next = next Player or HITL spec
- [x] 4.2 `docs/10-gaps.md` checklist slice 9
- [x] 4.3 Archive `openspec/changes/021-filesystem-player` and merge delta
      into `openspec/specs/filesystem-player/spec.md`
