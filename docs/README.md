# Documentacao do Runtgine

Ordem de leitura recomendada: 01 → 09, depois 10–11 (gaps e protocolo).
Autoridade de decisoes: [04-decisoes.md](04-decisoes.md).

| # | Documento | Conteudo |
|---|---|---|
| 00 | 00-rascunho.md | Discussoes em andamento (nao autoridade) |
| 01 | 01-visao.md | Visao, arquitetura, relacao com Chorus |
| 02 | 02-conceitos.md | Modelo conceitual completo |
| 03 | 03-principios.md | Principios de design |
| 04 | 04-decisoes.md | Decisoes com status (fonte de verdade) |
| 05 | 05-prd.md | PRD e requisitos |
| 06 | 06-glossario.md | Glossario de termos |
| 07 | 07-stack.md | Stack tecnologica e justificativas |
| 08 | 08-workflow-templates.md | Workflow Templates, TLC Spec-Driven |
| 09 | 09-mvp.md | Corte canônico do MVP |
| 10 | 10-gaps.md | Gaps e definicoes faltantes |
| 12 | 12-board-p1.md | Board / pipeline vertical (P1) |

## Fontes historicas (raiz do repo)

| Arquivo | Papel |
|---|---|
| brainstorm.md | Visao original; desatualizado em stack (Rust/GPUI) |
| conversas-empryo.md | Consolidacao de discussoes; desatualizado em stack |
| REVIEW.md | Resumo executivo; deve espelhar docs/ |

Nao usar fontes historicas para decisoes de implementacao.

Antes de codar: P0 (`11`) e P1 Board (`12`) estao **CONFIRMADOS**.
Gaps P2 (G-30+) ainda abertos em `10-gaps.md`.
