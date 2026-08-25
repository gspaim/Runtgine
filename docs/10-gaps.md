# 10 — Gaps e definicoes faltantes

Inventario do que a documentacao ainda nao fecha para iniciar o Core.
Complementa `04-decisoes.md` e `09-mvp.md`.

Status deste doc: inventario oficial. Itens P0 do protocolo foram
confirmados em [11-protocolo-v0.md](11-protocolo-v0.md) / `04-decisoes.md`.
P1+ ainda abertos.

---

## Como ler

| Severidade | Significado |
|---|---|
| P0 | Bloqueia codigo util do MVP Core |
| P1 | Bloqueia cenario Board / LLM vertical |
| P2 | Importante pos-Core; nao bloqueia `runtgine run` minimo |
| P3 | Futuro / ecossistema |

Em conflito de escopo, prevalece `09-mvp.md` + `04-decisoes.md`.

---

## P0 — Bloqueantes do MVP Core

| ID | Gap | Situacao atual | Proposta |
|---|---|---|---|
| G-01 | Task IR v0 schema | Conceito HYPOTHESIS; MVP exige | `11` § Task IR |
| G-02 | Player Manifest v0 | So exemplo verbal | `11` § Manifest |
| G-03 | Event envelope | Event CONFIRMED sem campos | `11` § Event |
| G-04 | Taxonomia minima de eventos | Nenhum catalogo | `11` § Event types |
| G-05 | Capability naming | So exemplos ad hoc | `11` § Capabilities |
| G-06 | Shell Player contract + sandbox | Player no MVP sem I/O nem limites | `11` § Shell Player |
| G-07 | Protocolo Entry Point → Core | Aberto em `00-rascunho` | `11` § Core API |
| G-08 | Result / Error model | Criterio de sucesso sem schema | `11` § Result |
| G-09 | Estados Task / Run | Sem maquina de estados | `11` § Run lifecycle |
| G-10 | Orchestrator no MVP | Em todo fluxo; status HYPOTHESIS | `11` § Runner v0 |
| G-11 | Execution Plan no MVP | CONFIRMED sem schema; sem Intent Engine | `11` § Plan v0 (passthrough) |
| G-12 | Queue semantica | CONFIRMED sem FIFO/prioridade/persistencia | `11` § Queue v0 |
| G-13 | Persistencia no MVP | SQLite CONFIRMED na stack; ausente do incluso MVP; P1 no PRD | `11` § Persistence |
| G-14 | JSON vs YAML | Stack = JSON Schema; MVP cita YAML | `11` § Encoding |
| G-15 | Driver SQLite | mattn vs modernc | `11` § Stack openers |
| G-16 | Logger | slog HYPOTHESIS | `11` § Stack openers |
| G-17 | Layout de pacotes Go | So em fonte historica | `11` § Repo layout |
| G-18 | Policy minima do Shell | Execution Policy fora do MVP; Shell inseguro sem corte | `11` § Shell Player |

### Tensoes meta (P0)

- Task IR / Validator / Orchestrator sao **HYPOTHESIS** mas **obrigatorios no MVP** → promover corte v0 a CONFIRMED apos revisao de `11`.
- **Capability Resolver** e **Planner** aparecem em `01-visao` sem entrada em `02-conceitos`.
- **Event Store** na arquitetura vs MVP “sem event sourcing” → esclarecer store minimo vs sourcing.

---

## P1 — Cenario Board / pipeline vertical

