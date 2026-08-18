# Design: 022-execution-policy

## Technical approach

### Package

`internal/core/policy` com:

- `type Verb string` — `allow`, `deny`, `approval-required`
- `Resolve(default Verb, manifestVerb, configVerb) (Verb, error)`
- `Table` carregada do `config.Config` + Manifest no Register

O Runner pergunta `Resolve(capability)` **depois** do Validator e
**antes** de `Player.Execute`.

Deny de qualquer step do Plan: falha em `SubmitTask` / admissão, sem
`InsertRun` (ou com rejeição equivalente a `task.rejected` hoje).

### Persistence

Estender `store.Status` com `waiting_approval`. Guardar no Run (coluna
JSON ou campos): `pending_step_id`, `pending_capability`, `pending_player`.

Restart: se status é `waiting_approval`, o Runner **não** despacha até
`ApproveRun`.

### Events

Reusar envelope de `11`. Types novos no switch do bus/store:

- `run.waiting_approval`
- `run.approval_granted`
- `run.approval_denied`

### API

`Core.ApproveRun(ctx, runID, decision)` com `grant` | `deny`.

CLI: novos subcomandos ao lado de `status` / `cancel`.
`run --wait` faz poll até terminal, incluindo espera de HITL.

### TUI

Aba RUNS/LIVE existentes. Token Amber para `waiting_approval`.
Key `a` / `d` só com seleção em espera. Seguir skill TUI: só Core API.

### Manifest

Campo JSON opcional por capability: `"execution_policy": "approval-required"`.
Ausente = herda default. Schema do Manifest no embed precisa aceitar o campo.

## Alternatives considered

| Alternativa | Por que não |
|---|---|
| Human Player `approval.grant` | Confunde Entry Point com Player |
| Wildcards `docker.*` | Superfície ambígua no v0 |
| Deny só no Execute | Viola filosofia de compilador (efeitos parciais) |
| Flag CLI `--auto-approve` | Ativa HITL bypass no produto; testes usam API |

## Risks

| Risco | Mitigação |
|---|---|
| Testes 1–9 quebram | Default allow |
| Duplo Execute após restart | Persistir pending; grant é idempotente no step |
| TUI scope | Sem aba nova; duas teclas |

## Packages touched

- `internal/core/policy` (novo)
- `internal/core/runner`, `store`, `api`, `event`
- `internal/config`
- Manifest schema / Register
- `internal/entrypoint/cli`, `internal/entrypoint/tui`
- `docs/11-protocolo-v0.md` (já estendido neste PR de spec)
