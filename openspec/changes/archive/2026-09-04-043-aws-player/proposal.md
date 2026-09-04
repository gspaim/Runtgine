# Proposal: 043-aws-player

## Why

G-41 continua em andamento. Depois do recorte Helm (`42`), o restante
documentado é **cloud / SQL**. Hoje a verificação de infra em nuvem
(`aws sts get-caller-identity`, `aws s3 ls`) depende de `shell.exec`.
Este recorte entrega a parte **cloud read-only** — "quem sou eu e o
que existe" — com argv fechado e JSON em todas as saídas. SQL
arbitrário e migrations continuam excluídos (G-206/G-209/G-216); a
exclusão de cloud SDKs de `41`/`42` (G-209/G-216) é levantada **só
para AWS em leitura** — GCP/Azure permanecem fora.

## What Changes

- Canonical `docs/43-aws-player-v0.md` (G-217..G-223 CONFIRMED)
- Player `aws` (`internal/players/aws`)
- Capabilities `aws.sts-identity`, `aws.s3-buckets`, `aws.s3-objects`
  — todas `allow`, saída `--output json`
- Credenciais exclusivamente via ambiente herdado; nunca no Task IR
- Intent heuristic `heuristic.aws` (`aws sts get-caller-identity`,
  `aws s3 ls [s3://...]`); mutantes (`s3 rm/cp`) não casam
- Examples JSON

## What Does Not Change

- Helm (`42`); K8s / Terraform / Postgres (`41`); Docker (`23`);
  Shell; HTTP Player
- Task IR schema; Claims / Blast tables
- Templates (`40`); MCP (`39`); Memory; HTTP API (`34`)
- Subcomandos mutantes (cp/mv/rm/sync/create/put/delete/tag)
- GCP / Azure; SQL arbitrário; migrations; NATS (G-36 DEFERRED)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `043-aws-player` |
| Doc canônico | `docs/43-aws-player-v0.md` |
| Gaps | G-217..G-223 **CONFIRMED** (recorte de G-41) |
| Código | Slice 36 |

## Approach

1. Um pacote no padrão Docker/NPM/infra/helm: Manifest +
   `ValidateStaticInput` + `ExecFunc` injetável.
2. Saída sempre `--no-cli-pager --output json`; parse JSON →
   `object` (string bruta + `truncated` se falhar, padrão `k8s`).
3. `safeRef` para `bucket`/`prefix`/`region`; credenciais só env
   herdado (modelo `PGPASSWORD`).
4. Heurística de alta confiança antes de `matchShell`; parse
   estático de `s3://bucket/prefix`; mutantes não casam.

## Impact

- `internal/players/aws`
- `api.Open`, `runner` admission, `intent`
- README Estágio: Slice 36
