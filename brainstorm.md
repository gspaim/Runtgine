> **HISTÓRICO — NÃO É AUTORIDADE.**  
> Fonte bruta original do projeto. Stack e UI neste arquivo (Rust, GPUI, etc.)
> estão **desatualizadas**. Fonte de verdade: `docs/04-decisoes.md`,
> `docs/09-mvp.md` e demais `docs/`. Use apenas como contexto narrativo.

Claro. E acho que vale fazer isso de um jeito **mais técnico do que um simples resumo**, porque a ideia é você entregar para outra LLM e ela conseguir entender o Runtgine sem precisar reconstruir toda a história do Chorus.

O ponto mais importante é: **Runtgine não deve ser tratado simplesmente como “um framework de agentes de IA”**. A evolução que vocês fizeram é justamente para algo mais fundamental: um **runtime universal de execução e orquestração**, onde LLMs são apenas um dos possíveis participantes.

# Runtgine — visão consolidada do projeto

## 1. O que é o Runtgine

**Runtgine é um runtime universal, leve, orientado a eventos e baseado em protocolo para orquestração e execução de trabalho.**

Ele nasceu dos conceitos do projeto **Chorus**, mas está sendo redesenhado para ser:

* mais genérico;
* mais simples;
* modular;
* desacoplado de LLMs;
* orientado a protocolos;
* orientado a eventos;
* determinístico sempre que possível;
* extensível sem transformar o core em um monólito.

A ideia central é que o Runtgine não deve saber necessariamente **quem está executando um trabalho**.

Pode ser:

* um LLM;
* um script;
* uma ferramenta;
* um processo local;
* um serviço remoto;
* uma pessoa;
* outro runtime;
* um workflow;
* uma combinação desses elementos.

O runtime fornece o ambiente, o protocolo e as regras para essas entidades colaborarem.

---

# 2. O problema que o Runtgine resolve

Ferramentas modernas permitem executar diversos agentes, CLIs, scripts e serviços, mas normalmente cada componente funciona isoladamente.

Isso cria problemas como:

* múltiplos terminais sem coordenação;
* agentes que não compartilham contexto adequadamente;
* comunicação baseada em convenções frágeis;
* workflows difíceis de reproduzir;
* excesso de dependência de LLMs;
* execução não determinística quando poderia ser determinística;
* pouca observabilidade;
* contexto espalhado;
* dificuldade para interromper, retomar ou auditar execuções;
* acoplamento entre o orquestrador e o tipo de executor.

O Runtgine pretende criar uma camada intermediária comum.

Em vez de:

```text
UI → Agent → Tool → Agent → Agent
```

a ideia é aproximar-se de:

```text
             ┌──────────────┐
             │   Runtgine   │
             │    Runtime   │
             └──────┬───────┘
                    │
             Protocol / Events
                    │
       ┌────────────┼────────────┐
       │            │            │
     LLM          Tool         Human
       │            │            │
     Service      Script      Workflow
```

O runtime coordena a execução sem precisar conhecer profundamente a implementação de cada participante.

---

# 3. Princípio fundamental: protocol-first

O Runtgine deve ser **protocol-first**.

O protocolo é mais importante que qualquer implementação específica.

Isso significa que:

* o core não deve depender de um LLM específico;
* Players não devem depender diretamente da UI;
* a UI não deve controlar diretamente os executores;
* componentes devem comunicar-se através de eventos/mensagens;
* capacidades devem ser declaradas;
* entradas e saídas devem possuir contratos claros;
* o runtime deve conseguir observar e controlar uma execução sem conhecer todos os detalhes internos dela.

O protocolo deve ser definido por contratos formais, idealmente utilizando:

* JSON Schema;
* JSON;
* Serde no lado Rust.

---

# 4. Player: abstração central

Uma das decisões conceituais mais importantes herdadas do Chorus é substituir o conceito tradicional de **Agent** por uma abstração mais genérica: **Player**.

Um Player é qualquer entidade capaz de participar de uma execução.

Exemplos:

```text
LLM Player
Human Player
Script Player
Tool Player
Service Player
Workflow Player
Deterministic Function Player
External Process Player
```

O runtime não deve assumir que um Player é inteligente.

Por exemplo:

```text
"somar dois números"
```