| ID | Gap | Notas |
|---|---|---|
| G-20 | Card GitHub Projects → Task IR | **CONFIRMED** — ver `12-board-p1.md` |
| G-21 | Write-back no board | **CONFIRMED** — status + comentario; sem criar subtasks |
| G-22 | Contratos por etapa do pipeline | **CONFIRMED** — `pipeline.*` + steps lineares; ver `12` |
| G-23 | Fronteira regras vs LLM | **CONFIRMED** — ver tabela em `12-board-p1.md` |
| G-24 | Context assembly basico | **CONFIRMED** — ContextPack v0; ver `12` |
| G-25 | LLM Player v0 | **CONFIRMED** — OpenAI-compat + Anthropic; ver `12` |
| G-26 | Task Router basico | **CONFIRMED** — regras deterministic-first; ver `12` |
| G-27 | Modelo de subtasks | **CONFIRMED** — SQLite + child runs; ver `12` |

Propostas detalhadas do Board ficam para um doc futuro apos confirmar `11`.
Ate la, o Core deve rodar so com CLI + Shell.

---

## P1b — Intent Engine (pos-Core)

| ID | Gap | Notas |
|---|---|---|
| G-50 | Papel Intent Engine | **CONFIRMED** — ver `17-intent-engine-v0.md` |
| G-51 | API CompileIntent / SubmitIntent | **CONFIRMED** — ver `17` |
| G-52 | Heuristicas deterministicas | **CONFIRMED** — shell \| pipeline |
| G-53 | Caminho LLM | **CONFIRMED** — Completer + offline heuristic |
| G-54 | CLI `runtgine intent` | **CONFIRMED** — dry-run + submit |

---

## P2 — Engenharia e produto (pos-Core minimo)

| ID | Gap |
|---|---|
| G-30 | Cancelamento, timeout, retry, concorrencia entre runs | **CONFIRMED** — retry configuravel por step; ver `13` |
| G-31 | Observabilidade alem da TUI (niveis de log, correlacao) | **CONFIRMED** — slog + SQLite; ver `13` |
| G-32 | Fronteira Runtgine ↔ Chorus | **CONFIRMED** — MVP independente; ver `13` |
| G-33 | Workspaces / worktrees | **CONFIRMED** — um root + `.runtgine/`; ver `13` |
| G-34 | Estrategia de testes (unit vs integracao) | **CONFIRMED** — ver `13` |
| G-35 | Wails: Svelte vs React | **CONFIRMED** — Wails + Svelte; ver `13` |
| G-36 | NATS / Event Bus distribuido (OPEN QUESTION) | **DEFERRED** — Bus plugavel; sem NATS no MVP; ver `13` |
| G-37 | Modulo path Go + versao minima de Go |
| G-38 | Config do runtime (arquivo, env, defaults) | **CONFIRMED** — defaults < file < env < flags; ver `13` |

---

## P3 — Futuro / ecossistema

| ID | Gap |
|---|---|
| G-40 | Workflow Templates loading (nativo vs repo externo) — ver `08` |
| G-41 | Biblioteca ampla de Players | Em andamento — Git (`20`), Filesystem (`21`), Docker (`23`), HTTP (`28`), Test (`30`), **NPM (`36`, G-166..G-171; slice 29 feito)**; resto: pytest/yarn/infra |
| G-42 | Human-in-the-loop / Approvals | **CONFIRMED v0** — recorte G-81..G-86 em `22` |
| G-43 | Resource Claims / Blast Radius | **Claims CONFIRMED v0** — `24`. **Blast CONFIRMED v0** — `25`. **Walk Blast←Graph CONFIRMED v0** — recorte G-111..G-116 em `27` |
| G-44 | MCP integration — transporte da Project Memory | **CONFIRMED v0 spec** — recorte G-187..G-193 em `39` (servidor MCP read-only; slice 32 a fazer) |
| G-45 | API HTTP / webhooks | **CONFIRMED v0** — recorte G-153..G-158 em `34`; distinto do HTTP Player (`28`) |
| G-46 | Project Memory (conceito + ContextPack + validade + hierarquia) | **CONFIRMED v0** — recorte G-123..G-128 em `29`; esboço em `16` |
| G-47 | Modelo de acesso Memory Provider vs Memory Player — Provider **CONFIRMED v0** (`29`); Player **OPEN QUESTION** (fora do v0) |

