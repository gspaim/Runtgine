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

O desktop futuro continua Wails v3 + Svelte + shadcn-svelte (spec `35`).
A TUI valida a linguagem de interacao antes do desktop. View **INTENT**
no Wails espelha a aba INTENT (`32` / G-144).

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
[ INTENT ] [ RUNS ] [ LIVE ] [ BOARD ] [ EVENTS ] [ GRAPH ] [ CONFIG ]
```

INTENT e a primeira aba (Entry Point visual). Spec: [32-intent-surface-v0.md](32-intent-surface-v0.md).

Footer:

```text
tab/shift+tab navigate · enter inspect · c cancel · / filter · q quit
(+ a approve · d deny when selected run is waiting_approval)
(+ Ctrl+p preview · Ctrl+Enter submit on INTENT tab)
```

## Aba INTENT

Superficie de Entry Point (Mission Brief). Nao e chatbot — compila intencao
em Task IR e submete Run via Core (`CompileIntent` / `SubmitIntent`).

Mostra:

- campo NL multilinha (ou modo JSON Task IR);
- preview Task IR + `method` (`heuristic.*` | `llm`);
- erros de compilacao/Validator;
- historico curto da sessao (ultimas submissoes: `run_id` + resumo).

Fluxo v0:

1. operador digita NL (ou cola JSON);
2. `Ctrl+p` → preview (`CompileIntent`, source `tui`);
3. `Ctrl+Enter` → submit (`SubmitIntent`) → `run_id`;
4. TUI seleciona run e abre **LIVE**.

Regras:

- nunca chama Player direto;
- mesma soberania Validator/Registry que CLI `runtgine intent`;
- sem thread conversacional infinita;
- HITL continua em LIVE/RUNS (`a`/`d`), nao nesta aba.

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

## Aba GRAPH

Read-only sobre o Runtime Graph do workspace (spec `26`, G-105..G-110).
Não substitui LIVE (LIVE = trajetória de **um** Run).

Mostra:

- counts `nodes` / `edges` e por `node_kind`;
- lista `kind` + `id` (ordem G-61, depois id);
- detalhe do nó selecionado: attrs + arestas incidentes (texto).

`r` na aba chama `RefreshGraph` e recarrega o snapshot. `/` filtra
kind/id. Em `< 80` colunas: lista vertical; sem diagrama horizontal.

Core APIs: `GetGraphSnapshot`, `RefreshGraph`. Sem Player, sem canvas 2D,
sem Blast/Hits nesta aba.

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

## Fora do MVP da TUI (implementacao pendente)

- aba **INTENT** (slice 21; spec confirmada em `32`) — docs prontas, codigo depois;
- tuios / terminal multiplexer;
- PTY interativo embutido;
- edicao rica de config;
- Runtime Graph “completo” (genome, AST contínuo, grafo federado);
- walk Blast←Graph na TUI;
- acesso web/SSH;
- Wails v0 (spec `35`; slices 27–28; inclui INTENT desktop espelhando `32`).

## Skill

Antes de criar ou alterar a TUI, ler:

`.cursor/skills/runtgine-tui-design/SKILL.md`

## Implementacao do Slice 3

- Entry Point: `internal/entrypoint/tui/`
- Comando: `runtgine tui`
- Stack: Charm v2 (`charm.land/bubbletea/v2`, `lipgloss/v2`, `bubbles/v2`)
- Core APIs: `ListRuns`, `GetRun`, `ListRecentEvents`, `Subscribe`,
  `CancelRun`, `ConfigSnapshot`, `GetGraphSnapshot`, `RefreshGraph`
- Config permanece read-only; o snapshot nao expoe tokens ou API keys
- Cancelamento exige confirmacao e persiste o estado de runs orfaos de um
  processo CLI anterior
- Testes cobrem navegacao, resize, **seis** tabs (slice 3–14; **sete** apos slice 21 INTENT),
  filtro GRAPH/EVENTS, cancelamento e `NO_COLOR`
