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
| Wails | CONFIRMED | Desktop (Go + Svelte); corte v0 em `35` (Wails **v3** beta) |
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
| Intent Engine | CONFIRMED (v0) | NL → Task IR; ver `17-intent-engine-v0.md` |
| Task IR | CONFIRMED (v0) | Schema em 11-protocolo-v0; NL via Intent Engine v0 |
| Task Validator | CONFIRMED (v0) | Subset MVP: capabilities, inputs, schemas; ver 11 |
| Runtime Graph | CONFIRMED (v0) | Memoria estrutural; ver `18-runtime-graph-v0.md` |
| Context Engine | CONFIRMED (v0) | Semente `repo_hits`; ver `31-context-engine-v0.md` |
| Project Memory | CONFIRMED (v0) | Episodica; ver `29-project-memory-v0.md` |
| Player Router | CONFIRMED + implementado | Multi-model routing; ver `33-evolution-v0.md` (G-147); slice 22 |
| Execution Policy | CONFIRMED (v0) | allow/deny/approval-required; ver `22-execution-policy-v0.md` |
| Resource Claim | CONFIRMED (v0) | Bloqueio concorrente; ver `24-resource-claims-v0.md` |
| Blast Radius | CONFIRMED (v0) | Relatorio Task IR; ver `25-blast-radius-v0.md` |
| Blast Graph Walk | CONFIRMED (v0) | 1 hop mentions; ver `27-blast-graph-walk-v0.md` |
| TUI GRAPH tab | CONFIRMED (v0) | Aba read-only; ver `26-tui-graph-v0.md` + `14` |
| HTTP Player | CONFIRMED (v0) | Cliente GET/HEAD HTTPS; ver `28-http-player-v0.md` |
| Test Player | CONFIRMED (v0) | `test.go` no workspace; ver `30-test-player-v0.md` |
| Many deterministic Players | CONFIRMED | Estrategico; recortes G-41: infra em `41`, Helm em `42` |
| Workflow Templates v0 | CONFIRMED (v0) | JSON nativo no workspace; ver `40-workflow-templates-v0.md` |
| Runtgine + Chorus | CONFIRMED | Complementares |
| Event Bus in-process (MVP) | CONFIRMED | Canais Go |
| Nativo (nao Electron) | CONFIRMED | Wails |
| Runner v0 | CONFIRMED | Orchestrator minimo do MVP |
| HTTP API v0 | CONFIRMED + implementado | Entry Point `runtgine serve`; ver `34`; slices 25–26 |
| Desktop Wails v0 | CONFIRMED + implementado | Entry Point `runtgine desktop`; Wails **v3**; slices 27–28 |

## MVP (corte canônico)

Ver [09-mvp.md](09-mvp.md). Decisoes-chave:

| Decisao | Status | Notas |
|---|---|---|
| Shell Player no MVP | CONFIRMED | Prova deterministic-first |
| CLI + TUI minimas no MVP | CONFIRMED | Superficies, nao produto |
| Board no MVP | CONFIRMED | Primeiro Entry Point de produto |
| Entrada estruturada (Task IR v0) no MVP | CONFIRMED | Sem depender de Intent Engine NL |
| Intent Engine NL v0 (pos-Core) | CONFIRMED | Ver `17`; heuristicas Player = slice 19 feito (G-135..G-136) |
| Context Engine v0 no 1.0 magro | CONFIRMED | Ver `31`; slice 20 feito (G-137..G-139) |
| API HTTP (G-45) fora do 1.0 | CONFIRMED | Spec v0 em `34`; slices 25–26 feitas; 1.0 magro continua CLI |
| Wails fora do MVP / 1.0 | CONFIRMED | Fase 3; spec v0 em `35` (Wails **v3**); codigo slices 27–28 |
| Escopo detalhado em 09-mvp.md | CONFIRMED | Realizado (slices 1–26; 1.0 magro + INTENT/Evolution/HTTP) |

## Modelo conceitual

