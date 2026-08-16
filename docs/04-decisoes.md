# 04 — Decisoes arquiteturais

## Autoridade

Em conflito entre documentos:
1. Este arquivo (`04-decisoes.md`)
2. Documentos oficiais em `docs/` (exceto `00-rascunho.md`)
3. `AGENTS.md` / `README.md` / `REVIEW.md`
4. `brainstorm.md` e `conversas-empryo.md` (historicos; nao sao autoridade)

Status: CONFIRMED | HYPOTHESIS | OPEN QUESTION | REJECTED

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
| Rust (Core) | REJECTED | Adiado; stack atual e Go |
| GPUI | REJECTED | Exigiria Rust; preterido por Wails |
| Tauri | REJECTED | Preterido por Wails (dois runtimes) |
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
| Entry Point != Player | CONFIRMED | Entry Point traduz sinal externo |
| Intent Engine | HYPOTHESIS | Traduz intencao NL em Task IR |
| Task IR | HYPOTHESIS | Representacao intermediaria (v0 no MVP) |
| Task Validator | HYPOTHESIS | Valida antes de executar (basico no MVP) |
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

## MVP (corte canônico)

Ver [09-mvp.md](09-mvp.md). Decisoes-chave:

| Decisao | Status | Notas |
|---|---|---|
| Shell Player no MVP | CONFIRMED | Prova deterministic-first |
| CLI + TUI minimas no MVP | CONFIRMED | Superficies, nao produto |
| Board no MVP | CONFIRMED | Primeiro Entry Point de produto |
| Entrada estruturada (Task IR v0) no MVP | CONFIRMED | Sem depender de Intent Engine NL |
| Intent Engine NL fora do MVP Core | CONFIRMED | Permanece HYPOTHESIS / P1 |
| Wails fora do MVP | CONFIRMED | Fase 3 |
| Escopo detalhado em 09-mvp.md | CONFIRMED | Prevalece sobre rascunhos |

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
- Entry Point != Player
- Runtgine + Chorus complementares
- Nativo (nao Electron); GPUI/Tauri/Rust-Core rejeitados para o caminho atual
- Biblioteca grande de Players deterministicos (visao)
- MVP: Core + Shell + CLI/TUI + Board (ver 09-mvp.md)
