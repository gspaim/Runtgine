# 15 — Git workflow e releases

Fluxo canônico do repositório. Autoridade: [04-decisoes.md](04-decisoes.md).

```text
feat/<NNN>-<slug>  →  develop  →  release/x.y.z (+ RC tags)  →  main (+ tag estável)
```

## Branches

| Branch | Papel | Quem mergeia |
|---|---|---|
| `feat/<NNN>-<slug>` (e `fix/` `docs/` `chore/`) | Trabalho isolado | Autor → `develop` via PR |
| `develop` | Integração contínua (instável ok) | Maintainers |
| `release/x.y.z` | Congela escopo; só bugfix / docs / version bump | Maintainers |
| `main` | Código liberável / última estável | Maintainers (só via `release/*`) |

`main` permanece a default branch do GitHub. `develop` é a base usual de PRs de feature.

### Bootstrap (uma vez)

Se `develop` ainda não existir:

```bash
git checkout main
git pull origin main
git checkout -b develop
git push -u origin develop
```

## Naming de branches de trabalho

Formato: `<tipo>/<NNN>-<slug>`

| Parte | Regra | Exemplo |
|---|---|---|
| `tipo` | `feat`, `fix`, `docs`, `chore` | `feat` |
| `NNN` | Id da spec ou issue, zero-padded a 3+ dígitos | `001` |
| `slug` | kebab-case curto | `shell-player` |

Exemplos válidos:

- `feat/001-shell-player`
- `fix/042-queue-deadlock`
- `docs/015-git-workflow`

PRs de agentes Cursor (`cursor/...`) são exceção operacional; ao concluir, o conteúdo deve aterrissar em `develop` (ou `release/*` se for hotfix de RC).

## Fluxo de feature

1. Abrir/ter issue ou spec com id `NNN`.
2. Branch a partir de `develop`:
   ```bash
   git checkout develop && git pull
   git checkout -b feat/001-exemplo
   ```
3. Abrir PR **para `develop`** (não para `main`).
4. CI verde (`go test ./...`, `go vet ./...`).
5. Merge (squash ou merge commit — preferir squash em features pequenas).
6. Apagar a branch de trabalho.

## Release candidates

Quando `develop` estiver pronto para um corte `x.y.z`:

1. Criar `release/x.y.z` a partir de `develop`:
   ```bash
   git checkout develop && git pull
   git checkout -b release/x.y.z
   git push -u origin release/x.y.z
   ```
2. Nesta branch: só correções, docs de release e bumps de versão.
3. Publicar RC com tag anotada:
   ```bash
   git tag -a vX.Y.Z-rc.1 -m "Runtgine vX.Y.Z-rc.1"
   git push origin vX.Y.Z-rc.1
   ```
4. Tags `v*-rc.*` geram **GitHub Release em prerelease** (workflow `release.yml`).
5. RC seguintes: `vX.Y.Z-rc.2`, … na mesma `release/x.y.z`.

Hotfix durante RC: PR **para `release/x.y.z`**, depois back-merge para `develop`.

## Release estável

1. CI verde na `release/x.y.z`.
2. PR `release/x.y.z` → `main`.
3. Após merge em `main`, tag estável:
   ```bash
   git checkout main && git pull
   git tag -a vX.Y.Z -m "Runtgine vX.Y.Z"
   git push origin vX.Y.Z
   ```
4. Back-merge `main` → `develop` para não divergir.
5. Tag `vX.Y.Z` (sem `-rc`) gera release **não-prerelease**.

## Semver do produto

- Tags Git (`v0.1.0`, `v0.1.0-rc.1`) versionam o **binário/CLI**.
- Enquanto MVP sem estabilidade de API: série **`0.y.z`**.
- `schema_version` do Task IR (ex.: `"0.1.0"`) é contrato de protocolo — **não** precisa coincidir com a tag do produto.

## CI / Actions

| Workflow | Gatilho | Ação |
|---|---|---|
| `ci.yml` | PR e push em `develop`, `main`, `release/**` | `go test ./...`, `go vet ./...` |
| `release.yml` | Push de tag `v*` | Build multi-OS + GitHub Release |

## Proteção de branches (recomendado)

No GitHub → Settings → Branches, para `develop`, `main` e `release/*`:

- Require pull request before merging
- Require status checks: job `test` do workflow CI
- Restrict who can push (maintainers)
- Em `main`: dispensar force-push; idealmente só merges vindos de `release/*`

Branch protection é configuração do remoto (não versionada neste repo).

## Relação com segurança

Após a primeira release estável, vulnerabilidades devem citar a tag afetada.
Até lá, a política em `SECURITY.md` cobre o tip de `main` (e, após bootstrap, o tip de `develop` para pré-release).
