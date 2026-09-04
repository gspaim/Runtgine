# AGENTS.md — Guia para LLMs e contribuidores

## Papel
Arquiteto + engenheiro principal do Runtgine. Nao implemente antes
das decisoes estarem registradas em `docs/04-decisoes.md`.

## Norte absoluto

1. Runtgine e runtime que transforma intencao em execucao verificavel
2. Deterministic-first: deterministico quando possivel, LLM quando necessario
3. Player e a abstracao central (nao Agent)
4. Event-driven: Task -> Event -> Queue -> Player -> Result
5. Validacao antes da execucao (filosofia de compilador)
6. Core e o produto. Interface e superficie.
7. Muitos Players deterministicos sao estrategicos

## Autoridade documental

1. `docs/04-decisoes.md`
2. Demais `docs/` oficiais (01–09; nao `00-rascunho`)
3. Este arquivo / README / REVIEW
4. `brainstorm.md` e `conversas-empryo.md` — historicos apenas

Mudancas de implementacao ativas: `openspec/changes/<NNN>-<slug>/`
(padrao OpenSpec; naming alinhado a `feat/<NNN>-<slug>`). Ver
`openspec/README.md` e `docs/15-git-workflow.md`. `openspec/` **nao**
substitui `04` — so organiza proposal/design/tasks/deltas.

MVP canônico: `docs/09-mvp.md`.
Gaps: `docs/10-gaps.md`.
Protocolo v0 (proposta): `docs/11-protocolo-v0.md` — confirmar antes de codar.
Design da TUI: `docs/14-tui-design.md`.
Git / releases: `docs/15-git-workflow.md`.
Project Memory (esboco): `docs/16-project-memory.md` — historico; corte v0 em `29`.
Intent Engine v0: `docs/17-intent-engine-v0.md` — CONFIRMED.
Runtime Graph: `docs/18-runtime-graph-v0.md` — CONFIRMED v0 (G-60..G-65).
Graph Hits: `docs/19-graph-hits-v0.md` + archive
`openspec/changes/archive/2026-08-17-019-graph-hits/` — CONFIRMED v0
(G-66..G-69); slice 7 feito.
Git Player: `docs/20-git-player-v0.md` + archive
`openspec/changes/archive/2026-08-17-020-git-player/` — CONFIRMED v0
(G-70..G-74); slice 8 feito.
Filesystem Player: `docs/21-filesystem-player-v0.md` + archive
`openspec/changes/archive/2026-08-17-021-filesystem-player/` — CONFIRMED v0
(G-75..G-80); slice 9 feito.
Execution Policy + HITL: `docs/22-execution-policy-v0.md` + archive
`openspec/changes/archive/2026-08-17-022-execution-policy/` — CONFIRMED v0
(G-81..G-86); slice 10 feito.
Docker Player: `docs/23-docker-player-v0.md` + archive
`openspec/changes/archive/2026-08-17-023-docker-player/` — CONFIRMED v0
(G-87..G-92); slice 11 feito.
Resource Claims: `docs/24-resource-claims-v0.md` + archive
`openspec/changes/archive/2026-08-18-024-resource-claims/` — CONFIRMED v0
(G-93..G-98); slice 12 feito.
Blast Radius: `docs/25-blast-radius-v0.md` + archive
`openspec/changes/archive/2026-08-18-025-blast-radius/` — CONFIRMED v0
(G-99..G-104); slice 13 feito.
TUI GRAPH: `docs/26-tui-graph-v0.md` + archive
`openspec/changes/archive/2026-08-18-026-tui-graph/` — CONFIRMED v0
(G-105..G-110); slice 14 feito.
Walk Blast←Graph: `docs/27-blast-graph-walk-v0.md` + archive
`openspec/changes/archive/2026-08-18-027-blast-graph-walk/` — CONFIRMED v0
(G-111..G-116); slice 15 feito.
HTTP Player: `docs/28-http-player-v0.md` + archive
`openspec/changes/archive/2026-08-19-028-http-player/` — CONFIRMED v0
(G-117..G-122); slice 16 feito (nao e G-45 API HTTP).
Project Memory: `docs/29-project-memory-v0.md` + archive
`openspec/changes/archive/2026-08-19-029-project-memory/` — CONFIRMED v0
(G-123..G-128); slice 17 feito (esboco em `16`; nao e MCP / Memory Player).
Test Player: `docs/30-test-player-v0.md` + archive
`openspec/changes/archive/2026-08-19-030-test-player/` — CONFIRMED v0
(G-129..G-134); slice 18 feito (nao e pytest/npm; nao e G-45).
MVP 1.0 magro: `docs/09-mvp.md` + `docs/31-context-engine-v0.md` +
archive `openspec/changes/archive/2026-08-19-031-mvp-1.0/` — CONFIRMED
(G-135..G-140); slices 19–20 feitos (nao e G-45 / NATS / Wails / MCP).
Intent Surface: `docs/32-intent-surface-v0.md` + archive
`openspec/changes/archive/2026-08-21-032-intent-surface/` — CONFIRMED
(G-141..G-146); slice 21 TUI feito; Wails = spec `35`.
Evolution v0: `docs/33-evolution-v0.md` + archive
`openspec/changes/archive/2026-08-21-033-evolution-v0/` — CONFIRMED
(G-147..G-152); slices 22–24 feitas.
HTTP API: `docs/34-http-api-v0.md` + archive
`openspec/changes/archive/2026-08-21-034-http-api/` — CONFIRMED
(G-153..G-158); slices 25–26 feitas (`runtgine serve` + webhooks outbound;
nao e HTTP Player `28`).
Desktop Wails: `docs/35-wails-v0.md` + archive
`openspec/changes/archive/2026-08-21-035-wails-v0/` — CONFIRMED
(G-159..G-165); slices 27–28 feitas (Wails v3).
NPM Player: `docs/36-npm-player-v0.md` + archive
`openspec/changes/archive/2026-08-22-036-npm-player/` — CONFIRMED
(G-166..G-171); slice 29 feita.
Memory Player: `docs/38-memory-player-v0.md` + archive
`openspec/changes/archive/2026-08-24-038-memory-player/` — CONFIRMED
(G-180..G-186); slice 31 fecha OPEN QUESTION G-47.
Pytest + Yarn Players: `docs/37-pytest-yarn-players-v0.md` + archive
`openspec/changes/archive/2026-08-24-037-pytest-yarn-players/` — CONFIRMED
(G-172..G-179); slice 30.
MCP Memory Server: `docs/39-mcp-memory-v0.md` + archive
`openspec/changes/archive/2026-08-25-039-mcp-memory/` — CONFIRMED
(G-187..G-193); slice 32 feito (fecha G-44; servidor read-only).
Workflow Templates: `docs/40-workflow-templates-v0.md` + archive
`openspec/changes/archive/2026-08-26-040-workflow-templates/` — CONFIRMED
(G-194..G-200); recorte de G-40; slice 33.
Infra Players: `docs/41-infra-players-v0.md` + archive
`openspec/changes/archive/2026-08-26-041-infra-players/` — CONFIRMED
(G-201..G-209); slice 34 (k8s / terraform / postgres).
Helm Player: `docs/42-helm-player-v0.md` + archive
`openspec/changes/archive/2026-09-04-042-helm-player/` — CONFIRMED
(G-210..G-216); slice 35 feito.
Cloud AWS Player: `docs/43-aws-player-v0.md` + archive
`openspec/changes/archive/2026-09-04-043-aws-player/` — CONFIRMED
(G-217..G-223); slice 36 feito.
Cloud Players GCP + Azure: `docs/44-cloud-gcp-azure-players-v0.md`
+ archive `openspec/changes/archive/2026-09-04-044-cloud-gcp-azure/`
— CONFIRMED (G-224..G-230); slice 37 feito.
Postgres EXPLAIN: `docs/45-pg-explain-v0.md` + archive
`openspec/changes/archive/2026-09-04-045-pg-explain/` — CONFIRMED
(G-231..G-237); slice 38 feito; fecha a fila nomeada de G-41.
Skill obrigatoria para TUI: `.cursor/skills/runtgine-tui-design/SKILL.md`.