| Conceito | Status | Notas |
|---|---|---|
| Task | CONFIRMED | Intencao/pedido do usuario |
| Workflow | CONFIRMED | Estrutura reutilizavel |
| Execution Plan | CONFIRMED | Plano para UMA execucao |
| Player | CONFIRMED | Entidade com capabilities |
| Event | CONFIRMED | Algo aconteceu |
| Queue | CONFIRMED | Trabalho aguardando |
| Intent Engine | CONFIRMED (v0) | Traduz intencao NL → Task IR; ver `17` |
| Task IR | CONFIRMED (v0) | Schema em 11-protocolo-v0 |
| Task Validator | HYPOTHESIS | Valida antes de executar |
| Runtime Graph | CONFIRMED (v0) | Memoria estrutural; ver `18-runtime-graph-v0.md` |
| Context Engine | CONFIRMED (v0) | Semente `repo_hits`; ver `31-context-engine-v0.md` |
| Project Memory | CONFIRMED (v0) | Episodica; ver `29-project-memory-v0.md` |

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
- MVP realizado (slices 1–18) + 1.0 magro (G-135..G-140; ver `09`)
- Runtime Graph v0 (G-60..G-65)
- Graph Hits v0 (G-66..G-69; ContextPack + Intent)
- Git Player v0 (G-70..G-74; recorte G-41)
- Filesystem Player v0 (G-75..G-80; recorte G-41)
- Execution Policy + HITL v0 (G-81..G-86; recorte G-42) — spec; codigo = slice 10
- Docker Player v0 (G-87..G-92; recorte G-41) — spec; codigo = slice 11; depende de 022
- Resource Claims v0 (G-93..G-98; recorte G-43) — spec; codigo = slice 12 — feito
- Blast Radius v0 (G-99..G-104; resto de G-43) — spec; codigo = slice 13 — feito
- TUI GRAPH v0 (G-105..G-110) — spec; codigo = slice 14 — feito; altera `14` + skill
- Walk Blast←Graph v0 (G-111..G-116) — spec; codigo = slice 15 — feito
- HTTP Player v0 (G-117..G-122; recorte G-41) — spec; codigo = slice 16 — feito
- Project Memory v0 (G-123..G-128; recorte G-46/G-47) — spec; codigo = slice 17 — feito
- Test Player v0 (G-129..G-134; recorte G-41) — spec; codigo = slice 18 — feito
- MVP 1.0 magro (G-135..G-140) — spec `09`/`31`/`17`; slices 19–20 feitos
- Intent Surface v0 (G-141..G-146) — spec `32`; TUI slice 21 feito; Wails = spec `35`
- Evolution v0 (G-147..G-152) — spec `33`; slices 22–24 feitas
- HTTP API v0 (G-153..G-158; recorte G-45) — spec `34`; slices 25–26 feitas
- Desktop Wails v0 (G-159..G-165; recorte G-35/G-144) — spec `35`; slices 27–28 feitas
- NPM Player v0 (G-166..G-171; recorte G-41) — spec `36`; codigo = slice 29 — feito
- Pytest + Yarn Players v0 (G-172..G-179; recorte G-41) — spec `37`; codigo = slice 30 — feito
- Memory Player v0 (G-180..G-186; recorte G-47) — spec `38`; codigo = slice 31 — feito
- MCP Memory Server v0 (G-187..G-193; recorte G-44) — spec `39`; codigo = slice 32 — feito
- Workflow Templates v0 (G-194..G-200; recorte G-40) — spec `40`; codigo = slice 33 — feito
- Infra Players v0 (G-201..G-209; recorte G-41) — spec `41`; codigo = slice 34 — feito

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

## Slice 4 — fidelidade do Validator / sandbox v0

Fecha o que o README listava como “próximo” apos slices 1–3: validacao
completa do protocolo v0 e sandbox Shell honrado, sem isolamento de OS.

| Decisao | Status | Notas |
|---|---|---|
| JSON Schema draft 2020-12 via `santhosh-tekuri/jsonschema/v6` | CONFIRMED | Schema canonico em `schemas/`; embed no Core |
| `schema_version` estrito `"0.1.0"` | CONFIRMED | Ausente/outro → `validation.schema`; sem fill silencioso |
| `task_id` UUID v7 | CONFIRMED | Omitido → Core gera v7; presente e nao-v7 → rejeita |
| `created_at` | CONFIRMED | Omitido → UTC; presente → RFC3339 |
| `input` vs `input_schema` no Validator | CONFIRMED | Antes de `Execute`; falha → `task.rejected` |
| Shell env omitido = heranca minima | CONFIRMED | PATH/HOME/USER/LANG/LC_*/TZ/TMP*; nunca tokens/`RUNTGINE_*` |
| Shell workdir resolve symlink | CONFIRMED | `EvalSymlinks`; path resolvido deve ficar no workspace |
| Allowlist de binarios | CONFIRMED | Default permissivo + `slog.Warn`; sem config obrigatoria |
| Isolamento OS / deny de rede | REJECTED (neste slice) | Continua fora do sandbox v0 (`11` §14) |

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
| Tabs Intent / Runs / Live / Board / Events / Graph / Config | CONFIRMED | Sete abas; INTENT = Entry Point visual (G-142) |
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

## Intent Engine (pos-MVP Core) — CONFIRMADO v0

Ver [17-intent-engine-v0.md](17-intent-engine-v0.md).