---

## Runtime Graph (pos-Intent) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-60 | Papel Runtime Graph / fronteiras | **CONFIRMED** — ver `18-runtime-graph-v0.md` |
| G-61 | Node kinds v0 | **CONFIRMED** — player, capability, task, run, path, symbol |
| G-62 | Edge kinds v0 | **CONFIRMED** — provides, executed, instance_of, mentions, child_of |
| G-63 | Persistência SQLite | **CONFIRMED** — mesmo `.runtgine/runtgine.db` |
| G-64 | Core API + CLI snapshot | **CONFIRMED** — sem tab TUI no v0 |
| G-65 | Sync boot / SyncFromRun | **CONFIRMED** — best-effort; nao falha Run |

## Graph Hits (pos-Graph estrutural) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-66 | Papel Graph Hits | **CONFIRMED** — ver `19-graph-hits-v0.md` |
| G-67 | Schema `graph_hits` + budget | **CONFIRMED** — extensao ContextPack |
| G-68 | API `QueryHits` | **CONFIRMED** — ranking deterministico |
| G-69 | Intent LLM + AssembleContext | **CONFIRMED** — heuristicas sem Graph |

## Git Player (recorte G-41) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-70 | Papel / pacote `git` | **CONFIRMED** — ver `20-git-player-v0.md` |
| G-71 | Capabilities status/diff/log/add/commit | **CONFIRMED** |
| G-72 | Sandbox mínima | **CONFIRMED** — sem rede; hooks off no commit |
| G-73 | Registry + exemplo | **CONFIRMED** |
| G-74 | Exclusões v0 | **CONFIRMED** — push/HITL/fora |

## Filesystem Player (recorte G-41) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-75 | Papel / pacote `filesystem` | **CONFIRMED** — ver `21-filesystem-player-v0.md` |
| G-76 | Capabilities read/write/list/stat | **CONFIRMED** |
| G-77 | Confinamento e symlink policy | **CONFIRMED** — workspace only |
| G-78 | Limites / UTF-8 / escrita atômica | **CONFIRMED** |
| G-79 | Registry + static validation + exemplo | **CONFIRMED** |
| G-80 | Exclusões v0 | **CONFIRMED** — delete/move/chmod/rede/HITL fora |

## Execution Policy + HITL (recorte G-42) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-81 | Papel / Core vs Player | **CONFIRMED** — ver `22-execution-policy-v0.md` |
| G-82 | Verbos + precedência | **CONFIRMED** — allow/deny/approval-required |
| G-83 | Lifecycle `waiting_approval` | **CONFIRMED** |
| G-84 | `ApproveRun` + CLI | **CONFIRMED** |
| G-85 | TUI RUNS/LIVE | **CONFIRMED** — sem aba nova |
| G-86 | Exclusões v0 | **CONFIRMED** — Claims/wildcards/Human Player/Docker |

## Docker Player (recorte G-41) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-87 | Papel / pacote `docker` | **CONFIRMED** — ver `23-docker-player-v0.md` |
| G-88 | Capabilities ps/inspect/logs/run/build | **CONFIRMED** |
| G-89 | Sandbox argv | **CONFIRMED** — pull never; network none |
| G-90 | Policy Manifest run/build | **CONFIRMED** — depende de `22` |
| G-91 | Registry + exemplo | **CONFIRMED** |
| G-92 | Exclusões v0 | **CONFIRMED** — push/compose/K8s/privileged |

## Resource Claims (recorte G-43) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-93 | Papel / pacote `claim` | **CONFIRMED** — ver `24-resource-claims-v0.md` |
| G-94 | Kinds `workspace` / `path` | **CONFIRMED** — exclusivo; overlap segmentado |
| G-95 | Tabela automática | **CONFIRMED** — mutadores Git/FS/Docker; `shell.exec` fora |
| G-96 | Lifecycle + SQLite | **CONFIRMED** — hold até terminal; órfãos no boot |
| G-97 | Conflito fail-fast | **CONFIRMED** — `claim.conflict`; sem wait |
| G-98 | Exclusões v0 | **CONFIRMED** — Blast/wait/Manifest claims[]/GRAPH |

