# Proposal: 044-cloud-gcp-azure

## Why

G-41 continua em andamento. Depois do recorte AWS (`43`), o restante
documentado é **cloud GCP/Azure** e **SQL**. Hoje a verificação de
identidade/config/projetos em GCP e Azure (`gcloud auth list`,
`az account show`) depende de `shell.exec`. Este recorte fecha a
parte **cloud read-only** com dois players de argv fechado e JSON.
SQL arbitrário e migrations continuam excluídos (G-206/G-209/G-216)
e exigem spec própria com desenho de segurança dedicado.

## What Changes

- Canonical `docs/44-cloud-gcp-azure-players-v0.md`
  (G-224..G-230 CONFIRMED)
- Player `gcp` (`internal/players/gcp`, binário `gcloud`):
  `gcp.identity` / `gcp.config` / `gcp.projects`
- Player `azure` (`internal/players/azure`, binário `az`):
  `azure.identity` / `azure.subscriptions` / `azure.groups`
- Todas `allow`, saída `--format=json` / `-o json`; sem input além
  de `timeout_ms`; credenciais via ambiente/config herdado
- Intent heuristics `heuristic.gcp` / `heuristic.az`; mutantes não
  casam
- Examples JSON

## What Does Not Change

- AWS Player (`43`); Helm (`42`); infra (`41`); Docker (`23`);
  Shell; HTTP Player
- Task IR schema; Claims / Blast tables
- Templates (`40`); MCP (`39`); Memory; HTTP API (`34`)
- Subcomandos mutantes (create/delete/set/IAM/roles/deploy)
- SQL arbitrário; migrations; NATS (G-36 DEFERRED)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `044-cloud-gcp-azure` |
| Doc canônico | `docs/44-cloud-gcp-azure-players-v0.md` |
| Gaps | G-224..G-230 **CONFIRMED** (recorte de G-41) |
| Código | Slice 37 |

## Approach

1. Dois pacotes no padrão infra/AWS: Manifest + `ValidateStaticInput`
   + `ExecFunc` injetável.
2. Saída sempre JSON (`--format=json` no gcloud, `-o json` no az);
   parse JSON → `object` (bruto + `truncated` se falhar).
3. Zero campos de input além de `timeout_ms` — projeto/subscription
   e credenciais vivem no ambiente/config herdado (nunca no IR).
4. Heurísticas de alta confiança antes de `matchShell`; mutantes
   (`projects create`, `group delete`) não casam.

## Impact

- `internal/players/gcp`, `internal/players/azure`
- `api.Open`, `runner` admission, `intent`
- README Estágio: Slice 37
