# Design: 041-infra-players

## Pacotes

| Player name | Package | Binary |
|---|---|---|
| `k8s` | `internal/players/k8s` | `kubectl` |
| `terraform` | `internal/players/tf` | `terraform` |
| `postgres` | `internal/players/pg` | `psql` |

Cada um expõe `SetRunner`/`SetExec` para testes. `defaultRunCmd`
usa `exec.LookPath` + `CommandContext` argv — nunca `sh -c`.

## K8s

`kubectl get RESOURCE [NAME] [-n NS] -o json`. Output: objeto JSON
parseado se possível; senão string em `object` + `truncated`.

## Terraform

Workdir resolvido com `shell.ResolveWorkdir`. Marker: pelo menos
um `*.tf` ou `*.tf.json` no diretório (não recursivo no v0).
`tf.plan` declara `ExecutionPolicy: approval-required` no Manifest
(G-82).

## Postgres

`psql` com flags fechadas. Env: herança mínima do Shell **mais**
`PGPASSWORD` e `PGSSLMODE`. Input sem `password`/`sql`.

## Intent

Parsers estreitos: primeiro token após `kubectl get` / `k8s get`
é `resource`; segundo, se presente, `name` → `k8s.get`.
`terraform validate|plan`. `pg ping` / `postgres ping` / `psql ping`
→ `pg.ping` sem host (operador preenche via Task IR / template).