## Blast Radius (resto de G-43) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-99 | Papel / pacote `blast` | **CONFIRMED** — ver `25-blast-radius-v0.md` |
| G-100 | Impact Report | **CONFIRMED** — touches / predicted_claims / risk / conflicts |
| G-101 | Tabelas de derivação | **CONFIRMED** — predicted = G-95; touches incluem leituras |
| G-102 | Overlay vs claims ativos | **CONFIRMED** — read-only; nunca Acquire |
| G-103 | Superfície CLI + API | **CONFIRMED** — `runtgine blast`; sem TUI GRAPH; sem auto no Runner |
| G-104 | Exclusões v0 | **CONFIRMED** — gate Execute, argv shell, persistência; walk 1-hop em `27` |

## TUI GRAPH (aba) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-105 | Papel da aba GRAPH | **CONFIRMED** — ver `26-tui-graph-v0.md` |
| G-106 | Seis tabs + keymap | **CONFIRMED** — GRAPH entre EVENTS e CONFIG |
| G-107 | Lista / counts / detalhe | **CONFIRMED** — sem canvas 2D |
| G-108 | Refresh | **CONFIRMED** — `r` → RefreshGraph + snapshot |
| G-109 | Filtro | **CONFIRMED** — substring kind/id |
| G-110 | Exclusões v0 | **CONFIRMED** — Blast-from-graph, Hits UI, PTY, edit |

## Walk Blast←Graph — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-111 | Papel do walk | **CONFIRMED** — ver `27-blast-graph-walk-v0.md` |
| G-112 | Snapshot / degradação | **CONFIRMED** — erro → `affected=[]` |
| G-113 | Sementes e hop | **CONFIRMED** — path touches; inbound `mentions` |
| G-114 | Campo `affected` | **CONFIRMED** — `risk` intacto |
| G-115 | Superfície | **CONFIRMED** — `BlastTask` + CLI; sem TUI |
| G-116 | Exclusões v0 | **CONFIRMED** — GRAPH→blast, multi-hop, gate, Hits |

## HTTP Player (recorte G-41) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-117 | Papel / pacote `http` | **CONFIRMED** — ver `28-http-player-v0.md` |
| G-118 | Capabilities GET/HEAD | **CONFIRMED** — allowlist de headers; sem POST |
| G-119 | URL / destino | **CONFIRMED** — `https`; deny link-local e metadata |
| G-120 | Cliente / sandbox | **CONFIRMED** — TLS verify; testes offline |
| G-121 | Registry + Graph | **CONFIRMED** — sem claim/blast |
| G-122 | Exclusões v0 | **CONFIRMED** — POST, auth, G-45, MCP, Memory |

## Project Memory (recorte G-46/G-47) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-123 | Papel / pacote `memory` | **CONFIRMED** — ver `29-project-memory-v0.md` |
| G-124 | Episódio + validade | **CONFIRMED** — kinds 4; validade explícita |
| G-125 | API + CLI | **CONFIRMED** — Query lexical; sem embeddings |
| G-126 | ContextPack `memory_hits` | **CONFIRMED** — abaixo de graph_hits |
| G-127 | Captura | **CONFIRMED** — default off; opt-in failures |
| G-128 | Exclusões v0 | **CONFIRMED** — Player, MCP, RAG, TUI, Knowledge |

