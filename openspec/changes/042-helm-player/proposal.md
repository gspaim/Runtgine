# Proposal: 042-helm-player

## Why

G-41 continua em andamento. Depois do recorte infra (K8s / Terraform /
PostgreSQL em `41`), o restante documentado é **Helm / cloud / SQL**.
Hoje `helm lint`, `helm template`, `helm list` e `helm status` caem
em `shell.exec`. Este recorte entrega lint / render local / leitura
de cluster com argv e policy — sem install, sem upgrade, sem valores
por `--set`.

A spec `41` (G-209) excluiu Helm do corte v0 explicitamente; este
change levanta a exclusão como recorte próprio de G-41.

## What Changes

- Canonical `docs/42-helm-player-v0.md` (G-210..G-216 CONFIRMED)
- Player `helm` (`internal/players/helm`)
- Capabilities `helm.lint` / `helm.template` / `helm.list` /
  `helm.status` — todas `allow`
- Chart no workspace com marker `Chart.yaml`
- Intent heuristic `heuristic.helm`
- Examples JSON

## What Does Not Change

- K8s / Terraform / Postgres Players (`41`); Docker (`23`); Shell
- Task IR schema; Claims / Blast tables
- Templates (`40`); MCP (`39`); Memory; HTTP API (`34`)
- install / upgrade / rollback / uninstall; repo / OCI management
- Kustomize / k3s / kind; cloud SDKs; SQL
- NATS (G-36 DEFERRED); TUI tabs; Wails views

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `042-helm-player` |
| Doc canônico | `docs/42-helm-player-v0.md` |
| Gaps | G-210..G-216 **CONFIRMED** (recorte de G-41) |
| Código | Slice 35 |

## Approach

1. Um pacote no padrão Docker/NPM/infra: Manifest +
   `ValidateStaticInput` + `ExecFunc` injetável.
2. `safeRef` para `release` / `namespace`; `chart` relativo no
   workspace com marker `Chart.yaml` (padrão do marker `*.tf`).
3. Subcomandos/flags fechados por argv; sem `--set*` / `--values` /
   `--repo` no v0.
4. Heurística de alta confiança antes de `matchShell`;
   `helm install` não casa.

## Impact

- `internal/players/helm`
- `api.Open`, `runner` admission, `intent`
- README Estágio: Slice 35
