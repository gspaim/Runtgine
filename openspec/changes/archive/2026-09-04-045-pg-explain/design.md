# Design: 045-pg-explain

## Pacote

| Player name | Package | Binary |
|---|---|---|
| `postgres` | `internal/players/pg` (extendido) | `psql` |

`runCmdFunc` mantém assinatura `(ctx, timeout, env, args)` — env
allowlist herdado de `41` (`PATH`, `HOME`, `PGPASSWORD`,
`PGSSLMODE`, …). `SetRunner` para testes fake.

## Capability `pg.explain`

- Input: `sql` (obrigatório, ≤ 10 KiB, ver validação), `dbname`
  (obrigatório), `host?`, `port?`, `user?`, `timeout_ms?`
  (default 10s, teto 60s — limites do player). `additionalProperties:
  false`.
- Output: `plan` (JSON parseado), `truncated`.
- Argv: flags de conexão do `pg.ping` + `--command "EXPLAIN (FORMAT
  JSON) <sql>"`; `-t -A --pset pager=off` para saída limpa.

## Validação estática (G-232)

Ordem: trim não vazio → ≤ 10 KiB → 1ª palavra (case-insensitive,
via `strings.Fields`) é `select`/`with` → sem `;` em qualquer
posição → sem `\` em qualquer posição.

Justificativa: psql `-c` reconhece meta-comandos (`\!`) só no início
de um buffer de comando; com prefixo `SELECT`/`WITH` e nenhum `;`
existe um único buffer que não começa com `\`. Argv nunca shell.
EXPLAIN sem ANALYZE é planner puro (sem locks de dados, sem leitura).

## Intent

`explain select <rest>` / `explain with <rest>` (texto normalizado)
→ `pg.explain` com `sql` = rest e `dbname` default `postgres`
(mesmo padrão do `pg.ping`). Outros prefixos não casam.
