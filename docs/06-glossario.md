# 06 — Glossario

| Termo | Definicao |
|---|---|
| Background Player | Player coordenado por outro Player via eventos |
| Blast Radius | Analise de impacto deterministica de uma Task IR (touches + predicted claims; ver `25`) |
| Capability | O que um Player sabe fazer. Ex: deployment.update |
| Chorus | Protocolo/comunicacao entre componentes (complementar ao Runtgine) |
| Context Engine | Assembler do ContextPack; CONFIRMED v0 (semente `repo_hits`; ver `31`) |
| ContextPack | Pacote de contexto por step (G-24); inclui `graph_hits` e `memory_hits` v0 (`29`) |
| Deterministic-first | Preferir execucao deterministica a IA |
| Entry Point | Interface com o mundo externo (CLI, TUI, Board, API…). Nao e Player |
| Event | Algo aconteceu no sistema |
| Event Bus | Transporte de eventos entre componentes |
| Execution Plan | Plano especifico criado para UMA execucao |
| Execution Policy | Regras allow/deny/approval-required por capability (Core; ver `22`) |
| HITL | Humano aprova/rejeita um Run em `waiting_approval` via Entry Point |
| Intent Engine | Traduz intencao humana (NL) em Task IR |
| Manifest | Declaracao de capabilities, entradas e saidas de um Player |
| Memory Player | Player read-only `memory.recall` / `memory.check`; CONFIRMED v0 (ver `38`) |
| Workflow Template | JSON reutilizavel que compila para Task IR; CONFIRMED v0 (ver `40`) |
| Memory Provider | Fonte de memoria consultada pelo AssembleContext; CONFIRMED v0 local SQLite (`29`) |
| Orchestrator | Coordena o fluxo de execucao |
| Player | Entidade capaz de fornecer capabilities |
| Player Router | Seleciona o melhor Player para uma capability |
| Project Knowledge | Conhecimento consolidado do projeto (evolucao possivel); HYPOTHESIS — distinto de Memory |
| Project Memory | Memoria episodica entre runs/sessoes; CONFIRMED v0 (ver `29`; esboco `16`) |
| Queue | Trabalho aguardando processamento |
| Resource Claim | Bloqueio concorrente exclusivo de `workspace`/`path` (Core; ver `24`) |
| Run | Tentativa de execucao de uma Task aceita (tem run_id) |
| Runner | Orchestrator minimo do MVP; valida, planeja e despacha steps |
| Runtime Graph | Memoria estrutural: relacoes entre Players, Resources, Tasks |
| Task | Intencao/pedido do usuario |
| Task IR | Representacao intermediaria validavel de uma task |
| Task Validator | Valida capabilities, inputs, schemas, policies antes de executar |
| Test Player | Player deterministico `test.go` (Go); CONFIRMED v0 (ver `30`) |
| Workflow | Estrutura reutilizavel de execucao |