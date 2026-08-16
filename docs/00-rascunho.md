# 00 — Rascunho (discussoes em andamento)

Este documento captura discussoes e decisoes em andamento que ainda nao
foram formalizadas nos documentos oficiais. Quando uma discussao amadurece,
o conteudo migra para o documento apropriado.

**Nao e autoridade.** Em conflito, prevalece `04-decisoes.md`.

---

## Multi-entry-point architecture (parcialmente formalizado)

### Decidido (ver 04-decisoes / 09-mvp)

- Entry Point != Player
- Core unico; entry points variados (Board, CLI, TUI, API, Desktop/Wails, Web)
- MVP: CLI + TUI + Board; API e Wails pos-MVP
- Desktop = Wails (nao GPUI)

### Ainda em aberto

- O protocolo entre Entry Point e Core e o mesmo Runtgine Protocol
  ou um protocolo de adaptacao separado?
- Board Integration: polling vs webhook no longo prazo?
- Board e um Entry Point direto ou adapter sobre um Entry Point tipo API?

### Implicacao para serverless/cloud (design, nao MVP)

1. Event Bus interno em memoria no MVP
2. Cloud: Event Bus troca por fila externa sem mudar a interface
3. Persistence plugavel: SQLite (local) -> RDS (cloud)
4. Context assembly deve poder ser stateless no futuro
5. Players remotos via protocolo, nao carga direta

---

## Context Management (discussao)

O que temos hoje:
- AGENTS.md: entry point para LLMs
- docs/ numerados por ordem de leitura
- docs/04-decisoes.md: decisoes com status
- docs/00-rascunho.md: discussoes em andamento
- brainstorm.md / conversas-empryo.md: fontes historicas

Falta documentar (quando houver codigo):
- Estrutura de diretorios do projeto Runtgine
- Convencoes de codigo e documentacao
- Genome / mapa de simbolos
- Como o contexto e gerenciado entre modulos

---

## SDD / TLC Spec-Driven

Formalizado em [08-workflow-templates.md](08-workflow-templates.md).

Resumo da posicao:
- SDD nao e um Player; e Workflow Template (+ gates / playbook)
- Templates carregados de repositorios externos = opcao preferida (em aberto)

Questao em aberto permanece: templates nativos no Graph vs registro dinamico
a partir de repos externos.
