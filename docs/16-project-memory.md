# 16 — Project Memory (esboço)

Memória **episódica** / de projeto entre runs e entre “agentes”
(Players LLM ou harnesses externos) no **mesmo** workspace.

Inventário: [10-gaps.md](10-gaps.md) (G-46, G-47; depende de G-44).
Autoridade de status: [04-decisoes.md](04-decisoes.md).

**Status deste doc: OPEN / HYPOTHESIS.** Não autoriza código.
Não promove contratos a API definitiva. Não é RAG genérico.
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
| **Estrutural** | Runtime Graph — o que existe e como se relaciona | CONFIRMED v0 (`18`); hits ContextPack em `19` |
| **Episódica / de projeto** | Project Memory — episódios: decisões, falhas, handoffs, briefs | **HYPOTHESIS (este doc)** |

Project Memory **não** substitui Event Store nem Runtime Graph.
Alimenta o Context Engine / ContextPack com *memória episódica do
projeto*, com budget, hierarquia e autoridade explícitos.

### 2.1 “Isso aconteceu” ≠ “isso ainda é válido”

Fato histórico e orientação operacional são camadas distintas:

| Camada | Exemplo | Mutável? |
|---|---|---|
| Fato histórico | “Em T1 decidimos usar Redis.” | Não (append-only em espírito) |
| Status operacional | “Redis ainda é a orientação vigente?” | Sim |

Após migração para Valkey, a memória Redis continua **historicamente
verdadeira**, mas não deve competir como contexto operacional default.

**Modelo candidato de validade** (HYPOTHESIS — não API):

| Status | Significado | Recall operacional default |
|---|---|---|
| `active` | Orientação vigente | Sim — preferido em `memory_hits` |
| `superseded` | Substituído (link conceitual a sucessor) | Não — só query histórica / “por que mudamos?” |
| `archived` | Retido, fora do recall default | Não |

Alternativa mais leve (também HYPOTHESIS): dois layers de recall —
`operational` vs `historical` — sem enum rígido até experimentos.

Supersession / arquivo deve ser **explícita e opt-in** (humano, regra
determinística ou fluxo dedicado). Inferência LLM silenciosa no Core
que “atualiza a verdade” é rejeitada conceitualmente: viraria
autoridade disfarçada.

---

## 3. Memory ≠ Knowledge (evolução possível)

Não criar subsistema agora. Registrar a distinção para não colapsar
conceitos:

| Conceito | Natureza | Exemplo |
|---|---|---|
| **Project Memory** | Episódica / histórica | “Tentamos X e falhou por Y.” |
| **Project Knowledge** | Consolidada (futuro possível) | “Este projeto utiliza X por causa de Y.” |

Pipeline conceitual desejado (compile, not retrieve):

```text
Event Store (runtime)
  → observações estruturadas (sanitizadas, bounded)
  → Project Memory (episódica)
  → [possível] Project Knowledge (consolidado)
  → Context Engine escolhe o relevante
  → ContextPack
```

Contraste rejeitado como produto: indexar logs/transcripts e fazer
retrieval genérico (RAG). Runtgine permanece deterministic-first:
compilar observações úteis, não arquivar chat.

Project Knowledge, se um dia existir, pode viver em ADRs/docs humanos,
wiki do sidecar, ou camada Runtgine — **OPEN QUESTION**; fora de G-46.

---

## 4. Norte (alinhamento)

- Core = produto; memória é **fonte de contexto**, não runtime.
- Deterministic-first: recall lexical/índice antes de LLM de consolidação.
- Player ≠ Agent: memória não cria “agente com alma”.
- **Memory Provider** (HYPOTHESIS): abstração de fonte consultada pelo
  AssembleContext / Context Engine — sidecar MCP, store local, stub.
- **Memory Player** (OPEN QUESTION): só se evidência mostrar necessidade
  de steps do Plan com capabilities `memory.*`. Não é default.
- Validação antes da execução: memória **sugere**; Registry / Validator
  / policies **autorizam**.
- LLM-agnostic: o Provider é plugável.

### 4.1 Cadeia de autoridade (obrigatória)

```text
Memory (Provider)
  → Context Engine / AssembleContext
  → Player interpreta contexto
  → Plan
  → Validator / Registry / Policies
  → Runtime executa
```

Memória **nunca**:

- concede capability;
- altera permissões;
- bypassa Validator;
- altera Policy;
- autoriza execução “porque algo semelhante ocorreu antes”;
- transforma “evitar falha X” em enforcement automático (isso seria
  Policy; memória só informa).

---

## 5. Fora de escopo (explícito)

- Embutir o binário `ai-memory` (Rust) no Core Go.
- Transformar Runtgine em framework de agentes / chatbot.
- RAG genérico como produto principal.
- Plugin system amplo (já fora do MVP em `09-mvp.md`).
- Autorizar capabilities ou bypassar Validator via wiki/página.
- Substituir Event Store por “notas de chat”.
- Inferência silenciosa de supersession no Core.
- Cristalizar schema/API antes dos experimentos (Fases A/B).

