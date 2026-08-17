# Tasks: 022-execution-policy

## 1. Policy engine

- [ ] 1.1 Scaffold `internal/core/policy` (verbs, Resolve, table)
- [ ] 1.2 Load `execution_policy` from config.json + `RUNTGINE_POLICY_DEFAULT`
- [ ] 1.3 Optional `execution_policy` on Manifest capabilities
- [ ] 1.4 Reject unknown verbs / unknown capability keys at config load

## 2. Runner + store

- [ ] 2.1 Status `waiting_approval` + pending step fields
- [ ] 2.2 Deny at admission (`policy.denied`, no Execute)
- [ ] 2.3 Pause before Execute when approval-required
- [ ] 2.4 Events `run.waiting_approval` / `approval_granted` / `approval_denied`
- [ ] 2.5 Restart: do not re-execute until ApproveRun

## 3. API + CLI

- [ ] 3.1 `ApproveRun(grant|deny)` + `policy.not_waiting`
- [ ] 3.2 `runtgine approve` / `runtgine deny`
- [ ] 3.3 `status` shows pending; `run --wait` waits through HITL

## 4. TUI

- [ ] 4.1 RUNS/LIVE render `waiting_approval` (Amber + label)
- [ ] 4.2 Keys `a` / `d` call Core; footer when waiting

## 5. Tests + closeout

- [ ] 5.1 Default allow: hello.json still succeeds
- [ ] 5.2 deny / grant / deny-human / not_waiting
- [ ] 5.3 `go test ./...` and `go vet ./...`
- [ ] 5.4 README Estágio: Slice 10 Feito; next = 023 Docker
- [ ] 5.5 Archive this change into `openspec/specs/execution-policy/`
