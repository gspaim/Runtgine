# 05 — MVP: Runtime Minimo

MVP focado no cenario real: board Kanban + pipeline de analise
+ Task Router para Players LLM especializados.

## Ciclo principal

Board (Entry Point) -> Import task -> Technical Review ->
Spec Review -> Repo Search -> Effort Estimation ->
Difficulty Classification -> Task Decomposition ->
Task Router -> Context Assembly -> Entrega para Players

## Entry points no MVP

| Entry Point | Incluido? | Notas |
|---|---|---|
| Board (Github Projects) | MVP | Polling inicial |
| CLI | MVP | runtgine run, status |
| API | Pos-MVP | Para serverless/CI |
| TUI | Pos-MVP | Terminal interativo |
| Desktop (GPUI) | Pos-MVP | Apos TUI validar |
| Web | Futuro | Se houver demanda |

## Escopo do MVP

### Incluido

- Board Integration (ler tasks do Github Projects)
- Task model com subtasks
- Pipeline de analise: technical review, spec review
- Repo Search
- Effort Estimation + Difficulty Classification
- Task Decomposition (regras + LLM)
- Task Router
- Context assembly (prioridade)
- Player Registry com LLMs especializados
- Event Bus
- CLI minima

### Nao incluido

- Shell Player
- Workflow engine complexo
- Human-in-the-loop completo
- Policies / Approvals
- Plugin system
- GPUI
- MCP integration
- Event sourcing
- API HTTP

## Ordem de implementacao

1. Board Integration + Task model
2. Player Registry + Manifest
3. Event Bus
4. Context assembly
5. LLM Player (Technical Review)
6. Pipeline linear
7. Repo Search
8. Effort Estimation + Difficulty
9. Task Decomposition
10. Task Router
11. CLI
12. Demais Players LLM

## Criterios de sucesso

- Task do board passa pelo pipeline completo
- Subtasks distribuidas para Players corretos
- Cada Player recebe contexto relevante
- Tech lead ve progresso no board
- Falha retorna erro claro