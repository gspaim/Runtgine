# 07 — Stack Tecnologica

Decisoes de tecnologia para o Runtgine.

## Stack definida

| Camada | Tecnologia | Status |
|---|---|---|
| Linguagem | Go | CONFIRMED |
| CLI | Cobra | CONFIRMED |
| TUI | Bubble Tea + Lip Gloss + Bubbles | CONFIRMED |
| Desktop | Wails (Go + Svelte/React) | CONFIRMED |
| Event Bus | Canal Go (in-process) | CONFIRMED |
| Serializacao | JSON + JSON Schema | CONFIRMED |
| Store | SQLite (mattn/modernc) | CONFIRMED |
| Logger | slog | HYPOTHESIS |

## Por que Go

Runtgine e I/O-bound: recebe eventos, roteia para Players,
espera resposta (LLM, Docker, Git, shell), emite novos eventos.
Go e ideal para isso:
- Goroutines + channels modelam event-driven naturalmente
- Ecossistema maduro de SDKs de infraestrutura (Docker, K8s, Git)
- Single binary com cross-compile nativo
- Runtime leve, startup rapido, concorrencia barata

## Por que Cobra + Bubble Tea

Cobra: padrao de fato para CLI em Go (Docker, K8s, Hugo usam).
Bubble Tea: framework TUI Elm-architecture. Com Lip Gloss
(estilizacao) e Bubbles (componentes prontos: input, spinner,
tabela, lista, viewport, progress). Produtivo e maduro.

## Por que Wails (desktop)

Wails sobre Tauri porque mantem Go como runtime unico:
- Wails: Go apenas, chamada direta, binary unico
- Tauri: Go + Rust, IPC via gRPC/HTTP, dois binaries
Wails entrega mais rapido, com menos complexidade.
Frontend opcional (Svelte, React, Vue) — a TUI cobre 80%
dos casos, Wails cobre o resto.

## Por que Canal Go (Event Bus)

Pub/sub in-process com canais Go. Zero dependencias externas.
Suficiente para o MVP. Se precisar de fila distribuida no futuro,
troca-se a implementacao sem mudar a interface.

## Por que JSON + JSON Schema

JSON Schema para contratos. JSON para transporte.
encoding/json do Go cobre o caso de uso.
Se performance for critica: json-iterator ou protobuf.

## Por que SQLite

Persistencia local, zero config, biblioteca embutida.
mattn/go-sqlite3 (cgo) ou modernc.org/sqlite (pure Go).
Se precisar de PostgreSQL/cloud: store plugavel.

## Tecnologias REJECTED (caminho atual)

Ver status em [04-decisoes.md](04-decisoes.md).

- Rust (Core): I/O-bound para o caso; ecossistema SDKs de infra menos
  pratico que Go. REJECTED para o Core atual.
- GPUI: Exige Rust no Core. REJECTED; desktop = Wails.
- Tauri: Dois runtimes. REJECTED; Wails e mais simples.
- Electron: Pesado (+100MB).
- Python: Runtime pesado, concorrencia limitada.
- Node/TS: Single-thread, npm.