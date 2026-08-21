# wails-v0

## ADDED Requirements

### Requirement: Desktop Entry Point adapter

The system SHALL expose the existing Core API through a Wails v2 desktop
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

### Requirement: Wails v2 pin

The desktop v0 SHALL use Wails v2 (stable) and Svelte 5 + shadcn-svelte
with Constellation tokens.

#### Scenario: Wails v3 out of v0

- **WHEN** choosing the Wails major version
- **THEN** the implementation uses v2
- **AND** Wails v3 remains out of this cut

### Requirement: Seven views

The desktop SHALL offer the same seven views as the TUI, with INTENT first:
`INTENT RUNS LIVE BOARD EVENTS GRAPH CONFIG`.

#### Scenario: Order

- **WHEN** the app opens
- **THEN** INTENT is the default view

### Requirement: Thin Core bindings

Wails backend methods SHALL be a thin facade over `api.Core`. The desktop
MUST NOT reimplement the Event Bus, Validator, or Player dispatch.

#### Scenario: Preview uses CompileIntent

- **WHEN** the operator requests preview
- **THEN** the binding calls `CompileIntent` only
- **AND** unit tests use a fake Core without a display

### Requirement: Not HTTP client and not Player

The desktop SHALL NOT call `runtgine serve` as a client in v0 and SHALL
NOT invoke Players directly.

#### Scenario: In-process Core

- **WHEN** `runtgine desktop` starts
- **THEN** it opens `api.Core` in-process for one `workspace_root`
- **AND** Validator remains sovereign on submit
