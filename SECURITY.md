# Política de segurança

O Runtgine está em fase MVP e ainda não possui releases estáveis.

## Versões suportadas

| Canal | Suporte de segurança |
|---|---|
| Tip de `main` | Sim (código liberável) |
| Tip de `develop` | Pré-release; correções entram via PR |
| Tags `vX.Y.Z-rc.N` | Só até a estável correspondente |
| Tags `vX.Y.Z` | Após a primeira estável; ver fluxo em `docs/15-git-workflow.md` |

Enquanto não houver tag estável publicada, reporte contra o tip de `main`
(ou o commit afetado).

## Relatando uma vulnerabilidade

Prefira o
[Private Vulnerability Reporting do GitHub](https://github.com/gspaim/Runtgine/security/advisories/new).

Se o canal privado não estiver disponível, abra uma issue contendo apenas um
pedido de contato privado. Não publique exploits, credenciais, dados sensíveis
ou detalhes suficientes para reproduzir a vulnerabilidade.

Inclua no relato privado:

- componente e versão/commit afetado;
- impacto observado;
- passos mínimos para reprodução;
- mitigação sugerida, se houver.

## Escopo importante

O Shell Player do MVP não é um sandbox de isolamento (sem namespaces,
Landlock ou deny de rede). Tasks não confiáveis, isolamento entre tenants e
execução hostil estão fora das garantias atuais.

Proteções documentadas do sandbox v0: argv-only (sem shell implícito),
timeout, `workdir` resolvido com symlinks e confinado ao workspace, e
herança mínima de ambiente quando `input.env` é omitido (sem tokens /
`RUNTGINE_*`). Bypasses dessas proteções devem ser relatados.
