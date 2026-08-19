# 28 — HTTP Player v0

Player determinístico para **GET/HEAD HTTPS**: buscar um recurso pela
rede com contrato `http.get` / `http.head`, sem `shell.exec`/`curl`.

Inventário: [10-gaps.md](10-gaps.md) (G-117+; recorte de G-41).
Autoridade de status: [04-decisoes.md](04-decisoes.md).
Não é a API HTTP do Runtgine ([G-45](10-gaps.md) permanece P3).
Não é MCP ([G-44](10-gaps.md)). Não é Project Memory (`16`).

**Status deste doc: CONFIRMED (v0).** G-117..G-122 implementados no
slice 16. POST, `http://` claro, Authorization e download-para-arquivo
permanecem fora.

**Pacote OpenSpec:** arquivado em
[`openspec/changes/archive/2026-08-19-028-http-player/`](../openspec/changes/archive/2026-08-19-028-http-player/).
Deltas mergeados em `openspec/specs/http-player/`. Branch de implementação:
`feat/028-http-player`.

---

## 1. Problema

Tasks que precisam de um JSON/documentação remota hoje passam por
`shell.exec` + `curl`. Isso perde `input_schema`, limites de bytes,
política de URL e telemetria de capability.

O Runtgine já tem Players locais (Shell, Git, FS, Docker). Falta o
primeiro Player de **rede em leitura**. A API *servidor* do runtime
(G-45) é outro produto.

---

## 2. Fronteiras

| É | Não é |
|---|---|
| Player `http` determinístico | Agent / LLM / MCP client |
| Capabilities `http.get` / `http.head` | `http.post` / webhook inbound |
| Cliente `net/http` (Go) | Binário `curl`; Shell |
| HTTPS + TLS verificado | `http://`, `file://`, skip-verify |
| Corpo texto UTF-8 truncável | Gravar arquivo (`fs.write` é outro step) |
| Leitura (sem claim) | Mutação remota; lock de path |

Regras:

1. Validator / Registry continuam soberanos.
2. URL só `https`, sem userinfo (`user:pass@`).
3. Sem `Authorization`, `Cookie` ou `Host` no input (Task IR não é
   cofre). Segredos via env/config ficam para spec futura.
4. Redirects: no máximo 5; destino continua `https`.
5. Destino resolvido não pode ser link-local nem metadata de nuvem
   (`169.254.0.0/16`, `fd00:ec2::/128` / hostname `metadata.google.internal`).
   Loopback e RFC1918 **são permitidos** em HTTPS (dev local).
6. Blast: `http.*` não gera touch nem predicted claim (como `shell.exec`).
7. Policy default: **allow** (leitura). Sem HITL no v0.
8. Pacote Go: `internal/players/httpclient` (não colidir com `net/http`).
   Nome do Player no Manifest: `http`.

---

## 3. Cortes confirmados (G-117+)

### G-117 — Papel e pacote

**Status: CONFIRMED**

- Player name: `http`
- Pacote: `internal/players/httpclient`
- Kind: `deterministic`
- Registro em `api.Open` com os demais Players
- Recorte de G-41: só cliente HTTPS de leitura

### G-118 — Capabilities v0

**Status: CONFIRMED**

| Capability | Entrada | Saída |
|---|---|---|
| `http.get` | `url` (obrigatório), `headers?` (mapa string), `timeout_ms?`, `max_bytes?` | `status`, `url_final`, `headers` (subconjunto), `body`, `bytes`, `truncated`, `binary` |
| `http.head` | `url`, `headers?`, `timeout_ms?` | `status`, `url_final`, `headers`; sem `body` |

Defaults / tetos:

| Campo | Default | Máximo |
|---|---|---|
| `timeout_ms` | 15000 | 60000 |
| `max_bytes` (`get`) | 1 MiB | 4 MiB |

Schemas JSON no Manifest; `additionalProperties: false`.

`headers` de pedido: allowlist **Accept**, `Accept-Language`,
`User-Agent`. Qualquer outra chave (incl. `Authorization`, `Cookie`,
`Host`, `Connection`) → `validation.invalid_input`.

`User-Agent` omitido → `runtgine-http/0.1`.

`body`: UTF-8 válido. Se não for UTF-8: `body=""`, `binary=true`,
`status` ainda é o da resposta. Truncar em `max_bytes` com
`truncated=true`.

