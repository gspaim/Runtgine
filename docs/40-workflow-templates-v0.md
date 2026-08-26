# 40 — Workflow Templates v0

Templates **executáveis** no workspace: JSON em
`.runtgine/templates/*.json` que compilam para Task IR v0
(multi-step). O Validator / Registry continuam soberanos — o
template não inventa capability nem bypassa Policy.

Inventário: [10-gaps.md](10-gaps.md) (G-194+; recorte de G-40).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Esboço / TLC SDD: [08-workflow-templates.md](08-workflow-templates.md)
(histórico; Playbooks v0 em `33` G-149).

**Status deste doc: CONFIRMED v0 (slice 33 feito).** G-194..G-200.
Fecha G-40 como **carregamento nativo** no workspace (não repo
externo). Não é Playbook (markdown / `playbook_hits`). Não é
Player. Não é o pipeline Board (`pipeline.*`). Não é auto-sizing
nem Verifier.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-26-040-workflow-templates/`](../openspec/changes/archive/2026-08-26-040-workflow-templates/)
(arquivado após o slice 33).

---

## 1. Problema

Playbooks (`33`, G-149) são documentação executável: o ContextPack
recebe `playbook_hits`, mas **não** viram Task IR. Intent ainda
compila para um step (`git.status`, `test.go`, `shell.exec`) ou
para o pipeline Board linear.

Falta o registro reutilizável que `08` descreveu: um processo com
steps/capabilities já conhecidas do Registry, versionado no
projeto, disparável por CLI/Intent, validado na admissão.

G-40 perguntava nativo vs repo externo. Este corte confirma
**nativo**: arquivos no workspace, irmãos de `.runtgine/playbooks/`.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| JSON no workspace (`.runtgine/templates/`) | Repo git externo / marketplace |
| Compila para Task IR v0 | Player / capability no Registry |
| Steps com capabilities já registradas | Auto-sizing TLC / fases opcionais |
| CLI `runtgine template` + Intent | Verifier Player; autor ≠ verifier |
| Nó Graph `template` (best-effort) | Engine de Workflow / DAG runtime novo |
| Load best-effort no boot | Mutação via LLM; HTTP API própria |

Regras:

1. Template **nunca** executa direto. Sempre Task IR → Validator →
   Runner.
2. Capability desconhecida → rejeição na admissão (Registry).
3. Playbooks permanecem markdown + hits (`33`); templates não
   geram `playbook_hits`.
4. Falha de load (arquivo inválido) **degrada**: skip +
   `slog.Warn`; Core sobe.
5. Pacote Go: `internal/core/templates`. Não entra no Registry.
6. Sem novo `schema_version` de Task IR.

---

## 3. Cortes confirmados (G-194+)

### G-194 — Papel e pacote

- Nome: Workflow Template v0
- Pacote: `internal/core/templates` (Core, não Player, não Entry Point)
- Recorte de G-40: loading **nativo** no workspace
- Distinto de Playbooks (`33`) e do template Board (`pipeline.NewTaskIR`)

### G-195 — Schema JSON v0

Arquivo `.json` (um template por arquivo):

| Campo | Regra |
|---|---|
| `schema_version` | opcional; se presente, estrito `"0.1.0"` |
| `id` | obrigatório; `[a-zA-Z0-9._-]{1,64}` |
| `title` | obrigatório; 1–200 runes |
| `steps` | array 1–20; cada item espelha um step de Task IR |

Step:

| Campo | Regra |
|---|---|
| `step_id` | obrigatório; único no arquivo |
| `capability` | obrigatório; naming `domain.action` (G-05) |
| `input` | objeto JSON; default `{}` |
| `depends_on` | ids de steps do mesmo arquivo; ciclo rejeitado no load |

`additionalProperties: false` na raiz e no step. Sem `phases`,
`gates`, `verifier`, `auto_size` neste corte — um gate v0 é um
step determinístico (`test.go`, `git.status`, …).

### G-196 — Loading nativo (fecha G-40)

- Diretório: `<workspace>/.runtgine/templates/*.json`
- Boot: `api.Open` carrega best-effort (espelha Playbooks)
- Id duplicado: o **primeiro** arquivo em ordem de `ReadDir` vence;
  os demais são skip + warning
- JSON inválido / schema falho: skip + warning; não derruba o Core
- **Não** carrega de git remoto, HTTP, ou path fora do workspace

### G-197 — Compile → Task IR

`templates.Compile(tpl, entryPoint, ref, summary) → task.Task`:

- `source.entry_point` / `ref` do caller
- `intent.summary` = title do template (notes = summary do caller)
- `steps` copiados; `input` omitido vira `{}`
- `metadata.template` = `id`
- `task_id` UUID v7 gerado; `schema_version` `"0.1.0"`

Compile **não** consulta o Registry. Validator na admissão rejeita
capability inexistente / input inválido.

### G-198 — CLI + Intent

CLI:

| Comando | Efeito |
|---|---|
| `runtgine template list` | JSON dos templates carregados |
| `runtgine template show <id>` | JSON de um template |
| `runtgine template run <id>` | Compile + `SubmitTask` (`--wait` default true; `--dry-run` imprime Task IR) |

Intent (antes de `matchShell`, para `"run template X"` não virar
`shell.exec`):

| NL (PT/EN, case-insensitive) | Método |
|---|---|
| `run template <id>` | `heuristic.template` |
| `roda o template <id>` / `rodar template <id>` | idem |
| `template <id>` | idem |

Prefixo reconhecido + id desconhecido → `validation.invalid_input`
(não cai no shell).

### G-199 — Graph

Kind aditivo `template` (não substitui o conjunto G-61). Boot:
`RefreshFromTemplates` faz upsert de nó `{kind: template, id, attrs.title}`.
Sem aresta nova (G-62 intacto). Falha degrada (warning), não falha
o Open. TUI GRAPH não ganha filtro novo neste slice.

### G-200 — Exclusões v0

- Repo externo / git submodule / URL de template
- Auto-sizing, fases opcionais TLC, Verifier Player
- Gates como tipo distinto (são steps)
- Mutação de template via LLM / Lessons
- Template como Player (`template.run` no Registry)
- HTTP API / MCP / Wails views específicas
- Substituição do pipeline Board
- `depends_on` entre templates (só intra-arquivo)

---

## 4. Critérios de aceite

1. Template com capability inexistente → `task.rejected` na admissão.
2. `runtgine intent --dry-run "run template <id>"` → Task IR
   multi-step, method `heuristic.template`, **não** `shell.exec`.
3. Prefixo `run template` + id ausente → erro de validação, não
   shell.
4. Arquivo JSON inválido no dir → Core sobe; list omite o arquivo.
5. `go test ./internal/core/templates/...` verde sem binários extras.
6. `go test ./...` / `go vet ./...` verdes.
7. OpenSpec `040` arquivado após o **código** (slice 33).

---

## 5. Ordem do slice de código

1. Pacote `internal/core/templates` (Load / Compile / lookup)
2. `api.Open` carrega + Graph `RefreshFromTemplates`
3. Intent `heuristic.template` (antes de `matchShell`)
4. CLI `runtgine template list|show|run`
5. Example `examples/templates/verify.json`
6. Testes; README Estágio: Slice 33
7. Arquivar OpenSpec `040`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-194 Papel (`internal/core/templates`; não Player)
- [x] G-195 Schema JSON (id/title/steps ≤20)
- [x] G-196 Loading nativo (fecha G-40; sem repo externo)
- [x] G-197 Compile → Task IR (Validator soberano)
- [x] G-198 CLI + Intent `heuristic.template`
- [x] G-199 Graph kind `template` (aditivo)
- [x] G-200 Exclusões (auto-sizing, verifier, remoto, Player)
