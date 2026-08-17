# Proposal: 019-graph-hits

## Why

O Runtime Graph v0 (G-60..G-65) já persiste estrutura do workspace, mas
ContextPack e Intent Engine não a consultam. LLM steps só veem contexto
intra-run (`repo_hits` / `prior_outputs`). Falta um contrato versionado de
hits estruturais **entre runs**, separado de `repo_hits` e do futuro
`memory_hits` (HYPOTHESIS em `docs/16`).

## What Changes

- Extender ContextPack com `graph_hits` + budget (`graph_max_hits`,
  `graph_max_chars`) — G-67
- API `QueryHits` determinística no pacote graph — G-68
- Runner (AssembleContext) e Intent caminho LLM consomem hits — G-69
- Heurísticas shell\|pipeline do Intent **não** consultam Graph

## What Does Not Change

- Node/edge kinds do Graph estrutural
- Validator / Registry como autoridade de capability
- TUI (sem aba GRAPH)
- Project Memory / `memory_hits`
- CLI obrigatória `graph hits` (nice-to-have fora do aceite)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `019-graph-hits` |
| Branch alvo | `feat/019-graph-hits` |
| Doc canônico | [`docs/19-graph-hits-v0.md`](../../docs/19-graph-hits-v0.md) |
| `docs/04-decisoes.md` | G-66..G-69 **CONFIRMED** |
| Código | Ainda não — este pacote autoriza o slice 7 |

## Approach

1. `QueryHits` no `internal/core/graph` (seeds → mentions → capability → keywords)
2. Estender `internal/core/contextpack`
3. Wire em Runner (steps LLM) e Intent `compileLLM`
4. Degradação: falha de Graph → hits vazios; Run/compile continua

## Impact

- Packages: `graph`, `contextpack`, `runner`, `intent`
- Sem mudança de Task IR schema
- LLM Players passam a receber campo novo opcional no pack JSON
