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

3. Run migrations (required before auth endpoints work):

   ```bash
   make migrate-up
   ```

   Confirm tables exist: `users`, `refresh_tokens`. If `migrate-up` fails with an empty database URL, ensure `.env` exists and `DATABASE_URL` is set (the Makefile loads `.env` automatically).

4. Start the API:

   ```bash
   make run
   ```

5. Health check:

   ```bash
   curl -s http://localhost:8080/api/v1/health | jq
   ```

6. API docs (Swagger UI, disabled when `APP_ENV=production`):

   Open in a browser: [http://localhost:8080/api/v1/docs/](http://localhost:8080/api/v1/docs/)

   Raw OpenAPI spec: [http://localhost:8080/api/v1/docs/openapi.yaml](http://localhost:8080/api/v1/docs/openapi.yaml)

## Makefile commands

| Command | Description |
|---------|-------------|
| `make seed` | Populate demo data + write `docs/SEED_DATA.xlsx` |
| `make dev-up` | Start Docker, migrations, API, and all 4 web apps |
| `make dev-down` | Stop API and all web dev servers |
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

## API documentation

When `APP_ENV` is not `production`, the server serves:

| URL | Description |
|-----|-------------|
| `/api/v1/docs/` | Swagger UI (try endpoints in the browser) |
| `/api/v1/docs/openapi.yaml` | OpenAPI 3.0 spec |

The spec source lives at `internal/app/spec/openapi.yaml`. Update it when you add endpoints.

For `/auth/me` in Swagger UI: run **Login** or **Register**, copy `data.tokens.access_token`, click **Authorize**, paste `Bearer <token>` (or just the token if the UI adds the prefix).

## Phase 2

- Auth module (JWT access/refresh) — implemented
- Users and roles — implemented

## Later phases

- sqlc query generation wired to migrations
- Request validation helpers
- Graceful config for production secrets