| Item | Status | Notas |
|---|---|---|
| G-50 Papel / fronteira | CONFIRMED | Compilador NL→Task IR; nao e Player nem autoridade |
| G-51 API CompileIntent / SubmitIntent | CONFIRMED | Core; CLI `runtgine intent` |
| G-52 Heuristicas shell \| pipeline | CONFIRMED | Deterministic-first |
| G-53 Caminho LLM + heuristic offline | CONFIRMED | Reusa Completer; intermediario JSON |
| G-54 CLI `--dry-run` / `--wait` | CONFIRMED | Mesmo SubmitTask/Validator |

## Project Memory — CONFIRMED v0

Ver [29-project-memory-v0.md](29-project-memory-v0.md). Recorte de
G-46/G-47 (Provider local + ContextPack). Esboco em `16`.
Codigo = slice 17 feito. Nao e MCP (G-44). Nao e Player.

| Item | Status | Notas |
|---|---|---|
| G-123 Papel / pacote `memory` | CONFIRMED | Core Provider; `internal/core/memory` |
| G-124 Episodio + validade | CONFIRMED | kinds 4; `active`/`superseded`/`archived` |
| G-125 API + CLI | CONFIRMED | Record/List/Query/Supersede/Archive; lexical |
| G-126 ContextPack `memory_hits` | CONFIRMED | budget 8 / 2000; abaixo de graph_hits |
| G-127 Captura | CONFIRMED | default `off`; opt-in `failures` |
| G-128 Exclusoes v0 | CONFIRMED | Player, MCP, RAG, TUI, Knowledge |

Rejeicoes (inalteradas vs `16`):

| Decisao | Status | Notas |
|---|---|---|
| Embutir ai-memory (Rust) no Core | REJECTED | Dominio e stack ortogonais |
| Memoria como autoridade de execucao | REJECTED | Sugere contexto; lista negativa |
| Supersession silenciosa via LLM no Core | REJECTED | Validade so com opt-in explicito |
| RAG generico / indexar transcripts | REJECTED | Compile observations |
| Memory Player (`memory.*`) | CONFIRMED (v0) | Recorte G-180..G-186 em `38`; fecha G-47 |

## Runtime Graph — CONFIRMED v0

Ver [18-runtime-graph-v0.md](18-runtime-graph-v0.md) (estrutural) e
[19-graph-hits-v0.md](19-graph-hits-v0.md) (hits).

| Item | Status | Notas |
|---|---|---|
| G-60 Papel / fronteiras | CONFIRMED | Um graph por workspace; ≠ Event Store / Project Memory |
| G-61 Node kinds v0 | CONFIRMED | player, capability, task, run, path, symbol |
| G-62 Edge kinds v0 | CONFIRMED | provides, executed, instance_of, mentions, child_of |
| G-63 Persistencia | CONFIRMED | Mesmo SQLite do Core |
| G-64 Core API + CLI snapshot | CONFIRMED | Sem tab TUI no v0 estrutural |
| G-65 Sync boot / SyncFromRun | CONFIRMED | Best-effort; nao falha Run |
| G-66 Papel Graph Hits | CONFIRMED | Promovido; ver `19-graph-hits-v0.md` |
| G-67 Schema graph_hits + budget | CONFIRMED | Extensao ContextPack; hierarquia vs repo_hits |
| G-68 API QueryHits | CONFIRMED | Ranking deterministico; degrada vazio |
| G-69 Intent LLM + AssembleContext | CONFIRMED | Heuristicas shell\|pipeline nao consultam Graph |

## Git Player — CONFIRMED v0

Ver [20-git-player-v0.md](20-git-player-v0.md). Recorte de G-41.

| Item | Status | Notas |
|---|---|---|
| G-70 Papel / pacote `git` | CONFIRMED | `internal/players/git`; Player deterministic |
| G-71 Capabilities v0 | CONFIRMED | status, diff, log, add, commit |
| G-72 Sandbox minima | CONFIRMED | Sem rede; hooks off no commit; workdir no workspace |
| G-73 Registry + exemplo | CONFIRMED | `api.Open` + `examples/git-status.json` |
| G-74 Exclusoes v0 | CONFIRMED | push/pull/clone/HITL/fora |

## Filesystem Player — CONFIRMED v0

Ver [21-filesystem-player-v0.md](21-filesystem-player-v0.md). Recorte de G-41.

| Item | Status | Notas |
|---|---|---|
| G-75 Papel / pacote `filesystem` | CONFIRMED | `internal/players/filesystem`; Player deterministic |
| G-76 Capabilities v0 | CONFIRMED | `fs.read`, `fs.write`, `fs.list`, `fs.stat` |
| G-77 Confinamento | CONFIRMED | Workspace root; symlink externo rejeitado |
| G-78 Limites / atomicidade | CONFIRMED | UTF-8; 4 MiB; list limit; write atomic |
| G-79 Registry + static validation | CONFIRMED | `api.Open` + Runner + `examples/fs-read.json` |
| G-80 Exclusoes v0 | CONFIRMED | delete/move/chmod/rede/HITL fora |