## Test Player (recorte G-41) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-129 | Papel / pacote `test` | **CONFIRMED** — ver `30-test-player-v0.md` |
| G-130 | Capabilities v0 | **CONFIRMED** — `test.go` |
| G-131 | Sandbox / argv | **CONFIRMED** — `-mod=readonly`; allowlist |
| G-132 | Falha vs sucesso | **CONFIRMED** — teste vermelho falha o Run |
| G-133 | Registry + Graph | **CONFIRMED** — sem claim/blast |
| G-134 | Exclusões v0 | **CONFIRMED** — pytest/npm no v0 `30`; npm = spec `36`; race, G-45, MCP |

## MVP 1.0 magro — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-135 | Heurísticas Player no Intent | **CONFIRMED** — ver `17` / `09` |
| G-136 | Métodos + soberania | **CONFIRMED** — `heuristic.test\|git\|docker\|npm` |
| G-137 | Papel Context Engine | **CONFIRMED** — ver `31-context-engine-v0.md` |
| G-138 | Semente `repo_hits` | **CONFIRMED** — QueryHits path/symbol se vazio |
| G-139 | Ranking / pack | **CONFIRMED** — sem walk / embeddings / file body |
| G-140 | Exclusões 1.0 | **CONFIRMED** — G-45, NATS, Wails, MCP, Router |

## Intent Surface — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-141 | Papel Mission Brief | **CONFIRMED** — ver `32-intent-surface-v0.md` |
| G-142 | TUI aba INTENT | **CONFIRMED** — primeira aba; NL + preview + submit |
| G-143 | Fluxo Core | **CONFIRMED** — `CompileIntent` / `SubmitIntent` |
| G-144 | Wails INTENT | **CONFIRMED** — semântica em `32`; app = spec `35` |
| G-145 | Exclusões v0 | **CONFIRMED** — não chatbot; sem Player; sem transcript RAG |
| G-146 | Critérios de pronto | **CONFIRMED** — Ctrl+p preview; submit → LIVE |

## Evolution v0 — CONFIRMED (spec)

| ID | Gap | Notas |
|---|---|---|
| G-147 | Player Router v0 | **CONFIRMED** — ver `33-evolution-v0.md` |
| G-148 | Multi-provider / routing | **CONFIRMED** — config + effort/difficulty |
| G-149 | Playbooks v0 | **CONFIRMED** — skills projeto; `playbook_hits` |
| G-150 | Lessons / Postmortem | **CONFIRMED** — HITL; Memory + patch playbook |
| G-151 | Exclusões evolution | **CONFIRMED** — não agent framework |
| G-152 | Slices 22–24 | **CONFIRMED** — Router → Playbooks → Lessons |

## HTTP API (recorte G-45) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-153 | Papel / pacote `httpapi` | **CONFIRMED** — ver `34-http-api-v0.md` |
| G-154 | Listen / auth | **CONFIRMED** — loopback; Bearer; boot guard |
| G-155 | Rotas REST + SSE | **CONFIRMED** — Core API `11` §13 + Intent |
| G-156 | Webhooks outbound | **CONFIRMED** — terminais; HTTPS; best-effort |
| G-157 | Exclusões v0 | **CONFIRMED** — sem inbound GitHub; sem TLS no binário |
| G-158 | Slices 25–26 | **CONFIRMED** — serve → webhooks |

## Desktop Wails (recorte G-35 / G-144) — CONFIRMED v0

| ID | Gap | Notas |
|---|---|---|
| G-159 | Papel / pacote `desktop` | **CONFIRMED** — ver `35-wails-v0.md` |
| G-160 | Stack pin Wails v3 | **CONFIRMED** — Svelte 5 + shadcn-svelte; v2 fora |
| G-161 | App shell / sete views | **CONFIRMED** — INTENT primeiro |
| G-162 | Bindings Core API | **CONFIRMED** — in-process; CI sem display |
| G-163 | INTENT desktop | **CONFIRMED** — fecha G-144 |
| G-164 | Exclusões v0 | **CONFIRMED** — sem HTTP client; sem chat; sem PTY |
| G-165 | Slices 27–28 | **CONFIRMED** — INTENT/LIVE → demais views |