`headers` de resposta v0 (chave lowercase): `content-type`,
`content-length`, `etag`, `last-modified`, `cache-control`. Outras
omitidas (não vazar `set-cookie`).

### G-119 — URL e destino

**Status: CONFIRMED**

- Parse estrito: scheme `https` apenas; porta default 443.
- Hostname obrigatório; IP literal é permitido se passar o filtro de
  destino (abaixo).
- Após cada hop de redirect e no Dial: recusar link-local e metadata.
- Falha de URL/destino → `validation.invalid_input` **antes** do GET
  quando for estático (scheme/userinfo/header). Destino resolvido em
  runtime → `runtime.player_error` se o IP cair no filtro.

### G-120 — Cliente / sandbox

**Status: CONFIRMED**

- `http.Client` com TLS verificação ligada (nunca `InsecureSkipVerify`).
- Sem cookie jar persistente; sem proxy config no input (usa env
  `HTTP_PROXY`/`HTTPS_PROXY` do processo se o OS já define — documentar;
  testes unitários injetam `RoundTripper`).
- Sem seguir redirect para scheme ≠ `https`.
- Cancelamento pelo context do Runner.
- Testes da suíte **não** exigem rede: `RoundTripper` fake. Um exemplo
  `examples/http-get.json` pode apontar a um URL estável de docs; CI
  não é obrigada a executá-lo.

### G-121 — Integração Core

**Status: CONFIRMED**

1. `api.Open` registra `httpclient.New()`
2. `ValidateStaticInput` no admission (espelha Git/FS/Docker)
3. Graph: `RefreshFromRegistry` cria nós `http` / `http.get` / `http.head`
4. Blast / Claims: nenhuma linha nova nas tabelas G-95/G-101
5. Exemplo: `examples/http-get.json`
6. Intent heuristics HTTP: nice-to-have; **não** bloqueia o slice

### G-122 — Exclusões v0

**Status: CONFIRMED** (como exclusões)

- `http.post` / `put` / `patch` / `delete`; body de pedido
- Scheme `http://`, `file://`, `data:`
- `Authorization` / `Cookie` / mTLS / client certs
- Skip TLS verify; HTTP/3 obrigatório
- Download direto para path (compor `http.get` + `fs.write`)
- HTML parse / scraping / SSE / WebSocket
- API HTTP do Runtgine (G-45); webhooks inbound
- MCP (G-44); Project Memory (`16`)
- Human Player; HITL neste Player
- TUI dedicada; blast-from-URL
- Proxy no Task IR; redirect ilimitado

---

## 4. Critérios de aceite

1. Manifest registra `http.get` e `http.head`; `http.post` ausente →
   Validator rejeita.
2. `url` `http://example.com` → `validation.invalid_input`.
3. Header `Authorization` no input → `validation.invalid_input`.
4. Fake transport: GET 200 `application/json` UTF-8 devolve `body` e
   `status=200`; corpo > `max_bytes` marca `truncated`.
5. Fake transport com bytes não UTF-8: `binary=true`, `body` vazio.
6. `http.head` não inclui `body` na saída.
7. Destino `169.254.169.254` (metadata) falha sem vazar o corpo.
8. `runtgine run examples/http-get.json` com transport de teste ou skip
   de rede documentado; `go test ./internal/players/httpclient/...`
   verde **offline**.
9. `go test ./...` e `go vet ./...` verdes.
10. OpenSpec `028-http-player` arquivado após o **código** (slice 16).

---

## 5. Ordem do slice de código

Feito no slice 16:

1. Pacote `internal/players/httpclient` + Manifest
2. `ValidateStaticInput` + Dial/redirect policy
3. Registrar no Core; exemplo `examples/http-get.json`
4. Testes com `RoundTripper` fake; README Estágio: Slice 16
5. Arquivar OpenSpec `028` após o código

---

## Checklist de confirmação humana

Marcado em `04-decisoes.md`:

- [x] G-117 Papel (`http` / `httpclient`)
- [x] G-118 GET + HEAD; allowlist de headers
- [x] G-119 HTTPS + filtro metadata/link-local
- [x] G-120 Cliente TLS; testes offline
- [x] G-121 Registry + Graph; sem claim/blast
- [x] G-122 Exclusões (POST, auth, G-45, MCP)
