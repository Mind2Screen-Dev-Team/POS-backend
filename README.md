# POS Backend

Backend API untuk POS Penglaris — Go + PostgreSQL.

## Tech Stack

- **Go 1.26**
- **PostgreSQL** (driver: pgx/v5)
- **net/http** (stdlib HTTP server)

## Quick Start

```bash
# 1. Copy env file
cp .env.example .env

# 2. Edit .env with your database credentials

# 3. Build & run
make build
make run
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_HOST` | `0.0.0.0` | Server listen host |
| `APP_PORT` | `8080` | Server listen port |
| `APP_ENV` | `development` | Environment (development/staging/production) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `pos_penglaris` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode (disable/require/verify-full) |
| `DB_MAX_CONNECTIONS` | `25` | Max connection pool size |

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check (includes DB ping) |

## Makefile Targets

```bash
make build   # Compile binary to bin/server
make run     # Build & run
make test    # Run tests with race detector
make vet     # Run go vet
make fmt     # Check formatting
make clean   # Remove bin/
make lint    # fmt + vet + test (pre-commit gate)
```

## Project Structure

```
.
├── cmd/server/          # Application entrypoint
│   └── main.go
├── internal/
│   ├── config/          # Environment-driven configuration
│   │   └── config.go
│   ├── handler/         # HTTP handlers & router
│   │   └── router.go
│   └── repository/      # Database connection & queries
│       └── db.go
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## License

Proprietary — Mind2Screen Dev Team
