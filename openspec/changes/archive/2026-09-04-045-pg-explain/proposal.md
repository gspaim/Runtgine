# Proposal: 045-pg-explain

## Why

Último recorte nomeado de G-41: **SQL**. "Intenção → execução
verificável" não tem verificação de consulta — não há como pedir o
**plano** de um SELECT pelo runtime. Este recorte levanta a exclusão
"sem SQL no Task IR" (G-206/G-209/G-216) **somente para EXPLAIN sem
ANALYZE** no Player `postgres` existente, com validação estática
conservadora (prefixo `SELECT`/`WITH`, sem `;`, sem `\`) que fecha
os vetores conhecidos (meta-comandos psql, multi-statement).
Executar queries (linhas) e migrations continuam fora.

## What Changes

- Canonical `docs/45-pg-explain-v0.md` (G-231..G-237 CONFIRMED)
- Capability `pg.explain` no Player `postgres`
  (`internal/players/pg`): `EXPLAIN (FORMAT JSON) <SELECT/WITH>`
- Validação estática da SQL no `ValidateStaticInput`
- Runner admission + Intent heuristic (`explain select|with` →
  `heuristic.pg`)
- Example `pg-explain.json`

## What Does Not Change

- `pg.ping` e o resto do Player `postgres` (`41`)
- Outros Players (AWS/GCP/Azure/Helm/infra/Docker/Shell…)
- Task IR schema; Claims / Blast tables
- Templates (`40`); MCP (`39`); Memory; HTTP API (`34`)
- `EXPLAIN ANALYZE`, SELECT de linhas, migrations, dumps, outros
  SGBDs
- NATS (G-36 DEFERRED)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `045-pg-explain` |
| Doc canônico | `docs/45-pg-explain-v0.md` |
| Gaps | G-231..G-237 **CONFIRMED** (recorte de G-41) |
| Código | Slice 38 |

## Approach

1. Um pacote extendido (`pg`), padrão existente: Manifest +
   `ValidateStaticInput` + `ExecFunc` injetável (com env).
2. Player constrói o prefixo `EXPLAIN (FORMAT JSON) `; SQL do IR
   nunca é o comando inteiro.
3. Validador conservador: 1ª palavra `SELECT`/`WITH`, sem `;`, sem
   `\`, ≤ 10 KiB — falso-positivo aceito, falso-negativo fechado.
4. Credenciais: mesma allowlist de env de `41` (`PGPASSWORD`,
   `PGSSLMODE`).

## Impact

- `internal/players/pg`
- `runner` admission, `intent`
- README Estágio: Slice 38
