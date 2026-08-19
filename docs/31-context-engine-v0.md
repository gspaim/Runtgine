# 31 — Context Engine v0

Nome do assembler do ContextPack: semear contexto **relevante** para
o LLM sem dump do repositório e sem embeddings.

Inventário: [10-gaps.md](10-gaps.md) (G-137+; recorte do HYPOTHESIS
em `02`). Autoridade: [04-decisoes.md](04-decisoes.md).
Corte de produto: [09-mvp.md](09-mvp.md) (MVP 1.0 magro).

Não é Player. Não é Intent Engine. Não é Graph Hits (`19`) nem
Project Memory (`29`) — **consome** os dois. Não é a API HTTP (G-45).

**Status deste doc: CONFIRMED (v0).** G-137..G-139 autorizam o
slice 20. G-140 (exclusões) vale para o 1.0 magro inteiro.

**Pacote OpenSpec:** ativo em
[`openspec/changes/031-mvp-1.0/`](../openspec/changes/031-mvp-1.0/).
Branch de implementação do código: após merge desta spec
(`feat/031-mvp-1.0` / slice 20).

---

## 1. Problema

`AssembleContext` (G-24) já monta task, step, `prior_outputs`,
`graph_hits`, `memory_hits` e `repo_hits`. `repo_hits` só existe se
`pipeline.repo-search` rodou **neste** Run.

Um step LLM fora do pipeline Board (ou um Completer de Intent) vê
Graph/Memory, mas **zero arquivos** se ninguém fez repo-search.
O “Context Engine” da visão (`02`) ainda era HYPOTHESIS por isso.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Função do Core sobre o ContextPack | Player `context.*` |
| Semente de `repo_hits` a partir do Graph | Walk de disco / `fs.read` no pack |
| Ranking = `QueryHits` existente | Embeddings / RAG |
| Degrada para `[]` | Falha o Run se o Graph falhar |

Regras:

1. Validator / Registry continuam soberanos.
2. Não ler corpo de arquivo para o pack no v0 (compor `fs.read` se a
   Task pedir).
3. Hierarquia de truncamento **inalterada**: identidade task/step >
   `repo_hits` > `graph_hits` > `memory_hits`.
4. Não substitui Graph Hits nem Memory; só preenche o buraco de
   `repo_hits` vazio.

---

## 3. Cortes confirmados

### G-137 — Papel

**Status: CONFIRMED**

- Context Engine v0 = o assembler (`AssembleContext` + semente)
- Pacote: permanecer em `internal/core/contextpack` (sem Player novo)
- Chamado pelo Runner (e pelo Intent no caminho LLM, se `repo_hits`
  vazio no pack do Completer)
- Recorte do HYPOTHESIS “Context Engine completo” em `02` / `05`

### G-138 — Semente quando `repo_hits` vazio

**Status: CONFIRMED**

Quando `repo_hits.paths` (e symbols) saírem vazios de
`extractRepoHits`:

1. Chamar `QueryHits` com o texto do `intent.summary` (mesmo budget
   `max_files` / `graph_max_hits` já existente)
2. Copiar itens `kind=path` → `repo_hits.paths` (id do nó = path)
3. Copiar itens `kind=symbol` → `repo_hits.symbols`
4. Respeitar `budget.max_files`
5. Erro ou Graph vazio → `repo_hits` continua vazio; Run segue

Se `pipeline.repo-search` já preencheu `repo_hits`, **não** sobrescrever.

### G-139 — Ranking / o que não entra no pack

**Status: CONFIRMED**

- Ranking = o de `QueryHits` (G-68). Sem score novo.
- Sem walk `filepath.Walk` do workspace.
- Sem embeddings, sem tokenização ML, sem corpo de arquivo.
- Sem stream de eventos históricos no pack (só `prior_outputs` do Run).

### G-140 — Exclusões do 1.0 magro

**Status: CONFIRMED** (como exclusões)

Compartilhado com heurísticas de Intent (`09` / G-135..G-136):

- G-45 API HTTP; G-44 MCP; G-36 NATS; Wails
- Player Router; Workflow Templates; Memory Player
- `http.post`, `git.push`, pytest/npm, Players de infra
- Context Engine “completo” (`02`: relevant events globais, current
  state rico, previous decisions além de `memory_hits`)

---

## 4. Criterios de aceite (slice 20)

1. Run **sem** `pipeline.repo-search`: se o Graph tem nós `path`,
   `repo_hits.paths` no pack do LLM não é vazio (até `max_files`).
2. Run **com** repo-search: `repo_hits` permanece o do pipeline
   (semente não pisa).
3. Graph indisponível / vazio → `repo_hits` `[]`; Run não falha.
4. Nenhuma leitura de conteúdo de arquivo no assembler.
5. `go test ./internal/core/contextpack/...` e `go test ./...` /
   `go vet ./...` verdes.
6. OpenSpec `031` arquivado **após** slices 19 e 20 (código), não
   neste PR de spec.

---

## 5. Ordem do slice de código

Bloqueado até G-135..G-140 CONFIRMED (este PR):

1. Slice 19 — Intent player heuristics (`17` / G-135..G-136)
2. Slice 20 — semente `repo_hits` neste doc
3. README Estágio: 1.0 magro feito; arquivar `031`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-137 Papel (assembler, não Player)
- [x] G-138 Semente QueryHits → `repo_hits`
- [x] G-139 Sem walk / embeddings / file body
- [x] G-140 Exclusões do 1.0 magro
