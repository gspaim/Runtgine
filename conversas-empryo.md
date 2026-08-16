> **HISTÓRICO — NÃO É AUTORIDADE.**  
> Consolidação de discussões (incl. Empryo). Stack e UI neste arquivo
> (Rust, GPUI, etc.) estão **desatualizadas**. Fonte de verdade:
> `docs/04-decisoes.md`, `docs/09-mvp.md` e demais `docs/`.

Claro. Juntando o que definimos até aqui, o **Runtgine** está ficando com uma arquitetura bem clara. E a análise do Empryo ajudou a preencher principalmente a parte de **inteligência estrutural**, sem mudar o coração event-driven que já tínhamos definido.

# Runtgine — resumo das definições

## 1. Visão do projeto

O **Runtgine** é um **runtime/orquestrador de engenharia**, focado em executar tarefas através de componentes chamados **Players**.

A ideia central não é:

> "um sistema de agentes LLM"

e sim:

> **um runtime capaz de receber intenções/tarefas, transformá-las em execução estruturada e coordenar Players determinísticos, LLMs, serviços e humanos através de eventos, filas e capabilities.**

O projeto é separado do **Chorus**:

* **Runtgine** → runtime, execução, Players, recursos, workflows.
* **Chorus** → protocolo/comunicação/orquestração entre componentes.

---

# 2. Princípio fundamental: deterministic-first

Uma das decisões mais importantes:

> **Se uma tarefa pode ser executada deterministicamente, o Runtgine deve preferir isso ao uso de IA.**

Exemplo:

```text
"Execute os testes"
        ↓
Test Player
        ↓
cargo test
```

Não:

```text
"Execute os testes"
        ↓
LLM
        ↓
"acho que deveria rodar cargo test"
```

LLM entra quando existe necessidade de:

* interpretação;
* planejamento;
* raciocínio;
* diagnóstico;
* geração;
* revisão;
* decisão complexa.

Isso reduz custo, latência, imprevisibilidade e erros.

---

# 3. Player é a abstração central

**Player não é sinônimo de Agent.**

Um Player é qualquer entidade capaz de fornecer determinadas **capabilities**.

Pode ser:

```text
Deterministic Player
AI Player
Human Player
Service Player
Workflow Player
```

Exemplos:

```text
Git Player
Filesystem Player
Shell Player
Docker Player
Kubernetes Player
Terraform Player
PostgreSQL Player
Test Player
HTTP Player
Claude Player
GPT Player
Human Approval Player
```

Cada Player possui um **manifest**, descrevendo suas capabilities, entradas, saídas e comportamento.

Exemplo conceitual:

```text
Kubernetes Player

capabilities:
  deployment.create
  deployment.update
  deployment.status
  pod.logs
```

O runtime não pensa:

> "chame Kubernetes Player".

Ele pensa:

> **"preciso da capability `deployment.update`; qual Player consegue fornecê-la?"**

---

# 4. Muitos Players determinísticos são estratégicos

Isso virou uma parte importante do roadmap.

Queremos construir uma **biblioteca grande de Players determinísticos**, porque isso:

* aumenta a utilidade do Runtgine sem IA;
* reduz custos;
* aumenta confiabilidade;
* facilita adoção;
* cria um ecossistema extensível;
* permite que empresas criem Players próprios.

Começamos com poucos Players para provar a arquitetura, mas a visão é ter muitos.

---

# 5. Event-driven continua sendo o coração

A análise do Empryo **não substituiu isso**.

O fluxo fundamental continua:

```text
Task
 ↓
Event
 ↓
Queue
 ↓
Orchestrator
 ↓
Capability
 ↓
Player
 ↓
Result
 ↓
Event
 ↓
próximo passo
```

Conceitualmente:

```text
                    Event Bus
                       │
                       ▼
                   Orchestrator
                       │
                 Capability Resolver
                       │
                  Player Router
                       │
                      Queue
                       │
                     Player
                       │
                  Result/Event
                       │
                       └──────────► Event Bus
```

Uma distinção que definimos:

### Event

Algo aconteceu.

```text
deployment.failed
```

### Queue

Existe trabalho aguardando processamento.

```text
diagnosis.queue
```

