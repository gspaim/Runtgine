# 10 — Gaps e definicoes faltantes

Inventario do que a documentacao ainda nao fecha para iniciar o Core.
Complementa `04-decisoes.md` e `09-mvp.md`.

Status deste doc: inventario oficial. Itens P0 do protocolo foram
confirmados em [11-protocolo-v0.md](11-protocolo-v0.md) / `04-decisoes.md`.
P1+ ainda abertos.

---

## Como ler

| Severidade | Significado |
|---|---|
| P0 | Bloqueia codigo util do MVP Core |
| P1 | Bloqueia cenario Board / LLM vertical |
| P2 | Importante pos-Core; nao bloqueia `runtgine run` minimo |
| P3 | Futuro / ecossistema |

Em conflito de escopo, prevalece `09-mvp.md` + `04-decisoes.md`.

---

## P0 — Bloqueantes do MVP Core

| ID | Gap | Situacao atual | Proposta |
|---|---|---|---|
| G-01 | Task IR v0 schema | Conceito HYPOTHESIS; MVP exige | `11` § Task IR |
| G-02 | Player Manifest v0 | So exemplo verbal | `11` § Manifest |
| G-03 | Event envelope | Event CONFIRMED sem campos | `11` § Event |
| G-04 | Taxonomia minima de eventos | Nenhum catalogo | `11` § Event types |
| G-05 | Capability naming | So exemplos ad hoc | `11` § Capabilities |
| G-06 | Shell Player contract + sandbox | Player no MVP sem I/O nem limites | `11` § Shell Player |
| G-07 | Protocolo Entry Point → Core | Aberto em `00-rascunho` | `11` § Core API |
| G-08 | Result / Error model | Criterio de sucesso sem schema | `11` § Result |
| G-09 | Estados Task / Run | Sem maquina de estados | `11` § Run lifecycle |
| G-10 | Orchestrator no MVP | Em todo fluxo; status HYPOTHESIS | `11` § Runner v0 |
| G-11 | Execution Plan no MVP | CONFIRMED sem schema; sem Intent Engine | `11` § Plan v0 (passthrough) |
| G-12 | Queue semantica | CONFIRMED sem FIFO/prioridade/persistencia | `11` § Queue v0 |
| G-13 | Persistencia no MVP | SQLite CONFIRMED na stack; ausente do incluso MVP; P1 no PRD | `11` § Persistence |
| G-14 | JSON vs YAML | Stack = JSON Schema; MVP cita YAML | `11` § Encoding |
| G-15 | Driver SQLite | mattn vs modernc | `11` § Stack openers |
| G-16 | Logger | slog HYPOTHESIS | `11` § Stack openers |
| G-17 | Layout de pacotes Go | So em fonte historica | `11` § Repo layout |
| G-18 | Policy minima do Shell | Execution Policy fora do MVP; Shell inseguro sem corte | `11` § Shell Player |

### Tensoes meta (P0)

- Task IR / Validator / Orchestrator sao **HYPOTHESIS** mas **obrigatorios no MVP** → promover corte v0 a CONFIRMED apos revisao de `11`.
- **Capability Resolver** e **Planner** aparecem em `01-visao` sem entrada em `02-conceitos`.
- **Event Store** na arquitetura vs MVP “sem event sourcing” → esclarecer store minimo vs sourcing.

---

## P1 — Cenario Board / pipeline vertical

| ID | Gap | Notas |
|---|---|---|
| G-20 | Card GitHub Projects → Task IR | **CONFIRMED** — ver `12-board-p1.md` |
| G-21 | Write-back no board | **CONFIRMED** — status + comentario; sem criar subtasks |
| G-22 | Contratos por etapa do pipeline | **CONFIRMED** — `pipeline.*` + steps lineares; ver `12` |
| G-23 | Fronteira regras vs LLM | **CONFIRMED** — ver tabela em `12-board-p1.md` |
| G-24 | Context assembly basico | **CONFIRMED** — ContextPack v0; ver `12` |
| G-25 | LLM Player v0 | **CONFIRMED** — OpenAI-compat + Anthropic; ver `12` |
| G-26 | Task Router basico | **CONFIRMED** — regras deterministic-first; ver `12` |
| G-27 | Modelo de subtasks | Citado no sucesso do MVP; fora do Task model |

Propostas detalhadas do Board ficam para um doc futuro apos confirmar `11`.
Ate la, o Core deve rodar so com CLI + Shell.

---

## P2 — Engenharia e produto (pos-Core minimo)

| ID | Gap |
|---|---|
| G-30 | Cancelamento, timeout, retry, concorrencia entre runs |
| G-31 | Observabilidade alem da TUI (niveis de log, correlacao) |
| G-32 | Fronteira Runtgine ↔ Chorus (alem de “complementares”) |
| G-33 | Workspaces / worktrees |
| G-34 | Estrategia de testes (unit vs integracao) |
| G-35 | Wails: Svelte vs React |
| G-36 | NATS / Event Bus distribuido (OPEN QUESTION) |
| G-37 | Modulo path Go + versao minima de Go |
| G-38 | Config do runtime (arquivo, env, defaults) |

---

## P3 — Futuro / ecossistema

| ID | Gap |
|---|---|
| G-40 | Workflow Templates loading (nativo vs repo externo) — ver `08` |
| G-41 | Biblioteca ampla de Players |
| G-42 | Human-in-the-loop / Approvals |
| G-43 | Resource Claims / Blast Radius |
| G-44 | MCP integration |
| G-45 | API HTTP / webhooks |

---

## Ordem para fechar gaps

1. Revisar e confirmar [11-protocolo-v0.md](11-protocolo-v0.md)
2. Promover itens aceitos em `04-decisoes.md` (HYPOTHESIS → CONFIRMED v0)
3. Implementar Core na ordem de `09-mvp.md` / `AGENTS.md`
4. So entao especificar Board/LLM (P1) em doc dedicado

## Criterio de “pronto para codar”

**P0 do protocolo v0: CONFIRMADO** (sessao interativa; ver `04-decisoes` / `11`).

Pode iniciar implementacao do Core. Board/LLM (G-20+) ainda precisam de
especificacao antes do cenario vertical.