não deveria necessariamente envolver um LLM.

O runtime deve preferir:

```text
Deterministic Player
```

quando uma capacidade determinística estiver disponível.

Somente quando necessário:

```text
LLM Player
```

entra em ação.

Isso reduz:

* custo;
* latência;
* imprevisibilidade;
* consumo de tokens;
* complexidade.

---

# 5. Manifest / Capabilities

Players devem expor suas capacidades através de um **manifest**.

Conceitualmente:

```yaml
name: terraform-player
version: 0.1.0

capabilities:
  - terraform.plan
  - terraform.apply

inputs:
  - repository
  - workspace

outputs:
  - plan
  - result
```

O manifest permite que o runtime descubra:

* quem é o Player;
* versão;
* capacidades;
* tipos de entrada;
* tipos de saída;
* requisitos;
* eventualmente políticas/permissões.

A ideia é que o runtime possa fazer **capability-based routing**.

Exemplo:

```text
Task:
  "executar terraform plan"

Capabilities disponíveis:

terraform-player
shell-player
llm-player
```

O runtime deve preferir o executor determinístico adequado.

---

# 6. Event-driven architecture

A comunicação interna deve ser orientada a eventos.

Uma execução pode produzir eventos como:

```text
ExecutionCreated
TaskCreated
TaskStarted
TaskWaiting
PlayerSelected
InputRequested
OutputProduced
ToolCalled
ApprovalRequested
TaskCompleted
TaskFailed
TaskCancelled
ExecutionCompleted
```

O importante é que o estado do sistema possa ser compreendido através do fluxo de eventos.

Conceitualmente:

```text
Event
  ↓
Runtime
  ↓
State transition
  ↓
New Event
  ↓
Subscribers
```

Isso também favorece:

* observabilidade;
* replay;
* auditoria;
* debugging;
* UI reativa;
* execução assíncrona;
* integração externa.

---

# 7. Execution

A unidade operacional mais importante é uma **Execution**.

Uma Execution representa uma execução concreta de trabalho.

Ela pode possuir:

```text
Execution
 ├── Tasks
 ├── Players
 ├── Events
 ├── Context
 ├── Inputs
 ├── Outputs
 ├── Logs
 ├── Approvals
 └── Artifacts
```

Uma Execution pode ser:

* criada;
* iniciada;
* pausada;
* retomada;
* cancelada;
* concluída;
* falhada.

O runtime deve manter uma separação clara entre:

```text
Definition
```

e

```text
Execution
```

Uma definição descreve **o que deve ser executado**.

Uma execution representa **uma instância daquela definição**.

---

# 8. Tasks

Tasks representam unidades de trabalho.

Uma Task não necessariamente significa "chamar um agente".

Ela pode representar:

```text
Run command
Call API
Transform data
Ask human
Call LLM
Run workflow
Execute tool
Wait for event
Evaluate condition
```

Isso mantém o runtime genérico.

Uma Task pode declarar uma capacidade necessária:

```text
required_capability:
    terraform.plan
```

O runtime então encontra um Player capaz de satisfazê-la.

---

# 9. Graph / Workflow

Uma Execution pode ser representada como um grafo.

Exemplo:

```text
        ┌──────────────┐
        │ Analyze repo │
        └───────┬──────┘
                ↓
        ┌──────────────┐
        │ Create plan  │
        └───────┬──────┘
                ↓
        ┌──────────────┐
        │ Human review │
        └───────┬──────┘
                ↓
        ┌──────────────┐
        │ Apply change │
        └──────────────┘
```

O runtime deve ser capaz de trabalhar com:

* sequência;
* branching;
* condições;
* paralelismo;
* dependências;
* espera por eventos;
* retry;
* aprovação humana.

Mas isso deve ser construído incrementalmente.

Não assumir um workflow engine extremamente complexo no MVP.

---

# 10. Deterministic-first

Essa é uma filosofia importante do projeto:

> **Use determinism whenever possible; use intelligence when necessary.**

Exemplo:

```text
Task:
"verificar se arquivo existe"
```

Não:

```text
LLM → decidir → executar
```

Mas:

```text
filesystem capability → result
```

Já algo como:

```text
"analise esse código e proponha uma arquitetura"
```

pode utilizar:

```text
LLM Player
```