## NPM Player (recorte G-41) — CONFIRMED v0 spec

| ID | Gap | Notas |
|---|---|---|
| G-166 | Papel / pacote `npm` | **CONFIRMED** — ver `36-npm-player-v0.md` |
| G-167 | Capabilities v0 | **CONFIRMED** — `npm.test` |
| G-168 | Sandbox / argv | **CONFIRMED** — `npm test`; sem install/npx |
| G-169 | Falha vs sucesso | **CONFIRMED** — teste vermelho falha o Run |
| G-170 | Registry + Graph + Intent | **CONFIRMED** — `heuristic.npm`; sem claim/blast |
| G-171 | Exclusões v0 | **CONFIRMED** — install, yarn/pnpm, pytest, G-44, G-45 |

## Memory Player (recorte G-47) — CONFIRMED v0 spec

| ID | Gap | Notas |
|---|---|---|
| G-180 | Papel / pacote `memory` | **CONFIRMED** — Player read-only; `internal/players/memory` |
| G-181 | Capabilities v0 | **CONFIRMED** — `memory.recall`, `memory.check` |
| G-182 | Provider `Reader` | **CONFIRMED** — `Recall`, `Check`; exposto em `internal/core/memory` |
| G-183 | Sandbox | **CONFIRMED** — in-process; sem rede; sem MCP; sem shell |
| G-184 | Falha do Provider degrada | **CONFIRMED** — erro → vazio + warning; nunca `runtime.player_error` |
| G-185 | Registry + Graph | **CONFIRMED** — `api.Open`; `RefreshFromRegistry`; edge `provides` |
| G-186 | Exclusões v0 | **CONFIRMED** — escrita, MCP, embeddings, RAG |

## Pytest + Yarn Players (recorte G-41) — CONFIRMED v0 spec

| ID | Gap | Notas |
|---|---|---|
| G-172 | Papel / pacote `pytest` | **CONFIRMED** — `internal/players/pytst`; deterministic |
| G-173 | Capability `pytest.run` | **CONFIRMED** — argv allowlist; sem `-n`/`--cov*` |
| G-174 | Workdir + marker | **CONFIRMED** — `pyproject.toml` \| `pytest.ini` \| `tests/` |
| G-175 | Papel / pacote `yarn` | **CONFIRMED** — `internal/players/jstest`; deterministic |
| G-176 | Capability `yarn.test` | **CONFIRMED** — `yarn test`; flags de install negadas |
| G-177 | Workdir + `package.json` | **CONFIRMED** — padrão do `npm` |
| G-178 | Falha vs sucesso | **CONFIRMED** — exit != 0 → `runtime.player_error` |
| G-179 | Registry + Graph + Intent | **CONFIRMED** — `heuristic.pytest`, `heuristic.yarn`; sem claim/blast |

## MCP Memory Server (recorte G-44) — CONFIRMED v0 spec

| ID | Gap | Notas |
|---|---|---|
| G-187 | Papel / pacote | **CONFIRMED** — servidor MCP read-only; `internal/entrypoint/mcpserver`; não Player |
| G-188 | Tools v0 | **CONFIRMED** — `memory.query`, `memory.list`; só `active`; sem escrita |
| G-189 | Transporte stdio | **CONFIRMED** — `runtgine mcp`; JSON-RPC 2.0 stdin/stdout |
| G-190 | Transporte HTTP | **CONFIRMED** — `/mcp` no serve; mesma auth (bearer) + loopback |
| G-191 | Segurança / degradação | **CONFIRMED** — falha do Provider → vazio + warning; server vivo |
| G-192 | Exclusões v0 | **CONFIRMED** — escrita via MCP, cliente MCP, embeddings, cross-workspace, subscriptions/resources |
| G-193 | Interop + aceite | **CONFIRMED** — handshake/tools/list/tools-call bem-formados nos dois transportes |

---

## Ordem para fechar gaps

