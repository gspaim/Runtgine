# Design: 043-aws-player

## Pacote

| Player name | Package | Binary |
|---|---|---|
| `aws` | `internal/players/aws` | `aws` |

Expõe `SetRunner`/`SetExec` para testes. `defaultRunCmd` usa
`exec.LookPath` + `CommandContext` argv — nunca `sh -c`. Dir de
execução: workspace root (sem dependência de cwd).

## Capabilities

- `aws.sts-identity`: `aws sts get-caller-identity --no-cli-pager
  --output json`. Saída `object` (JSON parseado) + `truncated`.
- `aws.s3-buckets`: `aws s3api list-buckets --no-cli-pager --output
  json`. Mesma saída.
- `aws.s3-objects`: `aws s3api list-objects-v2 --bucket <b>
  [--prefix <p>] --no-cli-pager --output json`. Mesma saída.

Timeout default 30s, teto 120s. Truncamento em 1 MiB; JSON
não-parseável → string truncada em 4096 runes + `truncated: true`
(padrao `k8s`). Paginação é do AWS CLI v2 (auto-paginação).

## Sandbox

Argv por capability, sem flags livres. Negados no input (JSON Schema
`additionalProperties: false` + `ValidateStaticInput`): `query`,
`endpoint`, `profile`, `token`, `access_key`, `secret`. Flags globais
perigosas ausentes por construção do argv.

Credenciais: ambiente herdado do processo (`~/.aws`, `AWS_*`); nunca
no Task IR (modelo `PGPASSWORD` do `pg`). `region?` no input vira
`--region <v>` (`safeRef`).

## Intent

Parsers estreitos antes de `matchShell`:
- `aws sts get-caller-identity` → `aws.sts-identity`
- `aws s3 ls` (exato) → `aws.s3-buckets`
- `aws s3 ls s3://bucket[/prefix]` → `aws.s3-objects` (parse estático
  da URI: bucket até `/`, resto é prefix; `safeRef` em ambos)

Mutantes (`s3 rm|cp|mv|sync`) e flags após o subcomando não casam —
caem no shell/LLM (method `heuristic.aws` só nos casos da tabela).