### Workflow

Define uma estrutura/relação de execução.

```text
test → build → deploy → verify
```

---

# 6. Task ≠ Workflow ≠ Execution Plan

Outra distinção importante.

### Task

A intenção/pedido do usuário.

> "Faça deploy da API no staging."

### Workflow

Uma estrutura reutilizável de execução.

```text
test
 ↓
build
 ↓
deploy
 ↓
verify
```

### Execution Plan

O plano específico criado para **aquela execução**.

Ele pode mudar dependendo do contexto, Players disponíveis, recursos, políticas etc.

---

# 7. Entrada de tarefas será flexível

O Runtgine não deve exigir que o usuário conheça sua linguagem interna.

Queremos suportar:

```text
CLI
TUI
API
Webhooks
GitHub
Slack
Scheduler
SDK
UI gráfica
```

Mas todos convergem para o mesmo protocolo interno:

```text
             CLI
              │
              ├── TUI
              ├── API
              ├── Webhook
              ├── Slack
              └── Scheduler
                     │
                     ▼
               Task Protocol
                     │
                     ▼
                  Runtime
```

---

# 8. Intent Engine

Essa foi uma das ideias mais importantes da conversa.

O usuário pode escrever:

> "Pega a última versão da API, roda os testes e coloca no staging."

Não queremos executar diretamente.

Criamos uma camada:

**Runtgine Intent Engine**

Ela funciona quase como um compilador:

```text
Human Intent
      ↓
Intent Engine
      ↓
Runtgine Task IR
      ↓
Validator
      ↓
Execution Plan
      ↓
Runtime
```

---

# 9. LLM especializada em Runtgine

Essa LLM de entrada seria diferente dos LLM Players que executam tarefas complexas.

Ela seria especializada em:

```text
Runtgine Protocol
Players
Capabilities
Task schemas
Policies
Resources
Runtime Graph
```

O objetivo é **traduzir a intenção humana para algo que o runtime consiga validar**.

Ela poderia descobrir:

```text
"deploy API"
```

↓

```text
git.checkout
test.run
container.build
image.push
deployment.update
deployment.verify
```

Mas ela **não é autoridade**.

Se inventar:

```text
database.magic_migration
```

o Registry rejeita.

---

# 10. Validação antes da execução

Criamos a ideia de um **Task Validator**.

Antes de executar:

```text
✓ capabilities existem
✓ inputs são válidos
✓ schemas corretos
✓ dependências resolvidas
✓ resources existem
✓ permissions permitidas
✓ policies respeitadas
✓ execution graph válido
```

Isso desloca muitos erros de:

```text
runtime error
```

para:

```text
compile/validation error
```

É uma filosofia quase de compilador:

```text
Human Intent
     ↓
Runtgine IR
     ↓
Validation
     ↓
Execution
```

---

# 11. Runtime Graph

Aqui entra a principal contribuição conceitual do Empryo.

O Runtgine deve ter um **Runtime Graph**, representando relações entre:

```text
Players
Capabilities
Tasks
Workflows
Resources
Repositories
Symbols
Events
Runs
Artifacts
Dependencies
```

O Graph responde:

> **"O que existe e como as coisas se relacionam?"**

Enquanto o Event Engine responde:

> **"O que está acontecendo agora?"**

Então temos:

```text
Runtime Graph = memória estrutural

Event Store/Bus = memória temporal
```

---

# 12. O que o Empryo pode trazer

A análise do Empryo revelou várias ideias interessantes.

### Genome → Runtime Graph

O conceito de Genome do Empryo, que representa estruturalmente um projeto, pode inspirar nosso Graph.

Mas não devemos copiá-lo literalmente.

Empryo:

```text
Code
Symbols
Dependencies
```

Runtgine:

```text
Code
Players
Capabilities
Resources
Workflows
Runs
Dependencies
```

---

### Symbol-level intelligence

A ideia do Empryo de trabalhar com **símbolos em vez de simples strings** pode ajudar o Runtgine a entender código estruturalmente.

Por exemplo:

```text
PaymentService
    ↓
processPayment()
    ↓
PaymentRepository
    ↓
PostgreSQL
```

