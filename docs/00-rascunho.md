# 00 — Rascunho (discussoes em andamento)

Este documento captura discussoes e decisoes em andamento que ainda nao
foram formalizadas nos documentos oficiais. Quando uma discussao amadurece,
o conteudo migra para o documento apropriado.

## Proxima discussao: Multi-entry-point architecture

### Problema

Como o Runtgine recebe tasks para processar? Board Kanban e o ponto de
entrada principal, mas nao o unico.

### Entry points identificados

```
Board Kanban (Github Projects, Todoist, Jira etc.)
    |— Polling ou webhook: task criada no Todo dispara o pipeline
    |— Runtgine atualiza status, cria subtasks, reporta progresso

API HTTP (REST/GraphQL)
    |— Recebe tasks via POST /tasks
    |— Permite execucao em serverless/cloud (Lambda, Cloud Run)
    |— Ideal para automacao: CI/CD, webhooks, integracoes
    |— Core roda stateless, eventos vao para fila

CLI (terminal)
    |— runtgine run task.yaml
    |— runtgine pipeline start --board github
    |— runtgine status <execution-id>

TUI (terminal interativo)
    |— Visualizacao do board no terminal
    |— Acompanhamento de execucoes em tempo real
    |— Ideal para devs que vivem no terminal

Desktop (GPUI)
    |— Interface nativa, visualizacao do grafo de execucao
    |— Instalacao local para devs, TLs, QAs
    |— Tema escuro, foco na execucao, nao em dashboard

Web (futuro)
    |— Interface web para acesso remoto
    |— Ideal para gestao centralizada
```

### Principio: Core unico, entry points variados

```
Board     API    CLI    TUI    Desktop    Web
  |        |      |      |       |        |
  +--------+------+------+-------+--------+
                   |
           Runtgine Core
          (Execution, Context,
           Router, Players)
                   |
            Persistence layer
```

O Core:
- Nao sabe se a task veio do board, API ou CLI
- Nao sabe se a saida vai para TUI, GPUI ou web
- So processa eventos e produz eventos

Cada entry point:
- Traduz sinal externo para o protocolo do Runtgine
- Escuta eventos do core e atualiza sua interface
- Pode ser adicionado sem modificar o core

### Implicacao para serverless/cloud

O Core precisa ser projetado para rodar em modo stateless:

1. Event Bus interno funciona em memoria no MVP
2. Para cloud: Event Bus troca por fila externa (SQS, RabbitMQ)
3. Persistence layer vira plugavel: SQLite (local) -> RDS (cloud)
4. Context assembly precisa ser stateless (repo search via API, nao

   filesystem local)
5. Players remotos comunicam via protocolo, nao por carga direta

Isso nao precisa ser implementado agora, mas o design do Core deve
permitir essa migracao sem reescrita.

### Questoes em aberto

- O Entry Point e um Player? Ou e uma camada separada?
- Rascunho inicial: Entry Point NAO e Player. Player executa trabalho.
  Entry Point e interface com o mundo externo.
- O protocolo entre Entry Point e Core e o mesmo Runtgine Protocol
  ou um protocolo separado?
- Board Integration: polling vs webhook? Board e um Entry Point ou
  e um adapter que se conecta a um Entry Point do tipo API?

### Proximos topicos

- TUI: o que mostrar, como navegar
- Desktop GPUI: o que torna a versao desktop diferente da TUI
- Instalacao e distribuicao
-HEREDOC
---

## Context Management (discussao)

O que temos hoje de documentacao sobre como o projeto se organiza:
- AGENTS.md: entry point para LLMs
- docs/ numerados por ordem de leitura
- docs/04-decisoes.md: decisoes com status
- docs/00-rascunho.md: discussoes em andamento
- brainstorm.md: fonte bruta original

Falta documentar:
- Estrutura de diretorios do projeto Runtgine
- Convencoes de codigo e documentacao
- Genome / mapa de simbolos (quando houver codigo)
- Como o contexto e gerenciado entre modulos

---

## SDD / TLC Spec-Driven (discussao)

Referencia: github.com/tech-leads-club/agent-skills
player/skills-catalog/skills/(development)/tlc-spec-driven/SKILL.md

### O que e

Skill/playbook de desenvolvimento com 4 fases adaptativas:
Specify -> Design (opcional) -> Tasks (opcional) -> Execute.

Caracteristicas principais:
- EARS notation para requisitos testaveis
- Gates deterministicos (scripts Python validam spec, tasks, commits, estado)
- Auto-sizing: profundidade ajustada por complexidade (small/medium/large/complex)
- Sub-agent delegation: batches de tasks delegados a sub-agentes
- Verifier independente (autor != verifier, evidence-or-zero, discrimination sensor)
- Blast radius: spec aprova impl local apenas; push/deploy exigem nova aprovacao
- Decision log (STATE.md com AD-NNN)
- Lessons layer (auto-distillation apos validacao)
- Context loading on-demand, target <40k tokens

### Minha opiniao: SDD nao e um Player. E um workflow template.

SDD encaixa no Runtgine em tres niveis:

**Nivel 1 — SDD como Workflow Template (recomendado)**
SDD vira um template registrado no Runtime Graph. O Intent Engine conhece
SDD e, quando o usuario diz 'implementa feature X seguindo SDD', gera um
Execution Plan com as 4 fases, gates deterministicos e Verifier ao final.
Cada fase do SDD e uma Task ou conjunto de Tasks no Runtgine.

**Nivel 2 — SDD como execution policy**
SDD define politicas: 'spec so avanca se validate_spec.py passar',
'commit so e valido se seguir Conventional Commits',
'implementacao so termina se validate_state.py passar'. O Task Validator
do Runtgine executa os scripts Python do SDD como gates.

**Nivel 3 — SDD como Player (nao recomendado)**
Um SDD Player que orquestra tudo internamente esconderia a visibilidade
do processo. O Runtgine e um runtime observavel — cada fase, cada gate,
cada decisao deve ser um evento visivel. Um Player opaco quebraria isso.

### Como ficaria no fluxo do Runtgine

Usuario: 'Implementa a feature de login seguindo SDD'

1. Intent Engine reconhece 'seguindo SDD' e consulta Runtime Graph
2. Graph retorna template SDD (4 fases, gates, verifier)
3. Intent Engine gera Execution Plan:
   a. Specify: Task para LLM Player (gerar spec) + Task para Validator (validate_spec.py)
   b. Design: Task para LLM Player (arquitetura) + Validator (opcional)
   c. Tasks: Task para Decomposition (quebrar em subtasks) + Validator (validate_tasks.py)
   d. Execute: N tasks para Players + Verifier (validate_state.py)
4. Orchestrator executa o plano, emitindo eventos em cada etapa
5. Cada gate deterministico e um Task Validator que trava se falhar
6. Verifier roda automaticamente ao final (autor != verifier)

### O que precisa ser adicionado ao modelo conceitual

- Workflow Template: registro reutilizavel no Runtime Graph (ex: SDD, CI/CD, deploy)
- Skill/Playbook: documentacao executavel que o Intent Engine conhece
- Deterministic Gate: script que valida um artefato (spec, tasks, commit, estado)
- Phase: etapa de um workflow template (Specify, Design, Tasks, Execute)
- Verifier: Player ou gate que valida o resultado final

### Questao em aberto

SDD vive no Runtime Graph como template nativo do Runtgine ou e carregado
de um repositorio externo (ex: baixado do github como skill) e registrado
dinamicamente? A segunda opcao e mais poderosa — qualquer skill/playbook
poderia ser registrada sem modificar o core.