---

## 6. Proposta de encaixe (fases)

### Fase A — Sidecar externo (operacional, sem Core)

Usar `ai-memory` (ou equivalente) **ao lado** do repo para o time e
harnesses (Cursor, Claude Code, Codex, …). Zero mudança no Core.
Útil já; não é feature do produto Runtgine.

**Objetivo empírico:** descobrir kinds úteis, queries reais, ruído vs
ganho, formato de observação — *antes* de APIs no Core.

### Fase B — Integração via Provider (MCP G-44) → ContextPack

Runtgine consulta um **Memory Provider** (candidato: cliente MCP):

1. Antes de steps LLM, `AssembleContext` (ou sucessor) pede hits /
   handoff com escopo = workspace/projeto, preferindo status
   operacional `active`.
2. Hits limitados entram no ContextPack (`memory_hits` candidato),
   com budget **próprio** e prioridade **menor** que task/estado atual.
3. Após run relevante, opcional: emitir observação sanitizada
   (summary/resultado/decisões) — **não** transcript cru.

Depende de **G-44** se o transporte for MCP. Alternativa OPEN QUESTION:
arquivos/briefs lidos pelo AssembleContext sem MCP no Core.

Falha do Provider **não** derruba o run (degradação: sem `memory_hits`).

### Fase C — Captura nativa + modelo de acesso (G-47)

Captura nativa candidata: Runner, em `run.succeeded` / `run.failed`
(e opcionalmente steps críticos), grava observação **estruturada**
derivada de Result + intent (“compile, not retrieve”).

**Acesso:**

1. Default HYPOTHESIS: **Memory Provider** injetando no ContextPack.
2. OPEN QUESTION: **Memory Player** com `memory.query` /
   `memory.record` / `memory.handoff` **somente** se steps do Task IR
   precisarem dessa capability explicitamente.

Não transformar “tudo que o Core consulta” em Player.

Consolidação (wiki) pode permanecer no sidecar; Core só precisa de
contrato de I/O estável *depois* dos experimentos.

### Fase D — Context Engine completo (já HYPOTHESIS)

Project Memory (e, se existir, Knowledge) como fontes oficiais:

`Task + Relevant Events + Symbols + Resources + Previous Decisions
+ Current State + **Project Memory hits (operacionais)**`

Runtime Graph continua ortogonal (relações estruturais).

---

## 7. ContextPack — hierarquia e budget (rascunho experimental)

**Não é contrato candidato até pós-Fase B.** Campos abaixo são
vocabulário de discussão.

### 7.1 Extensão candidata

```text
memory_hits[]:
  - id / path
  - kind        (decision | failure | handoff | preference | other)
  - validity    (active | superseded | archived)   # candidato
  - title
  - snippet     (bounded)
  - score       (opcional)
  - source      (mcp | provider | local)
budget.memory_max_chars / memory_max_hits
```

### 7.2 Hierarquia de prioridade (HYPOTHESIS)

Ordem conceitual de preservação sob truncamento:

1. `task` + `step` (intenção atual)
2. `prior_outputs` / estado do **run atual**
3. `repo_hits` (evidência do repo agora)
4. `memory_hits` operacionais (`active`)
5. `memory_hits` históricos / `superseded` (só sob necessidade)

Regras candidatas:

- Budget próprio; limite de hits; truncamento **determinístico**.
- Histórico não compete de forma ilimitada com o operacional.
- Escopo default = projeto atual (`workspace_root` / `.runtgine/`).
- Memória nunca eleva privilégio nem altera o Plan sozinha.

### 7.3 Capabilities `memory.*` (OPEN QUESTION — G-47)

Só relevantes se Memory Player for justificado por evidência:

| Capability | Intenção |
|---|---|
| `memory.query` | Busca limitada no escopo do projeto |
| `memory.record` | Grava observação / outcome estruturado |
| `memory.handoff` | Lê ou aceita brief “onde paramos” |

### 7.4 Captura (sanitização)

- Corpos limitados (KiB caps).
- Sem secrets / tokens / env bruto.
- Preferir Result estruturado + `intent.summary` a logs completos.
- Opt-in por config (`memory.capture = off | outcomes | outcomes+steps`).

---

## 8. Experimentos antes de cristalizar

Fases A/B são **aprendizado**, não estabilização de schema.

Medir empiricamente:

- quais tipos de memória são úteis;
- quais consultas aparecem;
- quanto recall é necessário;
- o que persistir vs descartar;
- quando memória melhora o contexto;
- quando adiciona ruído;
- qual formato de observação funciona;
- se Provider no AssembleContext basta (vs Player).

**Não** transformar hipóteses em APIs definitivas antes disso.
**Não** promover G-46/G-47 a CONFIRMED só por existir este doc.

---

## 9. Cenários de validação (modelo)

### 9.1 Decisão Redis → Valkey

