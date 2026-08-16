# 04 — Decisoes arquiteturais

## Stack

| Tecnologia | Status | Notas |
|---|---|---|
| Go | CONFIRMED | Linguagem principal do Core |
| Cobra | CONFIRMED | CLI |
| Bubble Tea | CONFIRMED | TUI |
| Lip Gloss + Bubbles | CONFIRMED | Estilizacao e componentes TUI |
| Wails | CONFIRMED | Desktop (Go + Svelte/React) |
| Canal Go (Event Bus) | CONFIRMED | Pub/sub in-process |
| JSON + JSON Schema | CONFIRMED | Protocolo e contratos |
| SQLite (mattn/modernc) | CONFIRMED | Persistencia local |
| log/slog | HYPOTHESIS | Logger padrao |
| NATS (futuro) | OPEN QUESTION | Event Bus distribuido |
| Rust (Core) | OPEN QUESTION | Nao descartado, adiado |
| GPUI | OPEN QUESTION | Exigiria Rust no Core |
| Tauri | OPEN QUESTION | Preterido por Wails |
| Git | CONFIRMED | Version control |

## Arquitetura

| Decisao | Status | Notas |
|---|---|---|
| Deterministic-first | CONFIRMED | Preferir deterministico |
| Player abstraction | CONFIRMED | Abstracao central |
| Event-driven | CONFIRMED | Task -> Event -> Queue -> Player |
| Capability routing | CONFIRMED | Runtime pensa em capabilities |
| Core = produto | CONFIRMED | Interface e superficie |
| Core independente de UI | CONFIRMED | Core funciona sem TUI/CLI |
| LLM-agnostic | CONFIRMED | Players LLM sao um tipo entre outros |
| Task != Workflow != ExecPlan | CONFIRMED | Tres conceitos distintos |
| Event != Queue != Workflow | CONFIRMED | Tres conceitos distintos |
| Intent Engine | HYPOTHESIS | Traduz intencao em Task IR |
| Task Validator | HYPOTHESIS | Valida antes de executar |
| Runtime Graph | HYPOTHESIS | Memoria estrutural |
| Context Engine | HYPOTHESIS | Monta contexto relevante |
| Player Router | HYPOTHESIS | Roteia por capability + custo |
| Execution Policy | HYPOTHESIS | Regras de seguranca |
| Resource Claim | HYPOTHESIS | Bloqueio concorrente |
| Blast Radius | HYPOTHESIS | Impact analysis |
| Many deterministic Players | CONFIRMED | Estrategico |
| Runtgine + Chorus | CONFIRMED | Complementares |
| Event Bus in-process (MVP) | CONFIRMED | Canais Go |
| Nativo (nao Electron) | CONFIRMED | Wails |

## Modelo conceitual

| Conceito | Status | Notas |
|---|---|---|
| Task | CONFIRMED | Intencao/pedido do usuario |
| Workflow | CONFIRMED | Estrutura reutilizavel |
| Execution Plan | CONFIRMED | Plano para UMA execucao |
| Player | CONFIRMED | Entidade com capabilities |
| Event | CONFIRMED | Algo aconteceu |
| Queue | CONFIRMED | Trabalho aguardando |
| Intent Engine | HYPOTHESIS | Traduz intencao |
| Task IR | HYPOTHESIS | Representacao intermediaria |
| Task Validator | HYPOTHESIS | Valida antes de executar |
| Runtime Graph | HYPOTHESIS | Memoria estrutural |
| Context Engine | HYPOTHESIS | Monta contexto relevante |

## Decisoes CONFIRMED (visao geral)

- Go, Cobra, Bubble Tea, Wails, JSON/JSON Schema, SQLite
- Canal Go como Event Bus (MVP)
- Core independente de UI
- Deterministic-first
- Player abstraction + Capability routing
- Task != Workflow != Execution Plan
- Event != Queue != Workflow
- Runtgine + Chorus complementares
- Nativo (nao Electron)
- Biblioteca grande de Players deterministicos (visao)