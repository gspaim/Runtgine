# 34 — HTTP API v0 (Entry Point)

Superfície HTTP do runtime: adapter que expõe a Core API (`11` §13)
para CI/CD e clientes remotos locais. Fecha UC-02 sem substituir CLI,
TUI, Board ou o HTTP Player cliente (`28`).

Inventário: [10-gaps.md](10-gaps.md) (G-45; recorte G-153+).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Protocolo: [11-protocolo-v0.md](11-protocolo-v0.md).
HTTP Player (cliente GET/HEAD): [28-http-player-v0.md](28-http-player-v0.md).
Board inbound continua polling: [12-board-p1.md](12-board-p1.md).

**Status deste doc: CONFIRMED (v0 spec).** Código slices 25–26 **feito**.

**Pacote OpenSpec:** [`openspec/changes/archive/2026-08-21-034-http-api/`](../openspec/changes/archive/2026-08-21-034-http-api/).

---

## 1. Problema

O Core já aceita Task IR, Intent, status, cancel, HITL e blast via
processo local (CLI/TUI/Board). CI/CD (UC-02) hoje só entra chamando o
binário. Isso impede um job remoto de submeter/observar um Run sem
empacotar a CLI, e mistura “cliente HTTPS de leitura” (`http.get`) com
“servidor do runtime”.

Falta um Entry Point HTTP **magro**: mesmo protocolo interno, JSON na
borda, auth por token, loopback por default.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Entry Point adapter (`internal/entrypoint/httpapi`) | Player; HTTP Player (`28`) |
| `net/http` stdlib sobre `api.Core` | Framework web; OpenAPI gerado no v0 |
| JSON Task IR / Intent → `SubmitTask` / `SubmitIntent` | YAML (YAML permanece borda CLI, G-14) |
| SSE de eventos de um Run | WebSocket; NATS; event sourcing |
| Webhook **outbound** de eventos terminais (slice 26) | Webhook **inbound** GitHub (Board continua polling) |
| Bearer token + bind loopback | SaaS multi-tenant; OAuth; mTLS no v0 |

Regras (inalteradas):

1. Entry Point ≠ Player. O servidor **não** chama Player.
2. Validator / Registry soberanos — HTTP não inventa capability.
3. Core não importa `entrypoint` (`11` §16).
4. Distinto de `http.get` / `http.head` (`28`).

---

## 3. Cortes confirmados (G-153+)

### G-153 — Papel e pacote

**Status: CONFIRMED**

- Nome: **HTTP API v0**. Comando: `runtgine serve`.
- Pacote: `internal/entrypoint/httpapi`.
- `source.entry_point` nas Tasks submetidas = `"http"` (salvo o cliente
  enviar `source` já preenchido e válido).
- `source.ref` default = path + método (ex.: `POST /v0/tasks`).
- Um processo = um `workspace_root` (G-33). O servidor não multiplexa
  workspaces.

### G-154 — Listen, auth, limites

**Status: CONFIRMED**

Config (precedência G-38: defaults < file < env < flags):

| Campo | Default | Env / flag |
|---|---|---|
| `api.listen` | `127.0.0.1:7420` | `RUNTGINE_API_LISTEN` / `--listen` |
| `api.token` | vazio | `RUNTGINE_API_TOKEN` (nunca no JSON de config) |
| `api.max_body_bytes` | `1048576` (1 MiB) | `RUNTGINE_API_MAX_BODY_BYTES` |

Regras v0:

- TLS **não** é do binário; termina no reverse proxy se precisar.
- Token: header `Authorization: Bearer <token>`. Comparação em tempo
  constante.
- `GET /v0/healthz` **não** exige token (liveness). Demais rotas exigem.
- Se o endereço de listen **não** for loopback (`127.0.0.1` / `::1`) e
  o token estiver vazio → **recusar boot**.
- Loopback + token vazio: boot permitido com `slog.Warn` (dev local).
- CORS: nenhum no v0 (não é API de browser).
- Body acima do limite → `413` + Error model (`validation.input`).

### G-155 — Rotas (slice 25)

**Status: CONFIRMED**

Prefixo `/v0/`. Content-Type `application/json; charset=utf-8`,
exceto SSE.

| Método | Path | Core | HTTP |
|---|---|---|---|
| GET | `/v0/healthz` | — | `200 {"ok":true}` |
| POST | `/v0/tasks` | `SubmitTask` | `202 {"run_id"}` |
| POST | `/v0/intent` | `SubmitIntent` | `202 {"run_id","method"}` |
| POST | `/v0/intent/preview` | `CompileIntent` | `200` Task IR + `method` (sem Run) |
| GET | `/v0/runs/{id}` | `GetRun` | `200` snapshot |
| GET | `/v0/runs` | `ListRuns` | `200`; query `limit` (default 20, max 100) |
| GET | `/v0/runs/{id}/events` | `Subscribe` filtrado | SSE `text/event-stream` |
| POST | `/v0/runs/{id}/cancel` | `CancelRun` | `202 {"ok":true}` |
| POST | `/v0/runs/{id}/approve` | `ApproveRun(grant)` | `202` |
| POST | `/v0/runs/{id}/deny` | `ApproveRun(deny)` | `202` |
| POST | `/v0/blast` | `BlastTask` | `200` Impact Report (não cria Run) |

`POST /v0/intent` body: `{"text":"…"}` (NL) **ou** Task IR JSON
completo (mesmo critério da aba INTENT / CLI: schema válido →
`SubmitTask` direto).

Mapeamento de erro (Error model de `11` §9):

