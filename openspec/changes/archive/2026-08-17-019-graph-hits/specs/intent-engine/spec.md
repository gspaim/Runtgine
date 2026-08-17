# Delta for Intent Engine

## ADDED Requirements

### Requirement: LLM compile path consumes QueryHits

When the Intent Engine uses the LLM Completer path, it SHALL call
`QueryHits` with the NL text and attach results as `graph_hits` on the
Completer ContextPack.

#### Scenario: LLM path queries graph
- GIVEN NL text that does not match shell or pipeline heuristics
- AND the Completer is invoked
- WHEN compile runs
- THEN QueryHits is called with Text set to the NL input

#### Scenario: Graph failure still compiles
- GIVEN QueryHits returns an empty result due to store error
- WHEN compile LLM runs
- THEN a Task IR is still produced when the Completer succeeds

## MODIFIED Requirements

### Requirement: LLM path uses ContextPack without graph

(Previously: LLM Completer ContextPack does not query the Runtime Graph.)

The LLM Completer ContextPack SHALL include `graph_hits` from QueryHits.
Heuristic shell and pipeline paths MUST NOT call QueryHits.

#### Scenario: Heuristic shell skips graph
- GIVEN text `echo hello-intent`
- WHEN CompileIntent runs
- THEN method is `heuristic.shell`
- AND QueryHits is not invoked
