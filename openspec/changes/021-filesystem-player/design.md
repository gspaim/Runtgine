# Design: 021-filesystem-player

## Technical approach

### Package

`internal/players/filesystem` com `New() *Player`, `Manifest()`,
`ValidateStaticInput` e `Execute`.

O Player usa `os.Open`, `os.ReadFile`, `os.OpenFile`, `os.ReadDir`,
`os.Stat` e `os.Rename`. Não chama `shell.exec` nem `exec.Command`.

### Path resolution

```text
workspace root
  -> reject absolute input
  -> lexical join + clean
  -> resolve existing components with EvalSymlinks
  -> verify relative path to workspace is not ".."
  -> reject symlink write destination
```

Para arquivos novos, o parent precisa existir ou ser criado somente quando
`create_parents=true`; cada parent é resolvido antes de mkdir/rename.

### Capability contracts

| Capability | Behavior |
|---|---|
| `fs.read` | abre texto UTF-8, lê até `max_bytes`, marca truncation |
| `fs.write` | valida UTF-8/size, escreve temp no mesmo parent e renomeia |
| `fs.list` | walk determinístico, ordena paths, para em `max_entries` |
| `fs.stat` | retorna file/directory/symlink, size, mode e mod time |

`fs.stat` usa `Lstat` para identificar symlink, mas rejeita symlink cujo
target escape o workspace. `fs.read`/`fs.list` não atravessam symlink externo.

### Static validation

`ValidateStaticInput(workspace, capability, raw)` valida schema-adjacent
limits, path lexical/confinement, `content` UTF-8 e destino write antes do
Runner admitir a Task. `Execute` repete as checks como boundary defense.

### Registration

`api.Open` registra `filesystem.New()` após Shell e Git. O Runner dispatcha
static validation por capability, como já faz para Shell/Git.

### Example

`examples/fs-read.json` executa `fs.read` em um arquivo versionado do
workspace, com `max_bytes` explícito.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Reusar Shell com `cat`/`find` | Perde schema, limites e confinement próprio |
| Permitir paths absolutos | Aumenta blast radius antes de Execution Policy |
| Seguir symlinks externos | Permite exfiltração fora do workspace |
| Implementar delete/move agora | Exige policy/claims e é fora de G-80 |

## Risks

| Risco | Mitigação |
|---|---|
| TOCTOU entre validação e I/O | revalidar no Execute; operações locais best-effort v0 |
| Arquivo muda durante read | limite de bytes + resultado explicitamente truncável |
| Atomic rename em Windows | testar OS alvo; fallback não deve produzir partial write |
| Caminho symlink | EvalSymlinks dos parents e rejeição de target externo |

## Packages touched

- `internal/players/filesystem` (novo)
- `internal/core/api` (register)
- `internal/core/runner` (static validation)
- `examples/fs-read.json`
