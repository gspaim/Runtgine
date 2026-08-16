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

- Board Integration: polling vs webhook no longo prazo?
- Detalhes do adapter GitHub Projects (auth, mapeamento card → Task IR) — gap P1

Resolvido em proposta (`11-protocolo-v0`, aguardando confirmacao):
- Protocolo Entry Point → Core = mesmo protocolo interno (adapters)
- Board e Entry Point/adapter, nao Player

### Implicacao para serverless/cloud (design, nao MVP)

1. Event Bus interno em memoria no MVP
2. Cloud: Event Bus troca por fila externa sem mudar a interface
3. Persistence plugavel: SQLite (local) -> RDS (cloud)
4. Context assembly deve poder ser stateless no futuro
5. Players remotos via protocolo, nao carga direta

---

## Gaps e protocolo v0

Formalizado em:
- [10-gaps.md](10-gaps.md) — inventario P0/P1/P2/P3
- [11-protocolo-v0.md](11-protocolo-v0.md) — schemas e cortes PROPOSED

Proximo passo humano: percorrer o checklist de confirmacao no fim de `11`
e promover itens em `04-decisoes.md`.

Questoes de Entry Point que `11` ja propoe resposta:
- Protocolo Entry Point → Core = mesmo protocolo interno (adapters)
- Board = Entry Point/adapter que emite Task IR (detalhe GitHub = gap P1)

Ainda em aberto apos `11` (nao bloqueia Core CLI+Shell):
- Board: polling vs webhook no longo prazo
- Detalhes GitHub Projects (G-20+)

## Context Management (discussao)

O que temos hoje:
- AGENTS.md: entry point para LLMs
- docs/ numerados por ordem de leitura
- docs/04-decisoes.md: decisoes com status
- docs/10-gaps.md + docs/11-protocolo-v0.md: gaps e propostas de contrato
- brainstorm.md / conversas-empryo.md: fontes historicas

Falta documentar (quando houver codigo):
- Convencoes de codigo alem do layout proposto em `11`
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
