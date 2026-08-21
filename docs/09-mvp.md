# 09 — MVP: execução verificável

Fonte de verdade deste escopo. `05-prd.md` lista requisitos; este
documento define o corte. Em conflito, prevalece este arquivo e
`04-decisoes.md`.

Dois andares:

| Andar | Significado |
|---|---|
| **MVP realizado** | Já no código (slices 1–20, incluindo 1.0 magro) |
| **MVP 1.0 magro** | Feito: NL → Player certo + semente `repo_hits` |

A frase do produto é **intenção → execução verificável**. O runtime
mínimo (Core + Shell + CLI/TUI + Board) já provou que o Core existe.
O 1.0 magro prova o loop sem `shell.exec` livre quando já há Player.

---

## Principio do corte

1. Core é o produto. CLI/TUI/Board são superfícies.
2. Deterministic-first: Player com schema antes de argv livre.
3. Validator / Registry continuam soberanos (Intent não inventa capability).
4. Magro: **não** puxar API HTTP (G-45), NATS, Wails, MCP, templates.

---

## MVP realizado (slices 1–20)

Já implementado. Não voltar a listar estes itens como “fora do MVP”.

### Core e superfícies

- Task IR v0, Validator (JSON Schema), Event Bus in-process, Registry
- Runner, SQLite, Execution Policy + HITL, Resource Claims, Blast Radius
- CLI (`run`, `status`, `intent`, `graph`, `memory`, `blast`, …)
- TUI Constellation (incluindo aba GRAPH)
- Board GitHub Projects + pipeline vertical (`12`)
- Intent Engine NL (`17`) — heuristics shell / pipeline / Players (slice 19)
- Runtime Graph + Graph Hits + Project Memory (`memory_hits`)

### Players determinísticos

Shell, Git, Filesystem, Docker, HTTP (`http.get`/`head`), Test (`test.go`).
Pipeline + LLM Players do cenário Board.

### Criterios de sucesso (já atendidos)

- `runtgine run examples/hello.json` → `run.succeeded`
- Validator rejeita capability inexistente / input inválido na admissão
- Task do board passa pelo pipeline quando o cenário vertical está ligado
- Falha retorna erro claro na CLI/TUI

---

## MVP 1.0 magro (feito)

O que falta para o loop ser honesto:

| Item | Gaps | Código |
|---|---|---|
| Heurísticas Intent → Players atuais | G-135..G-136 | slice 19 — feito |
| Context Engine v0 (seed `repo_hits`) | G-137..G-139 | slice 20 — feito |

Exclusões comuns: G-140.

Doc canônico do Context Engine: [31-context-engine-v0.md](31-context-engine-v0.md).
Intent: emenda em [17-intent-engine-v0.md](17-intent-engine-v0.md).
OpenSpec: [`openspec/changes/archive/2026-08-19-031-mvp-1.0/`](../openspec/changes/archive/2026-08-19-031-mvp-1.0/).

### 1. Heurísticas de Intent (slice 19)

Antes do prefixo genérico `go ` / argv `git` → `shell.exec`, o Engine
reconhece frases de alta confiança e emite Task IR com a capability
já registrada:

| NL (PT/EN, case-insensitive) | Capability |
|---|---|
| `go test`, `roda os testes`, `rodar testes`, `run tests` | `test.go` |
| `git status` | `git.status` |
| `git diff` | `git.diff` |
| `git log` | `git.log` |
| `docker ps` | `docker.ps` |

Ordem: vazio → **player** → shell → pipeline → LLM.
`go test` **não** pode virar `shell.exec`.
Caminho LLM continua `route: shell|pipeline` (não inventa Players).
Validator rejeita o que o Registry não conhece.

### 2. Context Engine v0 (slice 20)

Não é Player. É o nome do assembler do ContextPack (`AssembleContext`).

Hoje `repo_hits` só existe se `pipeline.repo-search` rodou neste Run.
No v0, se `repo_hits` estiver vazio, o Core **semeia** paths/symbols a
partir de `QueryHits` (nós `path` / `symbol` do Graph), no budget
`max_files` já existente. Falha → lista vazia; o Run não cai.
Sem walk do workspace, sem embeddings, sem corpo de arquivo no pack.

### Criterios extra do 1.0

- `runtgine intent "roda os testes"` → um step `test.go` (não `shell.exec`)
- `runtgine intent "git status"` → `git.status`
- Step LLM **sem** `repo-search` neste Run ainda pode ter `repo_hits`
  (semente do Graph) ou `[]` se o Graph estiver vazio
- `go test ./...` / `go vet ./...` verdes

---

## Explicitamente fora do 1.0

| Fica fora | Por quê |
|---|---|
| API HTTP / webhooks (G-45) | Superfície CI; não fecha o loop local |
| NATS / bus distribuído (G-36) | DEFERRED; um processo basta |
| Wails / desktop | Fase 3; TUI já é superfície |
| MCP (G-44) | Runtgine não é alternativa a MCP |
| Player Router completo | Fora do 1.0; spec v0 em `33` (slice 22) |
| Playbooks / Lessons / multi-model LLM | Fora do 1.0; spec v0 em `33` (slices 23–24) |
| Workflow Templates / TLC SDD (`08`) | Task ≠ Workflow; motor novo |
| Plugin system / event sourcing | Plataforma, não prova |
| Memory Player (G-47) | Provider já existe; Player OPEN |
| `http.post`, `git.push` / `add` / `commit` via NL | Escrita / rede; heuristicas só leitura |
| pytest / npm / `-race` / K8s / Terraform / PostgreSQL | Outros recortes G-41 |
| Embeddings / RAG / dump do repositório no pack | Fora do Context Engine v0 |

UC-02 (CI/CD via HTTP) é **pós-1.0**: spec [34-http-api-v0.md](34-http-api-v0.md)
(G-153..G-158). Até o código das slices 25–26, CI usa CLI.

---

## Entry points

| Entry Point | 1.0 magro | Notas |
|---|---|---|
| CLI | sim | Task IR + `runtgine intent` |
| TUI | sim | Observação hoje; aba **INTENT** confirmada em `32` (slice 21) |
| Board (GitHub Projects) | sim | Entry Point ≠ Player |
| API HTTP | spec `34` | Fora do 1.0 magro; slices 25–26 |
| Desktop (Wails) | não | Fase 3 |
| Web | não | Futuro |

---

## Ordem de codigo (a partir daqui)

1. Slice 19 — heurísticas Intent (G-135..G-136) — feito
2. Slice 20 — Context Engine v0 (G-137..G-139) — feito
3. OpenSpec `031` arquivado — feito
4. Slice 21 — Intent Surface TUI (G-141..G-146; ver `32`)
5. Slices 22–24 — Evolution v0 (G-147..G-152; ver `33`): Router, Playbooks, Lessons
6. Slices 25–26 — HTTP API v0 (G-153..G-158; ver `34`): `serve` + webhooks outbound
7. Depois: nova promoção em `04` (mais Players, Wails Fase 3 incl. INTENT desktop, MCP, …)

Histórico do runtime mínimo (Task IR → Shell → CLI → TUI → Board →
pipeline) está nos slices 1–4 / `11` / `12`. Não reabrir.
