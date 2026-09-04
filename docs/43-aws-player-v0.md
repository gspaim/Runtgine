# 43 — Cloud Player v0 (AWS read-only)

Player determinístico de **AWS CLI read-only**, recorte de G-41
depois de Git / FS / Docker / HTTP / Test / NPM / pytest / yarn /
infra (K8s, Terraform, PostgreSQL) / Helm. Levanta a exclusão de
cloud SDKs do corte v0 de `41`/`42` (G-209/G-216) **só para AWS em
leitura** — GCP/Azure continuam fora:

| Player | Pacote | Capabilities v0 |
|---|---|---|
| `aws` | `internal/players/aws` | `aws.sts-identity`, `aws.s3-buckets`, `aws.s3-objects` |

Inventário: [10-gaps.md](10-gaps.md) (G-217+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é `s3 cp` / `mv` / `rm` / `sync` / `mb`. Não é `ec2 run-instances`
/ `terminate`. Não é criação/alteração de nenhum recurso. Não é SQL
arbitrário. Não é GCP (`gcloud`) nem Azure (`az`). Não é NATS (G-36).
Não é MCP. Não é template (`40`).

**Status deste doc: CONFIRMED v0 (slice 36 feito).** G-217..G-223.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-09-04-043-aws-player/`](../openspec/changes/archive/2026-09-04-043-aws-player/)
(arquivado após o slice 36). Spec atual: [`openspec/specs/aws-player/`](../openspec/specs/aws-player/).

---

## 1. Problema

A verificação de infra em nuvem ainda depende de `shell.exec` (argv
livre) ou do caminho LLM. Sem schema, sem allowlist, sem política de
credenciais. O restante explícito de G-41 após Helm era **cloud /
SQL**; este corte entrega a parte **cloud read-only** — a superfície
mínima que responde "quem sou eu e o que existe" sem criar ou
alterar nada. SQL arbitrário e migrations continuam fora (mesmo
corte de exclusão de `41`/`42`).

---

## 2. Fronteiras

| É | Não é |
|---|---|
| CLI argv (`aws`), subcomandos fechados | SDK / boto3 / client library |
| `sts get-caller-identity` + `s3api list-*` | cp/mv/rm/sync/mb/rb, run-instances, create/put/delete/tag |
| Saída JSON (`--output json`) | Texto livre / `--query` / JMESPath no v0 |
| Credenciais via ambiente herdado (`~/.aws`, `AWS_*`) | Credentials / tokens / keys no Task IR |
| Região como input (`safeRef`) → `--region` | `--endpoint-url` / proxy no input |
| Binário no PATH | Daemon próprio do Runtgine |

Regras:

1. Validator / Registry / Policy soberanos.
2. Nunca shell string. Só argv.
3. `bucket` / `prefix` / `region`: `safeRef` (sem espaço, sem
   prefixo `-`, espelha Docker `safeRef`).
4. Todas as capabilities `allow` (leitura pura, sem HITL no v0).
5. Unit tests injetam o runner; CI **não** exige credenciais AWS nem
   o binário `aws`.
6. Blast / Claims: nenhuma linha nova (como `http.get` / `helm.*`).

---

## 3. Cortes confirmados (G-217+)

### G-217 — AWS Player

- Nome: `aws`; pacote `internal/players/aws`; kind `deterministic`
- Binário `aws` no PATH (`exec.LookPath` + `CommandContext`, argv)
- Global `--no-cli-pager` e `--output json` sempre presentes

### G-218 — Capabilities v0

| Capability | Entrada | Saída | Policy |
|---|---|---|---|
| `aws.sts-identity` | `region?`, `timeout_ms?` | `object` (JSON), `truncated` | allow |
| `aws.s3-buckets` | `region?`, `timeout_ms?` | `object` (JSON), `truncated` | allow |
| `aws.s3-objects` | `bucket` (obrigatório), `prefix?`, `region?`, `timeout_ms?` | `object` (JSON), `truncated` | allow |

Argv:

- `aws sts get-caller-identity --no-cli-pager --output json`
- `aws s3api list-buckets --no-cli-pager --output json`
- `aws s3api list-objects-v2 --bucket <b> [--prefix <p>]
  --no-cli-pager --output json`

Timeout default 30s, teto 120s. Objeto JSON truncado em 1 MiB
(espelha `k8s` em `41`). Paginação é do CLI (auto-paginação do
AWS CLI v2); sem token no input.

### G-219 — Sandbox / argv

- Subcomandos e flags fechados: só os da tabela G-218
- `bucket` / `prefix` / `region`: `safeRef`; sem `--endpoint-url`,
  `--query`, `--profile`, `--cli-binary-format`, `--page-size`,
  `--starting-token` no input
- Input sem campo `query` / `endpoint` / `profile`
  (`additionalProperties: false`)
- Nenhum subcomando mutante registrado: `s3 cp|mv|rm|sync|mb|rb`,
  `ec2 run-instances|terminate-instances`, `create-*`, `put-*`,
  `delete-*`, `update-*`, `tag-*`

### G-220 — Credenciais no ambiente

- Nunca no Task IR: sem campo `access_key` / `secret` / `session`
  / `token` / `profile` no input
- Credenciais via ambiente herdado do processo: `~/.aws`,
  `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`,
  `AWS_SESSION_TOKEN`, `AWS_PROFILE`, `AWS_CONFIG_FILE` — mesmo
  modelo do `PGPASSWORD` em `41` (env herdado, nunca no IR)
- `region` explícito no input vira flag `--region <v>`; ausente →
  região default do ambiente (`AWS_REGION` / config)

### G-221 — Falha vs sucesso

- exit 0 → step ok (JSON parseado em `object`)
- exit ≠ 0 (inclusive credencial ausente/inválida, bucket
  inexistente) / binário ausente → `runtime.player_error` com a
  mensagem do stderr
- timeout → `runtime.timeout`
- JSON não-parseável → string bruta truncada em `object` +
  `truncated: true` (padrão `k8s`)
- Testes injetam runner; `go test ./...` sem credenciais nem `aws`

### G-222 — Registry + Graph + Intent

- `api.Open` registra o Player `aws`
- Graph: `RefreshFromRegistry` cria nó/aresta `provides`
- Intent (alta confiança, antes de `matchShell`):

| NL | Capability | Method |
|---|---|---|
| `aws sts get-caller-identity` | `aws.sts-identity` | `heuristic.aws` |
| `aws s3 ls` | `aws.s3-buckets` | `heuristic.aws` |
| `aws s3 ls s3://<bucket>[/prefix]` | `aws.s3-objects` | `heuristic.aws` |

`aws s3 rm` / `cp` / `sync` **não** casam — caem no shell/LLM (sem
capability). Parse de `s3://bucket/prefix` é estático (sem rede).

### G-223 — Exclusões v0

- Todo subcomando mutante (cp/mv/rm/sync/mb/rb, create/put/delete/
  update/tag, run-instances/terminate)
- `--query` / JMESPath, `--endpoint-url`, `--profile` no input,
  paginação manual (`--starting-token` / `--page-size`)
- GCP (`gcloud`) e Azure (`az`) como Players
- SQL arbitrário / migrations (continuam exclusão de `41`/`42`)
- NATS (G-36, DEFERRED); MCP; templates (`40`)

---

## 4. Critérios de aceite

1. Manifest **não** registra `aws.s3-cp`, `aws.s3-rm`,
   `aws.ec2-run`.
2. `bucket` / `prefix` / `region` começando com `-` ou com espaço →
   `validation.invalid_input`.
3. `s3-objects` sem `bucket` → `validation.invalid_input`.
4. Argv de todas as capabilities termina com `--output json` (e
   contém `--no-cli-pager`).
5. `runtgine intent --dry-run "aws s3 ls"` → `aws.s3-buckets`,
   method `heuristic.aws`.
6. `runtgine intent --dry-run "aws s3 rm s3://b/x"` → **não** é
   `heuristic.aws`.
7. Fake exec exit 1 → `runtime.player_error`.
8. `go test ./...` / `go vet ./...` verdes sem credenciais AWS.
9. OpenSpec `043` arquivado após o código (slice 36).

---

## 5. Ordem do slice de código

1. Pacote `aws` + Manifest + `ValidateStaticInput`
2. Registrar em `api.Open` + runner admission
3. Heurística Intent `heuristic.aws`
4. Examples `examples/aws-sts-identity.json`, `aws-s3-buckets.json`,
   `aws-s3-objects.json`
5. Testes fake; README Estágio: Slice 36
6. Arquivar OpenSpec `043`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-217 AWS Player (`internal/players/aws`)
- [x] G-218 Capabilities v0 (sts-identity / s3-buckets / s3-objects)
- [x] G-219 Sandbox / argv (mutantes negados; sem endpoint/query/profile)
- [x] G-220 Credenciais no ambiente (nunca no IR)
- [x] G-221 Falha vs sucesso
- [x] G-222 Registry + Graph + Intent (`heuristic.aws`)
- [x] G-223 Exclusões v0
