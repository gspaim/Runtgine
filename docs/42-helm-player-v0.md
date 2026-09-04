# 42 — Helm Player v0

Player determinístico de **Helm via CLI**, recorte de G-41 depois de
Git / FS / Docker / HTTP / Test / NPM / pytest / yarn / infra (K8s,
Terraform, PostgreSQL). Levanta a exclusão de Helm registrada no
corte v0 de `41` (G-209) como recorte próprio:

| Player | Pacote | Capabilities v0 |
|---|---|---|
| `helm` | `internal/players/helm` | `helm.lint`, `helm.template`, `helm.list`, `helm.status` |

Inventário: [10-gaps.md](10-gaps.md) (G-210+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é `helm install` / `upgrade` / `rollback` / `uninstall`. Não é
`helm get` nem `helm test`. Não é repo/OCI (add/push/package). Não é
`kubectl` (`41`). Não é Docker (`23`). Não é MCP. Não é template (`40`).

**Status deste doc: CONFIRMED v0 (código = slice 35, pendente).**
G-210..G-216.

**Pacote OpenSpec:** [`openspec/changes/042-helm-player/`](../openspec/changes/042-helm-player/)
(mudança ativa; arquivar após o slice 35).

---

## 1. Problema

`helm lint`, `helm template`, `helm list` e `helm status` ainda caem
em `shell.exec` (argv livre) ou no caminho LLM. Sem schema, sem
allowlist, sem marker de chart. O restante explícito de G-41 após o
recorte infra era **Helm / cloud / SQL**; Helm é o recorte seguinte
por continuidade com a família infra (`41`).

Este corte entrega lint / render local / leitura de cluster — a
superfície que o runtime consegue exercer offline (`lint`,
`template`) ou em leitura (`list`, `status`) com `ExecFunc`
injetável.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| CLI argv (`helm`) | SDK / client-go / biblioteca Helm |
| `lint` / `template` (offline) + `list` / `status` (leitura) | install, upgrade, rollback, uninstall, get, test |
| Chart no workspace (`Chart.yaml` marker) | Chart remoto (`--repo` / OCI) no contrato |
| Values default do chart | `--set*` / values inline / values file no v0 |
| Policy allow (todas as capabilities) | approval-required no v0 |
| Binário no PATH | Daemon próprio do Runtgine |

Regras:

1. Validator / Registry / Policy soberanos.
2. Nunca shell string. Só argv.
3. `release` / `namespace`: `safeRef` (sem espaço, sem prefixo `-`,
   espelha Docker `safeRef`).
4. `chart` é caminho relativo no workspace e **deve** conter
   `Chart.yaml` (mesmo padrão do marker `*.tf` de `41`).
5. Unit tests injetam o runner; CI **não** exige cluster nem helm.
6. Blast / Claims: nenhuma linha nova (como `http.get` / `test.go`).

---

## 3. Cortes confirmados (G-210+)

### G-210 — Helm Player

- Nome: `helm`; pacote `internal/players/helm`; kind `deterministic`
- Binário `helm` no PATH (`exec.LookPath` + `CommandContext`, argv)

### G-211 — Capabilities v0

| Capability | Entrada | Saída | Policy |
|---|---|---|---|
| `helm.lint` | `chart` (obrigatório), `timeout_ms?` | `ok`, `exit_code`, `log` | allow |
| `helm.template` | `chart` (obrigatório), `release?`, `namespace?`, `timeout_ms?` | `output` (YAML renderizado), `truncated` | allow |
| `helm.list` | `namespace?`, `timeout_ms?` | `releases` (JSON), `truncated` | allow |
| `helm.status` | `release` (obrigatório), `namespace?`, `timeout_ms?` | `ok`, `exit_code`, `log` | allow |

Argv:

- `helm lint <chart-path>`
- `helm template [release] <chart-path> [-n ns]`
- `helm list [-n ns] -o json`
- `helm status <release> [-n ns]`

Timeouts: `lint`/`template` default 120s, teto 600s; `list`/`status`
default 30s, teto 120s. `output` / `releases` truncados em 1 MiB
(espelha `k8s` em `41`).

### G-212 — Sandbox

- Capabilities, subcomandos e flags **fechados** por argv: só os da
  tabela G-211
- Sem `install`, `upgrade`, `rollback`, `uninstall`/`delete`, `get`,
  `test`, `package`, `push`, `repo`, `plugin`
- Sem `--set`, `--set-string`, `--set-json`, `--values`, `-f`,
  `--repo`, `--version`, `--kubeconfig`, `--kube-as-user` no input
- Input sem campo `values` / `set` (`additionalProperties: false`)
- `list` sempre `-o json` (saída parseável; falha de parse → string
  bruta + `truncated`)

### G-213 — Chart no workspace + cluster read

- `chart`: caminho relativo resolvido com `shell.ResolveWorkdir`
  (fora do workspace → `validation.invalid_input`); marker: precisa
  de `Chart.yaml` no diretório (não recursivo no v0)
- `helm.list` / `helm.status` usam o kubeconfig herdado do processo
  (mesmo modelo do `kubectl` em `41`); sem override de kubeconfig no
  input
- `lint` / `template` são offline: sem acesso a cluster no contrato

### G-214 — Falha vs sucesso

- exit 0 → step ok
- exit ≠ 0 (inclusive `helm lint` com falha de lint) / binário
  ausente → `runtime.player_error`
- timeout → `runtime.timeout`
- Testes injetam runner; `go test ./...` sem cluster nem helm

### G-215 — Registry + Graph + Intent

- `api.Open` registra o Player `helm`
- Graph: `RefreshFromRegistry` cria nó/aresta `provides`
- Intent (alta confiança, antes de `matchShell`):

| NL | Capability | Method |
|---|---|---|
| `helm lint <path>` | `helm.lint` | `heuristic.helm` |
| `helm template <path>` | `helm.template` | `heuristic.helm` |
| `helm list` | `helm.list` | `heuristic.helm` |
| `helm status <release>` | `helm.status` | `heuristic.helm` |

`helm install` / `upgrade` **não** casam — caem no shell/LLM (sem
capability).

### G-216 — Exclusões v0

- `helm install|upgrade|rollback|uninstall|delete|get|test`
- Repo management (`helm repo add/update/remove`) e OCI
  (`package` / `push`)
- `plugin`; `--set*` / values file / values inline
- Kustomize / k3s / kind como Players (recortes futuros de G-41)
- Cloud SDKs (AWS/GCP/Azure) e SQL/migrations
- MCP; templates (`40`); NATS (G-36, DEFERRED)

---

## 4. Critérios de aceite

1. Manifest **não** registra `helm.install`, `helm.upgrade`,
   `helm.get`.
2. `release` / `namespace` começando com `-` →
   `validation.invalid_input`.
3. `chart` sem `Chart.yaml` ou fora do workspace →
   `validation.invalid_input`.
4. `helm list` argv contém `-o json`.
5. `runtgine intent --dry-run "helm template charts/demo"` →
   `helm.template`, method `heuristic.helm`.
6. Fake exec exit 1 → `runtime.player_error`.
7. `go test ./...` / `go vet ./...` verdes sem cluster.
8. OpenSpec `042` arquivado após o código (slice 35).

---

## 5. Ordem do slice de código

1. Pacote `helm` + Manifest + `ValidateStaticInput`
2. Registrar em `api.Open` + runner admission
3. Heurística Intent `heuristic.helm`
4. Examples `examples/helm-lint.json`, `helm-template.json`,
   `helm-list.json`
5. Testes fake; README Estágio: Slice 35
6. Arquivar OpenSpec `042`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-210 Helm Player (`internal/players/helm`)
- [x] G-211 Capabilities v0 (`lint` / `template` / `list` / `status`)
- [x] G-212 Sandbox (sem install/upgrade; sem `--set*`)
- [x] G-213 Chart no workspace + cluster read
- [x] G-214 Falha vs sucesso
- [x] G-215 Registry + Graph + Intent (`heuristic.helm`)
- [x] G-216 Exclusões v0
