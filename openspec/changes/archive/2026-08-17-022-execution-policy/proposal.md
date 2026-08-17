# Proposal: 022-execution-policy

## Why

Sandbox por Player não decide *se* uma capability pode correr. Sem
allow/deny/approval-required no Core, HITL não existe e o Docker Player
não tem porta de aprovação.

## What Changes

- Motor de Execution Policy no Core (não é Player)
- Verbos `allow` | `deny` | `approval-required` por capability exata
- Precedência: default allow < Manifest < config.json < env default
- Estado de Run `waiting_approval` persistido + eventos
- Core API `ApproveRun` e CLI `runtgine approve` / `runtgine deny`
- TUI RUNS/LIVE: status + teclas `a`/`d` (sem aba nova)

## What Does Not Change

- Schemas Task IR (além de campos opcionais no Manifest)
- Sandbox Shell/Git/FS
- Resource Claims / Blast Radius
- Human Player
- Docker Player (023)
- Tab GRAPH / Board write-back de approve

## Status / autoridade

| Item | Valor |
|---|---|
| Change id | `022-execution-policy` |
| Doc canônico | [`docs/22-execution-policy-v0.md`](../../docs/22-execution-policy-v0.md) |
| Gaps | G-81..G-86 **CONFIRMED** (recorte de G-42) |
| Código | Ainda não — este pacote autoriza o slice 10 |

## Approach

1. Pacote `internal/core/policy` resolve o verbo efetivo
2. Runner: deny na admissão; approval-required pausa antes de Execute
3. Persistir pending approval no store
4. Entry points CLI e TUI chamam só `ApproveRun`

## Impact

- Runner e store ganham estado novo
- Manifest pode declarar `execution_policy` opcional
- Config ganha bloco `execution_policy`
- Eventos novos no envelope
