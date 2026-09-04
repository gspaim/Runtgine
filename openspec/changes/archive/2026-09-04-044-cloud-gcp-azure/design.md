# Design: 044-cloud-gcp-azure

## Pacotes

| Player name | Package | Binary |
|---|---|---|
| `gcp` | `internal/players/gcp` | `gcloud` |
| `azure` | `internal/players/azure` | `az` |

Cada um expõe `SetRunner`/`SetExec` para testes. `defaultRunCmd` usa
`exec.LookPath` + `CommandContext` argv — nunca `sh -c`. Dir de
execução: workspace root (sem dependência de cwd).

## Capabilities

- `gcp.identity`: `gcloud auth list --format=json` (local)
- `gcp.config`: `gcloud config list --format=json` (local)
- `gcp.projects`: `gcloud projects list --format=json` (rede, leitura)
- `azure.identity`: `az account show -o json`
- `azure.subscriptions`: `az account list -o json`
- `azure.groups`: `az group list -o json`

Timeout default 30s, teto 120s. Truncamento em 1 MiB; JSON
não-parseável → string truncada em 4096 runes + `truncated: true`
(padrao `k8s`/`aws`).

## Sandbox

Input único aceito: `timeout_ms` (`additionalProperties: false`).
Nenhum campo de projeto/subscription/credentials no Task IR —
ambiente/config herdado (`GOOGLE_APPLICATION_CREDENTIALS`,
`CLOUDSDK_CORE_PROJECT`, `AZURE_CONFIG_DIR`, `az login`; modelo
`AWS_*` de `43`). Flags de formato fixas no argv; sem
`--filter`/`--query`/`--project`/`--subscription`.

## Intent

Parsers estreitos antes de `matchShell` (prefixo exato, nada
depois): `gcloud auth list` → `gcp.identity`; `gcloud config list`
→ `gcp.config`; `gcloud projects list` → `gcp.projects`;
`az account show` → `azure.identity`; `az account list` →
`azure.subscriptions`; `az group list` → `azure.groups`. Mutantes
(`projects create`, `group delete`) não casam.