Isso permite contexto muito mais preciso para Players de IA.

---

### Blast Radius

Essa foi uma das ideias que mais gostamos.

Antes de executar uma mudança:

```text
Change
 ↓
Graph
 ↓
Affected Players
Affected Workflows
Affected Resources
Affected Symbols
```

Podemos apresentar:

```text
Impact Analysis

Affected workflows: 4
Affected resources: 7
Risk: HIGH
```

Isso pode ser muito poderoso no futuro.

---

### Semantic Search

Em vez de somente:

```text
grep "deploy"
```

o Runtgine poderia responder:

> "Qual workflow faz deploy dessa API?"

ou:

> "Quais recursos dependem desse banco?"

Consultando o Graph.

---

### Context Compaction / Context Engine

Outra contribuição importante.

O LLM não deve receber todo o projeto.

O Runtgine monta:

```text
Task
+
Relevant Events
+
Relevant Symbols
+
Relevant Resources
+
Previous Decisions
+
Current State
```

Isso reduz tokens e melhora a qualidade.

---

### Model Task Router

O conceito do Empryo reforça nosso **Player Router**.

Mas vamos generalizar:

```text
Task
 ↓
Required Capability
 ↓
Player candidates
 ↓
Router
 ↓
best Player
```

Pode escolher:

```text
deterministic
Claude
GPT
Gemini
local LLM
human
```

Dependendo de:

* capability;
* complexidade;
* custo;
* latência;
* contexto;
* policy.

---

### Agent Modes → Execution Policies

O Empryo possui modos de operação de agentes.

Nós podemos transformar essa ideia em algo mais genérico:

```text
Execution Policy
```

Exemplo:

```text
filesystem: read
shell: deny
network: deny
production.deploy: approval-required
```

Isso se encaixa muito melhor com a arquitetura de Players.

---

### Background Agents → Background Players

Em vez de tratar isso como "subagents", podemos simplesmente ter:

```text
Coordinator Player
       │
       ├── Research Player
       ├── Test Player
       └── Review Player
```

Coordenados através de eventos.

---

### File Claims → Resource Claims

Outra ideia muito boa.

Em vez de apenas bloquear arquivos:

```text
Player A claims file X
```

podemos generalizar:

```text
Player A claims resource X
```

Recursos podem ser:

```text
file
repository
database
environment
deployment
workspace
```

Isso será útil para concorrência.

---

# 13. O fluxo completo que definimos

Um exemplo complexo:

```text
Usuário
 ↓
"Faça deploy da API..."
 ↓
Intent Engine
 ↓
Runtime Catalog
 ↓
Task IR
 ↓
Validator
 ↓
Execution Plan
 ↓
Orchestrator
 ↓
Capability Resolver
 ↓
Deterministic Player
 ↓
Event
 ↓
próximo Player
 ↓
LLM Player quando necessário
 ↓
Policy
 ↓
Deterministic execution
 ↓
Verification
 ↓
Task completed
```

Um exemplo real poderia envolver:

```text
Git Player
 ↓
Migration Analyzer
 ↓
LLM Review Player
 ↓
Test Player
 ↓
Docker Player
 ↓
Kubernetes Player
 ↓
Failure Analyzer
 ↓
LLM Diagnosis Player
 ↓
Kubernetes Player
 ↓
Verification Player
```

Tudo conectado por eventos.

---

# 14. Arquitetura do Core

A decisão mais recente foi:

> **O Core é o produto. A interface é uma superfície sobre ele.**

Arquitetura:

```text
                    RUNTGINE CORE
                         │
        ┌────────────────┼────────────────┐
        │                │                │
      Tasks            Events          Players
        │                │                │
        ├── Queue        ├── Event Bus   ├── Registry
        ├── Planner      ├── Event Store ├── Capabilities
        ├── Validator    └── Streams     └── Execution
        │
        ├── Policy
        ├── Runtime Graph
        └── Intent Engine
                         │
                  Public Protocol
                         │
             ┌───────────┼───────────┐
             │           │           │
            CLI         TUI         GPUI
```

**UI nunca chama Player diretamente.**

---

# 15. Interfaces

Queremos três possibilidades:

### CLI

