# wails-v0

Status: comportamento atual (pós-slices 27–28 / `runtgine desktop`).

## Requirements

### Requirement: Desktop Entry Point adapter

The system SHALL expose the existing Core API through a Wails v3 desktop
app as an Entry Point. The app MUST NOT call Players directly and MUST
NOT be an HTTP client of `runtgine serve`.

#### Scenario: Submit from INTENT

- **WHEN** the operator submits NL from the INTENT view
- **THEN** the app calls `SubmitIntent` with `source.entry_point` `wails`
- **AND** the UI navigates to LIVE with the returned `run_id`

#### Scenario: Preview does not create a Run

- **WHEN** the operator previews `git status`
- **THEN** the app calls `CompileIntent` only
- **AND** no Run is inserted

### Requirement: Wails v3 pin

The desktop v0 SHALL use Wails v3 (`github.com/wailsapp/wails/v3`) and
Svelte 5 with Constellation tokens. Wails v2 SHALL NOT be used.

#### Scenario: Major version

- **WHEN** choosing the Wails major version
- **THEN** the implementation uses v3
- **AND** the exact beta tag is pinned in `go.mod`
- **AND** Wails v2 remains out of this cut

### Requirement: Seven views

The desktop SHALL offer the same seven views as the TUI, with INTENT first:
`INTENT RUNS LIVE BOARD EVENTS GRAPH CONFIG`.

#### Scenario: Order

- **WHEN** the app opens
- **THEN** INTENT is the default view
- **AND** a single window is used (v3 multi-window unused)

#### Scenario: Remaining views

- **WHEN** the operator opens RUNS, BOARD, EVENTS, GRAPH, or CONFIG
- **THEN** each view is populated from the Core service (not placeholders)
- **AND** GRAPH is a read-only `GetGraphSnapshot`
- **AND** BOARD shows only runs with `source` `board` (display-only; no poll from desktop)
- **AND** selecting a run on RUNS or BOARD opens LIVE

### Requirement: Thin Core service

Wails v3 service methods SHALL be a thin facade over `api.Core`. The
desktop MUST NOT reimplement the Event Bus, Validator, or Player dispatch.

#### Scenario: Preview uses CompileIntent

- **WHEN** the operator requests preview
- **THEN** the service calls `CompileIntent` only
- **AND** unit tests use a fake Core without a display

### Requirement: CONFIG hides secrets

CONFIG SHALL render `ConfigSnapshot` only. It MUST NOT print `api.token`,
LLM API keys, or GitHub tokens.

#### Scenario: Snapshot fields

- **WHEN** CONFIG loads
- **THEN** credentials appear only as connected / not configured booleans
- **AND** the JSON of `ConfigSnapshot` has no token or key fields

### Requirement: Lessons HITL

CONFIG SHALL list Lessons and allow approve / reject via
`ListLessons` / `ApproveLesson` / `RejectLesson`.

#### Scenario: Approve pending lesson

- **WHEN** the operator approves a pending lesson
- **THEN** the service calls `ApproveLesson`
- **AND** the list refreshes

### Requirement: Not HTTP client and not Player

The desktop SHALL NOT call `runtgine serve` as a client in v0 and SHALL
NOT invoke Players directly.

#### Scenario: In-process Core

- **WHEN** `runtgine desktop` starts
- **THEN** it opens `api.Core` in-process for one `workspace_root`
- **AND** Validator remains sovereign on submit
