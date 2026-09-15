# 47 — Desktop Wails operator v1 (Hits / Blast / RESULT)

Fecha o gap que a TUI v1 (`46`) deixou de propósito: **Hits e Blast no
desktop**, e torna o **RESULT** (stdout do step) parte do contrato da
view LIVE. O Entry Point continua o mesmo (`runtgine desktop`, spec
[`35`](35-wails-v0.md)). Não substitui a TUI.

Inventário: [10-gaps.md](10-gaps.md) (G-245+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisitos: Desktop v0 ([35](35-wails-v0.md)), TUI v1 Hits/Blast
([46](46-tui-v1.md)), Graph Hits ([19](19-graph-hits-v0.md)), Blast
([25](25-blast-radius-v0.md) + walk [27](27-blast-graph-walk-v0.md)),
INTENT ([32](32-intent-surface-v0.md) G-144).

**Status deste doc: CONFIRMED v0 spec.** Código = **slice 40 feito**.
PTY/tuios, canvas 2D e mobile **permanecem fora**.

**Pacote OpenSpec:**
[`openspec/changes/archive/2026-09-14-047-wails-operator/`](../openspec/changes/archive/2026-09-14-047-wails-operator/)
(spec publicada em [`openspec/specs/wails-operator/spec.md`](../openspec/specs/wails-operator/spec.md)).

---

## 1. Problema

O desktop Wails v0 (slices 27–28) espelha as sete views e submete
INTENT, mas a verificação some na janela: LIVE listava tipos de evento,
não o stdout; Hits e Blast existem no Core e na TUI v1 e **não** no
Wails (exclusão G-244). O operador que abandona a TUI (atalhos, `q` no
draft, Paseo no celular) precisa da mesma frase do produto na janela:
intenção → preview → submit → **resultado visível**, com Hits e Blast
opcionais sem criar Run extra.

O stdout da LIVE já começou a aparecer no follow-up `#68`
(`internal/entrypoint/liveout` + painel RESULT). Este recorte **não
reverte** isso: canoniza RESULT e completa Hits/Blast + bindings.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Mesmo `runtgine desktop`, pacote `internal/entrypoint/desktop` | App nova; Wails v2; Tauri; Electron |
| Mesmas sete views (`35` G-161) | Oitava view; aba HITS; aba BLAST |
| RESULT no LIVE (stdout/stderr/`step.failed`) | Reabrir o payload JSON como único output |
| Hits inline (LIVE + preview INTENT) | QueryHits na view GRAPH |
| Blast painel/drawer (`BlastTask`) em INTENT/LIVE | Blast a partir de nó GRAPH; gate de Execute |
| Botões + atalhos já do G-163 | Substituir a TUI; mobile iOS/Android |
| Tokens Constellation (`14`) | Tema novo |

Regras:

1. Desktop continua Entry Point. Só APIs públicas do Core. Nunca Player.
2. `source.entry_point = "wails"` inalterado.
3. GRAPH **não** muda de papel (G-110 / G-115): sem Hits, sem Blast.
4. LIVE = trajetória de **um** Run; RESULT é o output daquele Run.
5. CI **não** exige display. Bindings com Core fake.
6. Linux: GTK4 + WebKitGTK 6 (já G-160); stub CGO permanece.

---

## 3. Cortes confirmados (G-245+)

### G-245 — Papel / stack

**Status: CONFIRMED**

- Operator v1 é o **mesmo** Entry Point Wails v3 + Svelte 5.
- Recorte: RESULT canônico + Hits/Blast na janela. Sem Player novo,
  sem protocolo Task IR novo, sem trocar a TUI.
- Alternativas REJECTED neste ciclo: reescrever o desktop em outra
  stack; promover PTY/tuios; servir a UI via `runtgine serve`.

### G-246 — LIVE RESULT

**Status: CONFIRMED**

Após submit (ou seleção em RUNS), LIVE mostra um bloco **RESULT**
acima da lista crua de tipos de evento:

| Evento | O que renderiza |
|---|---|
| `step.succeeded` | `stdout` (e `stderr` se não vazio); senão JSON do `output` |
| `step.failed` | `error` (ou output) |
| Run `error` | linha de falha do snapshot |
| ainda running, sem output | `waiting for step output…` |
| succeeded sem output | `No step output.` |

Fonte: `GetRun` → `events` (mesmo extrator `liveout` da TUI). Não
chamar Player. Não persistir um documento RESULT separado.

Já parcialmente no `develop` (`#68`). Slice 40 **não pode regressar**.

### G-247 — Hits inline (desktop)

**Status: CONFIRMED**

Não há view HITS. GRAPH **não** lista QueryHits.

| Onde | Fonte | O que mostra |
|---|---|---|
| LIVE | ContextPack nos eventos (`graph_hits`, `memory_hits`, `playbook_hits`) | kind/id/score; vazio = `No hits.` |
| INTENT Preview | `QueryHits` do Core com o texto do draft | budget de `19`; vazio não falha o preview |

Mesma semântica da TUI G-241, com botão **Preview** (e `Ctrl/Cmd+P`).

### G-248 — Blast painel (desktop)

**Status: CONFIRMED**

Não há view BLAST. GRAPH **não** dispara Blast.

- INTENT: botão **Blast** (e `Ctrl/Cmd+B`) — `CompileIntent` / JSON
  parse + `BlastTask`. **Não** cria Run, **não** Acquire.
- LIVE: botão **Blast** — `BlastTask` sobre o Task IR do snapshot.
- Painel: `risk` (símbolo+texto), `touches`, `conflicts`,
  `predicted_claims` se vierem, `affected` (walk `27`).
- Erro de blast: linha no painel; a janela não crasha.

Mesma semântica da TUI G-242, adaptada a mouse.

### G-249 — Bindings

**Status: CONFIRMED**

O service Wails (`35` G-162) ganha o que a TUI já chama:

| Binding | Core |
|---|---|
| `QueryHits` | wrapper sobre Graph (vazio se Graph nil) |
| `BlastTask` | já existe em `api.Core`; expor no service |

Frontend não importa `internal/core/graph` além dos DTOs JSON.
Testes de service com fake Core (Hits empty, Blast sem submit).

`liveout.FromEvents` permanece compartilhado TUI/desktop.

### G-250 — Chrome INTENT / LIVE

**Status: CONFIRMED**

- INTENT: textarea + **Preview** / **Submit** / **Blast** visíveis
  (não só atalhos). Preview mostra Task IR + Hits. Blast não navega
  para LIVE.
- Submit bem-sucedido → view LIVE com o `run_id` (G-144 / G-163).
- LIVE: status + **RESULT** no topo; Hits; Blast; events como
  detalhe secundário. HITL `Approve`/`Deny`/`Cancel` inalterados.
- Sem oitava view. Sem editar CONFIG.

### G-251 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- PTY, tuios, multiplexer, terminal vivo
- GRAPH canvas 2D; Blast/Hits a partir de GRAPH
- Mobile (iOS/Android); Wails como cliente HTTP de `serve`
- Trocar a TUI por este app; oitava view
- Gate de Execute via Blast; persistir report
- Wails v2; Tauri; Electron
- Edição de CONFIG; NATS (G-36)

---

## 4. Critérios de aceite

1. `runtgine desktop` continua uma janela, sete views, ordem G-161.
2. LIVE mostra RESULT com stdout de `step.succeeded` (fixture nos
   testes de `liveout` / service; smoke manual na janela).
3. Preview INTENT chama `QueryHits`; empty hits não quebra o Task IR.
4. Blast INTENT não chama `SubmitIntent` / `SubmitTask`.
5. Blast LIVE usa o Task do snapshot; GRAPH não tem controle de Blast.
6. Service testes verdes sem display (`go test ./internal/entrypoint/desktop`).
7. `go test ./...` e `go vet ./...` verdes (`CGO_ENABLED=0` no gate
   local; CI Linux pode ter WebKit).
8. Nenhuma chamada a Player a partir do desktop.

---

## 5. Ordem do slice de código (slice 40)

1. Expor `QueryHits` + `BlastTask` no service + fake
2. INTENT: Hits no Preview; botão Blast sem submit
3. LIVE: RESULT estável (não regressar `#68`); Hits do ContextPack;
   botão Blast
4. GRAPH sem Hits/Blast
5. Testes de binding; README Estágio: Slice 40
6. Arquivar OpenSpec `047`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-245 Papel / stack Wails v3
- [x] G-246 LIVE RESULT
- [x] G-247 Hits inline desktop
- [x] G-248 Blast painel desktop
- [x] G-249 Bindings `QueryHits` / `BlastTask`
- [x] G-250 Chrome INTENT / LIVE
- [x] G-251 Exclusões v0