## Execution Policy + HITL — CONFIRMED v0

Ver [22-execution-policy-v0.md](22-execution-policy-v0.md). Recorte de G-42.

| Item | Status | Notas |
|---|---|---|
| G-81 Papel (Core, nao Player) | CONFIRMED | `internal/core/policy`; HITL = ApproveRun |
| G-82 Verbos + precedencia | CONFIRMED | allow/deny/approval-required; exact capability |
| G-83 Lifecycle HITL | CONFIRMED | `waiting_approval`; eventos; persistencia SQLite |
| G-84 API + CLI | CONFIRMED | `ApproveRun`; `runtgine approve` / `deny` |
| G-85 TUI RUNS/LIVE | CONFIRMED | Amber; teclas `a`/`d`; sem aba nova |
| G-86 Exclusoes v0 | CONFIRMED | Claims, wildcards, Human Player, Docker |

## Docker Player — CONFIRMED v0

Ver [23-docker-player-v0.md](23-docker-player-v0.md). Recorte de G-41.
Codigo = slice 11 apos slice 10 (022).

| Item | Status | Notas |
|---|---|---|
| G-87 Papel / pacote `docker` | CONFIRMED | `internal/players/docker`; binario `docker` |
| G-88 Capabilities v0 | CONFIRMED | ps, inspect, logs, run, build |
| G-89 Sandbox argv | CONFIRMED | `--pull=never --network=none --rm`; sem privileged |
| G-90 Policy Manifest | CONFIRMED | run/build = approval-required (usa 22) |
| G-91 Registry + exemplo | CONFIRMED | `api.Open` + `examples/docker-ps.json` |
| G-92 Exclusoes v0 | CONFIRMED | push/compose/K8s/privileged |

## Resource Claims — CONFIRMED v0

Ver [24-resource-claims-v0.md](24-resource-claims-v0.md). Recorte de G-43
(Claims só; Blast Radius em `25`). Slice 12 feito.

| Item | Status | Notas |
|---|---|---|
| G-93 Papel / pacote `claim` | CONFIRMED | Core, nao Player; Validator → Policy → Claim → Execute |
| G-94 Kinds v0 | CONFIRMED | `workspace` e `path`; overlap segmentado; exclusivo |
| G-95 Tabela automatica | CONFIRMED | fs.write, git.add/commit, docker.build, docker.run+mount; `shell.exec` fora |
| G-96 Lifecycle | CONFIRMED | Acquire no step; hold ate terminal; SQLite; eventos |
| G-97 Conflito fail-fast | CONFIRMED | `claim.conflict`; sem wait/`waiting_claim`; TUI sem aba nova |
| G-98 Exclusoes v0 | CONFIRMED | Blast, wait, Manifest `claims[]`, GRAPH, locks distribuidos |

## Blast Radius — CONFIRMED v0

Ver [25-blast-radius-v0.md](25-blast-radius-v0.md). Resto de G-43
(análise de impacto; Claims já em `24`). Slice 13 feito.

| Item | Status | Notas |
|---|---|---|
| G-99 Papel / pacote `blast` | CONFIRMED | Core, nao Player; analise ≠ lock; nao entra no Runner |
| G-100 Impact Report | CONFIRMED | touches, predicted_claims, risk, conflicts, images |
| G-101 Tabelas | CONFIRMED | predicted = G-95; touches incluem fs.read/list/stat |
| G-102 Overlay | CONFIRMED | read-only vs claims ativos; nunca Acquire |
| G-103 Superficie | CONFIRMED | `BlastTask` + `runtgine blast`; sem TUI GRAPH |
| G-104 Exclusoes v0 | CONFIRMED | Gate Execute, shell argv, persistencia; walk 1-hop promovido em `27` |

## TUI GRAPH — CONFIRMED v0

Ver [26-tui-graph-v0.md](26-tui-graph-v0.md). Aba da TUI sobre o Graph
de `18`. Altera [14-tui-design.md](14-tui-design.md) e a skill.
Slice 14 feito.

| Item | Status | Notas |
|---|---|---|
| G-105 Papel | CONFIRMED | Aba TUI; SoT = `GetGraphSnapshot`; nao e Core novo |
| G-106 Tabs + teclas | CONFIRMED | Seis abas; GRAPH entre EVENTS e CONFIG; sem tecla `g` |
| G-107 Conteudo | CONFIRMED | Counts + lista + detalhe; sem canvas 2D |
| G-108 Refresh | CONFIRMED | `r` → `RefreshGraph` + snapshot |
| G-109 Filtro | CONFIRMED | `/` substring kind/id; independente de EVENTS |
| G-110 Exclusoes v0 | CONFIRMED | Blast-from-graph, Hits UI, PTY, editar grafo |

## Walk Blast←Graph — CONFIRMED v0

