# 35 — Desktop Wails v0 (Entry Point)

Superfície desktop nativa do runtime: adapter Wails sobre a Core API
(`11` §13), mesma semântica da TUI Constellation. Fecha G-144 (INTENT
desktop) sem substituir CLI, TUI, Board ou HTTP API (`34`).

Inventário: [10-gaps.md](10-gaps.md) (G-35; recorte G-159+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Stack: [07-stack.md](07-stack.md). Visual: [14-tui-design.md](14-tui-design.md).
INTENT (semântica): [32-intent-surface-v0.md](32-intent-surface-v0.md) G-144.

**Status deste doc: CONFIRMED (v0 spec).** Slices 27–28 **feitas**
(app + sete views + Lessons HITL).

**Pacote OpenSpec:** [`openspec/specs/wails-v0/spec.md`](../openspec/specs/wails-v0/spec.md)
(archive [`openspec/changes/archive/2026-08-21-035-wails-v0/`](../openspec/changes/archive/2026-08-21-035-wails-v0/)).

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
| Wails **v3** (beta aceite) + Svelte 5 + shadcn-svelte | Wails v2; Tauri; Electron |
| Services Go → `api.Core` in-process | Cliente HTTP de `runtgine serve` |
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
| Wails | **v3** (módulo `github.com/wailsapp/wails/v3`; CLI `wails3`) |
| Frontend | Svelte 5 + TypeScript + Vite (já G-35; template `-t svelte`) |
| Componentes | shadcn-svelte: Sidebar, Card, Badge, Scroll Area; Command ⌘K opcional |
| Tokens | Constellation (`14`): Space, Panel, Starlight, Violet, Telemetry, Amber, Anomaly, Muted |
| Tema | honrar `prefers-color-scheme`; sem inventar tokens fora de `14` |
| Pin | tag beta corrente no `go.mod` no slice 27 (não dual-stack v2) |

G-35 (Wails + Svelte) permanece. O pin **v3** (em vez de v2 stable) é
produto: a API desktop da v3 está estável o bastante para produção
segundo o anúncio oficial da beta, e começar em v2 forçaria um port
completo (lifecycle, services, bindings) antes do GA.

v2 **não** entra neste recorte — nem como fallback, nem como
implementação temporária. Churn de beta é aceite; o `go.mod` trava a
revisão usada.

Docs: https://v3.wails.io/ · CLI:
`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`.

### G-161 — App shell e navegação

**Status: CONFIRMED**

Uma janela. Sete views, mesma ordem da TUI:

`INTENT` · `RUNS` · `LIVE` · `BOARD` · `EVENTS` · `GRAPH` · `CONFIG`

INTENT é o Entry Point visual (Mission Brief, `32`). Após submit
bem-sucedido → view LIVE com o `run_id` selecionado (G-144).

Não há overlapping windows, tray app, nem “workspace switcher”.
Wails v3 oferece multi-window; o v0 **não** usa.

### G-162 — Bindings (Core API)

**Status: CONFIRMED**

O backend Wails é um **service** v3 (não o modelo `Bind` da v2):
fachada fina sobre `api.Core`. Métodos mínimos:

| Binding | Core |
|---|---|
| `CompileIntent` / `SubmitIntent` / `SubmitTask` | iguais à TUI |
| `GetRun` / `ListRuns` / `Subscribe` | LIVE / RUNS / EVENTS |
| `CancelRun` / `ApproveRun` | HITL |
| `BlastTask` | CONFIG ou ACTION; não cria Run |
| `ConfigSnapshot` | aba CONFIG (sem dump de secrets) |
| `GetGraphSnapshot` / `RefreshGraph` | aba GRAPH |
| `ListLessons` / `ApproveLesson` / `RejectLesson` | HITL Lessons (slice 28 ok) |

Subscribe: o service emite eventos Wails v3 (`runtgine:event`) com o
JSON `event.Event`. Não reimplementar o Event Bus.

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
| Wails v2 | Dual-stack / port posterior; o recorte é v3 |
| Cliente de `runtgine serve` | Superfície local = in-process |
| Chamar Player / pular Validator | Entry Point ≠ Player |
| Thread de chat / RAG de transcript | `32` G-145 |
| Multi-window / tray / auto-update | Explode o v0 (v3 tem multi-window; não usamos) |
| Mobile (iOS/Android) | Experimental na v3; fora do desktop v0 |
| Server-mode Wails / plugins | Outro produto |
| Assinatura de loja (App Store, etc.) | Distribuição depois |
| Embed PTY / terminal multiplexer | Mesma regra da TUI |
| MCP, NATS, Memory Player | Outros tracks |
| Webhook inbound GitHub | Board continua polling (G-20) |
| Trocar a TUI por este app | TUI permanece |

### G-165 — Ordem e critérios

**Status: CONFIRMED**

| Slice | Entrega | Depende de |
|---|---|---|
| **27** | Scaffold Wails v3 + service Core + views INTENT e LIVE | Core API, `32` G-144 |
| **28** | RUNS, BOARD, EVENTS, GRAPH, CONFIG (+ Lessons HITL na UI) — **feito** | Slice 27 |

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

1. Spec deste doc (G-159..G-165) — feito.
2. **Slice 27** — app + INTENT/LIVE — feito.
3. **Slice 28** — demais views + Lessons HITL — feito.
4. Depois: MCP (G-44), mais Players (G-41), templates (`08`) — nova promoção.

---

## 6. Referências

- INTENT: [32](32-intent-surface-v0.md) G-144
- TUI design / tokens: [14](14-tui-design.md)
- Stack: [07](07-stack.md) · G-35 em [13](13-p2.md)
- Core API: [11](11-protocolo-v0.md) §13 / §16
- OpenSpec: [`openspec/specs/wails-v0/spec.md`](../openspec/specs/wails-v0/spec.md) · archive [`2026-08-21-035-wails-v0`](../openspec/changes/archive/2026-08-21-035-wails-v0/)
