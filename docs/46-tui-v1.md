# 46 — TUI v1: Charm Mission Control

Redesenho profissional da TUI **Constellation Mission Control** sobre a
stack Charm **já confirmada** (Bubble Tea v2 + Lip Gloss v2 +
Bubbles v2). Não troca de linguagem nem de framework. O v0 (slices 3,
14, 21) usa Charm só como motor Elm; o v1 **usa os componentes famosos
do ecossistema** (`table`, `viewport`, `textarea`, `list`, `help`) e
fecha o loop de verificação na UI: Hits + Blast, hoje só na CLI/API.

Inventário: [10-gaps.md](10-gaps.md) (G-238+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisitos: TUI Slice 3 ([14-tui-design.md](14-tui-design.md)),
GRAPH ([26-tui-graph-v0.md](26-tui-graph-v0.md)), INTENT
([32-intent-surface-v0.md](32-intent-surface-v0.md)), Graph Hits
([19-graph-hits-v0.md](19-graph-hits-v0.md)), Blast
([25-blast-radius-v0.md](25-blast-radius-v0.md) + walk
[27-blast-graph-walk-v0.md](27-blast-graph-walk-v0.md)).
Este PR **altera `14`** e a skill
`.cursor/skills/runtgine-tui-design/SKILL.md`.

**Status deste doc: CONFIRMED v0 (slice 39 feito).**
G-238..G-244. PTY/tuios e canvas 2D do GRAPH **permanecem fora**.

**Pacote OpenSpec:**
[`openspec/changes/archive/2026-09-04-046-tui-v1/`](../openspec/changes/archive/2026-09-04-046-tui-v1/).
Spec atual: [`openspec/specs/tui-v1/`](../openspec/specs/tui-v1/).

---

## 1. Problema

A TUI já é Entry Point (INTENT submete, LIVE observa, GRAPH lê o
snapshot). Visualmente é uma casca fina: input de INTENT é concatenação
de runes, RUNS/EVENTS são `fmt.Sprintf`, Bubbles quase não entra
(só spinner/progress). O operador percebe “simples e ruim” — não porque
a stack está errada, mas porque os componentes Charm nunca foram o
layout.

Ao mesmo tempo, a frase do produto (“intenção → execução verificável”)
quebra na UI: `graph_hits` / `memory_hits` / `playbook_hits` e o Impact
Report (`runtgine blast`, com `affected` do walk) existem no Core e na
CLI/HTTP, e **não aparecem** na Mission Control.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Mesmo binário `runtgine tui`, mesmo pacote `internal/entrypoint/tui` | App nova, Ratatui, Textual, Ink, Bubble Tea v1 |
| Charm v2 de verdade: Bubbles table/list/viewport/textarea/help | Trocar Lip Gloss por outra lib de estilo |
| Sete abas atuais (INTENT…CONFIG) | Oitava aba; aba BLAST; aba HITS |
| Hits inline (LIVE + preview INTENT) | QueryHits como motor novo; Hits na aba GRAPH |
| Blast drawer via `BlastTask` (INTENT/LIVE) | Blast a partir de um nó GRAPH; gate de Execute |
| Tokens Constellation (`14`) | Tema arcade; HUD de jogo |
| Superfície sobre Core APIs | Chamar Player; Store SQL direto |

Regras:

1. TUI continua Entry Point. Só APIs públicas do Core. Nunca Player.
2. Uma janela fullscreen. Tabs **não** aumentam.
3. GRAPH **não** muda de papel (lista/detalhe estrutural; sem canvas 2D;
   sem disparar Blast — G-110).
4. LIVE **não** muda de papel (trajetória de **um** Run); ganha painel
   de Hits do ContextPack daquele Run.
5. `< 80` colunas: layout vertical; Hit/Blast viram drawer abaixo, não
   colunas extras.
6. `NO_COLOR` / `RUNTGINE_ASCII` iguais às outras abas.
7. Zero dependência nova se Bubbles v2 já cobre o componente. Sem `huh`
   no v0 (confirmações continuam no keymap atual).

---

## 3. Cortes confirmados (G-238+)

### G-238 — Papel / stack

**Status: CONFIRMED**

- TUI v1 é o **mesmo** Entry Point Charm. Go permanece a linguagem.
- Stack: `charm.land/bubbletea/v2`, `lipgloss/v2`, `bubbles/v2`.
- Recorte: redesenho da superfície + Hits/Blast. Sem protocolo Task IR
  novo, sem Player novo, sem Wails neste slice.
- Alternativas REJECTED neste ciclo: Ratatui (Rust), Textual (Python),
  reescrever em Wails no lugar da TUI.

### G-239 — Shell visual (componentes Charm)

**Status: CONFIRMED**

Layout persistente (tokens `14`):

1. **Header** — marca, workspace, fonte (`local` / status).
2. **Tabs** — sete abas; foco visível sem depender só de cor.
3. **Corpo** — painéis Lip Gloss; conteúdo via Bubbles:
   | Superfície | Componente |
   |---|---|
   | RUNS | `table` |
   | EVENTS | `table` ou `viewport` + filtro |
   | GRAPH lista | `list` |
   | GRAPH/LIVE detalhe, payload JSON | `viewport` |
   | INTENT draft | `textarea` |
   | Footer / overlay | `help` (`?` abre cheatsheet) |
   | LIVE progresso | `progress` + `spinner` (já usados) |
4. **Footer** — keymap da aba + `? help`.

Tema continua centralizado em `theme.go`. Sem cores cruas nas views.

### G-240 — Tabs, RUNS, INTENT, LIVE chrome

**Status: CONFIRMED**

Ordem inalterada:

```text
[ INTENT ] [ RUNS ] [ LIVE ] [ BOARD ] [ EVENTS ] [ GRAPH ] [ CONFIG ]
```

- RUNS: tabela (status símbolo+texto, `run_id` curto, source, summary,
  elapsed). Seleção starlight/violet. `enter` abre LIVE.
- INTENT: `textarea` multilinha (substitui o editor rune-a-rune);
  preview do Task IR em `viewport`; toggle JSON (`Ctrl+j` permanece).
- LIVE: lista vertical de steps + viewport de telemetria; HITL `a`/`d`
  e cancel `c` inalterados.
- BOARD / CONFIG: mesmos dados; chrome de painel alinhado ao v1.
- GRAPH: continua G-107 (lista+detalhe, sem canvas); lista passa a
  `list`/`viewport` em vez de strings soltas.

### G-241 — Hits inline

**Status: CONFIRMED**

Não há aba HITS. GRAPH **não** lista QueryHits (G-110 intacto).

| Onde | Fonte | O que mostra |
|---|---|---|
| LIVE | ContextPack nos eventos do Run (`graph_hits`, `memory_hits`, `playbook_hits`) | kind/id/score (texto); vazio = `No hits.` |
| INTENT preview | `QueryHits` do Core com o texto do draft (após `Ctrl+p`) | até o budget de `19`; degrada vazio, nunca falha o preview |

CoreAPI ganha `QueryHits` (wrapper fino sobre `graph.Service`). TUI
não importa `internal/core/graph` além dos DTOs já usados no snapshot.

### G-242 — Blast drawer

**Status: CONFIRMED**

Não há aba BLAST. GRAPH **não** dispara Blast (G-110 / G-115).

- Tecla `Ctrl+b` na aba INTENT: compila o draft (mesmo caminho de
  preview) e chama `BlastTask` — sem criar Run, sem Acquire.
- Tecla `b` na aba LIVE (Run selecionado): `BlastTask` sobre o Task IR
  do snapshot.
- Drawer/painel: `risk`, `touches`, `conflicts`, `predicted_claims`,
  `affected` (walk `27` já vem no report). Símbolo+texto para risk
  (não só cor).
- Erro de blast: linha no painel; TUI não crasha.

CoreAPI ganha `BlastTask` (já existe em `api.Core`; só expor na
interface da TUI).

### G-243 — Keymap e help

**Status: CONFIRMED**

Keymap v0 permanece. Acrescentos:

| Tecla | Ação |
|---|---|
| `?` | Overlay `help` (fecha com `?` ou `esc`) |
| `Ctrl+b` | Blast do draft (INTENT) |
| `b` | Blast do Run (LIVE; ignorado nas outras abas) |

`Ctrl+p` / `Ctrl+Enter` / `Ctrl+j` / `tab` / `c` / `a` / `d` / `/` /
`r` / `q` inalterados. Footer lista os atalhos da aba ativa.

### G-244 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- PTY, tuios, multiplexer, terminal vivo
- Canvas 2D / force-layout / zoom no GRAPH
- Ratatui, Textual, Ink, GPUI, reescrita Wails-no-lugar-da-TUI
- Oitava aba; Hits/Blast como tabs
- Blast a partir de nó GRAPH
- Gate de Execute via Blast; persistir report
- `huh` / dependências Charm além das já no `go.mod`
- Edição de CONFIG; Wails Hits/Blast neste slice
- NATS (G-36)

---

## 4. Critérios de aceite

1. `runtgine tui` continua com **sete** tabs na ordem G-240.
2. INTENT usa `textarea`; RUNS usa `table`; help `?` renderiza.
3. LIVE mostra Hits quando o Run tem ContextPack com `graph_hits`
   (fake Core nos testes); empty state `No hits.` caso contrário.
4. `Ctrl+p` no INTENT anexa Hits do `QueryHits` (ou vazio).
5. `Ctrl+b` no INTENT chama `BlastTask` e mostra `risk` no painel;
   não cria Run.
6. `b` no LIVE chama `BlastTask` do Task do snapshot.
7. GRAPH não ganha tecla de Blast; canvas 2D continua ausente.
8. Largura 70: View não pânica; Hits/Blast empilham verticalmente.
9. `NO_COLOR` / ASCII: status e risk continuam legíveis em texto.
10. Nenhuma chamada a Player a partir da TUI.
11. `go test ./internal/entrypoint/tui/...` cobre tabs, help, Hits
    empty, Blast na interface fake.
12. `go test ./...` e `go vet ./...` verdes.

---

## 5. Ordem do slice de código (slice 39)

1. Estender `CoreAPI` + fake: `QueryHits`, `BlastTask`
2. Shell: theme/layout + `help` overlay + textarea INTENT
3. RUNS `table`; EVENTS/GRAPH viewport/list
4. LIVE Hits a partir dos eventos
5. INTENT Hits no preview; Blast `Ctrl+b` / LIVE `b`
6. Testes de resize, `NO_COLOR`, keymap; README Estágio: Slice 39
7. Arquivar OpenSpec `046`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-238 Papel / stack Charm v2
- [x] G-239 Shell visual (Bubbles)
- [x] G-240 Tabs + chrome RUNS/INTENT/LIVE
- [x] G-241 Hits inline
- [x] G-242 Blast drawer
- [x] G-243 Keymap + help
- [x] G-244 Exclusões v0
