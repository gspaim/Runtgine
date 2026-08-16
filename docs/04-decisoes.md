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
| Wails | CONFIRMED | Desktop (Go + Svelte) |
| Canal Go (Event Bus) | CONFIRMED | Pub/sub in-process |
| JSON + JSON Schema | CONFIRMED | Protocolo e contratos |
| SQLite (modernc.org/sqlite) | CONFIRMED | Persistencia local; pure Go |
| log/slog | CONFIRMED | Logger padrao — sessao de fechamento |
| NATS (futuro) | DEFERRED | Event Bus distribuido; interface plugavel no Core |
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
| Task IR | CONFIRMED (v0) | Schema em 11-protocolo-v0; NL Intent Engine ainda HYPOTHESIS |
| Task Validator | CONFIRMED (v0) | Subset MVP: capabilities, inputs, schemas; ver 11 |
| Runtime Graph | HYPOTHESIS | Memoria estrutural |
| Context Engine | HYPOTHESIS | Monta contexto relevante |
| Project Memory | HYPOTHESIS | Memoria episodica / de projeto; ver `16` |
| Player Router | HYPOTHESIS | Roteia por capability + custo |
| Execution Policy | HYPOTHESIS | Regras de seguranca |
| Resource Claim | HYPOTHESIS | Bloqueio concorrente |
| Blast Radius | HYPOTHESIS | Impact analysis |
| Many deterministic Players | CONFIRMED | Estrategico |
| Runtgine + Chorus | CONFIRMED | Complementares |
| Event Bus in-process (MVP) | CONFIRMED | Canais Go |
| Nativo (nao Electron) | CONFIRMED | Wails |
| Runner v0 | CONFIRMED | Orchestrator minimo do MVP |

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
| Task IR | CONFIRMED (v0) | Schema em 11-protocolo-v0 |
| Task Validator | HYPOTHESIS | Valida antes de executar |
| Runtime Graph | HYPOTHESIS | Memoria estrutural |
| Context Engine | HYPOTHESIS | Monta contexto relevante |
| Project Memory | HYPOTHESIS | Memoria episodica / de projeto; ver `16` |

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

## Protocolo v0 — confirmado (sessao de fechamento)

Inventario: [10-gaps.md](10-gaps.md).
Texto completo: [11-protocolo-v0.md](11-protocolo-v0.md).

**Bloco P0 CONFIRMADO.** Liberado iniciar implementacao do Core.
Gaps P1 (Board/LLM) permanecem abertos.

| Proposta | Status | Notas |
|---|---|---|
| JSON canonico; YAML so na borda CLI | CONFIRMED | G-14 |
| Capability naming `domain.action` | CONFIRMED | G-05 |
| IDs UUID v7; schema_version semver | CONFIRMED | time-ordered; trocado de v4 na sessao |
| Task IR v0 schema | CONFIRMED | G-01 |
| Manifest v0 schema | CONFIRMED | G-02 |
| Event envelope + tipos minimos | CONFIRMED | G-03/G-04 |
| Result/Error + Run lifecycle | CONFIRMED | G-08/G-09 |
| Runner v0 (Orchestrator minimo) | CONFIRMED | G-10; Plan passthrough G-11 |
| Queue in-memory FIFO | CONFIRMED | G-12 multi-run |
| Persistencia MVP Core = SQLite cedo | CONFIRMED | G-13 variante B |
| Core API SubmitTask/GetRun/Subscribe | CONFIRMED | G-07 |
| Shell sandbox v0 (argv, workdir, timeout) | CONFIRMED | G-06/G-18 |
| log/slog | CONFIRMED | G-16 |
| SQLite via modernc.org/sqlite | CONFIRMED | G-15 |
| Go 1.25+; module github.com/gspaim/Runtgine | CONFIRMED | G-37; atualizado pelo Charm v2 no Slice 3 |
| Layout `cmd/` + `internal/core|players|entrypoint` | CONFIRMED | G-17 |

### Desvios em relacao a proposta inicial

- Queue: multi-run concorrente (nao 1 run so)
- Persistencia: SQLite cedo (nao so memoria)
- IDs: UUID **v7** (nao v4) — melhor localidade em SQLite / ordem temporal

### Resolvido nesta sessao

- Capability Resolver / Planner → absorvidos no Runner v0
- Event Store no MVP → events append-only em SQLite (nao event sourcing)
- Task IR / Validator basico / Runner v0 → CONFIRMED

## Engenharia (P2) — CONFIRMADO

Ver [13-p2.md](13-p2.md).

| Item | Status | Notas |
|---|---|---|
| G-30 Cancel/timeout/retry/concorrencia | CONFIRMED | Retry automatico configuravel por step (B) |
| G-31 Observabilidade | CONFIRMED | slog + correlacao; sem OTel no MVP |
| G-32 Runtgine ↔ Chorus | CONFIRMED | Independencia total no MVP |
| G-33 Workspaces / worktrees | CONFIRMED | Um workspace_root; store em `.runtgine/` |
| G-34 Testes | CONFIRMED | Unit + integracao + smoke; LLM mockado |
| G-35 Wails Svelte vs React | CONFIRMED | Wails mantido; frontend Svelte |
| G-36 NATS | DEFERRED | Bus plugavel; sem NATS no MVP |
| G-38 Config runtime | CONFIRMED | defaults < file < env < flags |

