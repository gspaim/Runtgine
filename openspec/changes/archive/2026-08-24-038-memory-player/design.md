# Design: 038-memory-player

## Pacote

`internal/players/memory` com `New()`, `Manifest()`,
`ValidateStaticInput(workspace, capability, input)`, `Execute`.

Player name no Manifest: `memory`. Pacote Go: `memory` (não há
colisão interna; `internal/core/memory` é o Provider, não o
Player — são pacotes diferentes).

## Capability contract

| Capability | Input | Output (sucesso) |
|---|---|---|
| `memory.recall` | `query` (string UTF-8, 1–512), `kind?` (enum 4), `limit?` (1–20, default 5) | `hits[]` (id, kind, title, body truncado, created_at) |
| `memory.check` | `pattern` (string UTF-8, 1–256) | `has_failure` (bool), `match?` (id, title) |

`hits[].body` é truncado em 1 KiB (não é stream). Schema
`additionalProperties: false`.

## Fronteiras (reforço de `29` §2)

1. **Read-only.** Player **nunca** grava, supersede ou arquiva. Provider
   (`internal/core/memory`) é o único gravador.
2. Provider soberano. Falha do store (`provider` indisponível,
   SQLite busy) → step `succeeded` com `hits: []` / `has_failure: false`.
   Erro vira `warning`, não `runtime.player_error`.
3. Sem rede. Sem MCP. Sem indexação externa.
4. Sem Authority. Resultado do Player **não** altera Policy,
   Validator ou Claims. Mesmo princípio de `repo_hits` /
   `graph_hits`.
5. Sandbox: workdir resolvido + fica no workspace (não usa, mas
   mantém invariante dos outros Players).

## Integração com Provider existente

`internal/core/memory` hoje exporta Provider (slice 17). Esta spec
introduz a **interface `Reader`** mínima:

```go
type Reader interface {
    Recall(ctx context.Context, q Query) ([]Episode, error)
    Check(ctx context.Context, pattern string) (CheckResult, error)
}
```

Provider existente (`*provider`) já cobre essas operações (são
internas a `memory_hits` no ContextPack); a spec **expõe** a
interface sem mudar a implementação. Sem migração de dados.

## Blast / Claims

Nem `memory.recall` nem `memory.check` entram em `claim.Required`
nem em `blast.Touched`. Player é leitura do store, não touch em
filesystem.

## Pipeline / Intent

- **Sem** Intent heuristic. Decisão explícita: Pipeline é soberano
  e lê memória via Task IR explícita. Heurística shell|pipeline já
  não consulta Memory (G-69, preservado).
- Pipeline pattern (recorte do `pipeline.*` de G-22): o step LLM
  recebe `memory.recall` como output anterior e injeta no
  ContextPack do step subsequente (mesma forma de `test.go` →
  LLM review no G-23).

## Integração Core

1. `api.Open` registra `memory.New()`
2. `ValidateStaticInput` no admission (ranges, pattern não vazio)
3. Runner despacha static validation como Git/FS/HTTP/Docker
4. Graph: `RefreshFromRegistry` cria nós `memory`,
   `memory.recall`, `memory.check`; `provides` para `provider`
   (`internal/core/memory`)
5. Exemplos `examples/memory-recall.json`,
   `examples/memory-check.json`
6. `internal/core/memory` expõe `Reader` (sem mudar Provider)

## Testes (slice 31)

- Reader stub: 3 hits → `recall` devolve `hits[3]`
- Reader stub: 0 hits → `hits: []`, step `succeeded`
- Reader stub: erro → step `succeeded`, `hits: []`,
  warning slog
- `check` com `failure` ativo → `has_failure: true`
- `check` sem match → `has_failure: false`
- Pattern vazio → `validation.invalid_input`
- Limit fora de range → `validation.invalid_input`
- Graph: nó `memory` aparece após `RefreshFromRegistry`
- `go test ./...` / `go vet ./...` verdes

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Player que **escreve** | Bypassa HITL de Lessons (`33`); viola regra 2 de `29` §2 |
| Substituir `memory_hits` do ContextPack pelo Player | Quebra G-69 (heurísticas sem Graph/Memory); Core precisa consultar Memory sem ser Player |
| Intent heuristic `lembre-se que …` | Ambíguo; Pipeline deve ser explícito |
| `memory.list` (todos episódios) | Sem filtro; expõe o store inteiro |
| `memory.archive` / `memory.supersede` | Gravação → fora do v0 |
| MCP como transporte | G-44 permanece OPEN; Player fala com Provider in-process |
| Embeddings / busca semântica | Recorte `29` é lexical; embeddings fora |

## Risks

| Risco | Mitigação |
|---|---|
| Player vira "atalho" para bypassar Validator | Player emite contexto; Validator/Pipeline continuam soberanos |
| Loop "lembre → verifique → lembre" | Sem Intent heuristic; Pipeline é quem decide |
| Falha do Provider derruba Run | Degrada para vazio (mesmo padrão de Graph Hits) |
| Falsa sensação de autoridade | Read-only + Provider soberano + warning explícito no log |

## Critérios de aceite

1. `memory.list` / `memory.archive` ausentes → Validator rejeita.
2. `pattern` vazio ou `query` vazio → `validation.invalid_input`.
3. Reader stub com hits → `hits[]` populado.
4. Reader stub com erro → `hits: []`, step `succeeded`,
   warning slog.
5. `has_failure` reflete Provider (Provider é soberano).
6. Graph: nó `memory` aparece em `RefreshFromRegistry`.
7. `go test ./...` e `go vet ./...` verdes.
8. OpenSpec `038` arquivado após código (slice 31).
