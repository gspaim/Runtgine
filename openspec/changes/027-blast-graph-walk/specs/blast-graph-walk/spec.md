# Delta for blast-graph-walk

## ADDED Requirements

### Requirement: Walk is 1-hop mentions from path touches

After Task-IR analysis, Blast SHALL compute `affected` from
`GetGraphSnapshot` by seeding unique `touches` with `kind=path` and
walking inbound `mentions` edges one hop. Walk SHALL NOT use QueryHits,
SHALL NOT write graph rows, and SHALL NOT change `risk`.

#### Scenario: Mentions hop

- GIVEN a graph node `path` `notes.md`
- AND a `mentions` edge `run:prior → path:notes.md`
- WHEN `BlastTask` runs a Task that touches `path` `notes.md`
- THEN `affected` includes `{kind: path, id: notes.md, reason: seed}`
- AND `affected` includes `{kind: run, id: prior, reason: mentions, via: path:notes.md}`

### Requirement: Missing graph degrades

If the snapshot is empty, the path node is absent, or snapshot
read fails, `affected` MUST be an empty array and the rest of the
Impact Report MUST still be returned.

#### Scenario: Unknown path

- GIVEN no `path` node for `a.txt`
- WHEN `BlastTask` runs `fs.read` on `a.txt`
- THEN `affected` is `[]`
- AND `touches` still contain `path` `a.txt`

### Requirement: Hello stays empty

`examples/hello.json` SHALL produce `affected: []` because it has no
`path` touches.

#### Scenario: Shell only

- GIVEN `examples/hello.json`
- WHEN `runtgine blast` runs
- THEN `affected` is `[]`
- AND `risk` is `none`
- AND no Run is created
