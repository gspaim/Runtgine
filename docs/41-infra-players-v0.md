# 41 — Infra Players v0 (Kubernetes, Terraform, PostgreSQL)

Três Players determinísticos de **infra local via CLI**, recorte
restante de G-41 depois de Git / FS / Docker / HTTP / Test / NPM /
pytest / yarn:

| Player | Pacote | Capabilities v0 |
|---|---|---|
| `k8s` | `internal/players/k8s` | `k8s.get`, `k8s.list` |
| `terraform` | `internal/players/tf` | `tf.validate`, `tf.plan` |
| `postgres` | `internal/players/pg` | `pg.ping` |

Inventário: [10-gaps.md](10-gaps.md) (G-201+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é `kubectl apply` / `delete` / `exec`. Não é `terraform apply`
/ `destroy` / `init`. Não é SQL arbitrário. Não é Docker (`23`).
Não é MCP. Não é template (`40`).

**Status deste doc: CONFIRMED v0 (slice 34 feito).** G-201..G-209.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-26-041-infra-players/`](../openspec/changes/archive/2026-08-26-041-infra-players/)
(arquivado após o slice 34). Spec atual: [`openspec/specs/infra-players/`](../openspec/specs/infra-players/).

---

## 1. Problema

`kubectl`, `terraform` e `psql` ainda caem em `shell.exec` (argv
livre) ou no caminho LLM. Sem schema, sem allowlist, sem policy
HITL em `plan`. O restante explícito de G-41 após pytest/yarn era
**infra / K8s / TF / PG**.

Este corte entrega leitura / validação / ping — a superfície que
o runtime consegue exercer offline com `ExecFunc` injetável.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| CLI argv (`kubectl` / `terraform` / `psql`) | SDK / CGO / client-go / libpq |
| Leitura cluster + validate/plan + ping | Apply, destroy, exec, SQL livre |
| Workdir no workspace (TF) | Backend remoto obrigatório no contrato |
| Policy allow (leitura/ping); `tf.plan` HITL | Privileged / host network |
| Binário no PATH | Daemon próprio do Runtgine |

Regras:

1. Validator / Registry / Policy soberanos.
2. Nunca shell string. Só argv.
3. Refs (`resource`, `name`, `namespace`, `image`-like) não começam
   com `-` e não têm espaços (espelha Docker `safeRef`).
4. Unit tests injetam o runner; CI **não** exige cluster, terraform
   nem postgres.
5. Blast / Claims: nenhuma linha nova (como `http.get` / `test.go`).

---

## 3. Cortes confirmados (G-201+)

### G-201 — Kubernetes Player

- Nome: `k8s`; pacote `internal/players/k8s`; kind `deterministic`
- Capabilities:

| Capability | Entrada | Saída | Policy |
|---|---|---|---|
| `k8s.list` | `resource` (obrigatório), `namespace?`, `timeout_ms?` | `resource`, `object` (JSON), `truncated` | allow |
| `k8s.get` | `resource`, `name` (obrigatórios), `namespace?`, `timeout_ms?` | idem + `name` | allow |

Argv: `kubectl get <resource> [name] [-n ns] -o json`.
Timeout default 30s, teto 120s. JSON truncado 1 MiB.

### G-202 — K8s sandbox

- `resource` / `name` / `namespace`: `safeRef` (sem espaço, sem
  prefixo `-`)
- Sem `apply`, `delete`, `exec`, `logs`, `port-forward`, `--raw`,
  `-f`, stdin YAML
- Sem kubeconfig path no input (usa default do processo)

### G-203 — Terraform Player

- Nome: `terraform`; pacote `internal/players/tf`
- Capabilities:

| Capability | Entrada | Saída | Policy |
|---|---|---|---|
| `tf.validate` | `workdir?`, `timeout_ms?` | `ok`, `exit_code`, `log` | allow |
| `tf.plan` | `workdir?`, `timeout_ms?` | `ok`, `exit_code`, `log` | **approval-required** |

Argv: `terraform validate -no-color` /
`terraform plan -lock=false -input=false -no-color`.
Timeout default 120s, teto 10 min. `workdir` relativo no workspace
e **deve** conter `*.tf` ou `*.tf.json`.

### G-204 — Terraform sandbox

- Sem `apply`, `destroy`, `init`, `import`, `taint`, `-auto-approve`,
  `-var`, `-var-file` no input
- Sem `-chdir` que escape o workdir resolvido
- `terraform init` **fora** (se `.terraform` faltar, o binário
  falha → `runtime.player_error`, como `npm` sem `node_modules`)

### G-205 — PostgreSQL Player

- Nome: `postgres`; pacote `internal/players/pg`
- Capability única `pg.ping`

| Campo | Default | Regra |
|---|---|---|
| `dbname` | obrigatório | `safeRef` |
| `host` | `127.0.0.1` | `safeRef` |
| `port` | `5432` | 1–65535 |
| `user` | omitido | `safeRef` se presente |
| `timeout_ms` | 10000 | teto 60000 |

Argv: `psql --host --port --dbname [--username] --no-psqlrc
-t -A --pset pager=off --command SELECT 1`.
Saída: `ok`, `exit_code`, `log`.

### G-206 — Postgres sandbox

- **Sem** campo `sql` / `command` / `password` no input
  (`additionalProperties: false`)
- Senha só via env herdado `PGPASSWORD` (e `PGSSLMODE`); nunca no
  Task IR; nunca `RUNTGINE_*`
- Sem `\copy`, `-f`, `-c` livre, superuser flags

### G-207 — Registry + Graph + Intent

- `api.Open` registra os três Players
- Graph: `RefreshFromRegistry` cria nós/arestas `provides`
- Intent (alta confiança, antes de `matchShell`):

| NL | Capability | Method |
|---|---|---|
| `kubectl get <resource>` / `k8s get <resource>` | `k8s.list` | `heuristic.k8s` |
| `kubectl get <resource> <name>` | `k8s.get` | `heuristic.k8s` |
| `terraform validate` | `tf.validate` | `heuristic.tf` |
| `terraform plan` | `tf.plan` | `heuristic.tf` |
| `pg ping` / `postgres ping` / `psql ping` | `pg.ping` | `heuristic.pg` |

`kubectl apply` **não** casa — cai no shell/LLM (sem capability).

### G-208 — Falha vs sucesso

- exit 0 → step ok
- exit ≠ 0 / binário ausente / timeout → `runtime.player_error` /
  `runtime.timeout`
- `tf.plan` ainda exige HITL **antes** de executar (Manifest
  `approval-required`)
- Testes injetam runner; `go test ./...` sem cluster/terraform/psql

### G-209 — Exclusões v0

- `kubectl apply|delete|exec|logs|port-forward|cp`
- Helm / kustomize / k3s / kind como Players
- `terraform apply|destroy|init|import`
- SQL arbitrário, migrations, `pg_dump`
- AWS/GCP/Azure Players; Ansible
- MCP; templates (`40`); Compose (já fora do Docker v0)

---

## 4. Critérios de aceite

1. Manifest **não** registra `k8s.apply`, `tf.apply`, `pg.query`.
2. `resource` / `name` começando com `-` → `validation.invalid_input`.
3. `tf.plan` no Manifest = `approval-required`.
4. Workdir TF sem `*.tf` → `validation.invalid_input`.
5. `runtgine intent --dry-run "terraform validate"` → `tf.validate`,
   method `heuristic.tf`.
6. Fake exec exit 1 → `runtime.player_error`.
7. `go test ./...` / `go vet ./...` verdes sem cluster.
8. OpenSpec `041` arquivado após o código (slice 34).

---

## 5. Ordem do slice de código

1. Pacotes `k8s`, `tf`, `pg` + Manifest + `ValidateStaticInput`
2. Registrar em `api.Open` + runner admission
3. Heurísticas Intent
4. Examples `examples/k8s-list.json`, `tf-validate.json`, `pg-ping.json`
5. Testes fake; README Estágio: Slice 34
6. Arquivar OpenSpec `041`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-201 K8s Player (`k8s.get` / `k8s.list`)
- [x] G-202 K8s sandbox (sem apply/exec)
- [x] G-203 Terraform Player (`tf.validate` / `tf.plan`)
- [x] G-204 Terraform sandbox (sem apply/init)
- [x] G-205 Postgres Player (`pg.ping`)
- [x] G-206 Postgres sandbox (sem SQL/password no IR)
- [x] G-207 Registry + Graph + Intent
- [x] G-208 Falha vs sucesso
- [x] G-209 Exclusões
