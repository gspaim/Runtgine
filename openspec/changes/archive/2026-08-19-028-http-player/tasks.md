# Tasks: 028-http-player

Código = slice 16. Implementado e arquivado após merge da spec em `develop`.

## 0. Spec (PR #31)

- [x] 0.1 `docs/28-http-player-v0.md` (G-117..G-122)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / G-41 e espelhos

## 1. Player

- [x] 1.1 Package `internal/players/httpclient` + Manifest `http.get`/`http.head`
- [x] 1.2 `ValidateStaticInput` (https, headers allowlist)
- [x] 1.3 Client + redirect/IP filter; injectable RoundTripper

## 2. Core

- [x] 2.1 Register in `api.Open`; runner static dispatch
- [x] 2.2 `examples/http-get.json`
- [x] 2.3 Blast/Claims tables unchanged

## 3. Tests + closeout

- [x] 3.1 Fake 200 JSON body; truncation; binary
- [x] 3.2 `http://` and Authorization rejected
- [x] 3.3 Metadata IP denied
- [x] 3.4 `go test ./internal/players/httpclient/...` offline
- [x] 3.5 `go test ./...` and `go vet ./...`
- [x] 3.6 README Estágio: Slice 16 Feito
- [x] 3.7 Archive this change into `openspec/specs/http-player/`
