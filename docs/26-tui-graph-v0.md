# 26 — TUI GRAPH v0

Aba **GRAPH** na TUI Constellation Mission Control: superfície
read-only sobre `GetGraphSnapshot` / `RefreshGraph`. Não é Player, não
caminha Blast, não é o Graph Hits do ContextPack.

Inventário: [10-gaps.md](10-gaps.md) (G-105+; aba prometida em `18`).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Runtime Graph v0 ([18-runtime-graph-v0.md](18-runtime-graph-v0.md),
slice 6) e TUI Slice 3 ([14-tui-design.md](14-tui-design.md)).
Este PR **altera `14`** (sexta aba) e a skill
`.cursor/skills/runtgine-tui-design/SKILL.md`.
Walk Blast←Graph a partir **desta aba** permanece fora. O walk 1-hop
do Impact Report é a spec [27-blast-graph-walk-v0.md](27-blast-graph-walk-v0.md)
(CLI `runtgine blast`, não GRAPH).

**Status deste doc: CONFIRMED (v0).** G-105..G-110 implementados no
slice 14. Multiplexer, PTY, edição do Graph e Hits UI permanecem fora.

**Nota (032):** a ordem de tabs passou a incluir **INTENT** como primeira aba;
ver [32-intent-surface-v0.md](32-intent-surface-v0.md). Este doc descreve só a aba GRAPH.

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-18-026-tui-graph/`](../openspec/changes/archive/2026-08-18-026-tui-graph/).
Deltas mergeados em `openspec/specs/tui-graph/` e `openspec/specs/runtime-graph/`.
Branch de implementação: `feat/026-tui-graph`.

---

## 1. Problema

O Runtime Graph já persiste nós/arestas e a CLI imprime
`runtgine graph snapshot`. A TUI só tem RUNS / LIVE / BOARD / EVENTS /
CONFIG. LIVE mostra a trajetória **de um Run** (`depends_on`); não é o
grafo estrutural do workspace.

`18` deixou a aba GRAPH explícita para depois de `14` + skill. Sem a
aba, o operador precisa sair para a CLI para ver players, capabilities,
runs e paths no Graph.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Sexta aba da TUI existente | App nova, multiplexer, PTY, tuios |
| Superfície sobre Core `GetGraphSnapshot` | Motor novo; cópia do Graph na TUI como SoT |
| Lista + detalhe + counts | Force-directed / canvas / zoom infinito |
| `r` chama `RefreshGraph` e recarrega snapshot | Mutar nós; “editar grafo” |
| Filtro `/` em kind/id | QueryHits UI; `graph hits` CLI |
| Tokens Constellation (`14`) | BLAST tab; Claims tab; walk Blast←Graph |

Regras:

1. TUI continua Entry Point. Só APIs públicas do Core (`GetGraphSnapshot`,
   `RefreshGraph`). Nunca Player, nunca Store SQL direto.
2. Uma janela fullscreen. Tabs: `RUNS LIVE BOARD EVENTS GRAPH CONFIG`.
3. LIVE **não** muda de papel (trajetória do Run). GRAPH é o snapshot
   estrutural do workspace.
4. `< 80` colunas: só lista vertical; sem desenho horizontal de arestas.
5. `NO_COLOR` / `RUNTGINE_ASCII` iguais às outras abas.
6. Falha ao carregar snapshot: mensagem no painel; TUI não crasha;
   as outras abas continuam.

---

## 3. Cortes confirmados (G-105+)

### G-105 — Papel

**Status: CONFIRMED**

- GRAPH é aba da TUI (`internal/entrypoint/tui`), não pacote Core novo.
- SoT permanece SQLite Graph (`18`). A aba **lê** o snapshot; não
  duplica o grafo em memória além do cache de view (como RUNS cacheia
  `ListRuns`).
- Recorte: só a superfície. QueryHits, Blast e Claims não ganham aba.

### G-106 — Tab, ordem e teclas

**Status: CONFIRMED**

Ordem:

```text
[ RUNS ] [ LIVE ] [ BOARD ] [ EVENTS ] [ GRAPH ] [ CONFIG ]
```

- Navegação: `tab` / `shift+tab` (cíclico nas **seis** abas).
- Sem tecla dedicada `g` no v0 (evita colidir com filtro/git mental).
- GRAPH usa o keymap já confirmado: `j`/`k` ou setas, `/` filtro,
  `r` refresh, `enter` foca o detalhe do nó selecionado, `q` quit.
- `c` / `a` / `d` **não** agem no Graph (continuam ligados ao Run
  selecionado em RUNS, como hoje).
- Footer lista as seis abas; em GRAPH acrescenta `r refresh graph`.

`14` e a skill passam a listar seis abas obrigatórias.

### G-107 — Conteúdo da aba

**Status: CONFIRMED**

Layout (largura ≥ 80):

1. **Header da aba:** `GRAPH` + totais
   `nodes=<n> edges=<e>` e counts por `node_kind` (player, capability,
   task, run, path, symbol — zero omitido ou mostrado como 0; estável).
2. **Lista:** uma linha por nó, `kind` + `id` (id truncado com as mesmas
   regras de RUNS). Ordem: kind (ordem G-61), depois id lexicográfico.
3. **Detalhe** (nó selecionado): `kind`, `id`, `attrs` JSON (sem secrets;
   o snapshot já é “no secrets”), arestas incidentes
   (`kind from→to`).

Largura ≥ 120: lista | detalhe lado a lado.
80–119: lista; detalhe em drawer/abaixo.
< 80: lista compacta; detalhe só após `enter` (ou omitido até enter).

Grafo visual:

- **Não** há canvas 2D no v0.
- Em ≥ 120, o detalhe **pode** listar vizinhos como linhas
  `provides shell → shell.exec` (texto). Não é constelação de nós.

Empty state: `GRAPH` + `No graph nodes.` (workspace recém-aberto ainda
tem players/capabilities após `api.Open` — o empty real é falha de
load). Se snapshot vazio de verdade, a mensagem vale.

### G-108 — Refresh e lifecycle

**Status: CONFIRMED**

- Ao entrar na aba (primeira vez / `r`): `GetGraphSnapshot`.
- `r` em GRAPH: `RefreshGraph` **depois** `GetGraphSnapshot` (espelha
  `runtgine graph refresh` + snapshot).
- Subscribe de runs **não** é obrigatório para atualizar GRAPH em
  tempo real no v0. Operador dá `r`. (LIVE/RUNS já são live.)
- Erro de snapshot: `err` no painel GRAPH; último snapshot bom, se
  houver, permanece.

### G-109 — Filtro

**Status: CONFIRMED**

`/` filtra a **lista de nós** (não as outras abas). Match case-insensitive
em `kind` e `id`. Exemplos: `player`, `shell.exec`, `run:`.

Sem DSL `type:` obrigatória no v0 (EVENTS já tem a dela). Substring basta.

### G-110 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- Walk Blast←Graph; mostrar `runtgine blast` na aba
- QueryHits / ranking / `graph_hits` na TUI
- Editar, apagar ou criar nós/arestas
- Force-layout, zoom, mouse-only, SVG
- Tecla `g`; sétima aba; aba CLAIMS
- PTY, tuios, multiplexer
- Mutar Graph fora de `RefreshGraph`
- Wails / web
- Genome / AST contínuo (continua fora de `18`)

---

## 4. Critérios de aceite

1. `runtgine tui` mostra seis tabs; GRAPH está entre EVENTS e CONFIG.
2. GRAPH lista nós `player`/`capability` após boot (shell, git, fs, …).
3. Selecionar um nó `player` `shell` mostra aresta `provides` →
   `shell.exec` no detalhe.
4. `/` `capability` restringe a lista; limpar o filtro restaura.
5. `r` na aba GRAPH não crasha; após um `runtgine run` + `r`, o nó
   `run` aparece (best-effort; teste pode injetar snapshot no fake Core).
6. Largura 70: GRAPH renderiza sem pânico; sem bloco horizontal de arestas.
7. `NO_COLOR` / ASCII: GRAPH permanece legível (kind+id em texto).
8. Nenhuma chamada a Player a partir da TUI; `CoreAPI` ganha
   `GetGraphSnapshot` e `RefreshGraph`.
9. `go test ./internal/entrypoint/tui/...` cobre as seis abas e GRAPH.
10. `go test ./...` e `go vet ./...` verdes.
11. OpenSpec `026-tui-graph` arquivado após o **código** (slice 14).

---

## 5. Ordem do slice de código

Implementado no slice 14:

1. Estender `CoreAPI` + fakeCore com snapshot/refresh
2. `tabGraph` + `tabNames` (seis)
3. View lista/detalhe + counts; filtro `/`
4. `r` → RefreshGraph + GetGraphSnapshot
5. Testes de navegação/resize/`NO_COLOR`; README Estágio: Slice 14
6. Arquivar OpenSpec `026` após o merge do código

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-105 Papel (aba TUI; SoT = Graph Core)
- [x] G-106 Seis tabs; keymap existente
- [x] G-107 Lista + detalhe + counts; sem canvas 2D
- [x] G-108 Refresh explícito
- [x] G-109 Filtro substring kind/id
- [x] G-110 Exclusões (Blast-from-graph, Hits UI, PTY, edit)