## TUI — Constellation Mission Control

Ver [14-tui-design.md](14-tui-design.md).

| Decisao | Status | Notas |
|---|---|---|
| Sistema visual Constellation Mission Control | CONFIRMED | Mission Control + constelacoes |
| Bubble Tea + Lip Gloss + Bubbles | CONFIRMED | TUI moderna e responsiva |
| Tabs Runs / Live / Board / Events / Config | CONFIRMED | Estrutura principal |
| Tema espacial e visual, nao dominio | CONFIRMED | Manter Task/Run/Step/Event/Player |
| TUI usa apenas APIs do Core | CONFIRMED | Nunca chama Player diretamente |
| Charm stack v2 via `charm.land/*` | IMPLEMENTED | Requer Go 1.25+ |
| Config da TUI read-only e secrets mascarados | IMPLEMENTED | Snapshot publico contem apenas estado/config nao sensivel |
| tuios no MVP | REJECTED | Nao e multiplexer; PTY futuro exige nova decisao |

## Board / pipeline (P1) — CONFIRMADO

Ver [12-board-p1.md](12-board-p1.md).

| Item | Status | Notas |
|---|---|---|
| G-20 Card → Task IR (adapter + polling) | CONFIRMED | Mapeamento titulo/body/ref; token via env |
| G-21 Write-back no board | CONFIRMED | Status + comentario; sem subtasks no board |
| G-22 Contratos por etapa | CONFIRMED | capabilities `pipeline.*`; steps lineares |
| G-23 Regras vs LLM | CONFIRMED | repo-search/effort/difficulty det.; reviews LLM |
| G-24 Context assembly basico | CONFIRMED | ContextPack v0; AssembleContext no Core |
| G-25 LLM Player v0 | CONFIRMED | Interface unica; backends OpenAI-compat + Anthropic |
| G-26 Task Router basico | CONFIRMED | Regras: capability → deterministic → default AI |
| G-27 Subtasks | CONFIRMED | SQLite + child runs (`parent_run_id`) |

## Project Memory (P3) — HYPOTHESIS

Ver [16-project-memory.md](16-project-memory.md). Esboco reforçado (revisao
conceitual); **nao autoriza codigo**. Nada abaixo e CONFIRMED.

| Decisao | Status | Notas |
|---|---|---|
| Project Memory (episodica / de projeto) | HYPOTHESIS | Continuacao entre runs no mesmo projeto (G-46) |
| Tres memorias: temporal / estrutural / episodica | HYPOTHESIS | Event Store ≠ Runtime Graph ≠ Project Memory |
| Fato historico ≠ status operacional (validade) | HYPOTHESIS | Candidatos: `active`/`superseded`/`archived` ou layers operational/historical |
| Memory ≠ Knowledge (evolucao possivel) | HYPOTHESIS | Episodio vs consolidado; sem novo subsystem agora |
| Memory Provider → ContextPack | HYPOTHESIS | Default conceitual de acesso (AssembleContext) |
| Integracao inicial via sidecar / MCP | HYPOTHESIS | Fases A/B experimentais; depende G-44 se MCP |
| Extensao ContextPack (`memory_hits` + budget + hierarquia) | HYPOTHESIS | Rascunho experimental; prioridade menor que task/estado atual |
| Memory Player (`memory.*`) | OPEN QUESTION | So se steps do Plan exigirem; ver G-47 |
| Embutir ai-memory (Rust) no Core | REJECTED | Dominio e stack ortogonais; sidecar/Provider apenas |
| Memoria como autoridade de execucao | REJECTED | Sugere contexto; nunca capability/policy/Validator bypass |
| Supersession silenciosa via LLM no Core | REJECTED | Validade so com opt-in explicito |
| RAG generico / indexar transcripts como produto | REJECTED | Compile observations; nao chat retrieval |

## Git / release — fluxo de branches

Ver [15-git-workflow.md](15-git-workflow.md).

| Decisao | Status | Notas |
|---|---|---|
| Fluxo `feat → develop → release → main` | CONFIRMED | Integracao em `develop`; estabilizacao em `release/*`; estavel em `main` |
| Prefixo de feature `feat/<NNN>-<slug>` | CONFIRMED | `NNN` = id da spec/issue (ex.: `feat/001-shell-player`) |
| Outros prefixos `fix/` `docs/` `chore/` | CONFIRMED | Mesmo padrao numerico quando houver issue/spec |
| Release candidates | CONFIRMED | Branch `release/x.y.z` + tags `vX.Y.Z-rc.N` |
| Release estavel | CONFIRMED | Merge `release/*` → `main` + tag `vX.Y.Z`; back-merge para `develop` |
| Semver do produto | CONFIRMED | Tags Git do binario/CLI; distinto de `schema_version` do Task IR |
| CI em PR (test + vet) | CONFIRMED | GitHub Actions; ver `.github/workflows/` |
| Branch protection | CONFIRMED | Recomendado em `develop`/`main`/`release/*`; configurar no GitHub |
