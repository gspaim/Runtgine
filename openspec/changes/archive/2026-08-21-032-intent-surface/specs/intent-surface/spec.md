# intent-surface

## ADDED Requirements

### Requirement: Intent Surface Entry Point

The system SHALL provide a visual Entry Point named **INTENT** (Mission Brief)
where operators compile natural language or Task IR JSON into Runs without
calling Players directly.

#### Scenario: NL preview without submit

- **WHEN** the operator enters NL text and requests preview
- **THEN** the system calls `CompileIntent` and displays Task IR and `method`
- **AND** no Run is created

#### Scenario: NL submit creates run

- **WHEN** the operator submits NL text
- **THEN** the system calls `SubmitIntent` with source `tui` or `wails`
- **AND** the Run passes through the same Validator and `SubmitTask` as CLI
- **AND** the UI navigates to observe the Run (LIVE)

#### Scenario: Invalid capability rejected

- **WHEN** compiled Task IR references an unknown capability
- **THEN** Validator rejects before execution
- **AND** the UI shows a clear error without crashing

### Requirement: Not a chatbot

The Intent Surface SHALL NOT provide an unbounded conversational thread,
LLM prose replies as primary output, or transcript indexing in Project Memory.

#### Scenario: Single-shot compile semantics

- **WHEN** the operator uses INTENT
- **THEN** the primary output is Task IR preview and Run submission
- **AND** not free-form assistant chat

### Requirement: TUI INTENT tab

The TUI SHALL include INTENT as the first tab in order:
`INTENT RUNS LIVE BOARD EVENTS GRAPH CONFIG`.

#### Scenario: TUI keymap

- **WHEN** on the INTENT tab
- **THEN** `Ctrl+p` triggers preview and `Ctrl+Enter` triggers submit
- **AND** successful submit selects the run and switches to LIVE

### Requirement: Wails INTENT view

The Wails desktop (Fase 3) SHALL mirror TUI INTENT semantics using Core APIs
and shadcn-svelte components with Constellation tokens.

#### Scenario: Wails submit source

- **WHEN** submitting from Wails
- **THEN** `source` is `wails`
- **AND** behavior matches TUI aside from presentation
