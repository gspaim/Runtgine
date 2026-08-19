# 29 — Project Memory v0

Memória **episódica** do workspace: decisões, falhas, handoffs e
preferências **entre runs**, consultada pelo ContextPack. Não é Event
Store, não é Runtime Graph, não é RAG.

Inventário: [10-gaps.md](10-gaps.md) (G-123+; recorte de G-46/G-47).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Esboco conceitual: [16-project-memory.md](16-project-memory.md)
(`16` permanece histórico; este doc é o corte).
Não é MCP ([G-44](10-gaps.md)). Não é Memory Player. Não é Knowledge.

**Status deste doc: CONFIRMED (v0).** G-123..G-128 autorizam o slice 17
de código. MCP, embeddings, TUI e `memory.*` permanecem fora.

**Pacote OpenSpec:** ativo em
[`openspec/changes/029-project-memory/`](../openspec/changes/029-project-memory/).
Branch de implementação: `feat/029-project-memory`.

---

## 1. Problema

O ContextPack v0 monta contexto **intra-run** (`task`, `prior_outputs`,
`repo_hits`, `graph_hits`). Cada run LLM redescobre o que o projeto já
decidiu, o que falhou e onde a sessão anterior parou.

O Event Store responde “o que aconteceu”. O Graph responde “o que
existe”. Falta “o que ainda vale como orientação” — episódios
compilados, não transcripts.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Core `memory` (Provider) | Player / Agent / MCP client |
| SQLite no mesmo `.runtgine/runtgine.db` | Sidecar Rust `ai-memory` |
| Episódios `decision`/`failure`/`handoff`/`preference` | Knowledge consolidada; ADR wiki |
| Recall `active` no ContextPack (`memory_hits`) | `graph_hits` / `repo_hits` |
| Validade explícita | Supersession silenciosa via LLM |
| Sugestão de contexto | Autoridade de execução |

Regras (lista negativa de `16` §4.1, preservada):

1. Validator / Registry / Policy continuam soberanos.
2. Memória **nunca** concede capability, altera Policy ou bypassa
   Validator.
3. “Evitar falha X” informa o LLM; **não** proíbe reexecutar o comando.
4. Falha do store → `memory_hits.items = []`; o Run segue.
5. Heurísticas Intent `shell|pipeline` **não** consultam Memory
   (espelha G-69).
6. Pacote Go: `internal/core/memory`. Não é Player.

---

## 3. Cortes confirmados (G-123+)

### G-123 — Papel e pacote

**Status: CONFIRMED**

- Nome: Project Memory v0
- Pacote: `internal/core/memory`
- Acesso: **Memory Provider** no Core (G-47 resolvido para o v0:
  Provider; Memory Player fora)
- Persistência: mesma SQLite do Core (como Graph)
- Recorte de G-46: episódios + ContextPack; sem MCP, sem Knowledge
- `16` continua esboço; autoridade de corte = este doc + `04`

### G-124 — Episódio e validade

**Status: CONFIRMED**

Um episódio:

| Campo | Regra |
|---|---|
| `id` | UUID v7 |
| `kind` | `decision` \| `failure` \| `handoff` \| `preference` |
| `validity` | `active` \| `superseded` \| `archived` |
| `title` | UTF-8; 1–200 runes; obrigatório |
| `body` | UTF-8; 0–4096 bytes; sem secrets/env bruto |
| `created_at` | RFC3339 UTC |
| `run_id` / `task_id` | opcionais (ligação ao Event Store) |
| `successor_id` | preenchido quando `superseded` |

Recall operacional default: só `validity=active`.

Supersession é **explícita** (`Supersede(old_id, new)`): o antigo vira
`superseded` com `successor_id`; o novo nasce `active`. LLM no Core
não infere supersession.

Arquivar (`archived`) é explícito; v0 **não** apaga linhas.

### G-125 — Superfície Core + CLI

**Status: CONFIRMED**

API (Entry Point → Core):

```text
Record(ctx, EpisodeInput) -> Episode
List(ctx, Filter) -> []Episode
Query(ctx, text, limit) -> []Episode   // só active; ranking lexical
Supersede(ctx, oldID, EpisodeInput) -> Episode
Archive(ctx, id) -> Episode
```