O Runtgine, portanto, não é um "AI orchestrator".

É um **execution runtime que pode orquestrar AI**.

Essa distinção é fundamental.

---

# 11. LLMs

LLMs devem ficar atrás de uma abstração própria.

O runtime não deve conhecer diretamente:

```text
OpenAI
Anthropic
Google
Ollama
etc.
```

Deve existir uma camada de provider/model abstraction.

Conceitualmente:

```text
Runtgine
    ↓
LLM abstraction
    ↓
Provider
    ↓
Model
```

Isso permite futuramente:

* routing;
* BYOK;
* múltiplos providers;
* fallback;
* seleção por custo;
* seleção por latência;
* seleção por capacidade.

A ideia de um **LLM Router** veio do Chorus e continua sendo uma possibilidade importante, mas deve permanecer desacoplada do core do runtime.

---

# 12. MCP

MCP pode ser utilizado como mecanismo de integração com ferramentas e contexto.

Mas existe uma distinção importante:

```text
Runtgine Protocol
```

é o protocolo fundamental de execução/orquestração.

Enquanto:

```text
MCP
```

pode ser utilizado como mecanismo de integração com ferramentas/contexto.

Não transformar MCP no protocolo interno fundamental do Runtgine.

---

# 13. Context / Knowledge

O projeto também herdou do Chorus a preocupação com **contexto e memória**.

O objetivo não é simplesmente armazenar todo o histórico.

O runtime deve conseguir fornecer aos Players o contexto necessário para executar uma tarefa.

Conceitualmente:

```text
Execution
    ↓
Context
 ├── Task context
 ├── Shared context
 ├── Knowledge
 ├── Artifacts
 └── History
```

Uma preocupação importante é eficiência de tokens.

Não queremos:

```text
todo histórico → todo Player
```

Queremos:

```text
relevant context → Player
```

A implementação concreta dessa camada ainda deve ser tratada como decisão arquitetural separada.

---

# 14. Workspaces / Worktrees

O Chorus também introduziu a ideia de trabalhar com ambientes isolados, especialmente para desenvolvimento de software.

O Runtgine deve poder trabalhar com:

```text
Repository
Workspace
Git branch
Git worktree
Artifacts
```

Isso permite cenários como:

```text
Task A
  → worktree A

Task B
  → worktree B
```

Players podem trabalhar isoladamente e depois produzir:

* diff;
* commits;
* artifacts;
* resultados.

Essa capacidade é especialmente importante para workflows de desenvolvimento assistido por IA.

---

# 15. Policies / Approvals

Execuções podem precisar de autorização humana.

Exemplo:

```text
LLM proposes change
       ↓
Approval requested
       ↓
Human approves
       ↓
Execution continues
```

A arquitetura deve permitir policies e approval gates.

Isso será especialmente importante para operações sensíveis:

```text
terraform apply
git push
production deployment
delete resource
execute shell command
```

Mas novamente: o modelo exato de policies ainda deve ser definido durante a arquitetura.

---

# 16. Observability

Cada Execution deve ser observável.

A UI/runtime deve conseguir mostrar:

```text
Execution
 ├── status
 ├── current task
 ├── events
 ├── logs
 ├── players
 ├── context
 ├── artifacts
 ├── diff
 ├── cost
 └── approvals
```

O sistema deve favorecer **event sourcing/event history**, mas não assumir automaticamente uma implementação completa de Event Sourcing antes de um spike técnico.

---

# 17. UI

Uma decisão importante atual é que o Runtgine **não deve ser uma aplicação web por padrão**.

A visão atual é de um **software desktop nativo**.

A tecnologia considerada principal para a UI é:

**Rust + GPUI**, framework utilizado pelo Zed.

Motivos:

* mesma linguagem do core;
* performance;
* baixo overhead;
* arquitetura orientada a eventos;
* possibilidade de aplicação desktop realmente nativa;
* interesse em evitar Electron/web stack.

Entretanto, GPUI deve passar por um **spike técnico antes de ser considerada decisão definitiva**.

---

# 18. Conceito da interface

A UI não deve parecer um dashboard corporativo cheio de telas.

A visão é uma interface minimalista e focada na execução.

O **Execution Graph** deve ser o centro da experiência.

Ao selecionar uma execution/task, o usuário poderia visualizar:

