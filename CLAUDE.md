# CLAUDE.md — POS Backend

Backend API untuk POS Penglaris. Go 1.26 + PostgreSQL (pgx/v5), net/http stdlib.

## Struktur

```
cmd/server/main.go          # entrypoint HTTP server
internal/config/            # config via env (APP_HOST, APP_PORT, APP_ENV, DB_*)
internal/handler/           # router + GET /health (cek koneksi DB)
internal/repository/        # connection pool PostgreSQL
```

## Perintah

```sh
make build        # kompilasi binary
make run          # jalankan server
make test         # go test
make vet          # go vet
```

Verifikasi manual: `go build ./...`, `go vet ./...`, `go test -race ./...`.

## Env & konfigurasi

- `APP_HOST` (default 0.0.0.0), `APP_PORT` (8080), `APP_ENV` (development).
- Koneksi DB via env `DB_*` — contoh di `.env.example`. Salin ke `.env` untuk run lokal, jangan commit `.env`.

## Konvensi branch

- Mulai dari `develop`, target merge `develop`. `main` = merge manusia saja. `git pull` dulu sebelum mulai kerja.