CLI (JSON stdout, como `graph snapshot`):

```text
runtgine memory list
runtgine memory query "<text>"
runtgine memory record --kind decision --title "..." [--body "..."]
runtgine memory supersede <id> --kind decision --title "..." [--body "..."]
runtgine memory archive <id>
```

`Query` ranking determinístico:

1. Só `active`
2. Score = contagem de tokens do texto (lowercase, split em
   não-alfanuméricos) que ocorrem em `title+body`
3. Desempate: `created_at` desc, depois `id` asc
4. `limit` default 8 (igual `memory_max_hits`)

Sem embeddings. Sem rede.

### G-126 — ContextPack `memory_hits`

**Status: CONFIRMED**

Estender o pack (G-24/G-67) com:

```text
memory_hits.items[]:
  - id
  - kind
  - validity      # sempre active no pack v0
  - title
  - snippet       # body truncado
  - score
budget.memory_max_hits     default 8
budget.memory_max_chars    default 2000
```

`AssembleContext` (steps LLM) chama `Query` com `intent.summary` +
capability do step. Heurísticas `shell|pipeline` não chamam.

Hierarquia de truncamento (quando o pack estoura chars):

1. `task` + `step`
2. `prior_outputs`
3. `repo_hits`
4. `graph_hits`
5. `memory_hits` (corta score baixo primeiro)

Erro de Query → `memory_hits.items = []`; schema sempre presente
(`items` array, possivelmente vazio).

### G-127 — Captura

**Status: CONFIRMED**

Escrita principal = CLI / API `Record` (humano ou harness).

Captura automática, config:

```text
memory.capture = off | failures     # default: off
```

`failures`: em `run.failed`, best-effort `Record` com
`kind=failure`, `title` = `intent.summary` truncado, `body` = mensagem
de erro truncada, `run_id`/`task_id` preenchidos. **Nunca** falha o
Run se a captura falhar.

Sem captura de transcripts, stdout completo, env ou secrets.
Sem captura automática de `run.succeeded` no v0 (ruído).

### G-128 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- Memory Player / capabilities `memory.*`
- MCP (G-44); embutir `ai-memory`
- Project Knowledge; RAG; embeddings; indexar transcripts
- TUI aba MEMORY; Blast-from-memory
- Cross-workspace; sync remoto
- Supersession silenciosa via LLM
- Memória como Policy/Validator
- Captura `outcomes` de sucesso; kinds além dos quatro
- Delete físico; vector search

---

## 4. Critérios de aceite

1. Store SQLite no mesmo DB; `Record` + `List` + `Query` verdes offline.
2. `Query` não devolve `superseded` / `archived`.
3. `Supersede` marca o antigo e cria o sucessor `active`.
4. ContextPack JSON sempre inclui `memory_hits` e
   `budget.memory_max_hits`; vazio quando não há episódios.
5. Falha injetada no Provider → pack com `items: []`; Run segue.
6. Intent heurístico `echo hi` **não** consulta Memory.
7. `memory.capture=off` (default) não grava em `run.failed`.
8. `go test ./internal/core/memory/...` e `go test ./...` / `go vet`
   verdes.
9. OpenSpec `029-project-memory` arquivado após o **código** (slice 17),
   não neste PR de spec.

---

## 5. Ordem do slice de código

Bloqueado até G-123..G-128 CONFIRMED — este doc + `04` (este PR):

1. Pacote `internal/core/memory` + tabela SQLite
2. API Core + CLI `runtgine memory`
3. `memory_hits` em AssembleContext (LLM path) + budget
4. Captura `failures` opt-in; testes; README Estágio: Slice 17
5. Arquivar OpenSpec `029` após o merge do código

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-123 Papel (`memory` Provider, não Player)
- [x] G-124 Episódio + validade explícita
- [x] G-125 API + CLI; ranking lexical
- [x] G-126 ContextPack `memory_hits` + hierarquia
- [x] G-127 Captura opt-in (`off`/`failures`)
- [x] G-128 Exclusões (MCP, Player, RAG, TUI, Knowledge)
