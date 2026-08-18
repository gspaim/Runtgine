# 18 — Runtime Graph v0

Memória **estrutural** do workspace: o que existe e como se relaciona.

Inventário: [10-gaps.md](10-gaps.md) (G-60+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Ortogonal a: Event Store (temporal, G-13) e Project Memory (episódica, `16`).

**Status deste doc: CONFIRMED (v0).** G-60..G-65 autorizam o corte deste
slice. Integração ContextPack/Intent (`graph_hits`) está em
[19-graph-hits-v0.md](19-graph-hits-v0.md) (**CONFIRMED** G-66..G-69).
Aba TUI GRAPH permanece fora.

Pré-requisitos de produto (ordem AGENTS): Core estável; Intent Engine v0
especificado/implementado (`17`). Graph v0 **não** depende de Intent Engine
para o corte mínimo estrutural; hits no ContextPack/Intent são o slice `19`.

---

## 1. Problema

Hoje o Core sabe:

| Fonte | Responde |
|---|---|
| Event Bus + SQLite events | O que **aconteceu** nesta/nas runs |
| Registry (in-process) | Quais Players/capabilities existem **neste processo** |
| ContextPack v0 | Contexto **intra-run** (task, priors, repo_hits) |

Falta uma visão persistente e consultável de **entidades e arestas** do
workspace: quais capabilities já rodaram, quais paths/símbolos o pipeline
já viu, como Runs se ligam a Tasks e a artefatos.

Sem isso:

- Intent Engine / Context Engine não têm mapa estrutural para enriquecer
  planejamento (só texto + priors da run);
- TUI não tem superfície “o que existe” (só LIVE/EVENTS);
- Workflow Templates (`08`) não têm onde se ancorar.

---

## 2. Fronteiras (não misturar)

| Memória | Papel | Status |
|---|---|---|
| Temporal | Event Store / SQLite — fatos de lifecycle | CONFIRMED |
| **Estrutural** | **Runtime Graph — nós e arestas** | **CONFIRMED (v0)** |
| Episódica | Project Memory — decisões/handoffs | HYPOTHESIS (`16`) |

Regras:

1. Graph **não** é event sourcing nem replay de eventos
2. Graph **não** substitui Registry (autoridade de capability em runtime)
3. Graph **não** autoriza execução (Validator continua soberano)
4. Graph **não** é Project Memory (sem “ainda válido?” / supersession)
5. Entry Points e TUI só leem/escrevem via Core API — nunca Player direto

---

## 3. Cortes confirmados (G-60+)

### G-60 — Papel e nome

**Status: CONFIRMED**

- Nome: Runtime Graph (não “Genome” no protocolo; Genome fica glossário/UI)
- Escopo: **um graph por `workspace_root`**, persistido em SQLite
- Pacote sugerido: `internal/core/graph`

### G-61 — Modelo de nós (v0)

**Status: CONFIRMED**

Tipos mínimos:

| `node_kind` | Chave estável | Origem típica |
|---|---|---|
| `player` | manifest `name` | Registry no boot / refresh |
| `capability` | `domain.action` | Registry |
| `task` | `task_id` | Submit / Store |
| `run` | `run_id` | Store |
| `path` | path relativo ao workspace | `pipeline.repo-search` / index leve |
| `symbol` | string do output de repo-search (`func Foo`); sem `path#name` no v0 | repo-search |

Payload JSON opcional por nó (`attrs`): versão do player, status da run,
`intent.summary`, etc. Sem schema rígido além de `kind` + `id`.

### G-62 — Modelo de arestas (v0)

**Status: CONFIRMED**

| `edge_kind` | from → to | Semântica |
|---|---|---|
| `provides` | player → capability | Manifest |
| `executed` | run → capability | Step executado |
| `instance_of` | run → task | Run desta Task |
| `mentions` | run\|task → path\|symbol | repo-search / ContextPack hits |
| `child_of` | run → run | `parent_run_id` (G-27) |

Arestas são direcionadas, idempotentes na chave
`(edge_kind, from_kind, from_id, to_kind, to_id)`.

### G-63 — Persistência

**Status: CONFIRMED** — variante B (mesmo SQLite do Core)

- Tabelas no DB existente (`workspace/.runtgine/runtgine.db`)
- Sem Neo4j / sem processo separado no v0
- Migração versionada junto ao Store
- Append/upsert; sem event sourcing do graph

Esboço:

```sql
CREATE TABLE graph_nodes (
  kind TEXT NOT NULL,
  id   TEXT NOT NULL,
  attrs_json TEXT NOT NULL DEFAULT '{}',
  updated_at TEXT NOT NULL,
  PRIMARY KEY (kind, id)
);

CREATE TABLE graph_edges (
  kind TEXT NOT NULL,
  from_kind TEXT NOT NULL,
  from_id TEXT NOT NULL,
  to_kind TEXT NOT NULL,
  to_id TEXT NOT NULL,
  attrs_json TEXT NOT NULL DEFAULT '{}',
  updated_at TEXT NOT NULL,
  PRIMARY KEY (kind, from_kind, from_id, to_kind, to_id)
);
```

### G-64 — Core API

**Status: CONFIRMED**

```text
UpsertNode(kind, id, attrs) -> error
UpsertEdge(kind, from, to, attrs) -> error
GetNode(kind, id) -> Node
QueryNeighbors(kind, id, edge_kind?, direction) -> []Node
GetGraphSnapshot(filter) -> GraphSnapshot   // TUI / CLI debug
RefreshFromRegistry() -> error              // players + capabilities
SyncFromRun(run_id) -> error                // após run terminal
```

- Escrita só pelo Core (Runner/Store hooks) + refresh explícito
- CLI: `runtgine graph snapshot` (read-only) no primeiro slice de código
- TUI tab dedicada: **fora do v0** (requer atualizar `14` + skill)

### G-65 — Quando popular

**Status: CONFIRMED**

| Momento | Ação |
|---|---|
| `api.Open` / boot | `RefreshFromRegistry` |
| Run `succeeded` / `failed` / `cancelled` | `SyncFromRun` (run, task, capabilities, child_of) |
| Step `pipeline.repo-search` succeeded | `mentions` path/symbol |
| Manual | `runtgine graph refresh` |

Não indexar o repo inteiro em background no v0 (custo/IO). Só o que o
pipeline ou refresh explícito já produziu.

### G-66 — Integração com ContextPack / Intent

**Status: CONFIRMED** — especificação completa em
[19-graph-hits-v0.md](19-graph-hits-v0.md) (G-66..G-69).

Este doc (`18`) cobre só o Graph estrutural (G-60..G-65). Hits,
`QueryHits`, budget e wiring Intent/AssembleContext vivem em `19`.

---

## 4. Fora do v0 (estrutural)

- Workflow Templates nativos no Graph (`08` / G-40)
- Genome completo / indexação AST contínua
- Multi-workspace / graph federado
- NATS / sync distribuído
- Aba TUI GRAPH (exige decisão em `14` + skill)
- Project Memory / validade operacional (`16`)
- Policies / Blast Radius derivados do graph (Claims v0 e Core, spec `24`; Blast v0 e Core, spec `25`; nenhum dos dois deriva do Graph)
- Substituir Registry ou Event Store

Hits no ContextPack/Intent: ver [19-graph-hits-v0.md](19-graph-hits-v0.md)
(não faz parte deste corte estrutural).

---

## 5. Critérios de aceite

1. Boot registra nós `player`/`capability` e arestas `provides`
2. Um `runtgine run` bem-sucedido cria/atualiza `run`, `task`, `executed`, `instance_of`
3. Pipeline com repo-search cria `mentions` para paths
4. `runtgine graph snapshot` imprime JSON estável (sem secrets)
5. Falha de graph **não** falha a Run (best-effort log; execução continua)
6. `go test ./...` cobre upsert idempotente e SyncFromRun

---

## 6. Ordem deste slice

1. G-60..G-65 CONFIRMED em `04` — feito
2. Migração SQLite + pacote `internal/core/graph` — feito
3. Hooks no Runner + `RefreshFromRegistry` em `api.Open` — feito
4. CLI `graph snapshot` / `graph refresh` — feito
5. Graph Hits (`19`, G-66..G-69) — próximo slice de código
6. Aba TUI GRAPH — só após decisão em `14` + skill

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-60 Papel / um graph por workspace
- [x] G-61 Node kinds v0
- [x] G-62 Edge kinds v0
- [x] G-63 SQLite no mesmo DB
- [x] G-64 Core API + CLI read-only
- [x] G-65 Momentos de sync (best-effort)
- [x] G-66+ ContextPack/Intent — ver `19` (CONFIRMED)
