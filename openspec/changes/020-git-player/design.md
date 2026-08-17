# Design: 020-git-player

## Technical approach

### Package

`internal/players/git` com `New() *Player`, `Manifest()`, `Execute`.

Cada capability tem `input_schema` / `output_schema` no Manifest
(`additionalProperties: false`).

### Execution

```text
resolve workdir ⊆ workspace
build argv = ["git", ...]  // never shell
optional -c user.name / user.email / core.hooksPath for commit
exec.CommandContext + timeout
capture stdout/stderr/exit
map to capability output JSON
```

Helpers internos permitidos na allowlist de invocação: `rev-parse`,
`symbolic-ref`, `status`, `diff`, `log`, `add`, `commit`.

### Static validation

- `paths` não vazios; cada path após join(workdir, path) + EvalSymlinks
  deve ter prefixo do workspace
- `message` trim não vazio para commit
- `max` em log clamp 1..50

Prefer espelhar Shell: exportar `ValidateStaticInput(workspace, capability, raw)`
e plugar no admission se o Runner já despacha por capability; senão validar
no início de `Execute` no primeiro PR.

### Identity / hooks (commit)

```text
git -c core.hooksPath=/dev/null \
    -c user.name=runtgine \
    -c user.email=runtgine@local \
    commit -m <message>
```

Se o repo já tiver `user.name`/`user.email`, ainda assim forçar os `-c`
no v0 para determinismo em CI/sandbox. Documentar; v1 pode ler config
existente.

### Registration

`api.Open`: `reg.Register(git.New())` após shell.

### Example

`examples/git-status.json` — um step `git.status` com `workdir: "."`.

## Risks

| Risco | Mitigação |
|---|---|
| `git` ausente no PATH | `runtime.player_error` claro; testes skip ou require git |
| Hooks maliciosos no repo | `core.hooksPath=/dev/null` no commit |
| Escape via symlink | EvalSymlinks + prefix check (Shell) |
| Scope creep (push/HITL) | G-74 exclusões explícitas |

## Packages touched

- `internal/players/git` (novo)
- `internal/core/api` (register)
- `examples/git-status.json`
- testes no pacote git (+ api smoke opcional)
