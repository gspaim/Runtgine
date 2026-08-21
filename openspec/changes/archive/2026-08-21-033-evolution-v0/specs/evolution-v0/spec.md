# evolution-v0

## ADDED Requirements

### Requirement: Player Router multi-model

The system SHALL route LLM steps to a configured provider and model based on
capability, effort, and difficulty signals — without introducing Agent personas.

#### Scenario: Route by capability prefix

- **WHEN** a step uses `pipeline.spec-review` and a routing rule matches that prefix
- **THEN** the LLM Player uses the configured provider and model for that rule

#### Scenario: Fallback to default provider

- **WHEN** no routing rule matches
- **THEN** the system uses the workspace default LLM provider

### Requirement: Workspace playbooks

The system SHALL load project playbooks from `.runtgine/playbooks/` and expose
relevant excerpts in the ContextPack as `playbook_hits`.

#### Scenario: Playbook indexed by capability

- **WHEN** a playbook declares `capabilities: [test.go]`
- **AND** a step uses `test.go`
- **THEN** capped playbook content MAY appear in `playbook_hits`

### Requirement: Lessons postmortem with HITL

The system SHALL analyze failed runs and create improvement **proposals** that
require explicit human approval before updating Project Memory or playbooks.

#### Scenario: Failure capture creates proposal only

- **WHEN** `lessons.capture=failures` and a run fails
- **THEN** a lesson proposal is stored in pending state
- **AND** Project Memory is not updated until approval

#### Scenario: Approved lesson becomes memory

- **WHEN** an operator approves a lesson proposal
- **THEN** an episodic memory record is created with kind appropriate to the proposal
- **AND** future ContextPacks MAY include it in `memory_hits`

### Requirement: Not an agent framework

The evolution features SHALL NOT implement autonomous agents, conversational
persona registries, or silent LLM mutation of authoritative project state.

#### Scenario: No silent playbook write

- **WHEN** a lesson is generated
- **THEN** playbooks are not modified without explicit human approval
