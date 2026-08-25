# mcp-memory

## ADDED Requirements

### Requirement: MCP Memory Server is a read-only entrypoint over the Project Memory Provider

The MCP Memory Server SHALL be an entrypoint package
(`internal/entrypoint/mcpserver`) exposing the Project Memory
Provider (`internal/core/memory`) to external MCP clients. It MUST
be read-only: only `memory.query` and `memory.list` tools are
exposed, and the server MUST NOT record, supersede, or archive
episodes. The server MUST NOT register Players or capabilities in
the Registry.

#### Scenario: Write tool absent from tools/list

- **WHEN** an MCP client requests `tools/list`
- **THEN** only `memory.query` and `memory.list` appear
- **AND** no write tool exists on any transport

#### Scenario: Unknown tool call rejected

- **WHEN** a client calls an unregistered tool such as
  `memory.record`
- **THEN** the response is a well-formed JSON-RPC error

### Requirement: memory.query returns active episodes via lexical ranking

`memory.query` MUST delegate to the Provider's deterministic
lexical query and MUST return `hits[]` with `id`, `kind`, `title`,
truncated `snippet` (≤ 200 runes), and `score`. Only episodes with
`validity=active` MUST ever be returned.

#### Scenario: Query with hits

- **WHEN** the Provider returns 3 episodes for `text: "deploy"`
- **THEN** the tool result contains `hits` with 3 entries
- **AND** every hit has `id`, `kind`, `title`, `snippet`, `score`

#### Scenario: Empty text rejected

- **WHEN** `text` is `""` or longer than 512 characters
- **THEN** the tool responds with a validation error

#### Scenario: Superseded episode never returned

- **WHEN** a matching episode has `validity=superseded`
- **THEN** it does not appear in `hits`

### Requirement: memory.list returns filtered active episodes

`memory.list` MUST return `episodes[]` (`id`, `kind`, `title`,
`created_at`) filtered by optional `kind` (enum of the four kinds)
and `limit` (1–100, default 20), restricted to `validity=active`.

#### Scenario: Filter by kind

- **WHEN** `memory.list` is called with `kind: "failure"`
- **THEN** only active `failure` episodes are returned

#### Scenario: Invalid kind rejected

- **WHEN** `kind` is `"transcript"` (not one of the four kinds)
- **THEN** the tool responds with a validation error

### Requirement: Provider failure degrades, never kills the server

If the Provider returns an error (store unavailable, SQLite busy),
each tool MUST respond with a well-formed empty result, and a
warning MUST be emitted via `slog.Warn`. The server process MUST
NOT exit or return a transport-level failure.

#### Scenario: Provider busy

- **WHEN** the Provider returns `store: busy` during
  `memory.query`
- **THEN** the tool returns an empty `hits[]`
- **AND** the server remains alive for subsequent calls

### Requirement: stdio transport via runtgine mcp

The CLI SHALL expose `runtgine mcp`, which speaks MCP over
JSON-RPC 2.0 on stdin/stdout as a child process of the MCP client,
scoped to the workspace Core instance.

#### Scenario: Initialize handshake

- **WHEN** a client sends the MCP `initialize` request on stdio
- **THEN** the server replies with its capabilities and server
  info

#### Scenario: Tools list on stdio

- **WHEN** a client sends `tools/list` after initialization
- **THEN** the response lists exactly the two read-only tools

### Requirement: HTTP transport under existing auth

The HTTP API (`runtgine serve`) SHALL expose route `/mcp` behind
the same bearer-token middleware and loopback-only binding used by
the existing endpoints.

#### Scenario: Missing token rejected

- **WHEN** a request reaches `/mcp` without a bearer token
- **THEN** the response is `401` before the Provider is touched

#### Scenario: Authorized call succeeds

- **WHEN** a request reaches `/mcp` with a valid token and calls
  `tools/list`
- **THEN** the response lists the two read-only tools

### Requirement: No authority over execution semantics

The MCP server MUST NOT alter Policy, Validator, Registry, Claims,
or Blast Radius. Hits are informational guidance for the external
agent; they MUST NOT grant capabilities or change execution
behavior of the Core.

#### Scenario: Guidance is informational

- **WHEN** an external agent receives a `failure` episode saying
  "avoid command X"
- **THEN** nothing in the Core changes; re-execution of X remains
  subject to Policy/Validator exactly as before
