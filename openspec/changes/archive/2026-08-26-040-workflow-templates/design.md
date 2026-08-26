# Design: 040-workflow-templates

## Pacote

`internal/core/templates` com `Load(dir)`, `Compile(tpl, ep, ref,
summary)`, `Lookup(list, id)`. Não é Player: não entra no Registry.

Espelha `internal/core/playbooks` (boot best-effort, skip inválido).

## Schema

```json
{
  "schema_version": "0.1.0",
  "id": "verify",
  "title": "Verify workspace",
  "steps": [
    {
      "step_id": "status",
      "capability": "git.status",
      "input": {"workdir": "."}
    },
    {
      "step_id": "test",
      "capability": "test.go",
      "input": {"workdir": "."},
      "depends_on": ["status"]
    }
  ]
}
```

Validação no load: id/title, 1–20 steps, `step_id` único,
`depends_on` intra-arquivo acíclico, `additionalProperties` false
via unmarshal estrito (unknown field → skip arquivo).

## Compile

Gera `task.Task` com `metadata.template = id`. Não resolve Player
nem consulta Registry. `SubmitTask` / Validator fazem o resto.

## Intent

`Engine.Templates` preenchido em `api.Open`. `matchTemplate` roda
depois de `matchPlayer` e **antes** de `matchShell`. Prefixo
conhecido + id ausente → `validation.invalid_input`.

## Graph

`graph.KindTemplate = "template"`. `RefreshFromTemplates` no boot
após `RefreshFromRegistry`. Falha → warning.

## CLI

Subcomandos em `root.go`, padrão Memory (`list`/`show` JSON;
`run` reusa `SubmitTask` + `--wait` / `--dry-run`).
