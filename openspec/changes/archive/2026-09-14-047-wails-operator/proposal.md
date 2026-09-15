# Proposal: 047-wails-operator

## Why

A TUI v1 (`46`) ganhou Hits/Blast, mas o desktop Wails v0 (`35`) — a
superfície que o operador consegue usar com mouse e campo de texto —
ainda não verifica a execução: LIVE mal mostrava stdout (parcialmente
corrigido em `#68`) e Hits/Blast ficaram fora de propósito no G-244.
A frase do produto quebra na janela.

## What Changes

- Canonical `docs/47-wails-operator-v1.md` (G-245..G-251 CONFIRMED)
- `docs/04-decisoes.md`: seção Desktop operator v1
- Desktop `internal/entrypoint/desktop` (**slice 40 — not this spec PR**)
- Service Wails expõe `QueryHits` + `BlastTask`
- RESULT no LIVE permanece canônico (`liveout`; não regressar `#68`)

## What Does Not Change

- Sete views e respectivos papéis (`35` G-161)
- Validator / Runner / Event Bus / Players
- Task IR schema; Claims
- TUI Charm (continua; não é substituída)
- PTY / tuios / multiplexer (REJECTED)
- GRAPH canvas 2D; Blast-from-GRAPH
- Mobile; Wails como cliente de `runtgine serve`
- NATS (G-36 DEFERRED)

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `047-wails-operator` |
| Doc canônico | [`docs/47-wails-operator-v1.md`](../../../docs/47-wails-operator-v1.md) |
| Gaps | G-245..G-251 **CONFIRMED** |
| Código | slice 40 — **bloqueado** até este pacote + `04` |

## Approach

1. Manter Wails v3 + Svelte; não nova stack
2. RESULT no LIVE a partir dos eventos (`liveout`)
3. Hits inline (LIVE ContextPack; INTENT Preview via `QueryHits`)
4. Blast painel (INTENT sem submit; LIVE no Task do snapshot); GRAPH não dispara
5. Testes de service com fake Core; CI sem exigir janela

## Impact

- `internal/entrypoint/desktop` (slice 40)
- Docs `35` (ponte), `04`, `10`, `AGENTS`, README