Ver [27-blast-graph-walk-v0.md](27-blast-graph-walk-v0.md). 1 hop
`mentions` inbound a partir de `touches` path do Blast (`25`).
Codigo = slice 15 — feito. Nao e QueryHits; nao dispara a partir da aba GRAPH.

| Item | Status | Notas |
|---|---|---|
| G-111 Papel | CONFIRMED | `blast.Walk`; 1 hop; nao e Player |
| G-112 Snapshot | CONFIRMED | `GetGraphSnapshot`; erro → `affected=[]` |
| G-113 Sementes / hop | CONFIRMED | path touches; inbound `mentions` only |
| G-114 Campo `affected` | CONFIRMED | Aditivo; `risk` intacto; schema `0.1.0` |
| G-115 Superficie | CONFIRMED | Mesmo `BlastTask` / CLI; sem TUI |
| G-116 Exclusoes v0 | CONFIRMED | GRAPH→blast, multi-hop, gate, Hits |

## HTTP Player — CONFIRMED v0

Ver [28-http-player-v0.md](28-http-player-v0.md). Recorte de G-41
(cliente HTTPS de leitura). Codigo = slice 16 — feito.
Nao e a API HTTP do Runtgine (G-45 recorte v0 em `34`).

| Item | Status | Notas |
|---|---|---|
| G-117 Papel / pacote `http` | CONFIRMED | `internal/players/httpclient`; Player deterministic |
| G-118 Capabilities v0 | CONFIRMED | `http.get`, `http.head`; allowlist de headers |
| G-119 URL / destino | CONFIRMED | `https` so; deny link-local e metadata |
| G-120 Cliente / sandbox | CONFIRMED | TLS verify on; RoundTripper injetavel; sem curl |
| G-121 Registry + Graph | CONFIRMED | `api.Open`; sem claim/blast; `examples/http-get.json` |
| G-122 Exclusoes v0 | CONFIRMED | POST, auth, `http://`, G-45, MCP, Memory |

## Test Player — CONFIRMED v0

Ver [30-test-player-v0.md](30-test-player-v0.md). Recorte de G-41
(`go test` no workspace). Codigo = slice 18 — feito.
Nao e pytest; NPM = spec `36`. Nao e G-45.

| Item | Status | Notas |
|---|---|---|
| G-129 Papel / pacote `test` | CONFIRMED | `internal/players/gotest`; Player deterministic |
| G-130 Capabilities v0 | CONFIRMED | `test.go` so |
| G-131 Sandbox / argv | CONFIRMED | allowlist; `-mod=readonly`; `-json` |
| G-132 Falha vs sucesso | CONFIRMED | exit != 0 → `runtime.player_error` |
| G-133 Registry + Graph | CONFIRMED | `api.Open`; sem claim/blast; `examples/test-go.json` |
| G-134 Exclusoes v0 | CONFIRMED | outros runners, `-race`, fuzz, G-45, MCP |

## MVP 1.0 magro — CONFIRMED v0

Ver [09-mvp.md](09-mvp.md). Heuristicas Intent em
[17-intent-engine-v0.md](17-intent-engine-v0.md). Context Engine em
[31-context-engine-v0.md](31-context-engine-v0.md).
Codigo = slices 19–20 feitos. Nao e G-45; nao e NATS/Wails/MCP.

| Item | Status | Notas |
|---|---|---|
| G-135 Heuristicas Player no Intent | CONFIRMED | slice 19 feito; `test.go` / git status|diff|log / `docker.ps` |
| G-136 Metodos + soberania | CONFIRMED | `heuristic.test|git|docker|npm`; LLM route inalterado |
| G-137 Papel Context Engine | CONFIRMED | slice 20 feito; assembler; nao e Player |
| G-138 Semente `repo_hits` | CONFIRMED | slice 20 feito; QueryHits path/symbol se vazio; nao pisa repo-search |
| G-139 Ranking / pack | CONFIRMED | sem walk, embeddings, file body |
| G-140 Exclusoes 1.0 | CONFIRMED | G-45, NATS, Wails, MCP, Router, templates |

## Intent Surface — CONFIRMED v0

Ver [32-intent-surface-v0.md](32-intent-surface-v0.md). Superficie visual de
Entry Point (Mission Brief / aba INTENT). Compilador = Intent Engine (`17`).
Codigo TUI = slice 21 feito; Wails INTENT = spec `35` (slices 27–28 feitas).

| Item | Status | Notas |
|---|---|---|
| G-141 Papel / Mission Brief | CONFIRMED | Entry Point visual; nao chatbot |
| G-142 TUI aba INTENT | CONFIRMED | Primeira aba; NL + preview + submit |
| G-143 Fluxo Core | CONFIRMED | `CompileIntent` / `SubmitIntent`; source `tui`\|`wails` |
| G-144 Wails INTENT | CONFIRMED | Semântica em `32`; app desktop = spec `35` |
| G-145 Exclusoes v0 | CONFIRMED | Sem thread chat; sem Player; sem transcript RAG |
| G-146 Criterios de pronto | CONFIRMED | Preview Ctrl+p; submit → LIVE |

