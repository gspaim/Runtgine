# Delta for blast-radius

## MODIFIED Requirements

### Requirement: CLI surface only

The CLI SHALL expose `runtgine blast` and print the full Impact
Report including `affected`. The Runner MUST NOT auto-blast on
`runtgine run`. The TUI MUST NOT add a Blast tab or trigger walk
from GRAPH (spec `26` remains a structural snapshot).

#### Scenario: run unchanged

- GIVEN `examples/hello.json`
- WHEN `runtgine run` executes
- THEN no blast report is required for `run.succeeded`

#### Scenario: JSON includes affected

- GIVEN any valid Task IR
- WHEN `runtgine blast` prints JSON
- THEN the object has an `affected` array (possibly empty)
