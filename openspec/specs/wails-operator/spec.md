# wails-operator

Status: comportamento atual (pós-slice 40 / desktop Hits, Blast e RESULT).

## ADDED Requirements

### Requirement: Desktop stays Wails v3 with seven views

The desktop Entry Point SHALL remain `runtgine desktop` in
`internal/entrypoint/desktop` using Wails v3 and Svelte 5. It MUST NOT
replace the TUI, switch to Wails v2, or add an eighth view.

#### Scenario: Window still opens Mission Control

- **WHEN** the operator runs `runtgine desktop` with CGO and WebView
- **THEN** one window shows INTENT, RUNS, LIVE, BOARD, EVENTS, GRAPH,
  CONFIG in that order
- **AND** `source.entry_point` on submitted tasks remains `wails`

### Requirement: LIVE shows RESULT from step output

LIVE SHALL render a RESULT pane from the selected run's events:
stdout/stderr of `step.succeeded`, error text of `step.failed`, and
snapshot `error`. RESULT MUST appear above the raw event-type list.
The extractor SHALL be `liveout` (shared with the TUI). The desktop
MUST NOT call a Player to obtain stdout.

#### Scenario: Succeeded shell step

- **GIVEN** a run snapshot with `step.succeeded` and
  `payload.output.stdout` = `hello-runtgine`
- **WHEN** LIVE renders
- **THEN** RESULT includes `hello-runtgine`

#### Scenario: Still running

- **GIVEN** a run with status `running` and no step output yet
- **WHEN** LIVE renders
- **THEN** RESULT shows a waiting state, not a blank panel

### Requirement: Hits are inline, not a view

LIVE SHALL display `graph_hits`, `memory_hits`, and `playbook_hits`
from the selected run's ContextPack events. INTENT Preview SHALL
display `QueryHits` for the draft text. GRAPH MUST NOT render
QueryHits. There SHALL NOT be an eighth HITS view. QueryHits failure
MUST degrade to an empty list without failing the Task IR preview.

#### Scenario: LIVE empty hits

- **GIVEN** a run snapshot with no ContextPack hits
- **WHEN** LIVE renders
- **THEN** the hits pane shows `No hits.`

#### Scenario: INTENT preview hits

- **WHEN** the operator clicks Preview or presses Ctrl/Cmd+P on INTENT
- **THEN** Core `QueryHits` is invoked with the draft text
- **AND** the preview still shows the Task IR even if hits are empty

### Requirement: Blast panel without a BLAST view

INTENT Blast SHALL compile the draft and call `BlastTask` without
creating a Run. LIVE Blast SHALL call `BlastTask` on the selected
run's Task IR. The panel SHALL show `risk` with text/symbol plus
touches / conflicts / affected. GRAPH MUST NOT start a blast. There
SHALL NOT be an eighth BLAST view.

#### Scenario: INTENT blast does not submit

- **GIVEN** a valid INTENT draft
- **WHEN** the operator clicks Blast or presses Ctrl/Cmd+B
- **THEN** Core `BlastTask` is invoked
- **AND** `SubmitIntent` / `SubmitTask` are not invoked

#### Scenario: GRAPH has no blast control

- **GIVEN** GRAPH is the active view
- **THEN** no Blast button or shortcut starts `BlastTask`

### Requirement: Bindings without a display

Desktop `Service` SHALL expose `QueryHits` and `BlastTask` over the
existing Wails service. Unit tests MUST pass without a GUI.

#### Scenario: Fake Core blast

- **WHEN** `Service.BlastIntent` is invoked in tests
- **THEN** the fake Core records the `BlastTask` call
- **AND** no Run is inserted
