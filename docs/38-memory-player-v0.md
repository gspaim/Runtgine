# 38 — Memory Player v0

Player **read-only** sobre o Project Memory Provider (`29`; slice 17
feito): `memory.recall` e `memory.check`. Não escreve, não supersede,
não arquiva. Falha do Provider **degrada** (vazio + warning), nunca
faz o Run falhar.

Inventário: [10-gaps.md](10-gaps.md) (G-180+; recorte de G-47 sobre o
Provider já CONFIRMED em `29`). Autoridade de status:
[04-decisoes.md](04-decisoes.md). Não é MCP (G-44). Não é Knowledge
base. Não é RAG. Não é indexação de transcripts. Não é escrita.

**Status deste doc: CONFIRMED v0 (slice 31 a fazer).** G-180..G-186.
Fecha a OPEN QUESTION "Memory Player" que `04` carregava desde a
sessão de `29` §128.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-24-038-memory-player/`](../openspec/changes/archive/2026-08-24-038-memory-player/).
Spec atual: [`openspec/specs/memory-player/`](../openspec/specs/memory-player/).

---

## 1. Problema

`29` fechou G-47 pelo lado Provider: o Core consulta episódios via
`memory_hits` no ContextPack. Mas **não há superfície para um step
declarativo** ler memória. Hoje a única forma é via Provider
in-process (Core), o que obriga o LLM a entrar cedo e limita o uso
em pipelines deterministas.

## 2. Fronteiras

| É | Não é |
|---|---|
| Player `memory` read-only | Player de escrita / Archive / Supersede |
| `memory.recall` (lexical) | Embeddings / RAG / Knowledge |
| `memory.check` (booleano `failure` ativa) | Sweep de Pattern global |
| Provider soberano (Core) | MCP / rede / shell |
| Falha degrada (vazio + warning) | Falha derruba Run |

Regras (lista negativa de `29` §2, preservada):

1. Player **nunca** grava, supersede ou arquiva.
2. Provider (`internal/core/memory`) é o único gravador.
3. Falha do store → `hits: []` / `has_failure: false` + `slog.Warn`.
4. Heurísticas Intent `shell|pipeline` **não** consultam Memory
   (espelha G-69).
5. Pacote Go: `internal/players/memory`. Não é Provider.
6. Sem MCP, sem rede, sem shell, sem embeddings.

## 3. Cortes confirmados (G-180+)

### G-180 — Papel e pacote

- Player name: `memory`
- Pacote: `internal/players/memory`
- Kind: `deterministic`
- Read-only sobre `internal/core/memory` (já CONFIRMED v0; slice 17)
- Recorte de G-47: Player sobre Provider CONFIRMED

### G-181 — Capabilities v0

| Capability | Input | Output |
|---|---|---|
| `memory.recall` | `query` (1–512), `kind?` (4 kinds), `limit?` (1–20, default 5) | `hits[]` (`id`, `kind`, `title`, `snippet`, `created_at`) |
| `memory.check` | `pattern` (1–256) | `has_failure` (bool), `match?` (`id`, `title`) |

`hits[].snippet` truncado em 1 KiB. Schemas JSON no Manifest com
`additionalProperties: false`.

### G-182 — Provider `Reader` interface

`internal/core/memory` expõe a interface mínima:

```go
type Reader interface {
    Recall(ctx context.Context, q RecallQuery) ([]Hit, error)
    Check(ctx context.Context, pattern string) (CheckResult, error)
}

type RecallQuery struct {
    Text  string
    Kind  string // opcional
    Limit int
}

type CheckResult struct {
    HasFailure bool
    Match      *Episode // opcional
}
```

A `*Service` existente já cobre `Recall` (mecanismo de `Query`) e
ganha `Check` (lexical sobre `active` `failure` episódios). Sem
mudança de schema SQLite; sem migração de dados.

### G-183 — Sandbox

- In-process: Player fala com `Reader` via interface Go
- Sem rede, sem MCP, sem shell string
- `workdir` resolvido (`EvalSymlinks` + stay-in-workspace) —
  invariante comum dos Players (memória é global, mas mantemos
  validação)

### G-184 — Falha do Provider degrada

`Reader.Recall` ou `Reader.Check` retorna erro → step `succeeded`
com vazio + `slog.Warn` com `err`. **Nunca** `runtime.player_error`.
Mesma regra de `graph_hits` / `repo_hits` (degrada, não falha).

### G-185 — Registry + Graph

1. `api.Open` registra `memoryplayer.New(c.Memory)`
2. `ValidateStaticInput` no admission (ranges, pattern não vazio)
3. Runner despacha static validation como Git/FS/HTTP/Docker/npm
4. `RefreshFromRegistry` cria nós `memory`, `memory.recall`,
   `memory.check`; edge `provides` aponta para
   `internal/core/memory`
5. Exemplos `examples/memory-recall.json`,
   `examples/memory-check.json`

### G-186 — Exclusões v0

- `memory.record`, `memory.supersede`, `memory.archive`,
  `memory.list` (todos → gravação)
- MCP como transporte (G-44)
- Embeddings / RAG / vector search
- Captura automática de transcripts
- Pipeline de "lembre-se que …" via Intent heuristic
- TUI aba MEMORY dedicada (continua nas ferramentas de debug)

## 4. Critérios de aceite

1. `memory.record` ausente → Validator rejeita.
2. `query` vazio ou `pattern` vazio → `validation.invalid_input`.
3. Reader stub com 3 hits → `hits[3]`, step `succeeded`.
4. Reader stub com erro → `hits: []` + `slog.Warn`, step
   `succeeded`.
5. `check` com `failure` ativo → `has_failure: true`.
6. `check` sem match → `has_failure: false`.
7. Graph: `RefreshFromRegistry` cria nó `memory` com edge
   `provides` para o Provider.
8. `go test ./...` e `go vet ./...` verdes sem MCP, sem rede.
9. OpenSpec `038` arquivado após o **código** (slice 31).

## 5. Ordem do slice de código

1. Adicionar `Reader` + `Recall`/`Check` em `internal/core/memory`
2. Pacote `internal/players/memory` + Manifest + `ValidateStaticInput`
3. `Execute`: chama Reader; degrada em erro
4. Registrar em `api.Open`; exemplo + Graph
5. Testes fake Reader
6. `go test ./...` / `go vet ./...` verdes
7. Arquivar OpenSpec `038`

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-180 Papel (`memory` / `internal/players/memory`)
- [x] G-181 Capabilities (`memory.recall`, `memory.check`)
- [x] G-182 `Reader` em `internal/core/memory`
- [x] G-183 Sandbox (in-process; sem rede/MCP/shell)
- [x] G-184 Falha degrada
- [x] G-185 Registry + Graph + `provides`
- [x] G-186 Exclusões (escrita, MCP, embeddings, RAG)