| Situação | HTTP |
|---|---|
| Token ausente / inválido | `401` `auth.unauthorized` |
| JSON / schema / capability | `400` (`validation.*`) |
| Run inexistente | `404` `runtime.not_found` |
| `ApproveRun` fora de `waiting_approval` | `409` `policy.not_waiting` |
| `claim.conflict` na admissão | `409` |
| Falha interna | `500` `runtime.internal` |

SSE v0: um evento JSON (envelope `11` §7) por linha `data:`; fecha no
terminal do Run (`run.succeeded` / `failed` / `cancelled`) ou no
disconnect do cliente. Heartbeat comentário `: ping` a cada 15s.
Query `?after=` **fora** do v0 (sem replay HTTP; SQLite continua SoT
via `GetRun` / CLI).

Fora do v0 de rotas: Graph, Memory CRUD, CONFIG dump, OpenAPI UI,
multipart upload.

### G-156 — Webhooks outbound (slice 26)

**Status: CONFIRMED**

Após o serve estável, notificar CI sem polling:

```json
{
  "webhooks": [
    {
      "id": "ci-main",
      "url": "https://example.invalid/hooks/runtgine",
      "events": ["run.succeeded", "run.failed", "run.cancelled"]
    }
  ]
}
```

Regras v0:

- Só eventos **terminais de Run** da lista acima.
- POST do envelope de Event (`11` §7); timeout 5s; 1 retry.
- Falha de entrega → `slog.Warn`; **não** falha o Run.
- Segredo opcional `RUNTGINE_WEBHOOK_SECRET` (ou `webhooks[].secret` só
  via env): header `X-Runtgine-Signature: sha256=<hex>` HMAC-SHA256 do
  body. Sem secret → sem header.
- Destino: `https` somente; deny link-local / metadata (reusar política
  de URL do HTTP Player `28` G-119).
- **Não** é webhook inbound do GitHub. Board permanece polling (G-20).

### G-157 — Exclusões v0

**Status: CONFIRMED**

| Fica fora | Por quê |
|---|---|
| HTTP Player POST / `http.get` via esta API | Player ≠ Entry Point (`28`) |
| Webhook inbound GitHub / Board | G-20 polling; outro recorte |
| TLS no binário / mTLS / OAuth | Reverse proxy; token basta no v0 |
| NATS / bus distribuído | G-36 DEFERRED |
| MCP | G-44 |
| Wails / TUI chamando HTTP em vez do Core | Superfícies locais continuam in-process |
| Graph / Memory REST | CLI cobre; explode a superfície |
| YAML na API | G-14: YAML só na borda CLI |
| Rate limit de admissão (`submit_per_min`) | Não está em `04`; concorrência = G-30 |
| Multi-workspace / multi-tenant | G-33: um root por processo |
| OpenAPI gerado / Swagger UI | Doc canônico desta spec basta no v0 |
| Replay SSE / `after=` | SQLite + `GetRun` |

### G-158 — Ordem e critérios

**Status: CONFIRMED**

| Slice | Entrega | Depende de |
|---|---|---|
| **25** | `runtgine serve` + rotas G-155 + auth G-154 | Core API (`11` §13), Intent (`17`) |
| **26** | Webhooks outbound G-156 | Slice 25 |

Não bloqueia nem é bloqueado por slices 21–24 (Intent Surface /
Evolution): superfícies independentes sobre o mesmo Core.

Critérios slice 25:

- `runtgine serve --listen 127.0.0.1:0` sobe; `GET /v0/healthz` → 200.
- Sem token em rota protegida → 401.
- Bind `0.0.0.0` sem token → processo **não** inicia.
- `POST /v0/tasks` com `examples/hello.json` → 202 + `run.succeeded`
  observável em SSE / `GET /v0/runs/{id}`.
- Capability inventada → 400, sem `InsertRun`.
- Preview Intent `git status` → 200, zero Run criado.
- Testes: `httptest` + Core fake/real de workspace temp; sem porta
  efêmera obrigatória em CI além de `httptest`.
- `go test ./...` / `go vet ./...` verdes.

Critérios slice 26:

- Config com um webhook; `run.failed` dispara POST HTTPS (RoundTripper
  injetável nos testes; sem rede real).
- Falha 5xx do destino não muda o estado do Run.
- URL `http://` ou link-local → não registra / não dispara (erro de
  config no boot ou skip + warn).

---

## 4. Relação com outras superfícies

| Superfície | Papel |
|---|---|
| CLI `runtgine run` / `intent` | Processo local; scripts |
| CLI `runtgine serve` | Sobe este Entry Point |
| TUI INTENT (`32`) | Humano no terminal; in-process |
| Board (`12`) | Cards GitHub → Task IR; polling |
| HTTP Player (`28`) | Step `http.get`/`head` **dentro** de um Run |
| HTTP API (`34`) | Fora do Run: admitir/observar Runs via HTTP |
| Wails (`35`) | Fase 3 / slices 27–28; in-process, não cliente desta API no v0 |

Todas convergem para o mesmo protocolo (`11`).

---

## 5. Ordem de implementação

1. **Slice 21** — TUI INTENT (`32`) — independente.
2. **Slices 22–24** — Evolution (`33`) — independente.
3. **Slice 25** — este doc, G-153..G-155, G-157 (serve + REST/SSE).
4. **Slice 26** — G-156 (webhooks outbound).

Código HTTP **não** entra neste PR de spec.

---

## 6. Referências

- Core API: [11](11-protocolo-v0.md) §13
- Error model: [11](11-protocolo-v0.md) §9
- HTTP Player (não confundir): [28](28-http-player-v0.md)
- OpenSpec: [`openspec/changes/archive/2026-08-21-034-http-api/`](../openspec/changes/archive/2026-08-21-034-http-api/)
