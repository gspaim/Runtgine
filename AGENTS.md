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

MVP canônico: `docs/09-mvp.md`.
Gaps: `docs/10-gaps.md`.
Protocolo v0 (proposta): `docs/11-protocolo-v0.md` — confirmar antes de codar.
Design da TUI: `docs/14-tui-design.md`.
Git / releases: `docs/15-git-workflow.md`.
Project Memory (esboco): `docs/16-project-memory.md` — HYPOTHESIS; nao codificar.
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
11. Intent Engine (NL) — apenas apos Core estavel; ainda HYPOTHESIS
12. Runtime Graph — proposta em `18-runtime-graph-v0.md`; codificar so apos CONFIRMED em `04`

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
  bibliotecas de sistema. `CGO_ENABLED=0` funciona.
- Nao ha servico de longa duracao: e um binario CLI/TUI. Cada `runtgine run`
  executa e encerra; nao ha servidor para deixar rodando.
- Estado persistido em `<workspace>/.runtgine/runtgine.db` (gitignored). Para
  um estado limpo, apague `.runtgine/` entre execucoes.
- `runtgine tui` exige um TTY interativo; nao roda em pipe/CI nao interativo.
  Use `RUNTGINE_ASCII=1` e `NO_COLOR=1` para glifos/cores simplificados.
- Pipeline e LLM Players funcionam offline sem credenciais: usam um completer
  heuristico. Credenciais LLM/GitHub sao opcionais e so via variaveis de
  ambiente (ver tabela em `README.md`).
- Smoke test rapido end-to-end: `./bin/runtgine run examples/hello.json` deve
  emitir `run.succeeded` com stdout `hello-runtgine`.
