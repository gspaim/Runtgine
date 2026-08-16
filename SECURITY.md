# Política de segurança

O Runtgine está em fase MVP e ainda não possui releases estáveis.

## Versões suportadas

Somente o código mais recente da branch `main` recebe correções de segurança.

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

O Shell Player do MVP não é um sandbox de segurança. Tasks não confiáveis,
isolamento entre tenants e execução hostil estão fora das garantias atuais.
Mesmo assim, bypasses que contradigam as proteções documentadas devem ser
relatados.
