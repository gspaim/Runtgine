# 33 — Evolution v0: Router, Playbooks, Lessons

Evolução P3 do Runtgine: roteamento inteligente de LLM, playbooks/skills
de projeto e loop de auto-melhoria assistida — **sem** virar framework de
agentes autônomos.

Inventário: [10-gaps.md](10-gaps.md) (G-147+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Referências: [08-workflow-templates.md](08-workflow-templates.md),
[12-board-p1.md](12-board-p1.md) (effort/difficulty), [16-project-memory.md](16-project-memory.md),
[29-project-memory-v0.md](29-project-memory-v0.md), [02-conceitos.md](02-conceitos.md).

**Status deste doc: CONFIRMED (v0 spec).** Código slices 22–24 **feito**.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-21-033-evolution-v0/`](../openspec/changes/archive/2026-08-21-033-evolution-v0/).

---

## 1. Problema

Hoje o Runtgine já tem:

- pipeline Board com `pipeline.effort` / `pipeline.difficulty` (heurístico);
- LLM Player com **um** backend default por workspace (`openai-compat` | `anthropic`);
- Task Router básico (G-26): capability → deterministic → default AI;
- Project Memory v0 (`memory_hits`) e capture opt-in de falhas;
- Playbooks / Lessons / Auto-sizing só como **HYPOTHESIS** em `08`.

Falta fechar o loop de produto que o time descreveu:

1. **Papéis especializados** (orquestração, dev, QA) como processo repetível
   — não como “agentes conversacionais”.
2. **Skills/playbooks** por projeto, versionados no workspace.
3. **Roteamento** effort/difficulty/tipo de step → provider/modelo adequado.
4. **Auto-melhoria assistida**: analisar falhas, propor episódio ou diff de
   playbook — com **HITL** antes de promover.

---

## 2. Princípios (não negociáveis)

1. **Player ≠ Agent** — especialização = capabilities + steps + playbooks
   ([02-conceitos.md](02-conceitos.md)).
2. **Deterministic-first** — gates, `test.go`, Players determinísticos antes
   de LLM ([03-principios.md](03-principios.md)).
3. **Validator soberano** — playbook/skill vira Task IR; nunca executa direto.
4. **Memória compila observações** — Project Memory episódica; **não** RAG de
   chat ([29](29-project-memory-v0.md)).
5. **Auto-melhoria com HITL** — proposta → revisão humana → `memory.record` ou
   merge de playbook; **sem** mutação silenciosa via LLM no Core.
6. **Runtgine não é framework de agentes** ([01-visao.md](01-visao.md)).

---

## 3. Três pilares

```text
┌─────────────────┐   ┌──────────────────┐   ┌─────────────────────┐
│ Player Router   │   │ Playbooks v0     │   │ Lessons v0          │
│ multi-model     │   │ skills projeto   │   │ postmortem + HITL   │
└────────┬────────┘   └────────┬─────────┘   └──────────┬──────────┘
         │                     │                        │
         └─────────────────────┴────────────────────────┘
                               │
                    Core: Validator, Runner, Memory, Graph
```

| Pilar | Pergunta que responde | Não é |
|---|---|---|
| **Player Router** | Qual provider/modelo para este step? | Agente orquestrador |
| **Playbooks** | Qual processo/skill usar neste projeto? | Cursor rules clone |
| **Lessons** | O que aprendemos com falhas deste run? | Chat log indexer |

---

## 4. Cortes confirmados (G-147+)

### G-147 — Player Router v0 (papel)

**Status: CONFIRMED**

- Evolução do Task Router (G-26) e do item **Player Router** (HYPOTHESIS em `04`).
- Entrada: `capability`, metadados do step (`effort`, `difficulty`, `step_kind`),
  policy, config do workspace.
- Saída: `{ provider_id, model_id, completer_config }` para o LLM Player.
- **Não** escolhe “persona”; escolhe **backend** para uma capability LLM.
- Fallback: backend default atual se nenhuma regra casar.
- Sem benchmark externo obrigatório no v0 — tabela de regras configurável
  primeiro; benchmarks como input opcional (G-148).

### G-148 — Model tiers e config multi-provider

**Status: CONFIRMED**

Config em `.runtgine/config.json` (secrets só env), precedência G-38:

```json
{
  "llm_providers": [
    {
      "id": "openai-main",
      "kind": "openai-compat",
      "base_url": "...",
      "default_model": "gpt-4.1-mini"
    },
    {
      "id": "anthropic-main",
      "kind": "anthropic",
      "default_model": "claude-sonnet-4-20250514"
    }
  ],
  "llm_routing": [
    {
      "match": { "capability_prefix": "pipeline.spec-review" },
      "provider_id": "anthropic-main",
      "model_id": "claude-sonnet-4-20250514"
    },
    {
      "match": { "effort_in": ["S", "M"], "capability_prefix": "pipeline." },
      "provider_id": "openai-main",
      "model_id": "gpt-4.1-mini"
    },
    {
      "match": { "difficulty_gte": 4 },
      "provider_id": "anthropic-main"
    }
  ]
}
```

Regras v0:

- `llm_backend` legado mapeia para provider default até migração explícita.
- Router consulta outputs de `pipeline.effort` / `pipeline.difficulty` quando
  presentes no Run (`12`).
- Pesquisa de benchmarks por tipo de tarefa = **input humano** à tabela
  (documentar rationale em playbook ou config); Core **não** scrapeia leaderboards
  automaticamente no v0.
- Ordem de match: regra mais específica vence; empate → primeira declarada.

### G-149 — Playbooks v0 (skills de projeto)

**Status: CONFIRMED**

- Playbooks = arquivos Markdown (ou YAML frontmatter + body) em
  `.runtgine/playbooks/*.md`.
- Campos mínimos (frontmatter): `id`, `title`, `capabilities[]`,
  `phases[]` opcional, `gates[]` opcional.
- **Não** são Players. São documentação executável consultada por:
  - Intent Engine (sugerir template/steps);
  - Context Engine (trecho capado no ContextPack como `playbook_hits`);
  - operador (referência humana).
- Registry local indexado no boot (best-effort); Graph pode registrar nó
  `playbook` (OPEN — não bloqueia v0).
- Exemplos de ids: `orchestrator`, `developer`, `qa`, `board-pipeline`
  (papéis = **processos**, não agentes nomeados).
- Workflow Templates completos (`08`) permanecem HYPOTHESIS; Playbooks v0 é
  recorte mínimo executável.

### G-150 — Lessons / Postmortem v0

**Status: CONFIRMED**

Loop assistido pós-falha:

```text
run.failed (+ opt-in config)
  → Lessons job (deterministic + optional LLM step)
  → propõe: episódio Project Memory OU diff sugerido em playbook
  → HITL: operador aprova/rejeita (CLI/TUI/Wails)
  → se aprovado: memory.record / arquivo playbook atualizado
```

Componentes v0:

| Peça | O quê |
|---|---|
| Trigger | `run.failed` + `lessons.capture = off \| failures` (default `off`) |
| Input | Event Store slice do run, `intent.summary`, step outputs, stderr |
| Output proposta | JSON: `{ kind, title, body, suggested_playbook_patch? }` |
| Promoção | **Só** após `ApproveLesson(proposal_id)` ou CLI explícita |
| Recall | Episódios aprovados entram em `memory_hits` normais |

Capability candidata (futuro): `pipeline.postmortem` — LLM com schema fixo;
alternativa slice 24: job Core sem capability pública.

Rejeições:

- LLM reescreve playbook/memória sem aprovação;
- indexar transcript de chat;
- “agente QA” autônomo que mergeia sozinho.

### G-151 — Exclusões v0

**Status: CONFIRMED**

| Fica fora | Por quê |
|---|---|
| Framework multi-agente / Agent registry | Contradiz visão Player-centric |
| Personas conversacionais persistentes | Chat ≠ runtime |
| Auto-promoção silenciosa de skills | HITL obrigatório |
| Benchmark crawler automático | Ops/research humano alimenta config |
| Memory Player (`memory.*`) | G-47 OPEN; Provider basta no v0 |
| Workflow Template engine completo | Escopo `08`; Playbooks são recorte |
| NATS / MCP / Wails | Outros tracks P3 |
| HTTP API / `runtgine serve` | Spec `34` (G-45); não é Evolution |

### G-152 — Ordem de implementação e critérios

**Status: CONFIRMED**

| Slice | Entrega | Depende de |
|---|---|---|
| **22** | Player Router + config multi-provider | G-26, LLM Player |
| **23** | Playbooks loader + `playbook_hits` no ContextPack | Context Engine |
| **24** | Lessons postmortem + HITL promote | Project Memory, Policy |

Critérios slice 22:

- Config com 2+ providers; rota por `capability_prefix` e `effort_in`.
- Step `pipeline.tech-review` usa provider roteado; fallback se regra ausente.
- `go test ./...` cobre match/order/fallback.

Critérios slice 23:

- `.runtgine/playbooks/qa.md` indexado; Intent ou pipeline referencia id.
- ContextPack inclui `playbook_hits` (budget dedicado, abaixo de `memory_hits`).
- Validator inalterado — playbook não inventa capability.

Critérios slice 24:

- `run.failed` + `lessons.capture=failures` gera **proposta** pendente.
- Aprovação grava episódio `active`; rejeição descarta.
- Nenhuma escrita em playbook sem segundo HITL ou flag explícita.

---

## 5. Relação com memória compartilhada

Project Memory v0 (**implementado**, slice 17) continua a camada episódica
compartilhada entre runs e harnesses:

- `memory_hits` no ContextPack;
- validade `active` / `superseded` / `archived`;
- capture `failures` opt-in.

Este doc **estende** o uso (Lessons → episódios; Playbooks → contexto
estruturado), **não** substitui Memory nem Graph.

Três memórias ([16](16-project-memory.md)):

| Memória | Evolution v0 usa como… |
|---|---|
| Temporal (events) | Input do postmortem |
| Estrutural (graph) | Contexto de impacto / deps |
| Episódica (memory) | Saída aprovada do Lessons loop |

---

## 6. Mapeamento “agentes” → Runtgine

| Linguagem comum | Modelo Runtgine |
|---|---|
| Agente orquestrador | Runner + Task IR + pipeline template / playbook `orchestrator` |
| Agente desenvolvedor | Steps `git.*`, `fs.*`, `test.go`, `pipeline.decompose` + playbook `developer` |
| Agente QA | `test.go`, gates determinísticos, `pipeline.spec-review` + playbook `qa` |
| Skill especializada | Arquivo `.runtgine/playbooks/<id>.md` |
| Aprendizado contínuo | Lessons v0 → Memory (+ patch playbook com HITL) |
| Roteamento inteligente | Player Router v0 (G-147/G-148) |

---

## 7. Referências

- Task Router atual: [12](12-board-p1.md) G-26
- Effort/difficulty: `internal/players/pipeline` (`pipeline.effort`, `pipeline.difficulty`)
- LLM config: `internal/config`, `internal/players/llm`
- OpenSpec: [`openspec/changes/archive/2026-08-21-033-evolution-v0/`](../openspec/changes/archive/2026-08-21-033-evolution-v0/)
