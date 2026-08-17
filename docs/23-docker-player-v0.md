# 23 — Docker Player v0

Player determinístico para o daemon Docker local, com sandbox de argv e
policy HITL nas operações que criam imagem/container.

Inventário: [10-gaps.md](10-gaps.md) (G-87+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Pré-requisito: Execution Policy + HITL v0 implementados
([22-execution-policy-v0.md](22-execution-policy-v0.md), slice 10).

**Status deste doc: CONFIRMED (v0).** G-87..G-92 autorizam o slice 11
de código. Compose, K8s, push/pull explícitos e privileged permanecem fora.

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-17-023-docker-player/`](../openspec/changes/archive/2026-08-17-023-docker-player/).
Deltas mergeados em `openspec/specs/docker-player/`. Branch de implementação:
`feat/023-docker-player`.

---

## 1. Problema

O Runtgine cobre Shell, Git e filesystem, mas não tem capability
`docker.*`. Tasks de container passam por `shell.exec`, sem schema,
sem policy HITL e sem argv controlado.

Docker é o primeiro Player em que `approval-required` deixa de ser
exemplo e vira default de Manifest para `run` / `build`.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Player `docker` determinístico | Agent / LLM / Compose / Kubernetes |
| Binário `docker` no `PATH` (argv) | SDK Docker / CGO / API HTTP própria |
| Leitura local + run/build gated | Push, pull explícito, registry login |
| Confinamento: workspace + daemon local | `--privileged`, host network, bind fora do workspace |

Regras:

1. Validator / Registry / Execution Policy continuam soberanos.
2. Nunca shell string (`sh -c`). Só `docker <subcommand> ...`.
3. `docker.run` não puxa imagem (`--pull=never`); imagem deve existir local.
4. `docker.run` usa `--network=none` e `--rm` no v0.
5. Bind mounts, se houver, só o workspace root (e somente se o input
   pedir `mount_workspace=true`; default false).
6. O daemon é uma fronteira que o Core não isola (análogo aos hooks Git):
   a mitigação é argv + policy, não seccomp no Runtgine.

---

## 3. Cortes confirmados (G-87+)

### G-87 — Papel e pacote

**Status: CONFIRMED**

- Nome do Player: `docker`
- Pacote: `internal/players/docker`
- Kind: `deterministic`
- Registro em `api.Open` após Filesystem
- Recorte de G-41: só Docker Engine CLI v0

### G-88 — Capabilities v0

**Status: CONFIRMED**

| Capability | Entrada (resumo) | Saída (resumo) | Policy Manifest |
|---|---|---|---|
| `docker.ps` | `all?` (bool, default false), `timeout_ms?` | `containers[]` `{id, image, names, status}` | omit / allow |
| `docker.inspect` | `id` obrigatório (container ou image), `timeout_ms?` | `id`, `object` (JSON truncável) | omit / allow |
| `docker.logs` | `id` obrigatório, `tail?` (default 100, max 1000), `timeout_ms?` | `id`, `logs`, `truncated` | omit / allow |
| `docker.run` | `image` obrigatório, `argv?[]`, `workdir?`, `mount_workspace?` (default false), `timeout_ms?` | `container_id`, `exit_code`, `stdout`, `stderr` | **approval-required** |
| `docker.build` | `context` (path relativo no workspace, default `.`), `tag?`, `dockerfile?` (relativo ao context), `timeout_ms?` | `image_id` ou `tag`, `stderr?` | **approval-required** |

Schemas JSON no Manifest; `additionalProperties: false`.

Timeouts: default **60s** em ps/inspect/logs; **120s** em run/build
(teto do input: 10 min).

`docker.inspect` / `docker.logs`: `id` é nome ou ID curto; rejeitar
strings com espaços ou flags (`-e`, `--`).

`docker.run` `workdir`: se presente, path relativo resolvido **dentro**
do workspace; passado como `-w` no container (não implica mount). Sem
`mount_workspace`, o `-w` só faz sentido se a imagem já tiver esse path.

`docker.build` `context` / `dockerfile`: confinement igual FS (symlink
externo rejeitado). `--pull=false`. Sem `-f` fora do context.

### G-89 — Sandbox / argv

**Status: CONFIRMED**

| Regra | Corte v0 |
|---|---|
| Invocação | `exec.Command` argv; nunca shell |
| Binário | `docker` no `PATH` (allowlist permissiva + warn, como Git) |
| Subcommands | allowlist: `ps`, `inspect`, `logs`, `run`, `build`, `images` (helper interno se preciso) |
| `docker.run` flags fixas | `--pull=never --network=none --rm` |
| `docker.run` proibido no input | `--privileged`, `--pid=host`, `--network`, `-p`/`--publish`, `--volume`/`-v` arbitrário, `--mount`, `--user=0` forçado pelo usuário via flag livre |
| Mount | só se `mount_workspace=true`: um `-v <workspace>:<workspace>:ro` |
| `docker.build` | `docker build --pull=false <context>`; `-t` se `tag`; `-f` só se dockerfile no context |
| Rede no build | daemon pode falar com registry (risco residual); sem `docker pull` capability |
| Env | herança mínima Shell; sem tokens / `RUNTGINE_*`; `DOCKER_HOST` do **input** proibido (usa o ambiente do processo) |

Falha de sandbox → `validation.invalid_input` (estático) ou
`runtime.player_error`.

Daemon ausente / `docker` missing → `runtime.player_error` (não
`validation.unknown_capability`).

### G-90 — Policy (usa spec 22)

**Status: CONFIRMED**

- Manifest declara `execution_policy: "approval-required"` em
  `docker.run` e `docker.build`.
- `docker.ps` / `inspect` / `logs` omitem o campo (default allow).
- Config do workspace pode endurecer (ex. deny em `docker.build`) ou
  relaxar (allow em `docker.run`) — precedência da spec 22.
- Testes do Player cobrem o caminho HITL: submit → `waiting_approval` →
  `ApproveRun(grant)` → succeeded. Sem grant, o daemon **não** recebe
  `run`/`build`.
- Sem a spec 22 no binário, este slice **não** começa.

### G-91 — Integração Core

**Status: CONFIRMED**

1. `api.Open` registra `docker.New()`.
2. Validator valida `input` contra o Manifest.
3. Runner chama `docker.ValidateStaticInput` na admissão (ids, paths,
   flags, limits).
4. Runtime Graph: nós `capability` `docker.*` via refresh do Registry.
5. Exemplos: `examples/docker-ps.json`; fixture de teste para run+HITL
   (não precisa de exemplo de `docker.run` sem aviso — se houver
   `examples/docker-run.json`, o Manifest já exige approve).

### G-92 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- `docker pull` / `push` / `login` / `tag` / `rmi` / `rm` / `exec` / `stop`
- Compose (`docker compose`) e Swarm
- Kubernetes
- `--privileged`, host network, bind fora do workspace
- GPU, devices, cgroup custom
- Intent heuristics Docker (nice-to-have)
- TUI dedicada / tab GRAPH
- Claims / Blast Radius

---

## 4. Critérios de aceite

1. `docker.ps` (daemon disponível) emite `run.succeeded` com lista JSON.
2. Sem daemon: erro de Player, não panic; testes unitários não exigem
   daemon (fake/exec stub ou skip documentado). Integração HITL pode
   usar stub do executor.
3. `docker.run` sem approve não invoca o binário; após grant, argv contém
   `--pull=never` e `--network=none`.
4. `mount_workspace=false` (default) não adiciona `-v`.
5. Path de build que escapa o workspace é rejeitado.
6. Capability inventada `docker.push` não está no Manifest → Validator rejeita.
7. `go test ./internal/players/docker/...` cobre argv, confinement, policy
   gating (com Core de teste).
8. `go test ./...` e `go vet ./...` verdes **sem** Docker instalado
   (testes que precisam do daemon são `testing.Short` skip ou build tag).
9. OpenSpec `023-docker-player` arquivado após o merge do código.

---

## 5. Ordem do slice de código

1. Slice 10 (022) mergeado
2. G-87..G-92 CONFIRMED — este doc + OpenSpec
3. Pacote `internal/players/docker` + Manifest + testes
4. Registrar no Core, static validation, exemplos
5. README Estágio: Slice 11 Feito

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-87 Papel / pacote `docker`
- [x] G-88 Capabilities ps/inspect/logs/run/build
- [x] G-89 Sandbox argv + network none + pull never
- [x] G-90 Manifest approval-required em run/build
- [x] G-91 Registry + static validation + exemplo
- [x] G-92 Exclusões (push/compose/privileged/K8s)