## Evolution v0 — CONFIRMED (Router, Playbooks, Lessons)

Ver [33-evolution-v0.md](33-evolution-v0.md). P3: roteamento LLM multi-provider,
playbooks de projeto, loop Lessons pós-falha com HITL. **Não** é framework de
agentes. Código = slices 22–24 (feito).

| Item | Status | Notas |
|---|---|---|
| G-147 Player Router v0 | CONFIRMED | Effort/difficulty/capability → provider/model |
| G-148 Multi-provider config | CONFIRMED | `llm_providers` + `llm_routing`; benchmarks = input humano |
| G-149 Playbooks v0 | CONFIRMED | `.runtgine/playbooks/`; `playbook_hits` |
| G-150 Lessons / Postmortem v0 | CONFIRMED | Proposta em `run.failed`; HITL antes de Memory/playbook |
| G-151 Exclusoes v0 | CONFIRMED | Sem Agent registry; sem promoção silenciosa; sem chat RAG |
| G-152 Ordem slices 22–24 | CONFIRMED | Router → Playbooks → Lessons |

## HTTP API — CONFIRMED v0 (Entry Point)

Ver [34-http-api-v0.md](34-http-api-v0.md). Recorte de G-45: servidor
do runtime (CI/UC-02), **não** o HTTP Player cliente (`28`). Código =
slices 25–26 (feito). Independente das slices 21–24.

| Item | Status | Notas |
|---|---|---|
| G-153 Papel / pacote `httpapi` | CONFIRMED | Entry Point; `runtgine serve`; `source=http` |
| G-154 Listen / auth | CONFIRMED | Loopback `:7420`; Bearer; recusa non-loopback sem token |
| G-155 Rotas REST + SSE | CONFIRMED | Tasks, Intent, Runs, cancel, HITL, blast; `/v0/` |
| G-156 Webhooks outbound | CONFIRMED | Eventos terminais; HTTPS; não falha o Run |
| G-157 Exclusoes v0 | CONFIRMED | Sem inbound GitHub; sem TLS no binário; sem Graph REST |
| G-158 Ordem slices 25–26 | CONFIRMED | Serve → webhooks |

## Desktop Wails — CONFIRMED v0 (Entry Point)

Ver [35-wails-v0.md](35-wails-v0.md). Recorte de G-35 / G-144: janela
nativa in-process sobre `api.Core`, **não** cliente da HTTP API (`34`).
Código = slices 27–28 (feitas). Wails **v3** (beta aceite; v2 fora).

| Item | Status | Notas |
|---|---|---|
| G-159 Papel / pacote `desktop` | CONFIRMED | Entry Point; `runtgine desktop`; `source=wails` |
| G-160 Stack pin | CONFIRMED | Wails v3 + Svelte 5 + shadcn-svelte; tokens `14` |
| G-161 App shell / sete views | CONFIRMED | INTENT primeiro; uma janela |
| G-162 Bindings Core API | CONFIRMED | Service v3; testes fake Core; CI sem display |
| G-163 INTENT desktop | CONFIRMED | Fecha G-144; preview sem Run |
| G-164 Exclusoes v0 | CONFIRMED | Sem v2; sem HTTP client; sem chat; sem PTY; sem multi-window |
| G-165 Ordem slices 27–28 | CONFIRMED | INTENT/LIVE (27) → demais views (28); ambas feitas |

## NPM Player — CONFIRMED v0

Ver [36-npm-player-v0.md](36-npm-player-v0.md). Recorte de G-41
(`npm test` no workspace). Código = slice 29 (feito).
Não é `test.go`; não é pytest; não é `npm install`.

| Item | Status | Notas |
|---|---|---|
| G-166 Papel / pacote `npm` | CONFIRMED | `internal/players/npm`; Player deterministic |
| G-167 Capabilities v0 | CONFIRMED | `npm.test` só |
| G-168 Sandbox / argv | CONFIRMED | `npm test`; sem install/npx/prefix |
| G-169 Falha vs sucesso | CONFIRMED | exit != 0 → `runtime.player_error` |
| G-170 Registry + Graph + Intent | CONFIRMED | `api.Open`; `heuristic.npm`; sem claim/blast |
| G-171 Exclusoes v0 | CONFIRMED | install, yarn/pnpm, pytest, G-44, G-45 |

## Memory Player — CONFIRMED v0

