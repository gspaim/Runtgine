# Design: 042-helm-player

## Pacote

| Player name | Package | Binary |
|---|---|---|
| `helm` | `internal/players/helm` | `helm` |

Expõe `SetRunner`/`SetExec` para testes. `defaultRunCmd` usa
`exec.LookPath` + `CommandContext` argv — nunca `sh -c`.

## Capabilities

- `helm.lint`: `helm lint <chart-path>`. Saída `ok`/`exit_code`/`log`.
- `helm.template`: `helm template [release] <chart-path> [-n ns]`.
  Saída `output` (YAML truncado 1 MiB) + `truncated`.
- `helm.list`: `helm list [-n ns] -o json`. Saída `releases` (JSON
  parseado; senão string bruta) + `truncated`.
- `helm.status`: `helm status <release> [-n ns]`. Saída
  `ok`/`exit_code`/`log`.

Timeouts: `lint`/`template` default 120s teto 600s; `list`/`status`
default 30s teto 120s.

## Chart / workdir

`chart` resolvido com `shell.ResolveWorkdir` (padrão do `tf` em
`41`). Marker: `Chart.yaml` no diretório (não recursivo no v0).
Fora do workspace → `validation.invalid_input`.

## Sandbox

Argv por capability, sem flags livres. Negados no input (JSON Schema
`additionalProperties: false` + `ValidateStaticInput`): `values`,
`set`, `repo`, `version`, `kubeconfig`. Flags de install/upgrade/
rollback/uninstall/get/test não existem como capabilities.

`helm.list` / `helm.status` herdam kubeconfig do processo (modelo
`kubectl`); sem override no input.

## Intent

Parsers estreitos antes de `matchShell`: `helm lint <path>` →
`helm.lint`; `helm template <path>` → `helm.template`; `helm list` →
`helm.list`; `helm status <release>` → `helm.status` (method
`heuristic.helm`). `helm install` / `upgrade` não casam.
