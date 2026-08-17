# 19 — Graph Hits v0 (G-66+)

Integração do Runtime Graph com ContextPack e Intent Engine:
hits estruturais consultáveis, com budget e degradação.

Inventário: [10-gaps.md](10-gaps.md) (G-66+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Runtime Graph v0 ([18-runtime-graph-v0.md](18-runtime-graph-v0.md),
G-60..G-65) **implementado e estável**.

**Status deste doc: CONFIRMED (v0).** G-66..G-69 autorizam o slice de
código seguinte. Aba TUI GRAPH permanece fora (exige `14` + skill).

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-17-019-graph-hits/`](../openspec/changes/archive/2026-08-17-019-graph-hits/).
Deltas mergeados em `openspec/specs/`. Branch de implementação:
`cursor/019-graph-hits-0ac1` (equiv. `feat/019-graph-hits`).

Ortogonal a: Project Memory / `memory_hits` ([16-project-memory.md](16-project-memory.md)
— HYPOTHESIS; não misturar).

---

## 1. Problema

O Graph v0 já persiste nós/arestas e expõe `snapshot` / `refresh`.
O ContextPack v0 (G-24) monta só contexto **intra-run**:

| Campo | Origem |
|---|---|
| `task` / `step` | Task IR atual |
| `prior_outputs` | steps do mesmo run |
| `repo_hits` | output de `pipeline.repo-search` neste run |
| `budget` | max_chars / max_files |

O Intent Engine (G-53) monta um ContextPack mínimo para o Completer
**sem** olhar o Graph.

Falta:

- LLM Player / Intent enriquecerem com mapa estrutural **entre runs**
  (paths/símbolos já mencionados, capabilities já executadas, runs
  relacionadas);
- Um contrato versionado de hits + budget, separado de `repo_hits`
  (intra-run) e de futuro `memory_hits` (episódico).

---

## 2. Fronteiras (não misturar)

| Fonte | Papel neste slice |
|---|---|
| `repo_hits` | Hits **desta** run (repo-search) — inalterado |
| **`graph_hits`** | Hits **estruturais** do Graph do workspace — **este doc** |
| `memory_hits` | Episódicos / Project Memory — **fora** (HYPOTHESIS em `16`) |

Regras:

1. Graph **não** autoriza execução (Validator / Registry soberanos)
2. Graph **não** bypassa heuristicas do Intent (deterministic-first)
3. Falha ao consultar Graph → `graph_hits` vazio; run/compile continua
4. Sem aba TUI GRAPH neste slice
5. Sem indexação full-repo nova — só o que G-65 já sincronizou

---

## 3. Cortes confirmados (G-66+)

### G-66 — Papel do slice

**Status: CONFIRMED** (promovido de DEFERRED em `18`)

Nome curto: **Graph Hits v0**.

Escopo:

1. Estender ContextPack com `graph_hits` + budget dedicado
2. API Core de consulta ranqueada (`QueryHits`)
3. Intent Engine (caminho LLM) e `AssembleContext` consomem `QueryHits`
4. Heurísticas shell\|pipeline do Intent **não** mudam por Graph

Fora: TUI tab, Workflow Templates no Graph, Policies/Blast Radius,
Project Memory, novos node/edge kinds.

### G-67 — Schema `graph_hits` + budget

**Status: CONFIRMED**

Extensão do ContextPack (JSON estável, sem secrets):

```json
{
  "task": { "task_id": "…", "summary": "…", "notes": "…" },
  "step": { "step_id": "…", "capability": "…" },
  "prior_outputs": [],
  "repo_hits": { "paths": [], "symbols": [] },
  "graph_hits": {
    "items": [
      {
        "kind": "path",
        "id": "internal/core/graph/graph.go",
        "reason": "mentions",
        "score": 3
      }
    ]
  },
  "budget": {
    "max_chars": 12000,
    "max_files": 40,
    "graph_max_hits": 20,
    "graph_max_chars": 4000
  }
}
```

| Campo | Regra |
|---|---|
| `graph_hits.items[].kind` | Um de: `path`, `symbol`, `capability`, `run`, `task` |
| `graph_hits.items[].id` | Chave estável do nó (igual G-61) |
| `reason` | Por que entrou: `seed`, `mentions`, `executed`, `instance_of`, `child_of`, `keyword` |
| `score` | Inteiro ≥ 0; maior = mais relevante; desempate por `kind`+`id` lexicográfico |
| `budget.graph_max_hits` | Default **20** |
| `budget.graph_max_chars` | Default **4000** (serialização JSON dos items após corte) |

`repo_hits` e defaults `max_chars` / `max_files` **permanecem** (G-24).

Hierarquia de truncamento determinística (quando o pack excede budget
global de chars):

1. `task` + `step` (sempre)
2. `prior_outputs` (já truncados hoje)
3. `repo_hits`
4. `graph_hits` (corta score baixo primeiro)
5. Futuro `memory_hits` — **não** neste slice

### G-68 — API `QueryHits`

**Status: CONFIRMED**

Pacote: `internal/core/graph` (estende o Service de `18`).

```text
QueryHits(ctx, Query) -> Hits
```

```text
Query:
  Text       string     // NL / summary (tokens simples)
  SeedPaths  []string   // paths já conhecidos (ex.: repo_hits)
  SeedSymbols []string
  SeedCapability string // capability do step atual (opcional)
  Limit      int        // default = budget.graph_max_hits

Hits:
  Items []Hit  // ordenados por score desc, depois kind, id
```

Algoritmo v0 (determinístico, sem LLM):

1. **Seeds diretos** — cada `SeedPaths` / `SeedSymbols` que existir como
   nó no Graph → hit `reason=seed`, score base 10
2. **Vizinhos `mentions`** — de runs/tasks que mencionam esses paths/symbols
   → path/symbol adicionais com `reason=mentions`, score 5
3. **Capability** — se `SeedCapability` ≠ "" e nó existe → hit
   `reason=executed` candidates: até N runs recentes com aresta
   `executed` para essa capability (attrs/`updated_at` do nó run),
   score 4; a própria capability score 8
4. **Keywords** — tokens do `Text` (lowercase, len ≥ 3, sem stopwords
   mínimas) que casam substring em `id` de nós `path`/`symbol`/
   `capability` → `reason=keyword`, score 2
5. Deduplicar por `(kind, id)` mantendo o **maior** score
6. Aplicar `Limit`; serializar e cortar por `graph_max_chars` se
   necessário (remove do fim da lista ordenada)

Stopwords mínimas v0 (PT/EN): `the`, `and`, `for`, `com`, `para`,
`uma`, `que`, `run`, `task`. Lista fixa no código; sem stemming.

Falha de I/O / SQL → retornar `Hits{}` + log; **nunca** erro fatal
para Runner / Intent.

Sem CLI nova obrigatória: `runtgine graph snapshot` continua; opcional
depois `graph hits --text "…"` (fora do aceite mínimo).

### G-69 — Intent Engine e AssembleContext

**Status: CONFIRMED**

#### AssembleContext (Runner → LLM Player)

Antes de `Complete`, o Core:

1. Monta pack como hoje (`task`, `step`, `prior_outputs`, `repo_hits`)
2. Chama `QueryHits` com:
   - `Text` = `intent.summary` (+ notes se curtas)
   - `SeedPaths` / `SeedSymbols` = `repo_hits`
   - `SeedCapability` = capability do step
3. Preenche `graph_hits` + budget defaults
4. Segue truncamento da hierarquia G-67

Players deterministicos **ignoram** `graph_hits` (sem mudança de contrato
de input_schema).

#### Intent Engine

| Caminho | Graph |
|---|---|
| `heuristic.shell` / `heuristic.pipeline` | **Não** consulta (ordem G-52 intacta) |
| `llm` (G-53) | Completer recebe ContextPack com `graph_hits` via `QueryHits(Text=NL)` |

- Intent **não** ganha capabilities novas do Graph
- Intent **não** escolhe route só porque o Graph tem histórico
- Saída continua Task IR → Validator → Registry

Assinatura sugerida (interna):

```text
Assemble(task, stepID, capability, priors, hits Query) Pack
```

ou `Assemble` + `WithGraphHits(pack, hits)` — detalhe de código livre desde
que o JSON do pack obedeça G-67.

`CompileIntent` / `SubmitIntent` (G-51) **sem** novos flags CLI neste
corte. `--dry-run` passa a poder incluir `graph_hits` no pack interno
só se o caminho LLM for usado; a Task IR impressa **não** embute
graph_hits (Task IR ≠ ContextPack).

---

## 4. Fora do v0

- Aba TUI GRAPH (`14` + skill)
- CLI `graph hits` (nice-to-have)
- `memory_hits` / Project Memory (`16`)
- Novos `node_kind` / `edge_kind`
- Ranking por embedding / LLM
- Policies, Blast Radius, Claims derivados do Graph
- Mudança de heuristicas Intent por histórico de runs
- Indexação background do repositório

---

## 5. Critérios de aceite

1. Após pelo menos um pipeline com `repo-search` sincronizado no Graph,
   um step LLM seguinte no **mesmo workspace** recebe `graph_hits.items`
   não vazio com `kind=path` (ou symbol) quando os seeds batem
2. `budget.graph_max_hits` / `graph_max_chars` respeitados (teste unitário)
3. Intent `heuristic.shell` **não** chama `QueryHits`
4. Intent caminho LLM chama `QueryHits`; falha do Graph → pack sem items,
   compile ainda produz Task IR válida
5. Graph Hits **não** altera Validator/Registry; capability inventada
   continua rejeitada
6. `go test ./...` cobre: ranking/dedup de `QueryHits`, Assemble com
   hits, Intent heuristic sem graph
7. Documentação cruzada: `04`, `10`, `12` (G-24 nota), `17`, `18`, README

---

## 6. Ordem deste slice (código, após merge desta spec)

1. G-66..G-69 CONFIRMED em `04` — feito
2. Pacote OpenSpec `openspec/changes/019-graph-hits/` — feito (este ciclo)
3. `QueryHits` + testes em `internal/core/graph`
4. Extender `contextpack.Pack` / `Assemble`
5. Wire no Runner (LLM steps) e Intent caminho LLM
6. Atualizar estágio README (Feito = Slice 7); arquivar change OpenSpec

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-66 Papel Graph Hits (promovido de DEFERRED)
- [x] G-67 Schema `graph_hits` + budget
- [x] G-68 API `QueryHits` determinística
- [x] G-69 Intent LLM + AssembleContext consomem hits
- [x] TUI GRAPH / memory_hits / CLI hits fora do v0
- [x] Pacote `openspec/changes/019-graph-hits/` criado
