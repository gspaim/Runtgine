# 16 — Project Memory (esboço)

Memória episódica / de projeto entre runs e entre “agentes”
(Players LLM ou harnesses externos) no **mesmo** workspace.

Inventário: [10-gaps.md](10-gaps.md) (G-46, G-47; depende de G-44).
Autoridade de status: [04-decisoes.md](04-decisoes.md).

**Status deste doc: OPEN / HYPOTHESIS.** Não autoriza código.
Referência externa de inspiração (não dependência):
[akitaonrails/ai-memory](https://github.com/akitaonrails/ai-memory).

---

## 1. Problema

O ContextPack v0 (G-24) monta contexto **intra-run**: task, step,
`prior_outputs`, `repo_hits`, budget.

Falta continuidade **entre runs** e entre sessões/harnesses no mesmo
projeto: decisões já tomadas, abordagens que falharam, handoff
“onde paramos”, preferências de projeto.

Sem isso, cada LLM Player (ou agente externo) redescobre o que o
runtime e o time já aprenderam.

---

## 2. Três memórias (não misturar)

| Memória | Papel | Status hoje |
|---|---|---|
| **Temporal** | Event Store / SQLite — o que aconteceu em runs | CONFIRMED (G-13) |
| **Estrutural** | Runtime Graph — o que existe e como se relaciona | HYPOTHESIS |
| **Episódica / de projeto** | Project Memory — decisões, falhas, handoffs, briefs | **HYPOTHESIS (este doc)** |

Project Memory **não** substitui Event Store nem Runtime Graph.
Alimenta o Context Engine / ContextPack com *conhecimento acumulado
do projeto*, com budget e autoridade explícitos.

---

## 3. Norte (alinhamento)

- Core = produto; memória é **fonte de contexto**, não runtime.
- Deterministic-first: recall lexical/índice antes de LLM de consolidação.
- Player ≠ Agent: memória não cria “agente com alma”; Players consomem
  capabilities `memory.*` ou o Core injeta hits no ContextPack.
- Validação antes da execução: memória **sugere**; Registry / Validator
  / policies **autorizam**.
- LLM-agnostic: provedor de memória é plugável (sidecar ou Player).

---

## 4. Fora de escopo (explícito)

- Embutir o binário `ai-memory` (Rust) no Core Go.
- Transformar Runtgine em framework de agentes / chatbot.
- RAG genérico como produto principal.
- Plugin system amplo (já fora do MVP em `09-mvp.md`).
- Autorizar capabilities ou bypassar Validator via wiki/página.
- Substituir Event Store por “notas de chat”.

---

## 5. Proposta de encaixe (fases)

### Fase A — Sidecar externo (operacional, sem Core)

Usar `ai-memory` (ou equivalente) **ao lado** do repo para o time e
harnesses (Cursor, Claude Code, Codex, …). Zero mudança no Core.
Útil já; não é feature do produto Runtgine.

### Fase B — Integração via MCP (G-44) → ContextPack

Runtgine como **cliente MCP** de um servidor de memória:

1. Antes de steps LLM, `AssembleContext` (ou sucessor) consulta
   `memory.query` / handoff com escopo = workspace/projeto.
2. Hits limitados entram no ContextPack como campo novo
   (`memory_hits` ou `project_briefs`), com budget próprio.
3. Após run relevante, opcional: emitir observação sanitizada
   (summary/resultado/decisões) para o servidor — **não** transcript cru.

Depende de **G-44** especificado (cliente MCP mínimo no Core ou em
Player dedicado).

### Fase C — Memory Player + captura nativa de runs

Player com capabilities `memory.query` / `memory.record` /
`memory.handoff` (G-47). Pode encapsular MCP ou store local.

Captura nativa: o Runner, em eventos `run.succeeded` / `run.failed`
(e opcionalmente step críticos), grava observação **estruturada**
derivada de Result + intent — alinhada à filosofia “compile, not
retrieve” do ai-memory, mas com schema Runtgine.

Consolidação (wiki markdown / páginas) pode permanecer no sidecar;
o Core só precisa de contrato de I/O estável.

### Fase D — Context Engine completo (já HYPOTHESIS)

Project Memory vira uma das fontes oficiais do Context Engine:

`Task + Relevant Events + Symbols + Resources + Previous Decisions
+ Current State + **Project Memory hits**`

Runtime Graph continua ortogonal (relações estruturais).

---

## 6. Contratos esboçados (não confirmados)

### 6.1 ContextPack — extensão candidata

```text
memory_hits[]:
  - id / path
  - kind        (decision | failure | handoff | preference | other)
  - title
  - snippet     (bounded)
  - score       (opcional)
  - source      (mcp | player | local)
budget.memory_max_chars / memory_max_hits
```

Regras candidatas:

- Truncamento determinístico se exceder budget.
- Escopo default = projeto atual (workspace_root / `.runtgine/`).
- Memória nunca eleva privilégio nem altera o Plan sozinha.

### 6.2 Capabilities candidatas (G-47)

| Capability | Intenção |
|---|---|
| `memory.query` | Busca limitada no escopo do projeto |
| `memory.record` | Grava observação / outcome estruturado |
| `memory.handoff` | Lê ou aceita brief “onde paramos” |

Manifest: Player `kind` deterministic ou AI conforme backend;
I/O JSON Schema; timeouts curtos; falha de memória **não** derruba
o run (degradação: ContextPack sem `memory_hits`).

### 6.3 Captura (sanitização)

Inspiração ai-memory, adaptada ao runtime:

- Corpos limitados (KiB caps).
- Sem secrets / tokens / env bruto.
- Preferir Result estruturado + `intent.summary` a logs completos.
- Opt-in por config (`memory.capture = off | outcomes | outcomes+steps`).

---

## 7. Gaps

| ID | Gap | Severidade | Notas |
|---|---|---|---|
| G-44 | MCP integration | P3 | Pré-requisito da Fase B |
| G-46 | Project Memory (conceito + ContextPack) | P3 | Este doc |
| G-47 | Memory Player / capabilities `memory.*` | P3 | Fase C; após G-46 |

Ordem sugerida para fechar:

1. Promover ou rejeitar este esboço em `04-decisoes.md`
2. Especificar G-44 (cliente MCP mínimo) se Fase B for o caminho
3. Confirmar schema `memory_hits` + budget (extensão G-24)
4. Confirmar G-47 (Manifest + I/O)
5. Só então implementar

---

## 8. Critérios de aceite (quando CONFIRMED)

- Docs `04` / `10` / glossário alinhados; este doc deixa de ser só esboço.
- Decisão explícita: sidecar MCP vs store local vs ambos.
- Degradação segura sem servidor de memória.
- Autoridade: memória não bypassa Validator / Registry.
- Testes: AssembleContext com hits mockados; Player com backend fake.

---

## 9. Decisão proposta (para `04`)

| Decisão | Status proposto | Notas |
|---|---|---|
| Project Memory (episódica / de projeto) | HYPOTHESIS | Continuação entre runs no mesmo projeto |
| Três memórias distintas (temporal / estrutural / episódica) | HYPOTHESIS | Evitar colapso conceitual |
| Integração inicial via MCP sidecar (ex.: ai-memory) | HYPOTHESIS | Não embutir no Core; G-44 |
| Extensão ContextPack com `memory_hits` | HYPOTHESIS | Após G-24; budget próprio |
| Memory Player (`memory.*`) | HYPOTHESIS | G-47 |
| Embutir ai-memory no Core | REJECTED (proposto) | Domínio e stack ortogonais |
| Memória como autoridade de execução | REJECTED (proposto) | Só contexto |

Até promoção formal, **não codificar** Fases B–D.
