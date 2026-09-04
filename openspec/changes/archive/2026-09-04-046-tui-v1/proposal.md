# Proposal: 046-tui-v1

## Why

A TUI Constellation já existe (slices 3, 14, 21) na stack Charm v2, mas
o layout é uma casca: INTENT concatena runes, RUNS/EVENTS são
`fmt.Sprintf`, Bubbles quase não entra. O operador descreve a UI como
simples demais. As “libs famosas” do Go TUI **já estão no `go.mod`**
(`bubbletea/v2`, `lipgloss/v2`, `bubbles/v2`); o v1 é usá-las de
verdade.

Hits (`graph_hits` / `memory_hits` / `playbook_hits`) e Blast
(`BlastTask`, com `affected`) já existem no Core e só aparecem na
CLI/API. Superficiá-los na Mission Control fecha a frase do produto
sem conceito novo e sem oitava aba.

## What Changes

- Canonical `docs/46-tui-v1.md` (G-238..G-244 CONFIRMED)
- `docs/04-decisoes.md`: seção TUI v1; tuios/canvas continuam REJECTED
- `docs/14-tui-design.md` + skill: componentes Bubbles, Hits/Blast
  inline, `?` help, `Ctrl+b` / `b`
- TUI `internal/entrypoint/tui` (**slice 39 — not this spec PR**)
- `CoreAPI` ganha `QueryHits` + `BlastTask`

## What Does Not Change

- Sete abas e respectivos papéis (INTENT submete; LIVE = um Run;
  GRAPH = snapshot estrutural)
- Validator / Runner / Event Bus / Players
- Task IR schema; Claims
- PTY / tuios / multiplexer (continuam REJECTED)
- GRAPH canvas 2D (G-107)
- Blast-from-GRAPH (G-110 / G-115)
- Wails desktop (spec `35`; Hits/Blast desktop fora deste slice)
- NATS (G-36 DEFERRED)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `046-tui-v1` |
| Doc canônico | [`docs/46-tui-v1.md`](../../../docs/46-tui-v1.md) |
| Gaps | G-238..G-244 **CONFIRMED** |
| Código | slice 39 — **bloqueado** até este pacote + `04` + `14` |

## Approach

1. Manter Charm v2; passar o corpo das abas para Bubbles
   table/list/viewport/textarea/help
2. Hits inline (LIVE via ContextPack dos eventos; INTENT preview via
   `QueryHits`); sem aba nova
3. Blast drawer (`BlastTask`) em INTENT (`Ctrl+b`) e LIVE (`b`); GRAPH
   não dispara
4. Testes com fake Core (TUI continua não-interativa no CI)

## Impact

- `internal/entrypoint/tui` (slice 39)
- Docs `14`, skill, `04`, `10`, `AGENTS`, README
