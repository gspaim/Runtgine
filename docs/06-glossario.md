# 06 — Glossario

| Termo | Definicao |
|---|---|
| Background Player | Player coordenado por outro Player via eventos |
| Blast Radius | Analise de impacto de uma mudanca no grafo |
| Capability | O que um Player sabe fazer. Ex: deployment.update |
| Chorus | Protocolo/comunicacao entre componentes (complementar ao Runtgine) |
| Context Engine | Monta contexto relevante para cada Player |
| Deterministic-first | Preferir execucao deterministica a IA |
| Entry Point | Interface com o mundo externo (CLI, TUI, Board, API…). Nao e Player |
| Event | Algo aconteceu no sistema |
| Event Bus | Transporte de eventos entre componentes |
| Execution Plan | Plano especifico criado para UMA execucao |
| Execution Policy | Regras de seguranca/permissao por Player/acao |
| Intent Engine | Traduz intencao humana (NL) em Task IR |
| Manifest | Declaracao de capabilities, entradas e saidas de um Player |
| Orchestrator | Coordena o fluxo de execucao |
| Player | Entidade capaz de fornecer capabilities |
| Player Router | Seleciona o melhor Player para uma capability |
| Queue | Trabalho aguardando processamento |
| Resource Claim | Bloqueio concorrente de recurso |
| Runtime Graph | Memoria estrutural: relacoes entre Players, Resources, Tasks |
| Task | Intencao/pedido do usuario |
| Task IR | Representacao intermediaria validavel de uma task |
| Task Validator | Valida capabilities, inputs, schemas, policies antes de executar |
| Workflow | Estrutura reutilizavel de execucao |