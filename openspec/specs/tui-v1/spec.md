# tui-v1

## ADDED Requirements

### Requirement: TUI v1 stays on Charm v2

The TUI SHALL remain a Go Bubble Tea v2 application in
`internal/entrypoint/tui` using Lip Gloss v2 and Bubbles v2. It MUST
NOT switch to Ratatui, Textual, or replace the TUI with Wails.

#### Scenario: Binary still opens Mission Control

- **WHEN** the operator runs `runtgine tui`
- **THEN** the process is the Charm Constellation Mission Control
- **AND** the seven tabs remain INTENT, RUNS, LIVE, BOARD, EVENTS,
  GRAPH, CONFIG in that order

### Requirement: Professional Bubbles chrome

RUNS SHALL render with a Bubbles `table`. INTENT draft SHALL use a
Bubbles `textarea`. Detail/JSON/telemetry panes SHALL use a `viewport`
or `list`. Pressing `?` SHALL toggle a `help` overlay.

#### Scenario: Help overlay

- **GIVEN** the TUI is on any tab
- **WHEN** the operator presses `?`
- **THEN** a help overlay lists the active keymap
- **AND** pressing `?` or `esc` closes it

#### Scenario: INTENT textarea

- **GIVEN** the INTENT tab is active
- **WHEN** the operator types multiple lines
- **THEN** the draft is held by the textarea component
- **AND** `Ctrl+p` still previews via `CompileIntent`

### Requirement: Hits are inline, not a tab

LIVE SHALL display `graph_hits`, `memory_hits`, and `playbook_hits`
from the selected run's ContextPack events. INTENT preview (`Ctrl+p`)
SHALL display `QueryHits` for the draft text. GRAPH MUST NOT render
QueryHits. There SHALL NOT be an eighth HITS tab. QueryHits failure
MUST degrade to an empty list.

#### Scenario: LIVE empty hits

- **GIVEN** a run snapshot with no ContextPack hits
- **WHEN** LIVE renders
- **THEN** the hits pane shows `No hits.`

#### Scenario: INTENT preview hits

- **WHEN** the operator presses `Ctrl+p` on INTENT
- **THEN** Core `QueryHits` is invoked with the draft text
- **AND** the preview still shows the Task IR even if hits are empty

### Requirement: Blast drawer without a BLAST tab

INTENT `Ctrl+b` SHALL compile the draft and call `BlastTask` without
creating a Run. LIVE `b` SHALL call `BlastTask` on the selected run's
Task IR. The drawer SHALL show `risk` with text/symbol plus touches /
conflicts / affected. GRAPH MUST NOT start a blast. There SHALL NOT be
an eighth BLAST tab.

#### Scenario: INTENT blast does not submit

- **GIVEN** a valid INTENT draft
- **WHEN** the operator presses `Ctrl+b`
- **THEN** Core `BlastTask` is invoked
- **AND** `SubmitIntent` / `SubmitTask` are not invoked

#### Scenario: GRAPH ignores blast keys

- **GIVEN** GRAPH is the active tab
- **WHEN** the operator presses `b` or `Ctrl+b`
- **THEN** `BlastTask` is not invoked

### Requirement: Narrow terminals and NO_COLOR

Width `< 80` MUST stack Hits/Blast vertically without panicking.
Status and blast `risk` MUST remain readable as text when `NO_COLOR`
is set.

#### Scenario: Width 70

- **GIVEN** width 70
- **WHEN** LIVE or INTENT renders with the Hits or Blast pane open
- **THEN** the view does not panic
