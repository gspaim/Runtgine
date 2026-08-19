# Tasks: 028-http-player

Código = slice 16. **Não implementar neste PR de spec.** Marcar só
depois de G-117..G-122 CONFIRMED em `04` (este change) e merge em
`develop`.

## 0. Spec (este PR)

- [x] 0.1 `docs/28-http-player-v0.md` (G-117..G-122)
- [x] 0.2 OpenSpec proposal / design / deltas
- [x] 0.3 Promover `04` / `10` / G-41 e espelhos

## 1. Player

- [ ] 1.1 Package `internal/players/httpclient` + Manifest `http.get`/`http.head`
- [ ] 1.2 `ValidateStaticInput` (https, headers allowlist)
- [ ] 1.3 Client + redirect/IP filter; injectable RoundTripper

## 2. Core

- [ ] 2.1 Register in `api.Open`; runner static dispatch
- [ ] 2.2 `examples/http-get.json`
- [ ] 2.3 Blast/Claims tables unchanged

## 3. Tests + closeout

- [ ] 3.1 Fake 200 JSON body; truncation; binary
- [ ] 3.2 `http://` and Authorization rejected
- [ ] 3.3 Metadata IP denied
- [ ] 3.4 `go test ./internal/players/httpclient/...` offline
- [ ] 3.5 `go test ./...` and `go vet ./...`
- [ ] 3.6 README Estágio: Slice 16 Feito
- [ ] 3.7 Archive this change into `openspec/specs/http-player/`
