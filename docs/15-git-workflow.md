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

`main` permanece a default branch do GitHub (clone e PRs novos apontam para ela).
**Troque a base do PR para `develop`.** `develop` já existe no remoto.

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

## OpenSpec (pacotes de mudança)

Layout e regras: [`openspec/README.md`](../openspec/README.md).

Cada mudança ativa vive em:

```text
openspec/changes/<NNN>-<slug>/
  proposal.md
  design.md
  tasks.md
  specs/<domain>/spec.md    # deltas ADDED|MODIFIED|REMOVED
```

O `NNN-slug` **é o mesmo** da branch `feat/<NNN>-<slug>` e do id da spec
de produto quando houver (`docs/NN-…` ou gaps G-xx promovidos).

| Papel | Onde |
|---|---|
| Status CONFIRMED / HYPOTHESIS | `docs/04-decisoes.md` (autoridade) |
| Comportamento atual do sistema | `openspec/specs/` |
| Pacote da próxima implementação | `openspec/changes/<NNN>-<slug>/` |

Ao concluir a mudança: merge dos deltas em `openspec/specs/` e mover a
pasta para `openspec/changes/archive/YYYY-MM-DD-NNN-slug/`.

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
2. No PR `release/x.y.z` → `main`, conferir e atualizar a seção
   **Estágio do projeto** do `README.md` para espelhar o que entra em `main`
   (Feito / Próximo / Depois).
3. PR `release/x.y.z` → `main`.
4. Após merge em `main`, tag estável:
   ```bash
   git checkout main && git pull
   git tag -a vX.Y.Z -m "Runtgine vX.Y.Z"
   git push origin vX.Y.Z
   ```
5. Back-merge `main` → `develop` para não divergir.
6. Tag `vX.Y.Z` (sem `-rc`) gera release **não-prerelease**.

## Semver do produto

- Tags Git (`v0.1.0`, `v0.1.0-rc.1`) versionam o **binário/CLI**.
- Enquanto MVP sem estabilidade de API: série **`0.y.z`**.
- `schema_version` do Task IR (ex.: `"0.1.0"`) é contrato de protocolo — **não** precisa coincidir com a tag do produto.

## CI / Actions

| Workflow | Gatilho | Ação |
|---|---|---|
| `ci.yml` | PR e push em `develop`, `main`, `release/**` | `go test ./...`, `go vet ./...` |
| `release.yml` | Push de tag `v*` | Build multi-OS + GitHub Release |

## Proteção de branches (enforced)

Ruleset Active: [runtgine-protected-branches](https://github.com/gspaim/Runtgine/rules/20913960)
(`Settings` → `Rules` → `Rulesets`). Cobre `main`, `develop` e `release/*`:

- PR obrigatório (0 approvals; maintainer solo pode mergear)
- Check `test` (GitHub Actions) obrigatório; branch atualizada com a base
- Sem force-push e sem delete
- Sem bypass (nem admin)
- Criar `release/x.y.z` a partir de `develop` é permitido; commits seguintes exigem PR

Restringir PRs para `main` só a partir de `release/*` é **processo** (o GitHub
não filtra a branch de origem). O ruleset é configuração do remoto, não do git.

Requer repo **público** no plano Free; se o repositório voltar a privado,
esta proteção some até haver GitHub Pro.

## Relação com segurança

Após a primeira release estável, vulnerabilidades devem citar a tag afetada.
Até lá, a política em `SECURITY.md` cobre o tip de `main` e o tip de `develop` (pré-release).
