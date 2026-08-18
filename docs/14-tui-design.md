# 14 — TUI Design: Constellation Mission Control

Direcionamento visual e funcional da TUI do Runtgine.

**Status: IMPLEMENTED (Slice 3).**

## Conceito

Nome do sistema visual: **Constellation Mission Control**.

A TUI combina:

- estrutura operacional de um centro de controle de missao;
- identidade visual inspirada em constelacoes;
- linguagem tecnica real do Runtgine.

O tema espacial e uma camada visual. Nao renomear conceitos do dominio:
continuar usando `Run`, `Task`, `Step`, `Event`, `Player` e `Board`.

## Principio de arquitetura

A TUI e uma superficie sobre o Core:

- usa `GetRun`, `Subscribe`, `CancelRun` e APIs publicas do Core;
- nunca chama Shell Player, LLM Player ou Board adapter diretamente;
- nao e terminal multiplexer;
- nao incorpora tuios no MVP;
- PTY/live terminal pode ser avaliado depois como widget isolado.

## Stack

- Bubble Tea
- Lip Gloss
- Bubbles

O desktop futuro continua Wails + Svelte. A TUI valida a linguagem de
interacao antes do desktop.

## Sistema visual

Paleta:

| Token | Cor | Uso |
|---|---|---|
| `space` | `#070B14` | Fundo principal |
| `panel` | `#101A2F` | Paineis |
| `starlight` | `#8FB8FF` | Selecao, links, tabs |
| `violet` | `#9D8CFF` | Destaque secundario |
| `telemetry` | `#73E2C2` | Sucesso / conexao |
| `amber` | `#FFB86B` | Running / atencao |
| `anomaly` | `#FF6B8A` | Falha / cancelamento |
| `muted` | `#64748B` | Texto secundario |

Regras:

- contraste e legibilidade acima de efeitos;
- brilho discreto apenas no foco/status ativo;
- bordas, grids, coordenadas e estrelas com moderacao;
- sem estetica arcade, holograma ou HUD de jogo;
- degradar corretamente em terminais sem true color;
- nao depender de glifos que quebrem largura em terminais comuns.

## Estrutura global

Header persistente:

```text
✦ RUNTGINE / CONSTELLATION MISSION CONTROL
workspace ~/proj · local ●
```

Tabs:

```text
[ RUNS ] [ LIVE ] [ BOARD ] [ EVENTS ] [ CONFIG ]
```

Footer:

```text
tab/shift+tab navigate · enter inspect · c cancel · / filter · q quit
(+ a approve · d deny when selected run is waiting_approval)
```

## Aba RUNS

Tabela de execucoes:

- sinal/status;
- `run_id` curto;
- intent/mission;
- estado;
- tempo decorrido.

Estados visuais:

- running: amber;
- waiting_approval: amber + label WAITING (HITL; teclas `a` grant / `d` deny);
- succeeded: telemetry;
- failed/cancelled: anomaly;
- selected: trilho violeta + starlight.

## Aba LIVE

Detalhe do Run selecionado:

- `run_id`, intent e status;
- steps como constelacao/trajectory legivel;
- ligacoes representam `depends_on`;
- concluido = telemetry;
- atual = amber com pulso discreto;
- waiting_approval = amber + step gated visivel ate grant/deny;
- pendente = starlight/muted;
- progress bar;
- Current Step (Player, capability, ContextPack);
- ticker de telemetria.

O grafo nao pode prejudicar a leitura em terminais estreitos: usar lista
vertical como fallback.

## Aba BOARD

Board terminal sincronizado com GitHub:

```text
INTAKE | IN FLIGHT | LANDED
```

Cada card mostra:

- issue/card;
- titulo;
- fonte;
- `run_id`;
- estado do pipeline.

Detalhes: polling, write-back e ultima sincronizacao. O Board nao cria
subtasks/cards filhos no MVP; subtasks permanecem no Core/SQLite.

## Aba EVENTS

Stream de telemetria com:

- UTC;
- event type;
- run;
- step;
- player;
- filtro `/` (ex.: `run:... type:step.*`);
- painel de payload JSON.

## Aba CONFIG

Somente leitura no v0:

- workspace;
- runtime / concorrencia;
- SQLite;
- backend LLM;
- GitHub;
- precedencia: `defaults < config.json < env < CLI flags`.

Secrets sempre mascarados.

## Responsividade

| Largura | Layout |
|---|---|
| >= 120 | Tabs + paineis lado a lado |
| 80–119 | Um painel principal + drawer/detalhe |
| < 80 | Lista vertical; tabs compactas; sem grafo horizontal |

## Acessibilidade

- status nunca depende apenas de cor: usar texto e simbolo;
- respeitar `NO_COLOR`;
- modo ASCII para terminais incompatíveis;
- foco sempre visivel;
- atalhos mostrados no footer;
- animacoes opcionais e desativaveis.

## Fora do MVP da TUI

- tuios / terminal multiplexer;
- PTY interativo embutido;
- edicao rica de config;
- Runtime Graph completo;
- acesso web/SSH;
- Wails.

## Skill

Antes de criar ou alterar a TUI, ler:

`.cursor/skills/runtgine-tui-design/SKILL.md`

## Implementacao do Slice 3

- Entry Point: `internal/entrypoint/tui/`
- Comando: `runtgine tui`
- Stack: Charm v2 (`charm.land/bubbletea/v2`, `lipgloss/v2`, `bubbles/v2`)
- Core APIs: `ListRuns`, `GetRun`, `ListRecentEvents`, `Subscribe`,
  `CancelRun` e `ConfigSnapshot`
- Config permanece read-only; o snapshot nao expoe tokens ou API keys
- Cancelamento exige confirmacao e persiste o estado de runs orfaos de um
  processo CLI anterior
- Testes cobrem navegacao, resize, cinco tabs, filtro, cancelamento e
  `NO_COLOR`
