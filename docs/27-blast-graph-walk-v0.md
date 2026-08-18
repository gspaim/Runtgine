# 27 — Walk Blast←Graph v0

Enriquecimento **determinístico** do Impact Report (`25`): depois das
tabelas Task IR, um hop no Runtime Graph a partir dos `touches` de
`path`. Não é Player, não é QueryHits, não dispara blast a partir da
aba GRAPH.

Inventário: [10-gaps.md](10-gaps.md) (G-111+; exclusão G-104 de `25`).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Blast Radius v0 ([25-blast-radius-v0.md](25-blast-radius-v0.md),
slice 13) e Runtime Graph v0 ([18-runtime-graph-v0.md](18-runtime-graph-v0.md),
slice 6). TUI GRAPH (`26`) permanece superfície estrutural — **não**
inicia walk.

**Status deste doc: CONFIRMED (v0).** G-111..G-116 autorizam o slice 15
de código. Gate de Execute, multi-hop, símbolos e Blast-from-GRAPH
permanecem fora.

**Pacote OpenSpec:** ativo em
[`openspec/changes/027-blast-graph-walk/`](../openspec/changes/027-blast-graph-walk/).
Branch de implementação: `feat/027-blast-graph-walk`.

---

## 1. Problema

`runtgine blast` já diz o que a Task **tocaria** e **claimaria**. Não
diz o que o Graph **já conhece** sobre esses paths: runs/tasks que
`mentions` o mesmo arquivo (ex.: um `pipeline.repo-search` anterior).

A visão longa em `02` (`Change → Graph → Affected Players, Workflows,
Resources, Symbols`) continua larga demais. Este corte é **1 hop**,
só a partir de `touches` `kind=path`, só aresta `mentions` inbound.

Não é o inverso (selecionar um nó na aba GRAPH e “blastar”). Isso
permanece fora (`26` G-110).

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Função no pacote `blast` (após `Analyze`) | Player `blast.*`; pacote Graph novo |
| 1 hop `mentions` inbound sobre snapshot | QueryHits / ranking / ContextPack |
| Campo `affected` no Impact Report | Mudar `risk` / predicted claims / overlay |
| Mesmo `BlastTask` + `runtgine blast` | Comando novo; flag obrigatória |
| Degrada para `affected: []` | Falhar o relatório se o Graph estiver vazio |
| CLI JSON | Aba GRAPH dispara walk; painel Blast na TUI |

Regras:

1. Validator / Registry / G-95 / overlay de `25` **não mudam**.
2. `risk` continua só de predicted claims (`none|path|workspace`).
3. Walk **lê** `GetGraphSnapshot`. Não chama `RefreshGraph` (o operador
   já tem `runtgine graph refresh` / `r` na TUI).
4. Runner **não** chama Blast (G-103 intacto). Gate de Execute continua
   fora.
5. Match de path: `touch.key` == `node.id` (ambos já normalizados pelo
   Core: claim path vs id do Graph). Sem glob, sem prefix-overlap.
6. `touches` `workspace` **não** semeiam (não há `node_kind=workspace`).

---

## 3. Cortes confirmados (G-111+)

### G-111 — Papel

**Status: CONFIRMED**

- Walk vive em `internal/core/blast` (ex.: `Walk(snapshot, touches)`).
- Não é Player, não é QueryHits (`19`), não é a aba GRAPH (`26`).
- Recorte da exclusão G-104: **só** este hop. O resto de G-104
  (gate, argv shell, persistência, evento) permanece exclusão.

### G-112 — Snapshot e degradação

**Status: CONFIRMED**

- Fonte: `GetGraphSnapshot` já existente (`18`).
- Snapshot vazio, path sem nó, ou erro de leitura: `affected` = `[]`.
  O relatório IR+overlay **ainda é válido**. Logar o erro; não
  transformar em ValidationError.
- Sem escrever nós/arestas.

### G-113 — Sementes e hop

**Status: CONFIRMED**

