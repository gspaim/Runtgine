# Delta for Runtime Graph

## ADDED Requirements

### Requirement: QueryHits ranked structural search

The Graph Service SHALL expose `QueryHits(ctx, Query) -> Hits` that returns
deduplicated, score-ranked hits from existing nodes/edges without writing
new kinds.

#### Scenario: Seed path hit
- GIVEN a `path` node already in the graph
- AND Query.SeedPaths contains that path
- WHEN QueryHits runs
- THEN an item with `kind=path`, `reason=seed`, and score ≥ 10 is returned

#### Scenario: Keyword match
- GIVEN a capability node id containing token `review` (len ≥ 3)
- AND Query.Text includes `review`
- WHEN QueryHits runs with Limit ≥ 1
- THEN a `keyword` reason hit for that capability MAY appear with score 2

#### Scenario: Best-effort errors
- GIVEN the store returns an error during QueryHits
- WHEN the caller is Runner or Intent
- THEN Hits are empty and execution continues

## MODIFIED Requirements

### Requirement: No ContextPack integration in structural v0

(Previously: structural v0 does not inject hits into ContextPack/Intent.)

After `019-graph-hits`, QueryHits MAY be consumed by ContextPack assembly and
Intent LLM path. Structural sync hooks (G-65) remain unchanged.