Automação, scripting, CI/CD.

### TUI

Primeira interface real do projeto.

Inspirada em:

* Claude Code;
* Aider;
* lazygit;
* IDEs.

Mas centrada em:

```text
Tasks
Runs
Events
Players
Capabilities
Graph
```

### GPUI

Futura interface desktop nativa em Rust.

A preferência é **GPUI/Zed**, evitando uma aplicação web/Electron.

---

# 16. MVP

A decisão atual é:

> **Core em Rust + TUI em paralelo.**

O Core precisa funcionar independentemente da TUI.

Stack proposta:

```text
Rust
Tokio
Serde
JSON/JSON Schema
tracing
```

Para o MVP:

* Event Bus inicialmente **in-process**;
* persistência local simples;
* SQLite posteriormente como evolução do storage;
* NATS JetStream fica para uma fase posterior;
* PostgreSQL/cloud também fica para depois;
* sem Kubernetes obrigatório;
* sem infraestrutura distribuída obrigatória;
* sem UI gráfica no primeiro estágio.

A ideia é provar o runtime primeiro.

---

# 17. Estrutura conceitual do repositório

Algo próximo disso:

```text
runtgine/
│
├── core/
│   ├── task/
│   ├── event/
│   ├── queue/
│   ├── player/
│   ├── capability/
│   ├── policy/
│   ├── orchestration/
│   ├── graph/
│   └── protocol/
│
├── players/
│   ├── filesystem/
│   ├── shell/
│   ├── git/
│   ├── test/
│   └── llm/
│
├── tui/
│   ├── app/
│   ├── screens/
│   ├── widgets/
│   └── input/
│
└── protocol/
    ├── events/
    └── schemas/
```

---

# 18. A TUI do MVP

A primeira experiência pode ser extremamente simples:

```text
┌────────────────────────────────────────────────────────────┐
│ RUNTGINE                              ● CORE ONLINE         │
├────────────────────────────────────────────────────────────┤
│                                                            │
│ > Analise e corrija os testes que estão falhando           │
│                                                            │
│ PLAN                                                       │
│                                                            │
│ ✓ Intent resolved                         deterministic     │
│ ✓ Tests executed                          Test Player       │
│ ● Analyze failure                         Claude Player     │
│ ○ Propose fix                             Claude Player     │
│ ○ Apply fix                               Filesystem Player │
│ ○ Run tests                               Test Player       │
│                                                            │
│ EVENTS                                                     │
│                                                            │
│ task.created                                                │
│ intent.resolved                                             │
│ tests.started                                               │
│ tests.failed                                                │
│ diagnosis.started                                           │
│                                                            │
├────────────────────────────────────────────────────────────┤
│ Players: 7 │ Queue: 1 │ Events: 42 │ Run: task_1842       │
└────────────────────────────────────────────────────────────┘
```

A TUI será também nosso **instrumento para validar o comportamento do Core**.

---

# 19. O grande diferencial que está emergindo

Acho que a síntese de tudo que discutimos é esta:

```text
                HUMAN INTENT
                     │
                     ▼
              Runtgine Intent
                     │
                     ▼
                Task IR
                     │
                     ▼
              Deterministic
                Validator
                     │
                     ▼
              Execution Plan
                     │
                     ▼
                 Event Bus
                     │
                     ▼
                Orchestrator
                     │
              Capability Resolver
                     │
                Player Router
                     │
          ┌──────────┴──────────┐
          │                     │
    Deterministic             AI
       Players              Players
          │                     │
          └──────────┬──────────┘
                     ▼
                   Events
                     │
                     ▼
                  Graph
                     │
                     ▼
                  State
```

**O Empryo ajuda principalmente a tornar esse runtime "consciente" da estrutura do que está executando.**

Mas o coração continua sendo nosso:

> **Tasks + Events + Queues + Orchestrator + Capabilities + Players + Policies.**

E a visão que eu considero mais forte que saiu dessas conversas é:

> **Runtgine não é um agente que executa coisas. É um runtime que transforma intenção em execução verificável, usando determinismo sempre que possível e inteligência quando necessário.**

Essa é, para mim, a definição que melhor amarra tudo que decidimos até agora.
