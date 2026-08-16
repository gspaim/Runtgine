# 18 — Runtime Graph v0 (PROPOSTA)

Memória **estrutural** do workspace: o que existe e como se relaciona.

Inventário: [10-gaps.md](10-gaps.md) (G-60+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Ortogonal a: Event Store (temporal, G-13) e Project Memory (episódica, `16`).

**Status deste doc: PROPOSED.** Não autoriza código até promoção
explícita em `04-decisoes.md` (HYPOTHESIS → CONFIRMED v0).

Pré-requisitos de produto (ordem AGENTS): Core estável; Intent Engine v0
especificado/implementado (`17` / PR aberto). Graph v0 **não** depende de
Intent Engine para o corte mínimo, mas o Intent não consulta o Graph neste
slice.

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
| **Estrutural** | **Runtime Graph — nós e arestas** | **HYPOTHESIS → este doc** |
| Episódica | Project Memory — decisões/handoffs | HYPOTHESIS (`16`) |

Regras:

1. Graph **não** é event sourcing nem replay de eventos
2. Graph **não** substitui Registry (autoridade de capability em runtime)
3. Graph **não** autoriza execução (Validator continua soberano)
4. Graph **não** é Project Memory (sem “ainda válido?” / supersession)
5. Entry Points e TUI só leem/escrevem via Core API — nunca Player direto

---

## 3. Cortes propostos (G-60+)

### G-60 — Papel e nome

**Proposta: CONFIRMED v0 após revisão**

- Nome: Runtime Graph (não “Genome” no protocolo; Genome fica glossário/UI)
- Escopo: **um graph por `workspace_root`**, persistido em SQLite
- Pacote sugerido: `internal/core/graph`

### G-61 — Modelo de nós (v0)

**Proposta**

Tipos mínimos:

| `node_kind` | Chave estável | Origem típica |
|---|---|---|
| `player` | manifest `name` | Registry no boot / refresh |
| `capability` | `domain.action` | Registry |
| `task` | `task_id` | Submit / Store |
| `run` | `run_id` | Store |
| `path` | path relativo ao workspace | `pipeline.repo-search` / index leve |
| `symbol` | `path#name` ou `name@path` | repo-search |

Payload JSON opcional por nó (`attrs`): versão do player, status da run,
`intent.summary`, etc. Sem schema rígido além de `kind` + `id`.

### G-62 — Modelo de arestas (v0)

**Proposta**

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

**Proposta: variante B (mesmo SQLite do Core)**

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

**Proposta**

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

**Proposta**

| Momento | Ação |
|---|---|
| `api.Open` / boot | `RefreshFromRegistry` |
| Run `succeeded` / `failed` / `cancelled` | `SyncFromRun` (run, task, capabilities, child_of) |
| Step `pipeline.repo-search` succeeded | `mentions` path/symbol |
| Manual | `runtgine graph refresh` |

Não indexar o repo inteiro em background no v0 (custo/IO). Só o que o
pipeline ou refresh explícito já produziu.

### G-66 — Integração com ContextPack / Intent

**Proposta: DEFERRED neste corte**

- ContextPack v0 permanece como está (G-24)
- Intent Engine v0 **não** consulta o Graph (`17` já exclui)
- Extensão futura: `graph_hits` no ContextPack + budget — só após CONFIRMED
  e um slice de código estável do graph

---

## 4. Fora do v0

- Workflow Templates nativos no Graph (`08` / G-40)
- Genome completo / indexação AST contínua
- Multi-workspace / graph federado
- NATS / sync distribuído
- Aba TUI GRAPH (exige decisão em `14`)
- Project Memory / validade operacional (`16`)
- Policies / Blast Radius derivados do graph
- Substituir Registry ou Event Store

---

## 5. Critérios de aceite (após CONFIRMED + código)

1. Boot registra nós `player`/`capability` e arestas `provides`
2. Um `runtgine run` bem-sucedido cria/atualiza `run`, `task`, `executed`, `instance_of`
3. Pipeline com repo-search cria `mentions` para paths
4. `runtgine graph snapshot` imprime JSON estável (sem secrets)
5. Falha de graph **não** falha a Run (best-effort log; execução continua)
6. `go test ./...` cobre upsert idempotente e SyncFromRun

---

## 6. Ordem sugerida pós-confirmação

1. Promover G-60..G-65 em `04` (G-66 DEFERRED)
2. Migração SQLite + pacote `internal/core/graph`
3. Hooks no Runner + `RefreshFromRegistry` em `api.Open`
4. CLI `graph snapshot` / `graph refresh`
5. Só então discutir ContextPack/`graph_hits` e TUI

---

## Checklist de confirmação humana

Marcar em `04-decisoes.md` após revisão:

- [ ] G-60 Papel / um graph por workspace
- [ ] G-61 Node kinds v0
- [ ] G-62 Edge kinds v0
- [ ] G-63 SQLite no mesmo DB
- [ ] G-64 Core API + CLI read-only
- [ ] G-65 Momentos de sync (best-effort)
- [ ] G-66 ContextPack/Intent DEFERRED

**Até lá: HYPOTHESIS. Não codificar.**