## Ordem de trabalho

1. Entender o dominio (docs 01 a 09)
2. Revisar gaps (`10`) e confirmar protocolo v0 (`11` → `04-decisoes`)
3. Criar arquitetura do Core (layout em `11` §16, apos confirmacao)
4. Implementar Event Bus + Task model (schemas confirmados)
5. Implementar Validator basico
6. Implementar Shell Player (+ sandbox v0)
7. CLI minima
8. TUI minima
9. Board Integration + pipeline vertical (ver 09-mvp; gaps P1)
10. Context assembly + LLM Player + Router
11. Intent Engine (NL) — CONFIRMED v0 em `17-intent-engine-v0.md`
12. Runtime Graph — CONFIRMED v0 em `18-runtime-graph-v0.md` (G-60..G-65) — feito
13. Graph Hits — CONFIRMED v0 em `19` + OpenSpec archive `019-graph-hits` — feito
14. Git Player — CONFIRMED v0 em `20` + OpenSpec archive `020-git-player` — feito
15. Filesystem Player — CONFIRMED v0 em `21` + OpenSpec archive `021-filesystem-player` — feito
16. Execution Policy + HITL — CONFIRMED v0 em `22` + OpenSpec archive `022-execution-policy` — feito
17. Docker Player — CONFIRMED v0 em `23` + OpenSpec archive `023-docker-player` — feito
18. Resource Claims — CONFIRMED v0 em `24` + OpenSpec archive `024-resource-claims` — feito
19. Blast Radius — CONFIRMED v0 em `25` + OpenSpec archive `025-blast-radius` — feito
20. TUI GRAPH — CONFIRMED v0 em `26` + OpenSpec archive `026-tui-graph` — feito
21. Walk Blast←Graph — CONFIRMED v0 em `27` + OpenSpec archive `027-blast-graph-walk` — feito
22. HTTP Player — CONFIRMED v0 em `28` + OpenSpec archive `028-http-player` — feito
23. Project Memory — CONFIRMED v0 em `29` + OpenSpec archive `029-project-memory` — feito
24. Test Player — CONFIRMED v0 em `30` + OpenSpec archive `030-test-player` — feito
25. MVP 1.0 magro — CONFIRMED em `09`/`31` + archive `031-mvp-1.0` — slices 19–20 feitos
26. Intent Surface — CONFIRMED v0 em `32` + archive `032-intent-surface` — slice 21 TUI feito; Wails = spec `35`
27. Evolution v0 — CONFIRMED em `33` + archive `033-evolution-v0` — slices 22–24 feitas
28. HTTP API — CONFIRMED em `34` + archive `034-http-api` — slices 25–26 feitas
29. Desktop Wails — CONFIRMED v0 em `35` + archive `035-wails-v0` — slices 27–28 feitas
30. NPM Player — CONFIRMED v0 em `36` + archive `036-npm-player` — slice 29 feito
31. Memory Player — CONFIRMED v0 em `38` + archive `038-memory-player` — slice 31
32. Pytest + Yarn — CONFIRMED v0 em `37` + archive `037-pytest-yarn-players` — slice 30
33. MCP Memory — CONFIRMED v0 em `39` + archive `039-mcp-memory` — slice 32
34. Workflow Templates — CONFIRMED v0 em `40` (G-194..G-200) — slice 33
35. Infra Players — CONFIRMED v0 em `41` (G-201..G-209) — slice 34
36. Helm Player — CONFIRMED v0 em `42` (G-210..G-216) — slice 35 feito
37. Cloud AWS Player — CONFIRMED v0 em `43` (G-217..G-223) — slice 36 feito
38. Cloud Players GCP + Azure — CONFIRMED v0 em `44` (G-224..G-230) — slice 37 feito
39. Postgres EXPLAIN — CONFIRMED v0 em `45` (G-231..G-237) — slice 38 feito; fila G-41 fechada
40. Depois — NATS G-36 DEFERRED; novas promocoes exigem decisao em `04`

