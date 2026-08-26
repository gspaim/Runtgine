# Proposal: 041-infra-players

## Why

G-41 continua em andamento. Depois de Git/FS/Docker/HTTP/Test/NPM/
pytest/yarn, o restante documentado é **infra: K8s / Terraform /
PostgreSQL**. Hoje `kubectl`, `terraform` e `psql` caem em
`shell.exec`. Este recorte entrega leitura / validate+plan / ping
com argv e policy — sem apply, sem SQL livre.

## What Changes

- Canonical `docs/41-infra-players-v0.md` (G-201..G-209 CONFIRMED)
- Players `k8s`, `terraform`, `postgres`
- Capabilities `k8s.get`/`k8s.list`, `tf.validate`/`tf.plan`, `pg.ping`
- `tf.plan` Manifest `approval-required`
- Intent heuristics `heuristic.k8s` / `heuristic.tf` / `heuristic.pg`
- Examples JSON

## What Does Not Change

- Docker Player (`23`); Shell; HTTP Player
- Task IR schema; Claims / Blast tables
- Templates (`40`); MCP (`39`); Memory
- Helm / apply / destroy / init / SQL
- NATS; TUI tabs; Wails views

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `041-infra-players` |
| Doc canônico | `docs/41-infra-players-v0.md` |
| Gaps | G-201..G-209 **CONFIRMED** (recorte de G-41) |
| Código | Slice 34 |

## Approach

1. Três pacotes no padrão Docker/NPM: Manifest + `ValidateStaticInput`
   + `ExecFunc` injetável.
2. `safeRef` para nomes; workdir TF exige `*.tf`.
3. Password PG só env `PGPASSWORD`, nunca no Task IR.
4. Heurísticas de alta confiança antes de `matchShell`.

## Impact

- `internal/players/k8s`, `tf`, `pg`
- `api.Open`, `runner` admission, `intent`
- README Estágio: Slice 34
