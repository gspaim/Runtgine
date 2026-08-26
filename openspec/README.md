# OpenSpec no Runtgine

Pasta de mudanças no padrão [OpenSpec](https://github.com/Fission-AI/OpenSpec),
adaptada às convenções já confirmadas em [`docs/15-git-workflow.md`](../docs/15-git-workflow.md).

## Layout

```text
openspec/
├── README.md                 # este arquivo
├── config.yaml               # contexto do projeto para agentes
├── specs/                    # comportamento atual (source of truth)
│   └── <domain>/spec.md
└── changes/
    ├── archive/              # mudanças concluídas
    └── <NNN>-<slug>/         # mudança ativa
        ├── proposal.md
        ├── design.md
        ├── tasks.md
        ├── .openspec.yaml
        └── specs/<domain>/spec.md   # deltas ADDED|MODIFIED|REMOVED
```

## Naming — obrigatório

Changes usam o **mesmo** id numérico das branches e specs de produto:

```text
openspec/changes/<NNN>-<slug>/
branch:              feat/<NNN>-<slug>   (ou docs/ fix/ chore/)
```

| Parte | Regra | Exemplo |
|---|---|---|
| `NNN` | Zero-padded ≥ 3 dígitos; id da spec | `019` |
| `slug` | kebab-case curto | `graph-hits` |

Exemplos válidos: `001-shell-player`, `017-intent-engine`, `019-graph-hits`.

**Próximo id livre:** `041`. Nenhuma mudança ativa.
Último archive:
[`changes/archive/2026-08-26-040-workflow-templates/`](changes/archive/2026-08-26-040-workflow-templates/)
(Workflow Templates v0; G-194..G-200; slice 33).

## Autoridade

| Camada | Papel |
|---|---|
| [`docs/04-decisoes.md`](../docs/04-decisoes.md) | Status CONFIRMED / HYPOTHESIS (autoridade) |
| `docs/01`–`18` (+) | Visão, protocolo, domínio |
| `openspec/specs/` | Comportamento **atual** do sistema (para agentes) |
| `openspec/changes/<NNN>-slug/` | Pacote de trabalho da próxima mudança |

Não codificar mudança cujo status em `04` não esteja **CONFIRMED**
(ou equivalente explícito). OpenSpec **não** substitui `04`.

## Fluxo

1. Escolher `NNN` + slug; criar `openspec/changes/NNN-slug/`
2. Preencher `proposal` → deltas em `specs/` → `design` → `tasks`
3. Promover cortes em `docs/04-decisoes.md` (e doc canônico se houver)
4. Branch `feat/NNN-slug` a partir de `develop`; implementar `tasks.md`
5. Ao concluir: merge deltas em `openspec/specs/`; mover pasta para
   `changes/archive/YYYY-MM-DD-NNN-slug/`

## Domínios atuais em `specs/`

| Domain | Cobre |
|---|---|
| `contextpack` | AssembleContext / ContextPack (+ graph_hits, memory_hits, playbook_hits; semente repo_hits = slice 20) |
| `intent-engine` | NL → Task IR (+ heuristicas Player = slice 19) |
| `runtime-graph` | Graph estrutural + QueryHits |
| `git-player` | Player `git.*` (status/diff/log/add/commit) |
| `docker-player` | Player `docker.*` (ps/inspect/logs/run/build) |
| `execution-policy` | allow/deny/HITL (`waiting_approval`); ordem Policy→Claim |
| `filesystem-player` | Player `fs.*` (read/write/list/stat) |
| `resource-claims` | Claims exclusivos `workspace`/`path`; auto-claim; `claim.conflict` |
| `blast-radius` | Impact Report Task IR; `runtgine blast`; overlay vs claims |
| `tui-graph` | Aba GRAPH da TUI sobre `GetGraphSnapshot` |
| `blast-graph-walk` | 1 hop `mentions` no Blast Report (`affected`) |
| `http-player` | Player `http.*` (GET/HEAD HTTPS) |
| `project-memory` | Provider episódico + `memory_hits` (não é Player) |
| `test-player` | Player `test.go` (`go test` no workspace; spec 030) |
| `intent-surface` | Aba INTENT TUI (slice 21; desktop = spec `35`) |
| `evolution-v0` | Player Router + Playbooks + Lessons (slices 22–24) |
| `http-api` | Entry Point HTTP `runtgine serve` + webhooks outbound (slices 25–26) |
| `wails-v0` | Entry Point desktop Wails v3 (`runtgine desktop`; slices 27–28) |
| `npm-player` | Player `npm.test` (`npm test` no workspace; spec 036) |
| `pytest-yarn-players` | Players `pytest.run` / `yarn.test` (spec 037) |
| `memory-player` | Player `memory.recall` / `memory.check` read-only (spec 038) |
| `mcp-memory` | Servidor MCP read-only sobre Project Memory (`runtgine mcp` + `/mcp`; spec 039) |
| `workflow-templates` | Templates JSON nativos → Task IR (`runtgine template`; spec 040) |
