# AGENTS.md — Guia para LLMs e contribuidores

## Papel
Arquiteto + engenheiro principal do Runtgine. Nao implemente antes
das decisoes estarem registradas em `docs/04-decisoes.md`.

## Norte absoluto

1. Runtgine e runtime que transforma intencao em execucao verificavel
2. Deterministic-first: deterministico quando possivel, LLM quando necessario
3. Player e a abstracao central (nao Agent)
4. Event-driven: Task -> Event -> Queue -> Player -> Result
5. Validacao antes da execucao (filosofia de compilador)
6. Core e o produto. Interface e superficie.
7. Muitos Players deterministicos sao estrategicos

## Autoridade documental

1. `docs/04-decisoes.md`
2. Demais `docs/` oficiais (01–09; nao `00-rascunho`)
3. Este arquivo / README / REVIEW
4. `brainstorm.md` e `conversas-empryo.md` — historicos apenas

MVP canônico: `docs/09-mvp.md`.

## Ordem de trabalho

1. Entender o dominio (docs 01 a 08 + 09-mvp)
2. Fechar Task IR v0 (schema JSON) — contrato minimo
3. Criar arquitetura do Core (pacotes Go)
4. Implementar Event Bus + Task model
5. Implementar Validator basico
6. Implementar Shell Player
7. CLI minima
8. TUI minima
9. Board Integration + pipeline vertical (ver 09-mvp)
10. Context assembly + LLM Player + Router
11. Intent Engine (NL) — apenas apos Core estavel; ainda HYPOTHESIS
12. Runtime Graph e demais HYPOTHESIS conforme promocao em 04-decisoes

## Conceitos chave (nao confundir)

- Task != Workflow != Execution Plan
- Event != Queue != Workflow
- Player != Agent
- Entry Point != Player
- Runtgine != Chorus (sao complementares)
- Intent Engine NAO e autoridade (Registry rejeita capabilities invalidas)

## O que NAO fazer

- Nao codificar antes das decisoes
- Nao tratar Runtgine como framework de agentes
- Nao confundir Runtgine com Chorus
- Nao pular o Validator (filosofia de compilador)
- Nao construir UI rica (Wails) antes do Core + CLI/TUI funcionarem
- Nao usar brainstorm/conversas-empryo como fonte de stack (Rust/GPUI estao REJECTED)
