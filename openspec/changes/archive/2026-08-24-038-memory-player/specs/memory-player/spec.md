# memory-player

## ADDED Requirements

### Requirement: Memory Player is a deterministic read-only memory reader

The Registry SHALL expose Player `memory` with capabilities
`memory.recall` and `memory.check` only. The Player MUST call the
Project Memory Provider (`internal/core/memory`) in-process and
MUST NOT call network, MCP, or a shell string. The Player MUST
NOT write, supersede, or archive episodes.

#### Scenario: Write capability rejected

- **WHEN** a Task step names `memory.archive`
- **THEN** admission fails because the capability is unregistered

#### Scenario: Network rejected

- **WHEN** ValidateStaticInput receives `transport: "mcp"`
- **THEN** the error is `validation.invalid_input` (schema
  rejects unknown field via `additionalProperties: false`)

### Requirement: Recall returns episodes from the Provider

`memory.recall` MUST return `hits[]` from `Provider.Recall`. Each
hit MUST expose `id`, `kind`, `title`, truncated `body`
(≤ 1024 bytes), and `created_at`.

#### Scenario: Query with hits

- **WHEN** the Provider returns 3 episodes for `query: "bloco"`
- **THEN** `Execute` returns `hits` with 3 entries
- **AND** the step is `succeeded`

#### Scenario: Empty query rejected

- **WHEN** `query` is `""`
- **THEN** the error is `validation.invalid_input`

### Requirement: Provider failure degrades, never fails the Run

If `Provider.Recall` or `Provider.Check` returns an error, the
step MUST be `succeeded` with empty result, and a warning MUST be
emitted via `slog.Warn`. The Run MUST NOT fail with
`runtime.player_error`.

#### Scenario: Provider busy

- **WHEN** the Provider returns `store: busy`
- **THEN** `recall` returns `hits: []`, step `succeeded`,
  warning logged

### Requirement: Check returns has_failure from the Provider

`memory.check` MUST call `Provider.Check(pattern)` and return
`has_failure` (bool) plus optional `match` (id + title) when
`has_failure` is true.

#### Scenario: Pattern with active failure

- **WHEN** the Provider has an `active` `failure` episode whose
  title or body contains the pattern
- **THEN** `has_failure` is `true`
- **AND** `match` is the matching episode

#### Scenario: No match

- **WHEN** no active episode matches the pattern
- **THEN** `has_failure` is `false`
- **AND** `match` is absent

### Requirement: Memory Player is not a write authority

`memory.recall` and `memory.check` MUST NOT mutate the Provider.
Writing, superseding, and archiving episodes continue to flow
through `Provider` and the Lessons HITL flow of `33-evolution-v0`.

#### Scenario: Player is read-only

- **WHEN** the input requests `memory.supersede`
- **THEN** the error is `validation.unregistered_capability`
  (capability absent from Manifest)