```text
Graph
 ├── Task
 ├── Player
 ├── Event
 ├── Context
 ├── Logs
 ├── Diff
 ├── Cost
 └── Approval
```

A UI funciona principalmente como uma camada de observação e controle do runtime.

Não deve conter lógica de negócio importante.

---

# 19. Stack atualmente considerada

A stack discutida até agora é aproximadamente:

```text
Language:
Rust

UI:
GPUI
(spike antes de compromisso definitivo)

Async runtime:
Tokio

Serialization:
Serde

Protocol:
JSON Schema
JSON
Serde

Persistence:
SQLite

Version control:
Git

Event infrastructure:
Event bus interno

LLM:
Provider abstraction

Tools/context:
MCP

Assets/configuration:
Markdown / YAML
```

Importante: **isso não deve ser tratado como uma lista de decisões irreversíveis**.

A regra do projeto é distinguir:

```text
CONFIRMED
HYPOTHESIS
OPEN QUESTION
```

---

# 20. Arquitetura conceitual

Uma primeira decomposição pode ser:

```text
┌──────────────────────────────────────────┐
│                  UI                      │
│                  GPUI                    │
└────────────────────┬─────────────────────┘
                     │
                     │ Protocol / Events
                     ↓
┌──────────────────────────────────────────┐
│                RUNTGINE                  │
│                                          │
│  Runtime                                 │
│  ├── Execution Engine                    │
│  ├── Task Engine                         │
│  ├── Player Registry                     │
│  ├── Capability Resolution               │
│  ├── Event Bus                           │
│  ├── Policy / Approval                   │
│  ├── Context                             │
│  └── Artifact Management                 │
│                                          │
└───────┬───────────┬───────────┬──────────┘
        │           │           │
        ↓           ↓           ↓
     Players      Tools       Services
        │
 ┌──────┼───────────────┐
 ↓      ↓       ↓       ↓
LLM   Script   Human   Workflow
```

O importante é que **Runtgine Core não dependa desses Players**.

---

# 21. Persistência

SQLite é a tecnologia atualmente considerada para persistência local.

Possíveis entidades:

```text
executions
tasks
players
events
artifacts
workspaces
approvals
policies
```

Mas o schema ainda não deve ser considerado fechado.

Primeiro deve-se definir:

1. modelo conceitual;
2. lifecycle;
3. eventos;
4. invariantes;
5. somente depois o schema físico.

---

# 22. CLI

Apesar da existência de uma UI desktop, a filosofia continua sendo:

> **CLI/runtime-first.**

A UI deve ser uma camada sobre o runtime.

Idealmente:

```text
CLI
  ↓
Runtgine Core
  ↑
GPUI
```

e não:

```text
GPUI
  ↓
lógica interna
```

Isso mantém o runtime testável e utilizável sem interface gráfica.

---

# 23. Filosofia de design

As decisões arquiteturais devem seguir estes princípios:

### Simplicidade

Não criar abstrações porque parecem interessantes.

### Baixo acoplamento

Core não conhece implementação concreta dos Players.

### Protocol-first

Comunicação através de contratos explícitos.

### Event-driven

Estado e comunicação devem ser orientados a eventos.

### Deterministic-first

Não utilizar IA quando uma operação determinística resolve o problema.

### LLM-agnostic

LLM é um Player, não o centro do sistema.

### Local-first

O runtime deve funcionar bem localmente.

### Extensible

Novos Players e capacidades devem poder ser adicionados sem alterar o core.

### Testable

Componentes devem ser testáveis isoladamente.

### Observable

Toda execução importante deve ser observável.

---

# 24. O que NÃO é o Runtgine

Isso é tão importante quanto definir o que ele é.

Runtgine **não deve ser inicialmente tratado como**:

* um chatbot;
* um IDE;
* um framework exclusivo de agentes;
* um wrapper de LLM;
* uma alternativa ao MCP;
* um workflow SaaS;
* um dashboard web;
* um sistema exclusivamente baseado em prompts;
* um sistema onde todo trabalho passa por IA.

Ele é fundamentalmente:

> **um runtime de execução e orquestração baseado em protocolo, capaz de coordenar diferentes tipos de Players e utilizar IA quando isso fizer sentido.**

---

# 25. Relação com Chorus

