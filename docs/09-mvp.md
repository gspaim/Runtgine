# 09 — MVP: Runtime Minimo

MVP canônico do Runtgine: provar o **Core** (event-driven, Players,
validação) com um Player determinístico real, superfícies mínimas
(CLI/TUI) e um primeiro cenário vertical via Board.

Fonte de verdade deste escopo. `05-prd.md` lista requisitos; este
documento define o corte do MVP. Em conflito, prevalece este arquivo
e `04-decisoes.md`.

## Principio do corte

1. Core antes de UI rica (CLI/TUI mínimas bastam)
2. Shell Player no MVP — prova deterministic-first
3. Entrada estruturada (Task IR v0 em JSON/YAML) antes de Intent Engine NL
4. Board como primeiro Entry Point de produto, não como substituto do Core

## Ciclo principal (cenario vertical)

Board (Entry Point) -> Import task -> Technical Review ->
Spec Review -> Repo Search -> Effort Estimation ->
Difficulty Classification -> Task Decomposition ->
Task Router -> Context Assembly -> Players

O ciclo vertical roda **sobre** o Core (Event Bus, Registry, Validator),
não no lugar dele.

## Entry points no MVP

| Entry Point | Incluido? | Notas |
|---|---|---|
| CLI | MVP | `runtgine run`, `status` — entrada estruturada |
| TUI | MVP | Superfície mínima para observar execuções |
| Board (Github Projects) | MVP | Polling inicial; Entry Point ≠ Player |
| API | Pos-MVP | Serverless/CI |
| Desktop (Wails) | Pos-MVP (Fase 3) | Após CLI/TUI validarem o Core |
| Web | Futuro | Se houver demanda |

## Escopo do MVP

### Incluido (Core)

- Task model + Task IR v0 (JSON/JSON Schema; entrada estruturada)
- Task Validator básico (capabilities, inputs, schemas)
- Event Bus in-process (canais Go)
- Player Registry + Manifest
- Shell Player (Player determinístico de referência)
- CLI mínima
- TUI mínima (observação de status/eventos)

### Incluido (cenario Board)

- Board Integration (ler tasks do Github Projects)
- Pipeline de analise: technical review, spec review
- Repo Search
- Effort Estimation + Difficulty Classification
- Task Decomposition (regras + LLM)
- Task Router básico
- Context assembly básico
- Pelo menos um LLM Player no pipeline

### Nao incluido

- Intent Engine de linguagem natural (permanece HYPOTHESIS; ver P1)
- Workflow engine completo / Workflow Templates
- Human-in-the-loop completo
- Policies / Approvals / Resource Claims / Blast Radius
- Plugin system
- Wails (desktop)
- MCP integration
- Event sourcing
- API HTTP
- NATS / Event Bus distribuído
- Biblioteca ampla de Players (Git, Docker, K8s…)

## Ordem de implementacao

Alinhada a `AGENTS.md`:

1. Task model + Task IR v0 (schema)
2. Player Registry + Manifest
3. Event Bus
4. Task Validator básico
5. Shell Player
6. CLI mínima
7. TUI mínima
8. Board Integration
9. Context assembly
10. LLM Player (Technical Review) + pipeline linear
11. Repo Search
12. Effort Estimation + Difficulty
13. Task Decomposition
14. Task Router
15. Demais Players LLM do cenário

## Criterios de sucesso

- `runtgine run` executa Task IR v0 via Shell Player com eventos observáveis
- Validator rejeita capability inexistente / input inválido antes de executar
- Task do board passa pelo pipeline completo quando o cenário vertical estiver ligado
- Subtasks distribuídas para Players corretos
- Cada Player recebe contexto relevante (quando Context assembly existir)
- Falha retorna erro claro na CLI/TUI

## Criterio de “pronto para codar”

Pode iniciar implementacao do Core quando G-01..G-18 estiverem
**CONFIRMED** (ou explicitamente REJECTED com alternativa) em `04-decisoes.md`.
Ver propostas em [11-protocolo-v0.md](11-protocolo-v0.md).

Nota: o incluso “Board” em `09-mvp.md` permanece no MVP de produto, mas
o **primeiro slice de codigo** e CLI + Shell sobre o protocolo v0; Board e P1
de especificacao (G-20+).
