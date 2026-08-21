# 35 — Desktop Wails v0 (Entry Point)

Superfície desktop nativa do runtime: adapter Wails sobre a Core API
(`11` §13), mesma semântica da TUI Constellation. Fecha G-144 (INTENT
desktop) sem substituir CLI, TUI, Board ou HTTP API (`34`).

Inventário: [10-gaps.md](10-gaps.md) (G-35; recorte G-159+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Stack: [07-stack.md](07-stack.md). Visual: [14-tui-design.md](14-tui-design.md).
INTENT (semântica): [32-intent-surface-v0.md](32-intent-surface-v0.md) G-144.

**Status deste doc: CONFIRMED (v0 spec).** Código **não** iniciado.
Duas fatias: slice 27 (app + bindings + INTENT/LIVE), slice 28
(demais views: RUNS, BOARD, EVENTS, GRAPH, CONFIG).

**Pacote OpenSpec:** [`openspec/changes/035-wails-v0/`](../openspec/changes/035-wails-v0/).

---

## 1. Problema

CLI, TUI e `runtgine serve` já cobrem scripts, terminal e CI. O stack
desktop (Wails + Svelte) está CONFIRMED desde `07` / G-35, e G-144
já fixou a semântica da view INTENT — mas **não há app**. Falta o
Entry Point nativo para operador em janela, in-process, com a paleta
Constellation.

A TUI valida a linguagem de interação. O desktop **espelha**; não
reinventa o protocolo.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Entry Point adapter (`internal/entrypoint/desktop`) | Player |
| Wails **v2** (stable) + Svelte 5 + shadcn-svelte | Wails v3 beta; Tauri; Electron |
| Bindings Go → `api.Core` in-process | Cliente HTTP de `runtgine serve` |
| Sete views alinhadas à TUI | Multiplexer / PTY / IDE |
| `source.entry_point = "wails"` | Chatbot; thread LLM na UI |

Regras (inalteradas):

1. Entry Point ≠ Player. O desktop **não** chama Player.
2. Validator / Registry soberanos.
3. Core não importa `entrypoint` (`11` §16).
4. Distinto da HTTP API (`34`) e do HTTP Player (`28`).

---

## 3. Cortes confirmados (G-159+)

### G-159 — Papel e pacote

**Status: CONFIRMED**

- Nome: **Desktop Wails v0**. Comando: `runtgine desktop`.
- Pacote: `internal/entrypoint/desktop` (Go + `frontend/` Svelte).
- Um processo = um `workspace_root` (G-33), o mesmo `api.Open` da CLI/TUI.
- `source.entry_point` = `"wails"` quando a UI submete (salvo o cliente
  enviar `source` já válido). Schema Task IR: enum ganha `wails`.

### G-160 — Stack pin

**Status: CONFIRMED**

| Peça | Corte v0 |
|---|---|
| Wails | **v2** (stable). Wails v3 (beta) **fora** do v0 |
| Frontend | Svelte 5 + TypeScript + Vite (já G-35) |
| Componentes | shadcn-svelte: Sidebar, Card, Badge, Scroll Area; Command ⌘K opcional |
| Tokens | Constellation (`14`): Space, Panel, Starlight, Violet, Telemetry, Amber, Anomaly, Muted |
| Tema | honrar `prefers-color-scheme`; sem inventar tokens fora de `14` |

G-35 (Wails + Svelte) permanece. Este recorte só **trava a major** v2
para não portar o v0 no meio de uma rewrite beta.

### G-161 — App shell e navegação

**Status: CONFIRMED**

Uma janela. Sete views, mesma ordem da TUI:

`INTENT` · `RUNS` · `LIVE` · `BOARD` · `EVENTS` · `GRAPH` · `CONFIG`

INTENT é o Entry Point visual (Mission Brief, `32`). Após submit
bem-sucedido → view LIVE com o `run_id` selecionado (G-144).

Não há overlapping windows, tray app, nem “workspace switcher”.

### G-162 — Bindings (Core API)

**Status: CONFIRMED**

O backend Wails é uma fachada fina sobre `api.Core`. Métodos mínimos:

| Binding | Core |
|---|---|
| `CompileIntent` / `SubmitIntent` / `SubmitTask` | iguais à TUI |
| `GetRun` / `ListRuns` / `Subscribe` | LIVE / RUNS / EVENTS |
| `CancelRun` / `ApproveRun` | HITL |
| `BlastTask` | CONFIG ou ACTION; não cria Run |
| `ConfigSnapshot` | aba CONFIG (sem dump de secrets) |
| `GetGraphSnapshot` / `RefreshGraph` | aba GRAPH |
| `ListLessons` / `ApproveLesson` / `RejectLesson` | HITL Lessons (slice 28 ok) |

Subscribe: o binding emite eventos para o frontend (callback / event
Wails). Não reimplementar o Event Bus.

Testes de binding: fake `api.Core` (mesmo espírito da TUI). CI **não**
exige janela GUI.

### G-163 — INTENT desktop (fecha G-144)

**Status: CONFIRMED**

Mesma semântica da TUI (`32` G-143):

- NL → preview (`CompileIntent`) **sem** `InsertRun`.
- NL → submit (`SubmitIntent`).
- JSON Task IR válido → `SubmitTask`.
- Não é chatbot: saída primária = Task IR + `run_id`.
- Capability inventada → erro de Validator visível; sem crash.

Atalhos v0 (espelho TUI, adaptados a desktop):

- Preview: `Ctrl+P` / `Cmd+P`
- Submit: `Ctrl+Enter` / `Cmd+Enter`
- JSON toggle: `Ctrl+J` / `Cmd+J`

### G-164 — Exclusões v0

**Status: CONFIRMED**

| Fica fora | Por quê |
|---|---|
| Wails v3 | Beta; v2 é o stable atual |
| Cliente de `runtgine serve` | Superfície local = in-process |
| Chamar Player / pular Validator | Entry Point ≠ Player |
| Thread de chat / RAG de transcript | `32` G-145 |
| Multi-window / tray / auto-update | Explode o v0 |
| Assinatura de loja (App Store, etc.) | Distribuição depois |
| Embed PTY / terminal multiplexer | Mesma regra da TUI |
| MCP, NATS, Memory Player | Outros tracks |
| Webhook inbound GitHub | Board continua polling (G-20) |
| Trocar a TUI por este app | TUI permanece |

### G-165 — Ordem e critérios

**Status: CONFIRMED**

| Slice | Entrega | Depende de |
|---|---|---|
| **27** | Scaffold Wails v2 + bindings + views INTENT e LIVE | Core API, `32` G-144 |
| **28** | RUNS, BOARD, EVENTS, GRAPH, CONFIG (+ Lessons HITL na UI) | Slice 27 |

Critérios slice 27:

- `runtgine desktop` abre uma janela (smoke manual).
- Preview `git status` → Task IR `git.status`, zero Run.
- Submit NL → `run_id` → LIVE.
- Bindings unit-tested com Core fake; `go test ./...` verde sem display.
- `source.entry_point` nas Tasks da UI = `wails`.

Critérios slice 28:

- Sete views navegáveis; GRAPH read-only via `GetGraphSnapshot`.
- CONFIG não revela tokens.
- HITL approve/deny e Lessons list/approve/reject na UI.
- `go test ./...` / `go vet ./...` verdes.

---

## 4. Relação com outras superfícies

| Superfície | Papel |
|---|---|
| CLI | Scripts / CI local |
| TUI | Operador no terminal; referência de interação |
| HTTP API (`34`) | CI/CD remoto; o desktop **não** é cliente dela no v0 |
| Board | Cards GitHub; polling |
| Desktop (`35`) | Operador em janela; in-process |

Todas convergem para o mesmo protocolo (`11`).

---

## 5. Ordem de implementação

1. Spec deste doc (G-159..G-165) — este PR.
2. **Slice 27** — app + INTENT/LIVE.
3. **Slice 28** — demais views.
4. Depois: MCP (G-44), mais Players (G-41), templates (`08`) — nova promoção.

Código Wails **não** entra neste PR de spec.

---

## 6. Referências

- INTENT: [32](32-intent-surface-v0.md) G-144
- TUI design / tokens: [14](14-tui-design.md)
- Stack: [07](07-stack.md) · G-35 em [13](13-p2.md)
- Core API: [11](11-protocolo-v0.md) §13 / §16
- OpenSpec: [`openspec/changes/035-wails-v0/`](../openspec/changes/035-wails-v0/)