1. Revisar e confirmar [11-protocolo-v0.md](11-protocolo-v0.md)
2. Promover itens aceitos em `04-decisoes.md` (HYPOTHESIS → CONFIRMED v0)
3. Implementar Core na ordem de `09-mvp.md` / `AGENTS.md`
4. Board/LLM (P1) — feito (`12`)
5. Intent Engine — spec/impl em `17` — feito
6. Runtime Graph — spec/impl em `18` — G-60..G-65 CONFIRMED — feito
7. Graph Hits — spec em `19` — G-66..G-69 CONFIRMED; codigo = slice 7 — feito
8. Git Player — spec em `20` — G-70..G-74 CONFIRMED; codigo = slice 8 — feito
9. Filesystem Player — spec em `21` — G-75..G-80 CONFIRMED; codigo = slice 9 — feito
10. Execution Policy + HITL — spec em `22` — G-81..G-86 CONFIRMED; codigo = slice 10
11. Docker Player — spec em `23` — G-87..G-92 CONFIRMED; codigo = slice 11 (apos 10)
12. Resource Claims — spec em `24` — G-93..G-98 CONFIRMED; codigo = slice 12 — feito
13. Blast Radius — spec em `25` — G-99..G-104 CONFIRMED; codigo = slice 13 — feito
14. TUI GRAPH — spec em `26` — G-105..G-110 CONFIRMED; codigo = slice 14 — feito
15. Walk Blast←Graph — spec em `27` — G-111..G-116 CONFIRMED; codigo = slice 15 — feito
16. HTTP Player — spec em `28` — G-117..G-122 CONFIRMED; codigo = slice 16 — feito
17. Project Memory — spec em `29` — G-123..G-128 CONFIRMED; codigo = slice 17 — feito
18. Test Player — spec em `30` — G-129..G-134 CONFIRMED; codigo = slice 18 — feito
19. MVP 1.0 magro — spec em `09`/`31`/`17` — G-135..G-140 CONFIRMED; slices 19–20 feitos
20. Intent Surface — spec em `32` — G-141..G-146 CONFIRMED; TUI slice 21 feito; Wails = spec `35`
21. Evolution v0 — spec em `33` — G-147..G-152 CONFIRMED; slices 22–24 feitas
22. HTTP API — spec em `34` — G-153..G-158 CONFIRMED; slices 25–26 feitas
23. Desktop Wails — spec em `35` — G-159..G-165 CONFIRMED; slices 27–28 feitas
24. NPM Player — spec em `36` — G-166..G-171 CONFIRMED; codigo = slice 29 — feito
25. Memory Player — spec em `38` — G-180..G-186 CONFIRMED; codigo = slice 31
26. Pytest + Yarn Players — spec em `37` — G-172..G-179 CONFIRMED; codigo = slice 30

## Criterio de “pronto para codar”

**P0 protocolo v0: CONFIRMADO.**  
**P1 Board/pipeline (G-20..G-27): CONFIRMADO.**  
**P1b Intent Engine (G-50..G-54): CONFIRMADO.**  
**P2 engenharia (G-30..G-38): CONFIRMADO** (G-36 DEFERRED).  
**Runtime Graph (G-60..G-65): CONFIRMADO.**  
**Graph Hits (G-66..G-69): CONFIRMADO** — slice 7 feito.  
**Git Player (G-70..G-74): CONFIRMADO** — slice 8 feito.
**Filesystem Player (G-75..G-80): CONFIRMADO** — slice 9 feito.  
**Execution Policy + HITL (G-81..G-86): CONFIRMADO** — slice 10 feito.  
**Docker Player (G-87..G-92): CONFIRMADO** — slice 11 feito.  
**Resource Claims (G-93..G-98): CONFIRMADO** — slice 12 feito.  
**Blast Radius (G-99..G-104): CONFIRMADO** — slice 13 feito.  
**TUI GRAPH (G-105..G-110): CONFIRMADO** — slice 14 feito.  
**Walk Blast←Graph (G-111..G-116): CONFIRMADO** — slice 15 feito.  
**HTTP Player (G-117..G-122): CONFIRMADO** — slice 16 feito.  
**Project Memory (G-123..G-128): CONFIRMADO** — slice 17 feito.  
**Test Player (G-129..G-134): CONFIRMADO** — slice 18 feito.  
**MVP 1.0 magro (G-135..G-140): CONFIRMADO** — slices 19–20 feitos.
**HTTP API (G-153..G-158): CONFIRMADO** — spec `34`; slices 25–26 feitas.
**Desktop Wails (G-159..G-165): CONFIRMADO** — spec `35`; slices 27–28 feitas.
**NPM Player (G-166..G-171): CONFIRMADO** — spec `36`; slice 29 feito.

