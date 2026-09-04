# 20 — Git Player v0

Primeiro Player determinístico além do Shell: operações Git locais
com contrato `domain.action`, sandbox alinhada ao Shell v0.

Inventário: [10-gaps.md](10-gaps.md) (G-70+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Core + Shell + Validator estáveis (slices 1–7).

**Status deste doc: CONFIRMED (v0).** G-70..G-74 autorizam o slice 8
de código. HITL / Approvals (G-42) e Execution Policy completa
permanecem fora (HYPOTHESIS / P3).

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-17-020-git-player/`](../openspec/changes/archive/2026-08-17-020-git-player/).
Deltas mergeados em `openspec/specs/git-player/`. Branch de implementação:
`cursor/020-git-player-0ac1`.

---

## 1. Problema

Hoje o único Player determinístico de I/O é `shell.exec`. Tasks que
precisam de status/diff/commit Git passam por argv genérico — sem
`input_schema` específico, sem política Git e sem capability estável
para o Router / Intent / Graph.

O protocolo já cita `git.commit` como exemplo de naming (`11` §5).
Falta o Manifest, o pacote e o sandbox.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Player `git` determinístico | Agent / LLM |
| Capabilities `git.*` no Registry | Substituto do Shell |
| Invoca binário `git` com argv | Shell string / `sh -c` |
| Workdir dentro do workspace | Clone de URL arbitrária / push remoto no v0 |
| Sandbox estática + timeout | Execution Policy / HITL completa |

Regras:

1. Validator / Registry continuam soberanos
2. Sem rede no v0: **sem** `push`, `pull`, `fetch`, `clone` remoto
3. Sem reescrita destrutiva: **sem** `reset --hard`, `rebase`, `force`, `clean -fdx`
4. Binário: `git` no `PATH` (mesmo padrão permissivo+warn do Shell allowlist)
5. Intent heuristics **podem** ganhar atalhos leves (`git status` → `git.status`);
   não obrigatório para o aceite mínimo do Player

---

## 3. Cortes confirmados (G-70+)

### G-70 — Papel e pacote

**Status: CONFIRMED**

- Nome do Player: `git`
- Pacote: `internal/players/git`
- Registro em `api.Open` junto de shell / pipeline / llm
- Kind: `deterministic`
- Recorte de **G-41** (biblioteca ampla): este slice entrega **só** Git v0

### G-71 — Capabilities v0

**Status: CONFIRMED**

| Capability | Entrada (resumo) | Saída (resumo) |
|---|---|---|
| `git.status` | `workdir?`, `timeout_ms?` | `branch`, `porcelain` (linhas), `clean` (bool) |
| `git.diff` | `workdir?`, `staged?` (bool), `paths?[]`, `timeout_ms?` | `diff` (string; truncável) |
| `git.log` | `workdir?`, `max?` (default 10, max 50), `timeout_ms?` | `entries[]` `{hash, subject, author, date}` |
| `git.add` | `paths[]` (obrigatório, ≥1), `workdir?`, `timeout_ms?` | `added[]`, `exit_code` |
| `git.commit` | `message` (obrigatório, não vazio), `workdir?`, `allow_empty?` (default false), `timeout_ms?` | `commit` (hash ou vazio), `exit_code`, `stderr?` |

Schemas JSON formais vivem no Manifest (como Shell). `additionalProperties: false`.

Paths em `git.add` / `git.diff`: relativos ao `workdir`; após resolve,
devem permanecer **dentro** do workspace root (mesma regra `EvalSymlinks`
do Shell). Negar `..` que escape.

### G-72 — Sandbox / policy mínima

**Status: CONFIRMED** (espelha G-18; sem Execution Policy completa)

| Regra | Corte v0 |
|---|---|
| Invocação | só argv → `git <subcommand> ...`; nunca shell string |
| Workdir | dentro do workspace root |
| Timeout | obrigatório; default **60s** |
| Env | herança mínima igual Shell (sem tokens / `RUNTGINE_*`); `GIT_DIR` / `GIT_WORK_TREE` do input **proibidos** |
| Subcommands | allowlist fixa: `status`, `diff`, `log`, `add`, `commit`, `rev-parse`, `symbolic-ref` (helpers internos) |
| Flags proibidas | qualquer `--upload-pack`, `--exec`, `-c core.hooksPath=…` arbitrário no input do usuário; commit usa `-m` apenas |
| Rede | não expor capabilities de rede; documentar que `git` do sistema ainda *pode* ter hooks — hooks path custom via input = **rejeitado** |
| Hooks | `git.commit` passa `-c core.hooksPath=/dev/null` (ou equivalente seguro no OS) para não executar hooks do repo no v0 |
| Identity | `git.commit` usa `-c user.email=runtgine@local -c user.name=runtgine` se identity ausente no repo (determinístico offline) |

Falha de sandbox → `validation.invalid_input` (estático) ou
`runtime.player_error` na execução.

### G-73 — Integração Core

**Status: CONFIRMED**

1. `api.Open` registra `git.New()`
2. Validator valida `input` contra `input_schema` do Manifest (já genérico)
3. Opcional v0: `ValidateStaticInput` no pacote git (paths/workdir) chamado
   no admission **se** o Runner/Validator já tiver gancho por capability —
   senão, checagens no `Execute` bastam no primeiro PR (preferir espelhar
   Shell: static no admission quando o gancho existir)
4. Runtime Graph: nós `capability` `git.*` via `RefreshFromRegistry` (já automático)
5. Exemplo: `examples/git-status.json`

### G-74 — Fora / deferido neste slice

**Status: CONFIRMED** (como exclusões)

- `git.push` / `pull` / `fetch` / `clone` / `remote`
- `git.checkout` / `branch` / `merge` / `rebase` / `reset` / `stash` / `tag`
- HITL / Approvals (G-42)
- Execution Policy genérica
- Intent heuristics Git (nice-to-have; não bloqueia)
- TUI dedicada
- Assinar commits / GPG
- Submodules

---

## 4. Critérios de aceite

1. `runtgine run examples/git-status.json` (num repo git) emite
   `run.succeeded` com output JSON contendo `branch` e `porcelain`/`clean`
2. `git.add` + `git.commit` num workspace temporário de teste cria commit
   sem hooks e sem rede
3. Path que escapa o workspace é rejeitado na validação/execução
4. Capability inventada `git.push` **não** está no Manifest → Validator rejeita
5. `go test ./internal/players/git/...` cobre status/diff/log/add/commit e
   path escape
6. `go test ./...` e `go vet ./...` verdes
7. OpenSpec `020-git-player` arquivado após merge do código

---

## 5. Ordem do slice de código

1. G-70..G-74 CONFIRMED — **este PR de docs/OpenSpec**
2. Pacote `internal/players/git` + Manifest + testes
3. Registrar em `api.Open` + exemplo JSON
4. Atualizar estágio README (Slice 8 Feito); arquivar OpenSpec

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-70 Papel / pacote `git`
- [x] G-71 Capabilities status/diff/log/add/commit
- [x] G-72 Sandbox mínima (sem rede, hooks off no commit)
- [x] G-73 Wire Registry + exemplo
- [x] G-74 Exclusões (push/HITL/…)
