# 45 — Postgres EXPLAIN (recorte SQL read-only)

Recorte de G-41 que fecha a parte **SQL** da fila: nova capability
`pg.explain` no Player `postgres` (`41`). **Levanta explicitamente**
a exclusão "sem SQL no Task IR" (G-206/G-209/G-216) — somente para
**EXPLAIN sem ANALYZE**, com validação estática conservadora:

| Player | Pacote | Capabilities v0 |
|---|---|---|
| `postgres` | `internal/players/pg` | `pg.ping` (existente), `pg.explain` (nova) |

Inventário: [10-gaps.md](10-gaps.md) (G-231+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é `EXPLAIN ANALYZE` (executa). Não é SELECT de dados (linhas).
Não é INSERT/UPDATE/DELETE/DDL/DCL. Não é migrations (`goose`,
`golang-migrate`, `flyway`). Não é `pg_dump` / `pg_restore`. Não é
outro SGBD (MySQL/SQLite). Não é NATS (G-36). Não é MCP.

**Status deste doc: CONFIRMED v0 (slice 38 feito).** G-231..G-237.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-09-04-045-pg-explain/`](../openspec/changes/archive/2026-09-04-045-pg-explain/)
(arquivado após o slice 38). Spec atual: [`openspec/specs/pg-explain/`](../openspec/specs/pg-explain/).

---

## 1. Problema

"Intenção → execução verificável" não tem verificação de consulta:
o desenvolvedor não consegue pedir o **plano** de um SELECT pelo
runtime. O corte SQL completo (executar queries, migrations) exige
execução de código arbitrário no banco e continua fora; o `EXPLAIN`
entrega o valor de verificação (plano, custo, índices usados) **sem
executar** a consulta.

Por que é seguro levantar a exclusão aqui:

1. `EXPLAIN (FORMAT JSON)` sem `ANALYZE` é **planner puro** — não
   executa a consulta, não lê dados, não adquire locks de dados.
2. A SQL passa por **validação estática conservadora** (G-232) antes
   do argv: prefixo obrigatório `SELECT`/`WITH`, proibição de `;` e
   de `\` em qualquer posição.
3. O Player constrói o prefixo `EXPLAIN (FORMAT JSON) ` — a SQL do
   IR nunca é o comando inteiro.
4. `psql -c` recebe **um elemento argv** (nunca shell string) e o
   texto proibido de `\` elimina meta-comandos do psql (`\!`), que
   só são reconhecidos no início de um buffer de comando — impossível
   com prefixo `SELECT`/`WITH` e sem `;`.
5. Credenciais seguem o modelo de `pg.ping`: só ambiente herdado
   (`PGPASSWORD`/`PGSSLMODE`), nunca no Task IR.

Risco residual assumido (v0): o validador é tokenizador simples, não
parser SQL completo — rejeita por excesso (ex.: `;` dentro de
comentário/string é rejeitado) e strings citadas com_case diferente
são baixadas pela heurística NL. Falso-positivo é aceitável;
falso-negativo para os casos listados não é (prefixo + `;` + `\`
fecham os vetores conhecidos).

---

## 2. Fronteiras

| É | Não é |
|---|---|
| `EXPLAIN (FORMAT JSON) <SELECT/WITH>` | `EXPLAIN ANALYZE` / `EXECUTE` |
| Planner puro (sem leitura de dados) | SELECT retornando linhas |
| SQL única, `SELECT`/`WITH`, sem `;` sem `\` | Múltiplos statements, meta-comandos psql |
| Credenciais só env herdado | `password`/`connstring` no Task IR |
| `psql` argv, flags fechadas (como `pg.ping`) | Outro SGBD / migrations / dump |

Regras:

1. Validator / Registry / Policy soberanos.
2. Nunca shell string. Só argv.
3. Capability `allow` (planner puro, sem custo de execução).
4. Unit tests injetam o runner; CI **não** exige postgres.

---

## 3. Cortes confirmados (G-231+)

### G-231 — Capability `pg.explain`

| Campo | Default | Regra |
|---|---|---|
| `sql` | obrigatório | ver G-232; teto 10 KiB |
| `dbname` | obrigatório | `safeRef` |
| `host` | `127.0.0.1` | `safeRef` |
| `port` | `5432` | 1–65535 |
| `user` | omitido | `safeRef` se presente |
| `timeout_ms` | 10000 | teto 60000 (herda limites do player) |

Saída: `plan` (JSON parseado de `EXPLAIN (FORMAT JSON)`),
`truncated` (string bruta truncada em 4096 runes se não parsear).

### G-232 — Validação estática da SQL (allowlist conservadora)

Ordem de checagem, qualquer falha → `validation.invalid_input`:

1. `sql` trim não vazio; ≤ 10 KiB.
2. Primeira palavra (case-insensitive) é `SELECT` ou `WITH`.
3. Nenhum `;` em qualquer posição (statement único; falso-positivo
   em string/comentário aceito).
4. Nenhuma `\` em qualquer posição (meta-comandos psql impossíveis).

O Player envia `"EXPLAIN (FORMAT JSON) " + sql` como **um** argv de
`--command`. A SQL do IR nunca contém o prefixo EXPLAIN (rejeitado
em 2).

### G-233 — Conexão + argv

Argv idêntico ao `pg.ping` (G-205), trocando o comando:
`psql --host --port --dbname [--username] --no-psqlrc -t -A --pset
pager=off --command "EXPLAIN (FORMAT JSON) <sql>"`.
Senha/SSL só via ambiente herdado (`PGPASSWORD`, `PGSSLMODE`) —
mesma allowlist de env de `41`; nunca no Task IR; nunca `RUNTGINE_*`.

### G-234 — Falha vs sucesso

- exit 0 → `plan` parseado (JSON array do EXPLAIN)
- exit ≠ 0 (SQL inválida, permissão, conexão) / binário ausente →
  `runtime.player_error` com a mensagem do stderr
- timeout → `runtime.timeout`
- Testes injetam runner; `go test ./...` sem postgres

### G-235 — Registry + admission + Graph

- `pg.Manifest` ganha `pg.explain` (schema com `sql` — única
  capability com SQL no IR, por decisão deste recorte)
- Runner admission inclui `pg.explain` na validação estática
- Graph: sem nó novo (`RefreshFromRegistry` já reflete a capability)

### G-236 — Intent

| NL | Capability | Method |
|---|---|---|
| `explain select <query>` | `pg.explain` | `heuristic.pg` |
| `explain with <query>` | `pg.explain` | `heuristic.pg` |

A heurística remove o prefixo `explain ` e emite `sql` com o resto
(normalizado: minúsculas, espaços colapsados — identificadores não
citados são case-insensitive no Postgres). `dbname` default
`postgres` (mesmo padrão do `pg.ping`; operador ajusta via Task IR).
`explain analyze select ...` **não** casa a capability `pg.explain`
pela validação estática se chegar via Task IR; pela NL, prefixo
obrigatório é `explain select ` / `explain with `.

### G-237 — Exclusões v0

- `EXPLAIN ANALYZE`, `EXECUTE`, `SELECT` retornando linhas
  (`pg.query`/`pg.exec` seguem não registrados)
- INSERT/UPDATE/DELETE/TRUNCATE/DDL/DCL (bloqueados por prefixo)
- Migrations (`goose`, `golang-migrate`, `flyway`, `prisma`),
  `pg_dump`/`pg_restore`, `COPY`
- Outros SGBDs (MySQL/SQLite/MSSQL); `connstring`/URI no IR
- NATS (G-36, DEFERRED); MCP; templates (`40`)

---

## 4. Critérios de aceite

1. Manifest registra `pg.ping` e `pg.explain`; **não** registra
   `pg.query`/`pg.exec`.
2. `sql` com `;`, com `\`, vazio, > 10 KiB ou sem prefixo
   `SELECT`/`WITH` → `validation.invalid_input`.
3. `sql` com `EXPLAIN ANALYZE` → rejeitado (prefixo não é
   `select`/`with`).
4. Argv contém `--command` com valor prefixado por
   `EXPLAIN (FORMAT JSON) `.
5. `runtgine intent --dry-run "explain select id from users"` →
   `pg.explain`, method `heuristic.pg`.
6. Fake exec exit 1 → `runtime.player_error`.
7. `go test ./...` / `go vet ./...` verdes sem postgres.
8. OpenSpec `045` arquivado após o código (slice 38).

---

## 5. Ordem do slice de código

1. `pg.Manifest` + `ValidateStaticInput` (G-231..G-233)
2. Runner admission (`pg.explain`)
3. Heurística Intent (`explain select|with`)
4. Example `pg-explain.json`
5. Testes fake; README Estágio: Slice 38
6. Arquivar OpenSpec `045`

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-231 Capability `pg.explain`
- [x] G-232 Validação estática (SELECT/WITH; sem `;` sem `\`)
- [x] G-233 Conexão + argv (credenciais só env)
- [x] G-234 Falha vs sucesso
- [x] G-235 Registry + admission + Graph
- [x] G-236 Intent (`heuristic.pg`)
- [x] G-237 Exclusões v0
