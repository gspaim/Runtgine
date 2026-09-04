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
- MVP: CLI + TUI + Board; API HTTP v0 spec `34` (pos-1.0, slices 25–26);
  Wails Fase 3
- Desktop = Wails (nao GPUI)

### Ainda em aberto

- Board Integration: webhook **inbound** GitHub no longo prazo?
  (polling G-20 permanece; outbound de Run = `34` G-156)
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
- Board: webhook inbound GitHub no longo prazo (G-20 = polling)
- Detalhes GitHub Projects (G-20+; em `12`)

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

---

## TUI v1 (discussao, pre-decisao)

Contexto: a TUI Constellation Mission Control ja existe e e autoridade
de design (`14-tui-design.md` + skill `.cursor/skills/runtgine-tui-design/`).
Implementada nos slices 3 (base), 14 (aba GRAPH) e 21 (aba INTENT).
Esta secao abre a discussao do **proximo ciclo de TUI** — nada aqui e
CONFIRMED; qualquer recorte precisa de decisao em `04-decisoes.md`
antes de spec/codigo.

Recortes candidatos (todos eram exclusoes ou REJECTED do v0, e por isso
exigem decisao nova):

1. **PTY / tuios — terminal vivo dentro da TUI** (maior valor, maior risco)
   - Hoje a TUI observa runs; nao executa nem multiplexa terminais.
     `tuios no MVP` foi REJECTED ("Nao e multiplexer; PTY futuro exige
     nova decisao").
   - Exige: multiplexador de sessoes PTY, lifecycle de processos
     filhos, seguranca (o que pode rodar dentro?), keymap novo,
     scrollback/resize.
   - Valor: mission control de verdade — acompanhar E intervir.

2. **GRAPH canvas 2D** (UX visual)
   - v0 da aba GRAPH e lista/detalhe por decisao (G-107: "sem canvas
     2D"). Um canvas de nos/arestas (player → capability → task → run)
     mudaria a forma de navegar a memoria estrutural.
   - Custo alto de UX/teclado vs. ganho incremental sobre a lista;
     candidate a ficar por ultimo.

3. **Hits + Blast surfaces** (menor, fecha o loop de verificacao)
   - Os dados de verificacao ja existem no Core mas hoje so aparecem
     em CLI/API: `graph_hits`/`memory_hits`/`playbook_hits` no detalhe
     do step, e o Impact Report do Blast (`runtgine blast`) navegavel
     a partir de um Run/Task.
   - Segue `14` + skill; nao cria conceito novo; reusa `QueryHits`,
     `BlastTask` e o Memory Provider.

Ordem sugerida pelo arquiteto: **3 → 1 → 2**. Hits+Blast primeiro
(pequeno, coerente com a frase do produto, zero conceito novo);
PTY/tuios como a grande aposta do v1 se a discussao confirmar valor;
canvas 2D por ultimo (puro UX, sem dado novo).

Questoes abertas para decidir antes de promover qualquer recorte:
- Qual(is) recorte(s) entram no ciclo TUI v1?
- PTY/tuios: qual o caso de uso concreto (ver saida de testes longos?
  intervir em shell?) e quais guardas de seguranca?
- Hits/Blast: surfacem inline no detalhe do step ou aba propria
  (`14` fixa seis abas — aba nova exigiria emenda de design)?
- Prioridade vs. estabilizar a release 0.1.0 (limitacoes conhecidas:
  steps sequenciais, cancelamento nao coordenado).
