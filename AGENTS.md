# AGENTS.md — Guia para LLMs e contribuidores

## Papel
Arquiteto + engenheiro principal do Runtgine. Nao implemente antes
das decisoes estarem registradas.

## Norte absoluto

1. Runtgine e runtime que transforma intencao em execucao verificavel
2. Deterministic-first: deterministico quando possivel, LLM quando necessario
3. Player e a abstracao central (nao Agent)
4. Event-driven: Task -> Event -> Queue -> Player -> Result
5. Validacao antes da execucao (filosofia de compilador)
6. Core e o produto. Interface e superficie.
7. Muitos Players deterministicos sao estrategicos

## Ordem de trabalho

1. Entender o dominio (docs 01 a 06)
2. Definir protocolo Task IR
3. Definir Intent Engine
4. Definir Runtime Graph
5. Criar arquitetura do Core
6. Implementar Event Bus + Task model
7. Implementar Validator
8. Implementar Shell Player
9. CLI minima
10. TUI minima
11. Intent Engine basico
12. Runtime Graph
13. Context Engine
14. Player Router
15. Demais Players

## Conceitos chave (nao confundir)

- Task != Workflow != Execution Plan
- Event != Queue != Workflow
- Player != Agent
- Runtgine != Chorus (sao complementares)
- Intent Engine NAO e autoridade (Registry rejeita capabilities invalidas)

## O que NAO fazer

- Nao codificar antes das decisoes
- Nao tratar Runtgine como framework de agentes
- Nao confundir Runtgine com Chorus
- Nao pular o Validator (filosofia de compilador)
- Nao construir UI antes do Core funcionar