# 44 — Cloud Players v0 (GCP + Azure read-only)

Dois Players determinísticos de **cloud CLI read-only**, fechando a
parte cloud de G-41 depois de AWS (`43`). Levanta a exclusão de
cloud SDKs de `41`–`43` (G-209/G-216/G-223) **só para GCP e Azure
em leitura**:

| Player | Pacote | Binário | Capabilities v0 |
|---|---|---|---|
| `gcp` | `internal/players/gcp` | `gcloud` | `gcp.identity`, `gcp.config`, `gcp.projects` |
| `azure` | `internal/players/azure` | `az` | `azure.identity`, `azure.subscriptions`, `azure.groups` |

Inventário: [10-gaps.md](10-gaps.md) (G-224+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é `gcloud projects create` / `iam` / `deploy`. Não é `az group
create` / `role` / `storage` mutante. Não é criação/alteração de
nenhum recurso. Não é SQL arbitrário (continua fora, corte próprio
futuro). Não é NATS (G-36). Não é MCP. Não é template (`40`).

**Status deste doc: CONFIRMED v0 (slice 37 feito).** G-224..G-230.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-09-04-044-cloud-gcp-azure/`](../openspec/changes/archive/2026-09-04-044-cloud-gcp-azure/)
(arquivado após o slice 37). Spec atual: [`openspec/specs/cloud-gcp-azure/`](../openspec/specs/cloud-gcp-azure/).

---

## 1. Problema

A verificação de identidade/projeto em GCP e Azure ainda depende de
`shell.exec` (argv livre) ou do caminho LLM. Sem schema, sem
allowlist, sem política de credenciais. O restante de G-41 após AWS
era **cloud GCP/Azure** e **SQL**; este corte fecha a parte cloud.
SQL arbitrário e migrations continuam excluídos (G-206/G-209/G-216)
e exigem spec própria com desenho de segurança dedicado.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| CLI argv (`gcloud` / `az`), subcomandos fechados | SDK / bibliotecas cliente |
| Identidade, config, listagem (leitura pura) | create/delete/set/iam/deploy/role |
| Saída JSON (`--format=json` / `-o json`) | Texto livre / `--filter` livre |
| Credenciais via ambiente/config herdado | Credentials/tokens no Task IR |
| Sem campos de projeto/subscription no IR | `--project` / `--subscription` no input |
| Binário no PATH | Daemon próprio do Runtgine |

Regras:

1. Validator / Registry / Policy soberanos.
2. Nunca shell string. Só argv.
3. Todas as capabilities `allow` (leitura pura; sem HITL no v0).
4. Nenhum campo de input além de `timeout_ms`
   (`additionalProperties: false`).
5. Unit tests injetam o runner; CI **não** exige gcloud, az nem
   credenciais.
6. Blast / Claims: nenhuma linha nova (como `helm.*` / `aws.*`).

---

## 3. Cortes confirmados (G-224+)

### G-224 — GCP Player

- Nome: `gcp`; pacote `internal/players/gcp`; binário `gcloud`
- Capabilities:

| Capability | Entrada | Saída | Policy |
|---|---|---|---|
| `gcp.identity` | `timeout_ms?` | `object` (JSON), `truncated` | allow |
| `gcp.config` | `timeout_ms?` | idem | allow |
| `gcp.projects` | `timeout_ms?` | idem | allow |

Argv: `gcloud auth list --format=json` / `gcloud config list
--format=json` / `gcloud projects list --format=json`.
`gcp.identity` e `gcp.config` são locais (sem rede); `gcp.projects`
é leitura de rede. Paginação/filtro são do CLI; sem `--filter` no
input.

### G-225 — GCP sandbox

- Subcomandos e flags fechados: só os de G-224 (`--format=json`
  fixo, sem `--filter`, sem `--project` no input)
- Projeto/região via ambiente/config herdado
  (`GOOGLE_APPLICATION_CREDENTIALS`, `CLOUDSDK_CORE_PROJECT`,
  `gcloud config`) — nunca no Task IR (modelo `AWS_*` de `43`)
- Nenhum subcomando mutante registrado: `create`/`delete`/`set`/
  `add-iam-policy-binding`/`deploy`/`run`

### G-226 — Azure Player

- Nome: `azure`; pacote `internal/players/azure`; binário `az`
- Capabilities:

| Capability | Entrada | Saída | Policy |
|---|---|---|---|
| `azure.identity` | `timeout_ms?` | `object` (JSON), `truncated` | allow |
| `azure.subscriptions` | `timeout_ms?` | idem | allow |
| `azure.groups` | `timeout_ms?` | idem | allow |

Argv: `az account show -o json` / `az account list -o json` /
`az group list -o json`.

### G-227 — Azure sandbox

- Subcomandos e flags fechados: só os de G-226 (`-o json` fixo,
  sem `--subscription`/`--query`/`--output` livre no input)
- Credenciais via `az login` do operador / ambiente herdado
  (`AZURE_CONFIG_DIR`) — nunca no Task IR
- Nenhum subcomando mutante registrado: `create`/`delete`/`update`/
  `role`/`storage` mutante

### G-228 — Falha vs sucesso

- exit 0 → step ok (JSON parseado em `object`)
- exit ≠ 0 (inclusive sem login/credencial) / binário ausente →
  `runtime.player_error` com a mensagem do stderr
- timeout → `runtime.timeout`
- JSON não-parseável → string bruta truncada em `object` +
  `truncated: true` (padrão `k8s`/`aws`)
- Testes injetam runner; `go test ./...` sem gcloud/az/credenciais

### G-229 — Registry + Graph + Intent

- `api.Open` registra `gcp` e `azure`
- Graph: `RefreshFromRegistry` cria nós/arestas `provides`
- Intent (alta confiança, antes de `matchShell`):

| NL | Capability | Method |
|---|---|---|
| `gcloud auth list` | `gcp.identity` | `heuristic.gcp` |
| `gcloud config list` | `gcp.config` | `heuristic.gcp` |
| `gcloud projects list` | `gcp.projects` | `heuristic.gcp` |
| `az account show` | `azure.identity` | `heuristic.az` |
| `az account list` | `azure.subscriptions` | `heuristic.az` |
| `az group list` | `azure.groups` | `heuristic.az` |

`gcloud projects create` / `az group delete` **não** casam — caem
no shell/LLM (sem capability).

### G-230 — Exclusões v0

- Todo subcomando mutante (create/delete/set/update, IAM, roles,
  deploy, storage mutante)
- `--filter` / `--query` / JMESPath / `--project` /
  `--subscription` no input
- Outros provedores (Alibaba, Oracle); Terraform Helm provider
- SQL arbitrário / migrations (corte próprio futuro)
- NATS (G-36, DEFERRED); MCP; templates (`40`)

---

## 4. Critérios de aceite

1. Manifest **não** registra `gcp.projects-create`,
   `azure.groups-create`, `gcp.iam`.
2. Input com qualquer campo além de `timeout_ms` →
   `validation.invalid_input`.
3. Argv de `gcp.*` contém `--format=json`; argv de `azure.*`
   contém `-o`, `json`.
4. `runtgine intent --dry-run "az account show"` →
   `azure.identity`, method `heuristic.az`.
5. `runtgine intent --dry-run "gcloud projects create x"` → **não**
   é `heuristic.gcp`.
6. Fake exec exit 1 → `runtime.player_error`.
7. `go test ./...` / `go vet ./...` verdes sem gcloud/az/credenciais.
8. OpenSpec `044` arquivado após o código (slice 37).

---

## 5. Ordem do slice de código

1. Pacotes `gcp` e `azure` + Manifest + `ValidateStaticInput`
2. Registrar em `api.Open` + runner admission
3. Heurísticas Intent `heuristic.gcp` / `heuristic.az`
4. Examples `gcp-identity.json`, `gcp-projects.json`,
   `az-identity.json`, `az-groups.json`
5. Testes fake; README Estágio: Slice 37
6. Arquivar OpenSpec `044`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-224 GCP Player (`gcp.identity` / `config` / `projects`)
- [x] G-225 GCP sandbox (env/config herdado; `--format=json` fixo)
- [x] G-226 Azure Player (`azure.identity` / `subscriptions` / `groups`)
- [x] G-227 Azure sandbox (`az login` herdado; `-o json` fixo)
- [x] G-228 Falha vs sucesso
- [x] G-229 Registry + Graph + Intent (`heuristic.gcp` / `heuristic.az`)
- [x] G-230 Exclusões v0