Sementes: `touches` com `kind=path`, ids únicos, ordem estável
(primeira aparição na lista de touches).

Para cada semente cujo nó `path` existe:

1. Emitir `affected` `{kind: path, id, reason: seed}`.
2. Arestas `mentions` com `to` = esse path: emitir o `from`
   (`run` ou `task`) com `reason: mentions` e `via: path:<id>`.

Ordem de `affected`: `path`, depois `task`, depois `run`; empate por
`id` lexicográfico. Deduplicar `(kind, id, reason, via)`.

Fora do hop v0: `provides`, `executed`, `instance_of`, `child_of`,
símbolos, segundo hop, prefix overlap de paths.

### G-114 — Campo `affected`

**Status: CONFIRMED**

```text
affected[]: { kind, id, reason, via? }
reason: seed | mentions
```

- Sempre presente no JSON (array, possivelmente vazio).
- `schema_version` do relatório permanece `0.1.0` (campo aditivo).
- `hello.json`: `affected: []` (sem touch path).
- Não altera `touches`, `predicted_claims`, `conflicts`, `images`, `risk`.

### G-115 — Superfície

**Status: CONFIRMED**

- `BlastTask` passa a preencher `affected` após `Analyze`.
- CLI `runtgine blast` imprime o JSON completo (incluindo `affected`).
- Sem `runtgine blast --graph` no v0 (walk sempre; é barato e degrada).
- Sem tecla / aba / drawer na TUI. GRAPH continua só snapshot.
- `runtgine run` inalterado.

### G-116 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- Blast a partir de nó selecionado na aba GRAPH (G-110)
- QueryHits / `graph_hits` / ranking / LLM
- Multi-hop; `executed` inbound em capabilities; `child_of`
- Símbolos; `node_kind` novos; glob / overlap de path no walk
- Gate Execute por `affected` ou `risk`; auto-blast no Runner
- Persistência; evento `blast.computed`
- Inferir `shell.exec` argv; kind `image` no Graph
- Painel Blast na TUI; Board write-back
- Project Memory (`16`)

---

## 4. Critérios de aceite

1. `runtgine blast examples/hello.json` continua `risk: none`,
   touches/claims vazios, e inclui `affected: []`. Não cria Run.
2. Task `fs.read` `notes.md` sem nó `path` no Graph: `affected: []`.
3. Graph com `path:notes.md` e `mentions` `run:prior → path:notes.md`:
   blast `fs.write` `notes.md` lista seed `path:notes.md` e
   `run:prior` `reason:mentions`. `risk` continua `path` (G-100).
4. Erro simulado de snapshot: relatório IR ok; `affected: []`.
5. `runtgine run examples/hello.json` sem chamar walk/blast.
6. TUI GRAPH não ganha tecla de blast; seis abas inalteradas.
7. `go test ./internal/core/blast/...` cobre seed + mentions + empty.
8. `go test ./...` e `go vet ./...` verdes.
9. OpenSpec `027-blast-graph-walk` arquivado após o **código**
   (slice 15), não neste PR de spec.

---

## 5. Ordem do slice de código

Bloqueado até G-111..G-116 CONFIRMED — este doc + `04` (este PR):

1. `Affected` + `Walk(snapshot, touches)` no pacote `blast`
2. `BlastTask` lê snapshot e preenche `affected` (degrada vazio)
3. Testes: hello vazio; seed sem mentions; seed+mentions; snap error
4. README Estágio: Slice 15; next = mais Players / Memory
5. Arquivar OpenSpec `027` após o merge do código

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-111 Papel (blast.Walk; 1 hop; não QueryHits)
- [x] G-112 Snapshot read-only; degrada vazio
- [x] G-113 Sementes = path touches; inbound mentions
- [x] G-114 Campo `affected`; `risk` intacto
- [x] G-115 Mesmo BlastTask / CLI; sem TUI
- [x] G-116 Exclusões (GRAPH→blast, multi-hop, gate, Hits)