## Conceitos chave (nao confundir)

- Task != Workflow != Execution Plan
- Event != Queue != Workflow
- Player != Agent
- Entry Point != Player
- Runtgine != Chorus (sao complementares)
- Intent Engine NAO e autoridade (Registry rejeita capabilities invalidas)

## O que NAO fazer

- Nao codificar antes das decisoes
- Nao tratar Runtgine como framework de agentes
- Nao confundir Runtgine com Chorus
- Nao pular o Validator (filosofia de compilador)
- Nao construir UI rica (Wails) antes do Core + CLI/TUI funcionarem
  (ja funcionam; codigo desktop so apos spec `35`)
- Nao usar brainstorm/conversas-empryo como fonte de stack (Rust/GPUI estao REJECTED)
- Protocolo v0 P0 esta CONFIRMED em `04-decisoes` / `11` — Core liberado
- Nao implementar gaps P1 (Board/LLM) sem especificar contratos (G-20+)
- Ao trabalhar na TUI, seguir `docs/14-tui-design.md` e a skill
  `.cursor/skills/runtgine-tui-design/SKILL.md`
- Nao transformar a TUI em multiplexer nem adicionar tuios/PTY sem nova decisao
- Nao abrir PR de feature direto para `main`; base = `develop` (a default do GitHub e `main`; trocar a base). Ver `15-git-workflow.md`
- Nao publicar tag estavel `vX.Y.Z` sem passar por `release/x.y.z` e RC

## Cursor Cloud specific instructions

Projeto Go puro (Go 1.25+, ja instalado na imagem). Comandos padrao de
lint/test/build/run estao no `README.md` (secao "Começando" e "CLI"):
`go vet ./...`, `go test ./...`, `go build -o bin/runtgine ./cmd/runtgine`.
O update script (startup) roda `go mod download`; nao ha build/test/servico no startup.

Notas nao obvias para desenvolver aqui:

- SQLite usa `modernc.org/sqlite` (Go puro): nao precisa de CGO nem de
  bibliotecas de sistema. `CGO_ENABLED=0` funciona (o comando `desktop`
  vira stub). O binario padrao no Linux com CGO precisa de GTK4 + WebKitGTK 6
  para `runtgine desktop`.
- Nao ha servico de longa duracao **hoje** no startup: e um binario CLI/TUI.
  `runtgine serve` sobe a HTTP API. `runtgine desktop` abre a janela Wails.
- Estado persistido em `<workspace>/.runtgine/runtgine.db` (gitignored). Para
  um estado limpo, apague `.runtgine/` entre execucoes.
- `runtgine tui` exige um TTY interativo; nao roda em pipe/CI nao interativo.
  Use `RUNTGINE_ASCII=1` e `NO_COLOR=1` para glifos/cores simplificados.
- Pipeline e LLM Players funcionam offline sem credenciais: usam um completer
  heuristico. Credenciais LLM/GitHub sao opcionais e so via variaveis de
  ambiente (ver tabela em `README.md`).
- Smoke test rapido end-to-end: `./bin/runtgine run examples/hello.json` deve
  emitir `run.succeeded` com stdout `hello-runtgine`.