Ver [38-memory-player-v0.md](38-memory-player-v0.md). Recorte de G-47
sobre o Provider já CONFIRMED em `29` (Project Memory v0; slice 17).
Fecha a OPEN QUESTION "Memory Player" que `04` carregava desde a
sessão de `29`. **Não** é MCP (G-44), **não** é Knowledge, **não**
RAG, **não** indexação de transcript. Read-only sobre o store.

| Item | Status | Notas |
|---|---|---|
| G-180 Papel / pacote `memory` | CONFIRMED | `internal/players/memory`; Player deterministic; read-only |
| G-181 Capabilities v0 | CONFIRMED | `memory.recall`, `memory.check` |
| G-182 Provider `Reader` interface | CONFIRMED | `Recall`, `Check`; exposta a partir de `internal/core/memory`; Provider já cobre |
| G-183 Sandbox | CONFIRMED | In-process; sem rede; sem MCP; sem shell |
| G-184 Falha do Provider degrada | CONFIRMED | erro → step succeeded com vazio + `slog.Warn`; nunca `runtime.player_error` |
| G-185 Registry + Graph | CONFIRMED | `api.Open`; `RefreshFromRegistry`; edge `provides` para `internal/core/memory` |
| G-186 Exclusoes v0 | CONFIRMED | escrita (record/supersede/archive), MCP, embeddings, RAG, intents `lembre` |

## Pytest + Yarn Players — CONFIRMED v0 spec

Ver [37-pytest-yarn-players-v0.md](37-pytest-yarn-players-v0.md).
Recorte de G-41. Mesmo padrão de `test.go` (`30`) e `npm.test`
(`36`). Não cobre pytest parametrizado, tox, `pytest-xdist`,
`yarn install`/`add`/`npx`, MCP, Knowledge.

| Item | Status | Notas |
|---|---|---|
| G-172 Papel / pacote `pytest` | CONFIRMED | `internal/players/pytst`; Player deterministic |
| G-173 Capability `pytest.run` | CONFIRMED | `pytest …`; argv allowlist; sem `-n`/`--cov*` |
| G-174 Workdir + marker | CONFIRMED | `pyproject.toml` \| `pytest.ini` \| `tests/` |
| G-175 Papel / pacote `yarn` | CONFIRMED | `internal/players/jstest`; Player deterministic |
| G-176 Capability `yarn.test` | CONFIRMED | `yarn test`; `--frozen-lockfile`/`--immutable`/`--parallel`/`add`/`install`/`dlx`/`npx` negados |
| G-177 Workdir + `package.json` | CONFIRMED | mesmo padrão do `npm` |
| G-178 Falha vs sucesso | CONFIRMED | exit != 0 → `runtime.player_error` (pytest e yarn) |
| G-179 Registry + Graph + Intent | CONFIRMED | `api.Open`; `heuristic.pytest`, `heuristic.yarn`; sem claim/blast; slices 30 |

## MCP Memory Server — CONFIRMED v0 spec

Ver [39-mcp-memory-v0.md](39-mcp-memory-v0.md). Recorte de G-44
("candidato a transporte futuro da Project Memory" desde `29`).
Servidor **MCP read-only** sobre o Provider (`internal/core/memory`):
tools `memory.query` / `memory.list`; transportes stdio (`runtgine
mcp`) e HTTP (`/mcp` no serve, mesma auth). Não é cliente MCP; não
é Player; não é alternativa ao MCP (`01`). Código = slice 32
(feito).

| Item | Status | Notas |
|---|---|---|
| G-187 Papel / pacote | CONFIRMED | `internal/entrypoint/mcpserver`; servidor read-only; não Player |
| G-188 Tools v0 | CONFIRMED | `memory.query`, `memory.list`; só `active`; sem escrita |
| G-189 Transporte stdio | CONFIRMED | `runtgine mcp`; JSON-RPC 2.0 stdin/stdout |
| G-190 Transporte HTTP | CONFIRMED | `/mcp` no serve; mesma auth (bearer) + loopback (`CheckBind`) |
| G-191 Segurança / degradação | CONFIRMED | Falha do Provider → vazio + warning; server vivo; token obrigatório no HTTP |
| G-192 Exclusoes v0 | CONFIRMED | Escrita via MCP, cliente MCP, embeddings/RAG/Knowledge, cross-workspace, subscriptions/resources |
| G-193 Interop + aceite | CONFIRMED | Handshake/tools/list/tools-call bem-formados nos dois transportes |

## Workflow Templates — CONFIRMED v0 spec

Ver [40-workflow-templates-v0.md](40-workflow-templates-v0.md). Recorte de
G-40 (nativo vs repo externo). JSON em `.runtgine/templates/`
compila para Task IR; Validator soberano. Distinto de Playbooks
(`33`). Código = slice 33 (feito).