| Camada | Conteúdo |
|---|---|
| Event Store | Runs, steps, resultados da decisão e da migração |
| Runtime Graph | Dependência **atual** (serviço → Valkey), se existir |
| Project Memory | Episódio Redis (depois `superseded`); episódio Valkey (`active`) |
| ContextPack | Task + repo_hits (código Valkey) + memory operacional Valkey; Redis só se query histórica |
| Validator / Registry / Policy | Mandam capabilities/sandbox; memória não libera nem bloqueia stacks |
| Player recebe | Intenção + evidência do repo + orientação `active` |
| Memória indisponível | Run segue; repo_hits ainda mostram Valkey; perde continuidade narrativa |

### 9.2 Falha de execução a compreender / evitar

| Camada | Conteúdo |
|---|---|
| Event Store | `step.failed`, result/stderr, run_id (verdade temporal) |
| Runtime Graph | Em geral nada (falha não é relação estrutural) |
| Project Memory | Episódio compilado: “flag F quebra; preferir G” (`failure`, `active`) |
| ContextPack | Entre runs: hit de failure capped; no mesmo run: `prior_outputs` |
| Validator / Registry / Policy | Continuam mandando; memória **não** proíbe reexecutar o comando ruim |
| Player recebe | Resumo bounded da falha anterior — interpretação, não enforcement |
| Memória indisponível | Pode repetir falha; Event Store/TUI ainda permitem inspeção humana |

### 9.3 Handoff entre runs / Players

| Camada | Conteúdo |
|---|---|
| Event Store | Run interrompido, outputs parciais persistidos |
| Runtime Graph | N/A para handoff episódico |
| Project Memory | Brief `handoff` `active` até consumido/arquivado |
| ContextPack | Brief + task + artefatos relevantes; não o run inteiro |
| Validator / Registry / Policy | Novo run valida do zero; handoff não herda privilégios |
| Player recebe | “Onde paramos” bounded |
| Memória indisponível | Degradação esperada; recuperação via Event Store / TUI / child runs |

Leitura: handoff encaixa melhor como **insumo do Provider → ContextPack**
do que como capability default de Player.

---

## 10. Gaps

| ID | Gap | Severidade | Notas |
|---|---|---|---|
| G-44 | MCP integration | P3 | Candidato a transporte da Fase B |
| G-46 | Project Memory (conceito + ContextPack + validade + hierarquia) | P3 | Este doc; perguntas abertas abaixo |
| G-47 | Modelo de acesso: Memory Provider vs Memory Player | P3 | Provider = HYPOTHESIS; Player = OPEN QUESTION |

Perguntas abertas em G-46 (não fechar sem evidência):

- modelo de validade (`active`/`superseded`/`archived` vs layers);
- kinds realmente úteis;
- quando capturar (outcomes vs steps);
- Memory vs futura Knowledge.

Ordem sugerida:

1. Experimentar Fase A (sidecar) e registrar achados
2. Decidir se Fase B precisa de G-44 ou outro Provider
3. Só então candidatar extensão ContextPack (ainda HYPOTHESIS)
4. Resolver G-47 (Provider vs Player) com evidência
5. Implementar somente após promoção explícita em `04`

---

## 11. Critérios de aceite (quando CONFIRMED)

- Docs `04` / `10` / glossário alinhados; este doc deixa de ser só esboço.
- Decisão explícita: sidecar MCP vs store local vs ambos.
- Decisão explícita: Provider default; Player só se justificado.
- Modelo de validade (ou layers) documentado.
- Hierarquia ContextPack + degradação sem memória.
- Autoridade: lista negativa (§4.1) preservada.
- Testes mínimos: AssembleContext + Provider stub (Player não obrigatório).

---

## 12. Decisão proposta (para `04`)

| Decisão | Status proposto | Notas |
|---|---|---|
| Project Memory (episódica / de projeto) | HYPOTHESIS | Continuação entre runs (G-46) |
| Três memórias: temporal / estrutural / episódica | HYPOTHESIS | Evitar colapso |
| Fato histórico ≠ status operacional (validade) | HYPOTHESIS | `active` / `superseded` / `archived` ou layers |
| Memory ≠ Knowledge (evolução possível) | HYPOTHESIS | Sem subsistema agora |
| Memory Provider → ContextPack | HYPOTHESIS | Default conceitual de acesso |
| Integração inicial via sidecar / MCP | HYPOTHESIS | Fase A/B; experimentos antes de API |
| Extensão ContextPack (`memory_hits` + budget + hierarquia) | HYPOTHESIS | Rascunho experimental; pós-evidência |
| Memory Player (`memory.*`) | OPEN QUESTION | Só se steps do Plan exigirem (G-47) |
| Embutir ai-memory no Core | REJECTED | Domínio e stack ortogonais |
| Memória como autoridade de execução | REJECTED | Sugere contexto; lista negativa §4.1 |
| Supersession silenciosa via LLM no Core | REJECTED | Opt-in explícito apenas |
| RAG genérico / indexar transcripts como produto | REJECTED | Compile observations, not chat |

Até promoção formal, **não codificar** Fases B–D.
Fase A (sidecar externo) pode ser usada operacionalmente pelo time
sem mudança no repositório Runtgine.
