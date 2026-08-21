# Design: 033-evolution-v0

## Slice 22 — Player Router

### Config model

Extend `internal/config.Config`:

- `LLMProviders []ProviderConfig` — id, kind, base_url, default_model
- `LLMRouting []RoutingRule` — match predicates + provider_id + optional model_id
- Keep `LLMBackend` as deprecated alias → default provider id

### Router package

`internal/core/router` (or extend existing selection in Runner):

```go
type RouteRequest struct {
    Capability string
    Effort     string // from prior pipeline.effort if any
    Difficulty int    // from prior pipeline.difficulty if any
}

type RouteResult struct {
    ProviderID string
    ModelID    string
}

func SelectLLM(cfg config.Config, req RouteRequest) RouteResult
```

Runner / LLM Player calls `SelectLLM` before `CompleterFromConfig`.

### Match precedence

1. Exact capability match
2. `capability_prefix`
3. `difficulty_gte` / `effort_in`
4. Default provider

First matching rule in config order wins.

### Tests

- Table-driven rules; fallback when no match
- Legacy `llm_backend` still works

---

## Slice 23 — Playbooks

### Layout

```text
.runtgine/playbooks/
  orchestrator.md
  developer.md
  qa.md
```

Frontmatter (YAML):

```yaml
---
id: qa
title: QA playbook
capabilities: [test.go, pipeline.spec-review]
---
```

### Loader

`internal/core/playbooks` — scan on `api.Open`, index by id and capability.

### ContextPack

Add optional `playbook_hits` (similar to `memory_hits`):

- Budget: default 2 hits / 1500 chars
- Selected when step capability intersects playbook.capabilities
- Hierarchy: after `repo_hits`, before or alongside `memory_hits` (doc 33)

Intent Engine may suggest playbook id in compile metadata (future); v0 =
Context Engine only.

---

## Slice 24 — Lessons

### Config

```json
{ "lessons": { "capture": "off" } }
```

Values: `off` | `failures` (mirror `memory.capture` pattern).

### Flow

1. On `run.failed`, if `lessons.capture=failures`, enqueue proposal (SQLite table
   `lesson_proposals`: id, run_id, payload JSON, status pending|approved|rejected).
2. Analyzer: deterministic extraction (failed step, error code, stderr tail) +
   optional `pipeline.postmortem` LLM step with fixed schema.
3. CLI: `runtgine lessons list|approve|reject <id>`
4. Approve → `memory.Record` + optional write playbook patch file to
   `.runtgine/playbooks/pending/<id>.patch` (human applies merge).

### HITL

No auto-write to playbooks or memory without explicit approve API.

---

## Alternatives rejected

| Alt | Why not |
|---|---|
| Agent registry with personas | Conflicts Player-centric model |
| Auto-apply LLM playbook edits | REJECTED in `29`/`16` |
| External benchmark API in Core | Human-curated routing table first |
| Memory Player for lessons | Provider + HITL API sufficient for v0 |

## Packages touched (future slices)

- `internal/config`
- `internal/core/router` (new)
- `internal/core/playbooks` (new)
- `internal/core/lessons` (new)
- `internal/core/contextpack`
- `internal/players/llm`
- `cmd/runtgine` (lessons subcommand)