| Item | Status | Notas |
|---|---|---|
| G-194 Papel / pacote | CONFIRMED | `internal/core/templates`; Core, não Player |
| G-195 Schema JSON v0 | CONFIRMED | `id`, `title`, `steps` 1–20; `additionalProperties: false` |
| G-196 Loading nativo | CONFIRMED | Fecha G-40: workspace only; best-effort skip+warn |
| G-197 Compile → Task IR | CONFIRMED | `metadata.template`; admissão valida capabilities |
| G-198 CLI + Intent | CONFIRMED | `runtgine template`; `heuristic.template` antes de shell |
| G-199 Graph | CONFIRMED | kind aditivo `template`; sem aresta nova |
| G-200 Exclusoes v0 | CONFIRMED | remoto, auto-sizing, verifier, Player, HTTP/MCP |

## Infra Players — CONFIRMED v0 spec

Ver [41-infra-players-v0.md](41-infra-players-v0.md). Recorte de
G-41 (K8s / Terraform / PostgreSQL). Código = slice 34 (feito).
Leitura / validate+plan / ping; sem apply, sem SQL livre.

| Item | Status | Notas |
|---|---|---|
| G-201 Kubernetes Player | CONFIRMED | `k8s.get` / `k8s.list`; `internal/players/k8s` |
| G-202 K8s sandbox | CONFIRMED | argv `kubectl get`; sem apply/exec; `safeRef` |
| G-203 Terraform Player | CONFIRMED | `tf.validate` / `tf.plan`; `internal/players/tf` |
| G-204 Terraform sandbox | CONFIRMED | `*.tf` no workdir; `tf.plan` approval-required; sem apply/init |
| G-205 Postgres Player | CONFIRMED | `pg.ping`; `internal/players/pg` |
| G-206 Postgres sandbox | CONFIRMED | sem SQL/password no IR; `PGPASSWORD` só env |
| G-207 Registry + Intent | CONFIRMED | `heuristic.k8s` / `heuristic.tf` / `heuristic.pg` |
| G-208 Falha vs sucesso | CONFIRMED | exit ≠ 0 → `runtime.player_error`; testes fake |
| G-209 Exclusoes v0 | CONFIRMED | apply/destroy/init/exec/SQL; Helm; cloud SDKs |

## Helm Player — CONFIRMED v0 spec

Ver [42-helm-player-v0.md](42-helm-player-v0.md). Recorte de G-41
(Helm). Levanta a exclusão de Helm do corte v0 de `41` (G-209) como
recorte próprio. Lint / render local / leitura de cluster; sem
install/upgrade. Código = slice 35 (pendente).

| Item | Status | Notas |
|---|---|---|
| G-210 Helm Player | CONFIRMED | `internal/players/helm`; deterministic |
| G-211 Capabilities v0 | CONFIRMED | `helm.lint` / `helm.template` / `helm.list` / `helm.status`; todas allow |
| G-212 Sandbox | CONFIRMED | argv fechado; sem install/upgrade/rollback/uninstall/get/test; sem `--set*`/values |
| G-213 Chart no workspace | CONFIRMED | `chart` relativo com marker `Chart.yaml`; cluster read via kubeconfig herdado |
| G-214 Falha vs sucesso | CONFIRMED | exit ≠ 0 → `runtime.player_error`; timeout → `runtime.timeout` |
| G-215 Registry + Intent | CONFIRMED | `heuristic.helm` antes de shell; `helm install` não casa |
| G-216 Exclusoes v0 | CONFIRMED | repo/OCI, plugin, kustomize/k3s/kind, cloud SDKs, SQL |

## Git / release — fluxo de branches

Ver [15-git-workflow.md](15-git-workflow.md).

| Decisao | Status | Notas |
|---|---|---|
| Fluxo `feat → develop → release → main` | CONFIRMED | Integracao em `develop`; estabilizacao em `release/*`; estavel em `main` |
| Prefixo de feature `feat/<NNN>-<slug>` | CONFIRMED | `NNN` = id da spec/issue (ex.: `feat/001-shell-player`) |
| Outros prefixos `fix/` `docs/` `chore/` | CONFIRMED | Mesmo padrao numerico quando houver issue/spec |
| OpenSpec em `openspec/` | CONFIRMED | Changes = `openspec/changes/<NNN>-<slug>/`; ver `15` + `openspec/README.md` |
| Release candidates | CONFIRMED | Branch `release/x.y.z` + tags `vX.Y.Z-rc.N` |
| Release estavel | CONFIRMED | Merge `release/*` → `main` + tag `vX.Y.Z`; back-merge para `develop` |
| Semver do produto | CONFIRMED | Tags Git do binario/CLI; distinto de `schema_version` do Task IR |
| CI em PR (test + vet) | CONFIRMED | GitHub Actions; ver `.github/workflows/` |
| Branch protection | CONFIRMED | IMPLEMENTED: ruleset `runtgine-protected-branches` em `main`/`develop`/`release/*` (repo publico; Free nao protege privado) |
