# Design: 019-graph-hits

## Technical approach

### QueryHits (`internal/core/graph`)

```text
QueryHits(ctx, Query) -> Hits
```

`Query`: `Text`, `SeedPaths`, `SeedSymbols`, `SeedCapability`, `Limit`.

Algoritmo determinístico (scores): seed=10, capability node=8, mentions=5,
executed-run neighbors=4, keyword=2. Dedup por `(kind,id)` mantém maior
score; ordenar score desc, depois kind, id; aplicar `Limit` e
`graph_max_chars`.

Stopwords mínimas fixas (PT/EN) como em `docs/19` §G-68.

Erros de store → `Hits{}` + `slog`; nunca propagar como falha de Run.

### ContextPack (`internal/core/contextpack`)

Novos campos:

```go
GraphHits GraphHits `json:"graph_hits"`
// Budget gains GraphMaxHits, GraphMaxChars
```

`Assemble` (ou helper `WithGraphHits`) preenche a partir de `Hits`.
Hierarquia de truncamento: task/step → prior_outputs → repo_hits →
graph_hits (corta score baixo primeiro).

### Runner

Antes de `Complete` em steps LLM:

1. Montar pack atual (priors + repo_hits)
2. `QueryHits` com summary/notes, repo seeds, capability do step
3. Anexar `graph_hits`

Players deterministicos ignoram o campo.

### Intent Engine

| Path | Graph |
|---|---|
| `heuristic.shell` / `heuristic.pipeline` | não chama `QueryHits` |
| `llm` | Completer recebe pack com `QueryHits(Text=NL)` |

Task IR impressa em `--dry-run` **não** embute graph_hits.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Ranking por embedding/LLM | Viola deterministic-first neste slice |
| Mudar heuristicas Intent via histórico | Escopo G-69 explícito: só caminho LLM |
| Novo node_kind | Desnecessário; reusa G-61 |

## Risks

- Graph vazio → hits sempre vazios até haver sync (aceitável; degradação)
- Pack JSON maior → budget `graph_max_chars` mitiga
- Drift docs/19 vs este design → `docs/19` permanece canônico de produto;
  este arquivo é o “how” de implementação

## Packages touched

- `internal/core/graph` (+ tests)
- `internal/core/contextpack` (+ tests)
- `internal/core/runner`
- `internal/core/intent` (+ tests heuristic sem QueryHits)