Ordem pratica de codigo:
1. Core CLI + Shell (+ SQLite) — slice 1 — feito
2. Pipeline deterministic + LLM + Board adapter — slice 2 — feito
3. TUI Constellation Mission Control — slice 3 — feito
4. Validator JSON Schema + IDs estritos + sandbox Shell v0 — slice 4 — feito
5. Intent Engine NL v0 — slice 5 — feito
6. Runtime Graph v0 — slice 6 (G-60..G-65) — feito
7. Graph Hits v0 — slice 7 (G-66..G-69) — feito
8. Git Player v0 — slice 8 (G-70..G-74) — feito
9. Filesystem Player v0 — slice 9 (G-75..G-80) — feito
10. Execution Policy + HITL v0 — slice 10 (G-81..G-86) — feito
11. Docker Player v0 — slice 11 (G-87..G-92) — feito
12. Resource Claims v0 — slice 12 (G-93..G-98) — feito
13. Blast Radius v0 — slice 13 (G-99..G-104) — feito
14. TUI GRAPH v0 — slice 14 (G-105..G-110) — feito
15. Walk Blast←Graph v0 — slice 15 (G-111..G-116) — feito
16. HTTP Player v0 — slice 16 (G-117..G-122) — feito
17. Project Memory v0 — slice 17 (G-123..G-128) — feito (spec `29`)
18. Test Player v0 — slice 18 (G-129..G-134) — feito (spec `30`)
19. Intent player heuristics — slice 19 (G-135..G-136) — feito (spec `031`)
20. Context Engine v0 — slice 20 (G-137..G-139) — feito (spec `031`)
21. Intent Surface TUI — slice 21 (G-141..G-146) — feito (spec `32`)
22. Evolution v0 — slices 22–24 (G-147..G-152) — feito (spec `33`)
23. HTTP API v0 — slices 25–26 (G-153..G-158) — feito (spec `34`)
24. Desktop Wails v0 — slices 27–28 feitas (G-159..G-165) — spec `35`
25. NPM Player v0 — spec `36` (G-166..G-171); slice 29 feito
26. Memory Player v0 — spec `38` (G-180..G-186); slice 31
27. Pytest + Yarn Players v0 — spec `37` (G-172..G-179); slice 30
28. MCP Memory Server v0 — spec `39` (G-187..G-193); slice 32 a fazer

P3 restante (G-40 templates; G-41 em andamento — infra / K8s / TF / PG).
MVP 1.0 magro: spec `09`/`31` (G-135..G-140); slices 19–20 feitos.
Test Player corte v0: spec `30` (G-129..G-134); slice 18 feito.
Project Memory corte v0: spec `29` (G-123..G-128); slice 17 feito.
Esboço conceitual permanece em `16`.
Walk Blast←Graph: spec `27` (G-111..G-116); slice 15 feito.
HTTP Player: spec `28` (G-117..G-122); slice 16 feito.
Aba TUI GRAPH: spec `26` — slice 14 feito.
Experimentos de sidecar (Fase A) nao exigem mudanca no Core.