A evolução conceitual é:

```text
Chorus
   ↓
Agent orchestration
   ↓
Generic Player
   ↓
Protocol-first
   ↓
Deterministic execution
   ↓
Runtgine
```

O Chorus tinha uma visão mais diretamente associada a:

* agentes;
* LLMs;
* memória;
* worktrees;
* Kanban;
* terminals;
* orchestration.

O Runtgine abstrai isso.

Em vez de perguntar:

> "Como coordenamos vários agentes?"

o Runtgine pergunta:

> **"Como executamos e coordenamos unidades de trabalho heterogêneas através de um protocolo comum?"**

Essa é provavelmente a mudança conceitual mais importante do projeto.

---

# 26. Estado atual do projeto

O projeto deve ser tratado como **greenfield**.

Não assumir que funcionalidades do Chorus foram automaticamente portadas.

Devemos separar tudo em:

### Confirmado

* Runtgine é o nome atual.
* É um runtime universal.
* Protocol-first.
* Event-driven.
* Player é a abstração central.
* LLM não é requisito do runtime.
* Deterministic-first.
* Desktop nativo é a direção atual.
* Rust é a linguagem central.
* GPUI é a principal candidata para UI.
* CLI/runtime-first.
* Arquitetura modular e desacoplada.

### Hipóteses atuais

* Tokio.
* SQLite.
* MCP.
* JSON Schema.
* Event bus interno.
* Worktree integration.
* LLM router.
* Context/knowledge architecture.
* Policy/approval system.

### Questões em aberto

* protocolo definitivo;
* lifecycle exato de Execution;
* lifecycle de Task;
* modelo definitivo de Player;
* formato final do Manifest;
* capability resolution;
* event model;
* persistência;
* replay/event sourcing;
* modelo de context;
* plugin architecture;
* segurança/sandbox;
* processo de instalação/distribuição;
* limites entre core e adapters;
* GPUI após spike técnico.

---

# 27. Como outra LLM deve trabalhar no projeto

A LLM que receber esse contexto deve assumir o papel de **arquiteto + engenheiro principal do Runtgine**, mas não deve sair implementando tudo imediatamente.

A ordem recomendada é:

```text
1. Entender o domínio
        ↓
2. Formalizar conceitos
        ↓
3. Identificar invariantes
        ↓
4. Definir protocolo mínimo
        ↓
5. Definir Execution lifecycle
        ↓
6. Definir Player + Manifest
        ↓
7. Definir Event model
        ↓
8. Criar arquitetura do core
        ↓
9. Implementar runtime mínimo
        ↓
10. Criar primeiro Player determinístico
        ↓
11. Criar primeiro LLM Player
        ↓
12. CLI
        ↓
13. Persistence
        ↓
14. UI/GPUI
```

A primeira versão não deve tentar implementar:

```text
LLM router
memory sofisticada
multi-agent planning
workflow engine completo
Kanban
plugin marketplace
cloud
distributed execution
```

antes de existir um runtime pequeno e funcional.

---

# 28. Primeiro MVP conceitual

O MVP deveria provar apenas isto:

```text
Task
  ↓
Capability
  ↓
Player resolution
  ↓
Player execution
  ↓
Events
  ↓
Result
```

Por exemplo:

```text
Task:
  capability = "shell.execute"

        ↓

Player Registry

        ↓

Shell Player

        ↓

execute("echo hello")

        ↓

OutputProduced

        ↓

TaskCompleted
```

Depois:

```text
Task
  ↓
LLM capability
  ↓
LLM Player
  ↓
response
  ↓
events
```

Se esse núcleo funcionar muito bem, todo o restante pode crescer em cima dele.

---

## 29. A frase que resume o Runtgine

Se fosse necessário explicar o projeto para uma LLM em uma única frase:

> **Runtgine is a lightweight, protocol-first, event-driven universal runtime for executing and orchestrating heterogeneous Players—LLMs, tools, scripts, humans, services and workflows—with deterministic execution preferred whenever possible.**

E a regra arquitetural mais importante seria:

> **Do not build an AI orchestration framework. Build an execution runtime in which AI is one possible capability.**

Esse é o ponto que eu usaria como **norte absoluto** para a documentação e para qualquer decisão de arquitetura daqui para frente.
