# 21 — Filesystem Player v0

Player determinístico para operações locais de filesystem, com confinamento
ao workspace e limites explícitos de leitura/listagem.

Inventário: [10-gaps.md](10-gaps.md) (G-75+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Core + Shell + Git Player estáveis (slices 1–8).

**Status deste doc: CONFIRMED (v0).** G-75..G-80 autorizam o slice 9
de código. HITL / Approvals e Execution Policy completa permanecem fora
(G-42 HYPOTHESIS / P3).

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-17-021-filesystem-player/`](../openspec/changes/archive/2026-08-17-021-filesystem-player/).
Deltas mergeados em `openspec/specs/filesystem-player/`. Branch de implementação:
`cursor/021-filesystem-player-0ac1`.

---

## 1. Problema

O Runtgine já executa Shell e Git, mas não possui uma capability
determinística para ler, escrever, listar e inspecionar arquivos. Usar
`shell.exec` para isso perde contratos de input/output, dificulta limites
de tamanho e amplia a superfície de execução.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Player `filesystem` determinístico | Agent / LLM |
| Capabilities `fs.*` no Registry | Substituto do Git ou Shell |
| APIs Go de filesystem, sem shell | `rm`, `mv`, `chmod` ou execução |
| Paths relativos confinados ao workspace | Acesso a paths absolutos arbitrários |
| Texto UTF-8 no v0 | Upload, rede ou armazenamento externo |

Regras:

1. Validator / Registry continuam soberanos.
2. Todo path deve permanecer dentro do `workspace_root` após resolução.
3. Symlink que escape o workspace é rejeitado, inclusive em componentes
   intermediários.
4. Não seguir symlink no destino de `fs.write`; substituir arquivo apontado
   por symlink é proibido.
5. Limites de bytes e entradas são obrigatórios para evitar leitura/listagem
   sem limite.
6. Não apagar, mover, alterar permissões ou executar arquivos neste slice.

---

## 3. Cortes confirmados (G-75+)

### G-75 — Papel e pacote

**Status: CONFIRMED**

- Nome do Player: `filesystem`
- Pacote: `internal/players/filesystem`
- Kind: `deterministic`
- Registro em `api.Open` junto de Shell e Git
- Recorte de G-41: apenas operações locais seguras de arquivo

### G-76 — Capabilities v0

**Status: CONFIRMED**

| Capability | Entrada | Saída |
|---|---|---|
| `fs.read` | `path` obrigatório, `max_bytes?` (default 1 MiB, max 4 MiB) | `path`, `content`, `bytes`, `truncated` |
| `fs.write` | `path`, `content` obrigatórios, `create_parents?` (default false) | `path`, `bytes`, `created` |
| `fs.list` | `path?` (default `.`), `recursive?` (default false), `max_entries?` (default 200, max 1000) | `path`, `entries[]`, `truncated` |
| `fs.stat` | `path` obrigatório | `path`, `type`, `size`, `mode`, `modified_at` |

Schemas JSON formais vivem no Manifest e usam
`additionalProperties: false`.

`fs.read` aceita somente texto UTF-8 válido; bytes inválidos retornam
`validation.invalid_input`. `fs.write` recebe texto UTF-8 e não cria
parents por padrão.

`fs.list` retorna entries ordenadas lexicograficamente por path relativo.
Tipos v0: `file`, `directory`, `symlink`. Symlink que escape o workspace
é erro, não entry.

### G-77 — Confinamento de paths

**Status: CONFIRMED**

- `path` é relativo ao workspace root; path absoluto é rejeitado.
- `..` só é aceito quando o resultado final ainda está dentro do workspace,
  mas a implementação pode rejeitar qualquer escape lexical antecipadamente.
- Componentes existentes são resolvidos com `EvalSymlinks`.
- Para `fs.write`, parents existentes são resolvidos e o destino existente
  não pode ser symlink.
- O root do workspace pode ser usado como `path="."` para `fs.list` e
  `fs.stat`, mas não como destino de `fs.write`.

Falha de confinement retorna `validation.invalid_input` antes de I/O.

### G-78 — Limites e escrita

**Status: CONFIRMED**

| Regra | Corte v0 |
|---|---|
| Read | default 1 MiB; máximo 4 MiB; `truncated=true` quando aplicável |
| Write | máximo 4 MiB; falha acima do limite antes de alterar arquivo |
| List | default 200 entries; máximo 1000 |
| Recursive list | permitida, mas limitada por `max_entries` |
| Write parents | `create_parents=false` por default |
| Atomicidade | escrever arquivo temporário no mesmo parent e `rename`; sem partial write |
| Permissões | não altera permissões de arquivo existente; modo default do OS em novo arquivo |

Não há limite de timeout separado: as operações são locais e limitadas por
bytes/entries. O contexto do Runner ainda pode cancelar a execução.

### G-79 — Integração Core

**Status: CONFIRMED**

1. `api.Open` registra `filesystem.New()`.
2. Validator valida `input` contra `input_schema` do Manifest.
3. Runner chama `filesystem.ValidateStaticInput` na admissão para paths,
   limits e destino de escrita.
4. Runtime Graph registra os nós `capability` `fs.*` automaticamente.
5. Exemplo: `examples/fs-read.json`.

### G-80 — Exclusões v0

**Status: CONFIRMED**

- `fs.delete`, `fs.remove`, `fs.move`, `fs.copy`
- `fs.chmod`, `fs.symlink`
- execução de conteúdo ou shell fallback
- paths fora do workspace
- leitura binária / upload / rede
- HITL, Claims, Blast Radius e Execution Policy genérica
- TUI dedicada e heurísticas de Intent específicas

---

## 4. Critérios de aceite

1. `fs.read` lê arquivo UTF-8 dentro do workspace e respeita `max_bytes`.
2. `fs.write` cria/sobrescreve arquivo dentro do workspace sem partial write.
3. `fs.list` ordena entries e respeita `max_entries`, inclusive recursivo.
4. `fs.stat` retorna tipo, tamanho, modo e timestamp do arquivo.
5. Paths com escape ou symlink externo são rejeitados antes do I/O.
6. Capability inventada `fs.delete` não está no Manifest e é rejeitada.
7. `go test ./internal/players/filesystem/...` cobre read/write/list/stat,
   limits, UTF-8, symlink escape e atomicidade.
8. `go test ./...` e `go vet ./...` verdes.
9. OpenSpec `021-filesystem-player` é arquivado após merge do código.

---

## 5. Ordem do slice de código

1. G-75..G-80 CONFIRMED — spec em `21` + OpenSpec
2. Pacote `internal/players/filesystem` + Manifest + testes — feito
3. Registrar no Core, adicionar static validation e exemplo — feito
4. Atualizar estágio README (Slice 9 Feito) e arquivar OpenSpec — este PR

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-75 Papel / pacote `filesystem`
- [x] G-76 Capabilities read/write/list/stat
- [x] G-77 Confinamento e symlink policy
- [x] G-78 Limites, UTF-8 e escrita atômica
- [x] G-79 Wire Registry + static validation + exemplo
- [x] G-80 Exclusões (delete/move/chmod/rede/HITL)
