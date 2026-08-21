# 17 — Intent Engine v0

Traduz intencao em linguagem natural para Task IR v0.
Inventario: `10-gaps.md` (G-50+).

**Status: CONFIRMED (v0)** — pos-MVP Core (slices 1–3+).

---

## Principios

1. Intent Engine **nao e Player** e **nao e autoridade**
2. Saida e sempre Task IR v0 → Validator → Runner (sem bypass)
3. Deterministic-first: heuristicas antes de LLM
4. Capabilities inventadas → rejeitadas pelo Registry/Validator
5. Sem Runtime Graph / Workflow Templates nas heuristicas v0; o caminho LLM
   passa a receber `graph_hits` apos o slice `19` (G-69)

---

## G-50 — Papel e fronteira

**Status: CONFIRMED**

| E | Nao e |
|---|---|
| Compilador NL → Task IR | Player / Agent |
| Entry-point helper no Core | Substituto do Validator |
| Superficie CLI (`runtgine intent`) | Autoridade de capability |

Fluxo:

```text
NL text -> Intent Engine -> Task IR v0 -> SubmitTask -> Validator -> Runner
```

---

## G-51 — API v0

**Status: CONFIRMED**

```text
CompileIntent(text, source) -> (TaskIR, method)
SubmitIntent(text, source)  -> run_id   // CompileIntent + SubmitTask
```

- `method`: `heuristic.shell` | `heuristic.pipeline` | `heuristic.test` |
  `heuristic.git` | `heuristic.docker` | `llm`
- Entry point tipico: `cli` (ref opcional)
- Pacote: `internal/core/intent`

---

## G-52 — Heuristicas deterministicas

**Status: CONFIRMED**

Ordem:

1. Texto vazio → erro de validacao
2. Padrao **Player** (G-135; MVP 1.0 magro) — ver tabela abaixo
3. Padrao shell (prefixos `run `, `exec `, `echo `, `go `, `make `, `npm `, `./`, `$ ` ou argv simples) → Task com um step `shell.exec`
4. Padrao de analise (`review`, `revisa`, `analisa`, `decompose`, `pipeline`, `estima`, `board`) → template `pipeline` linear (`12`)
5. Caso contrario → caminho LLM (G-53)

Shell argv: split whitespace simples (sem shell string / sem `sh -c`).
`go test` **nao** cai no prefixo `go ` — Player heuristic ganha.

---

## G-135 — Heuristicas de Player (MVP 1.0 magro)

**Status: CONFIRMED** — codigo = slice 19 (feito).

Antes de `matchShell`. Frases de alta confianca (PT/EN, case-insensitive):

| NL | Capability | Method |
|---|---|---|
| `go test`, `roda os testes`, `rodar testes`, `run tests` | `test.go` | `heuristic.test` |
| `git status` | `git.status` | `heuristic.git` |
| `git diff` | `git.diff` | `heuristic.git` |
| `git log` | `git.log` | `heuristic.git` |
| `docker ps` | `docker.ps` | `heuristic.docker` |

Fora desta tabela no v0: `git add/commit/push`, `http.get`, `fs.*`,
`docker run/build`, pytest/npm. Inputs = defaults do Manifest.

## G-136 — Metodos e soberania

**Status: CONFIRMED** — codigo = slice 19 (feito).

- `method` passa a incluir `heuristic.test` | `heuristic.git` |
  `heuristic.docker` alem de `heuristic.shell` | `heuristic.pipeline` | `llm`
- Caminho LLM (G-53) **nao** ganha `route` novo; continua `shell|pipeline`
- Registry/Validator rejeitam capability inventada (inalterado)

---

## G-53 — Caminho LLM

**Status: CONFIRMED**

- Reusa `Completer` (OpenAI-compat / Anthropic / heuristic offline)
- LLM emite JSON intermediario (nao executa):

```json
{
  "summary": "…",
  "notes": "…",
  "route": "shell" | "pipeline",
  "shell_command": ["echo", "hi"]
}
```

- Core monta Task IR a partir de `route` (shell.exec ou pipeline template)
- Sem chave LLM: `HeuristicCompleter` escolhe pipeline se houver palavras de analise; senao shell `echo` do resumo
- Retry 1x se JSON invalido (igual G-25)

---

## G-54 — CLI

**Status: CONFIRMED**

```text
runtgine intent "<nl>" [--dry-run] [--wait]
```

| Flag | Default | Efeito |
|---|---|---|
| `--dry-run` | false | Imprime Task IR JSON; nao submete |
| `--wait` | true | Aguarda terminal do run (como `run`) |

---

## Superficie visual (Intent Surface v0)

**Status: CONFIRMED** — ver [32-intent-surface-v0.md](32-intent-surface-v0.md).

- Aba **INTENT** na TUI (slice 21) e view equivalente no Wails (spec `35`)
- Reusa `CompileIntent` / `SubmitIntent`; source `tui` | `wails`
- Nao e chatbot; preview Task IR antes de submit

## Fora do v0 (Intent Engine)

- Workflow Templates / SDD auto-sizing
- Multi-step planning rico alem de shell|pipeline
- Intent Engine como Player (`intent.*`)
- Consulta Graph nas heuristicas (so caminho LLM; ver `19`)
- Heuristicas `git.add` / `commit` / `push`, `http.get`, `fs.*` (1.0 magro)

Nota: Graph Hits no Completer LLM e ContextPack e o slice
[19-graph-hits-v0.md](19-graph-hits-v0.md) (G-66..G-69), nao deste doc.

---

## Criterio de pronto

- `runtgine intent "echo hello-intent" --dry-run` emite Task IR com `shell.exec`
- `runtgine intent "revisar a arquitetura"` roteia para pipeline (heuristic ou LLM)
- Task gerada passa pelo mesmo Validator/`SubmitTask` que JSON manual
- Capability inventada pelo LLM e rejeitada antes de executar
