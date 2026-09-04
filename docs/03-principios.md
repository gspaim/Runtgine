# 03 — Principios de design

## 1. Deterministic-first
Se uma tarefa pode ser executada deterministicamente, prefira isso
ao uso de IA. LLM entra quando existe necessidade de interpretacao,
planejamento, raciocinio, diagnostico, geracao, revisao ou decisao
complexa. Reduz custo, latencia, imprevisibilidade.

## 2. Player e a abstracao central
Player nao e sinonimo de Agent. Player e qualquer entidade capaz de
fornecer capabilities. O runtime pensa em capabilities, nao em Players.

## 3. Muitos Players deterministicos sao estrategicos
Uma biblioteca grande de Players deterministicos aumenta a utilidade
do Runtgine sem IA, reduz custos, aumenta confiabilidade e facilita
adocao. A visao de longo prazo e ter muitos Players deterministicos.

## 4. Event-driven e o coracao
Task -> Event -> Queue -> Orchestrator -> Capability -> Player -> Result -> Event.
Tudo conectado por eventos. Distincao: Event (algo aconteceu),
Queue (trabalho aguardando), Workflow (estrutura de execucao).

## 5. Core e o produto. Interface e superficie.
O Core funciona independentemente da TUI, CLI ou Wails.
UI nunca chama Player diretamente — tudo passa pelo protocolo.
Entry Point != Player.

## 6. Validacao antes da execucao
Task Validator verifica capabilities, inputs, schemas, dependencias,
resources, permissions, policies antes de executar. Filosofia de
compilador: deslocar runtime errors para validation errors.

## 7. Entrada flexivel, protocolo unico
CLI, TUI, Board, API, Webhooks, Slack, Scheduler — todos convergem para
o mesmo Task Protocol interno.

## 8. Runtime Graph = memoria estrutural
O Graph conhece Players, Capabilities, Tasks, Workflows, Resources,
Symbols, Runs, Artifacts. Enquanto o Event Bus sabe o que esta
acontecendo agora, o Graph sabe o que existe e como se relaciona.

## 9. Contexto relevante, nao todo o projeto
Context Engine v0 (`31`) semeia `repo_hits` a partir do Graph quando
o pack nao tem repo-search neste Run. Sem dump do repositorio, sem
embeddings. O desenho amplo (events globais, current state rico)
permanece fora do 1.0.

## 10. LLM-agnostic
Players LLM sao um tipo de Player entre outros. O Core nao conhece
provedores especificos. Ha uma camada de abstracao.
