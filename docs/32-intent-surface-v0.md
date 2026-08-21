# 32 — Intent Surface v0 (Mission Brief)

Superfície visual de **Entry Point** para compilar intenção em NL (ou Task IR
JSON) e submeter Runs. Cobre a aba **INTENT** na TUI Constellation e a mesma
experiência no desktop Wails (spec [35](35-wails-v0.md); slices 27–28).

Inventário: [10-gaps.md](10-gaps.md) (G-141+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Intent Engine (compilador): [17-intent-engine-v0.md](17-intent-engine-v0.md).
TUI: [14-tui-design.md](14-tui-design.md) · Desktop: [35-wails-v0.md](35-wails-v0.md)
(Wails v3 + Svelte + shadcn; [07-stack.md](07-stack.md)).

**Status deste doc: CONFIRMED (v0).** Código TUI = slice 21 **feito**.
Wails INTENT = spec [35](35-wails-v0.md) (slices 27–28 feitas).

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-21-032-intent-surface/`](../openspec/changes/archive/2026-08-21-032-intent-surface/).

---

## 1. Problema

O Intent Engine v0 e `runtgine intent` já existem na CLI. A TUI e o futuro
desktop Constellation focam **observação** (RUNS, LIVE, EVENTS, …) — o
operador não tem superfície visual para **mandar intenção** sem sair para o
terminal.

Isso gera desalinhamento de expectativa: o produto recebe NL, mas a UI parece
só “dashboard de telemetria”.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Entry Point visual (TUI + Wails) | Chatbot / agente conversacional |
| Compilador NL → Task IR → Submit | Player; bypass do Validator |
| Preview (`CompileIntent` / dry-run) antes de submit | Thread infinita estilo ChatGPT |
| Uma aba/painel **INTENT** (Mission Brief) | Substituto da CLI `runtgine intent` |
| Reuso de `CompileIntent` / `SubmitIntent` | Novo `intent.*` Player |
| Após submit → observar em LIVE/RUNS | UI chamar Shell/LLM Player direto |

Regras (inalteradas vs [03-principios.md](03-principios.md)):

1. Fluxo: **NL ou JSON → Intent Engine → Task IR → Validator → Runner**.
2. Runtgine **não** é chatbot ([01-visao.md](01-visao.md)).
3. Project Memory **não** arquiva transcript de chat como produto ([16](16-project-memory.md)).
4. TUI/Wails usam **somente** APIs públicas do Core.

---

## 3. Cortes confirmados (G-141+)

### G-141 — Papel e nome

**Status: CONFIRMED**

- Nome de produto na UI: **INTENT** (subtítulo opcional: *Mission Brief*).
- É **Entry Point**, não Player. Traduz sinal humano → Task IR → `SubmitTask`.
- Visual pode lembrar “campo de chat”, mas o contrato é **compilar + executar**,
  não conversar.

### G-142 — TUI: aba INTENT

**Status: CONFIRMED**

Ordem das abas (sétima aba; **INTENT primeiro** — ponto de entrada):

```text
[ INTENT ] [ RUNS ] [ LIVE ] [ BOARD ] [ EVENTS ] [ GRAPH ] [ CONFIG ]
```

Conteúdo mínimo v0:

- Campo de texto multilinha (NL) **ou** modo JSON Task IR (toggle simples).
- Painel **Preview**: Task IR pretty-printed + `method` (`heuristic.*` | `llm`).
- Erros de `CompileIntent` / Validator na admissão — mensagem clara, sem crash.
- Histórico **curto** da sessão TUI (últimas N submissões, só `run_id` + resumo);
  **não** persistir transcript longo no SQLite.

Teclas v0 (footer atualizado):

| Tecla | Ação |
|---|---|
| `tab` / `shift+tab` | navegar abas |
| `Ctrl+p` | preview / dry-run (`CompileIntent`) |
| `Ctrl+Enter` | submit (`SubmitIntent` → `run_id`) |
| Após submit OK | selecionar run + ir para **LIVE** |
| `Esc` | limpar input (confirmar se preview dirty) |

Responsividade: em `< 80` colunas, preview colapsa abaixo do input; sem perder
submit/preview.

### G-143 — Fluxo Core

**Status: CONFIRMED**

```text
input text + source ("tui" | "wails")
  → CompileIntent(text, source)   // preview
  → SubmitIntent(text, source)    // submit (= Compile + SubmitTask)
  → run_id
  → TUI/Wails: GetRun + Subscribe; tab LIVE
```

- Preview **não** submete Run.
- Submit usa o **mesmo** Validator/`SubmitTask` que CLI e Board.
- Task IR JSON colado no modo JSON: validar schema; se válido, `SubmitTask`
  direto (sem Intent Engine), equivalente a `runtgine run -`.

CoreAPI na TUI (slice 21): estender interface com `CompileIntent` /
`SubmitIntent` (já existem em `api.Core`).

### G-144 — Desktop Wails (spec `35`)

**Status: CONFIRMED** — app e corte em [35-wails-v0.md](35-wails-v0.md)
(G-159..G-165; slices 27–28).

Mesma semântica da aba TUI INTENT; stack [07-stack.md](07-stack.md):

- **Svelte 5 + shadcn-svelte** ([Sidebar](https://www.shadcn-svelte.com/docs/components/sidebar),
  [Command](https://www.shadcn-svelte.com/docs/components/command) opcional ⌘K,
  [Card](https://www.shadcn-svelte.com/docs/components/card),
  [Badge](https://www.shadcn-svelte.com/docs/components/badge),
  [Scroll Area](https://www.shadcn-svelte.com/docs/components/scroll-area)).
- Paleta Constellation (`14` § tokens).
- Bindings Wails → `CompileIntent` / `SubmitIntent`; `source: "wails"`.
- Após submit: navegar para view **Live** com o `run_id` retornado.

Implementação Wails **não** bloqueia slice 21 da TUI.

### G-145 — Exclusões v0

**Status: CONFIRMED**

| Fica fora | Por quê |
|---|---|
| Thread multi-turn com memória de conversa | Produto ≠ chatbot |
| Respostas em prosa do LLM na superfície INTENT | Saída = Task IR, não chat |
| Intent Engine como Player (`intent.*`) | Compilador no Core |
| Indexar / buscar transcripts na Project Memory | [16](16-project-memory.md) |
| NL na TUI durante HITL | HITL continua em LIVE/RUNS (`a`/`d`) |
| Editar Execution Plan na UI | Fora do Entry Point v0 |
| Substituir `runtgine intent` CLI | CLI permanece para scripts/CI |

### G-146 — Critérios de pronto

**Status: CONFIRMED**

**TUI (slice 21):**

- Aba INTENT visível; ordem conforme G-142.
- `Ctrl+p` mostra Task IR para `git status` (heuristic) sem submit.
- `Ctrl+Enter` submete NL → `run_id` → LIVE com run selecionado.
- Capability inventada → erro de Validator antes de executar.
- Testes: model/update com fake Core; sem TTY interativo obrigatório em CI.

**Wails (spec `35`, slices 27–28):**

- Painel INTENT equivalente; submit → Live view.
- Mesmos critérios de Validator/soberania.

---

## 4. Relação com outras superfícies

| Superfície | Papel |
|---|---|
| CLI `runtgine intent` | NL + flags `--dry-run` / `--wait` |
| CLI `runtgine run` | Task IR JSON arquivo |
| Board | Cards GitHub → Task IR |
| TUI INTENT | NL/JSON visual → Run |
| Wails INTENT | Idem desktop |

Todas convergem para o mesmo protocolo ([11-protocolo-v0.md](11-protocolo-v0.md)).

---

## 5. Ordem de implementação

1. **Slice 21** — TUI aba INTENT (`internal/entrypoint/tui`) + skill + testes.
2. **Slices 27–28** — Desktop Wails (`35`) — view INTENT + demais views.

Não depende de HTTP API (`34` / G-45), NATS, nem novos Players.

---

## 6. Referências

- Intent Engine API: [17](17-intent-engine-v0.md) G-51
- TUI design system: [14](14-tui-design.md)
- Desktop Wails: [35](35-wails-v0.md)
- OpenSpec: [`openspec/changes/archive/2026-08-21-032-intent-surface/`](../openspec/changes/archive/2026-08-21-032-intent-surface/)
