# Go Connect — Backend (Phase 1)

Phase 1 foundation for the service marketplace modular monolith API.

## Stack

- Go
- [chi](https://github.com/go-chi/chi) router
- PostgreSQL (`pgx` driver via `database/sql`)
- [golang-migrate](https://github.com/golang-migrate/migrate) for SQL migrations
- Docker Compose for local PostgreSQL

## Project layout

```text
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── app/              # HTTP server and routes
│   ├── config/           # Environment configuration
│   ├── platform/         # Database, logger
│   └── shared/           # Response helpers, errors, middleware
├── migrations/           # Sequential SQL migrations
├── docs/                 # Backend documentation
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Prerequisites

- Go 1.26+
- Docker and Docker Compose
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation)

## Quick start

1. Copy environment variables:

   ```bash
   cp .env.example .env
   export $(grep -v '^#' .env | xargs)
   ```

2. Start PostgreSQL:

   ```bash
   docker compose up -d
   ```

3. Run migrations:

   ```bash
   make migrate-up
   ```

4. Start the API:

   ```bash
   make run
   ```

5. Health check:

   ```bash
   curl -s http://localhost:8080/api/v1/health | jq
   ```

## Makefile commands

| Command | Description |
|---------|-------------|
| `make run` | Start the API server |
| `make test` | Run all Go tests |
| `make migrate-up` | Apply migrations |
| `make migrate-down` | Roll back one migration |

## API

### `GET /api/v1/health`

Success (`200`):

```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "ok",
    "database": "up"
  }
}
```

Failure when database is unreachable (`503`):

```json
{
  "success": false,
  "message": "Service is unhealthy",
  "error": "HEALTH_CHECK_FAILED"
}
```

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | yes | — | PostgreSQL connection string |
| `HTTP_PORT` | no | `8080` | HTTP listen port |
| `APP_ENV` | no | `development` | Environment name |
| `LOG_LEVEL` | no | `info` | `debug`, `info`, `warn`, `error` |
| `DB_MAX_OPEN_CONNS` | no | `10` | Max open DB connections |
| `DB_MAX_IDLE_CONNS` | no | `5` | Max idle DB connections |
| `DB_CONN_MAX_LIFETIME_SEC` | no | `300` | Connection max lifetime |

## Phase 2 (planned)

- Auth module (JWT access/refresh)
- Users and roles
- sqlc query generation wired to migrations
- Request validation helpers
- Graceful config for production secrets
