# Design: 031-mvp-1.0

## Technical approach

This change is a **product cut**. Code comes in two slices after merge.

### Slice 19 — Intent player heuristics (G-135..G-136)

In `internal/core/intent`, run **before** `matchShell`:

```text
empty → player phrases → shell prefixes → pipeline keywords → LLM
```

Phrase table is exact-enough contains / prefix (PT+EN, case-insensitive).
`go test` must win over prefix `go `. `git status|diff|log` must win
over argv `git` → `shell.exec`.

New `CompileResult.Method` values:

- `heuristic.test`
- `heuristic.git`
- `heuristic.docker`

LLM `route` enum stays `shell|pipeline`. Do not emit `test.go` from
the Completer in this slice (Registry would accept it, but the cut is
heuristic-only).

Task IR inputs use Player defaults (`packages` omitted → `./...`;
git `workdir` `.`). No HITL change (those caps are allow).

Tests: table-driven Compile; `go test` must not be `shell.exec`.

### Slice 20 — Context Engine seed (G-137..G-139)

Keep package `internal/core/contextpack`. Add something like
`WithSeededRepoHits(p Pack, hits []GraphHit) Pack` used when
`len(paths)==0 && len(symbols)==0`.

Runner already calls QueryHits for LLM steps (`attachGraphHits`).
Reuse that result: filter `kind=path|symbol` into `repo_hits` if empty.
Intent LLM pack: same if repo_hits empty.

Do not `filepath.Walk`. Do not `os.ReadFile`. Graph error → empty.

Tests: fake hits seed paths; existing repo-search hits are preserved;
empty graph → empty repo_hits, no error.

### Blast / Claims / TUI

Unchanged. Heuristics emit existing caps. Seed does not execute Players.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Puxar G-45 no 1.0 | Superfície CI; não fecha o loop local |
| Context Engine lê arquivos | Vira `fs.read` escondido; budget de bytes novo |
| LLM route `test\|git\|docker` | Heurística já cobre; Completer inventaria caps |
| `git add/commit` via NL | Escrita; 1.0 magro é verificação / leitura |
| Pacote `contexteng` novo | Assemble já vive em `contextpack` |

## Risks

| Risco | Mitigação |
|---|---|
| `go test ./...` via Intent dispara suite | Default packages `./...` is the cap contract; user can pass Task IR with a tighter list. Phrase `go test` is explicit. |
| Semente Graph polui pack | Só quando `repo_hits` vazio; `max_files` |
| Falso positivo `run tests` em texto de review | Player heuristics before pipeline; if both match, **player wins** for the exact phrases in the table. Broader `review` still pipeline. |

## Packages touched (slices 19–20, not this PR)

- `internal/core/intent`
- `internal/core/contextpack` (+ runner / intent LLM pack)
