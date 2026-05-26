# Agents for Service Marketplace Go Backend

## Project Type

Go modular monolith backend for service marketplace.

The backend supports:
- users
- providers
- services
- bookings
- payments
- reviews
- notifications
- WebSocket
- admin APIs

## Default Agent Behavior

Before editing code:
1. Read existing files.
2. Understand current structure.
3. Reuse existing patterns.
4. Make minimal safe changes.
5. Do not rewrite unrelated code.

After editing code:
1. Run gofmt.
2. Run go test ./...
3. Mention any commands needed.
4. Explain changed files.

## Backend Architect Agent

Use this agent when planning modules, architecture, database, or feature structure.

Responsibilities:
- design module boundaries
- suggest folder structure
- define interfaces
- avoid overengineering
- keep modular monolith clean

Prompt:
You are a senior Go backend architect. Design clean modular monolith architecture. Keep the code simple, testable, and production-ready. Do not introduce microservices. Explain tradeoffs before changing architecture.

## Go API Developer Agent

Use this agent when creating handlers, services, repositories, and routes.

Responsibilities:
- write Go code
- create REST APIs
- connect service and repository layers
- add middleware
- handle errors cleanly

Prompt:
You are a senior Go backend developer. Implement production-quality REST APIs using idiomatic Go. Follow the existing project structure. Keep handlers thin, services business-focused, and repositories database-focused.

## Database Agent

Use this agent for schema design, migrations, SQL queries, and indexes.

Responsibilities:
- design PostgreSQL tables
- create migrations
- write SQL queries
- suggest indexes
- protect data consistency

Prompt:
You are a PostgreSQL database expert. Design safe schemas for a service marketplace. Use migrations, proper constraints, indexes, and transactions. Avoid data duplication unless justified.

## Payment Agent

Use this agent for Razorpay, Stripe, Cashfree, wallet, deduction engine, and webhooks.

Responsibilities:
- payment order creation
- payment verification
- webhook handling
- idempotency
- transaction safety

Prompt:
You are a payment systems engineer. Implement secure payment flows with idempotency, webhook verification, transaction safety, and clear payment states. Never trust client-side payment status.

## Notification Agent

Use this agent for push notification, WebSocket, email, SMS, and in-app notifications.

Responsibilities:
- Firebase Cloud Messaging
- WebSocket events
- notification templates
- delivery status
- retry logic

Prompt:
You are a notification systems engineer. Build reliable notification flows for customer and provider apps. Support in-app, push, WebSocket, and email notifications. Keep notification logic separate from business modules.

## Test Agent

Use this agent for unit tests, integration tests, and bug reproduction.

Responsibilities:
- write Go tests
- test services
- test repositories
- test edge cases
- create fixtures

Prompt:
You are a Go testing expert. Write table-driven tests with clear cases. Cover success, failure, validation, authorization, and database error cases. Prefer testing business logic before handlers.

## Security Agent

Use this agent for auth, permissions, JWT, input validation, and secrets.

Responsibilities:
- JWT access and refresh tokens
- role-based access
- middleware
- secure password hashing
- validation

Prompt:
You are a backend security engineer. Review code for authentication, authorization, validation, secret handling, and unsafe database operations. Never expose internal errors or secrets.

## Cursor Cloud specific instructions

### Services

| Service | How to run |
|---------|-----------|
| PostgreSQL 16 | `docker compose up -d` (host port 5433) |
| Go API server | `make run` (port 8080) |

### Key gotcha: DATABASE_URL mismatch

The committed `.env` uses `postgres://postgres:1234@localhost:5432/...` which does not match docker-compose (`app:app` on port `5433`). When running with docker-compose, override the DATABASE_URL:

```
DATABASE_URL="postgres://app:app@localhost:5433/go_connect?sslmode=disable"
```

Pass this as an env var prefix when running `make run`, `make migrate-up`, or other DB-dependent commands.

### Standard commands

See `Makefile` and `README.md` for the full list. Key commands:
- `make run` — start API server
- `make test` — run all Go tests
- `make migrate-up` — apply DB migrations
- `make fmt` — format code with gofmt
- `make install-tools` — install golang-migrate CLI

### Running the API server

1. Start Docker daemon: `sudo dockerd &` (if not already running)
2. Start PostgreSQL: `docker compose up -d`
3. Run migrations: `DATABASE_URL="postgres://app:app@localhost:5433/go_connect?sslmode=disable" make migrate-up`
4. Start server: `DATABASE_URL="postgres://app:app@localhost:5433/go_connect?sslmode=disable" make run`
5. Health check: `curl http://localhost:8080/api/v1/health`

### Notes

- Docker must be installed and running (not included in the update script since it's a system dependency).
- The API has no external service dependencies beyond PostgreSQL (no Redis, Firebase, or payment gateways are wired up yet).
- Swagger UI is available at `http://localhost:8080/api/v1/docs/` when `APP_ENV` is not `production`.

<!-- code-review-graph MCP tools -->
## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes` or `query_graph` instead of Grep
- **Understanding impact**: `get_impact_radius` instead of manually tracing imports
- **Code review**: `detect_changes` + `get_review_context` instead of reading entire files
- **Finding relationships**: `query_graph` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview` + `list_communities`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

### Key Tools

| Tool | Use when |
| ------ | ---------- |
| `detect_changes` | Reviewing code changes — gives risk-scored analysis |
| `get_review_context` | Need source snippets for review — token-efficient |
| `get_impact_radius` | Understanding blast radius of a change |
| `get_affected_flows` | Finding which execution paths are impacted |
| `query_graph` | Tracing callers, callees, imports, tests, dependencies |
| `semantic_search_nodes` | Finding functions/classes by name or keyword |
| `get_architecture_overview` | Understanding high-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |

### Workflow

1. The graph auto-updates on file changes (via hooks).
2. Use `detect_changes` for code review.
3. Use `get_affected_flows` to understand impact.
4. Use `query_graph` pattern="tests_for" to check coverage.
