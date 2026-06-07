# Go Connect — Web Frontends

| App | Path | Port (dev) | Purpose |
|-----|------|------------|---------|
| Public website | `website/` | 3000 | Marketing, discovery (links to customer app) |
| Admin dashboard | `admin/` | 3001 | Operations, approvals, management |
| Customer portal | `customer/` | 3002 | Bookings, profile, chat (planned) |
| Employee portal | `employee/` | 3003 | Services, schedule, KYC (planned) |

Screen/API inventory: [`docs/UI_INVENTORY.md`](../docs/UI_INVENTORY.md)

## Cursor MCP (UI)

The repo configures **shadcn MCP** in `.cursor/mcp.json` for installing UI components via the agent.

1. Open **Cursor → Settings → MCP**
2. Enable the `shadcn` server (green status)
3. Restart Cursor if tools do not appear

Use with prompts like: "Add shadcn table and dialog to web/customer for the bookings list."

## Quick start

**One command** (from repo root — starts Docker, migrations, API, and all 4 apps):

```bash
make dev-up
```

Stop API + all UIs: `make dev-down` (Docker keeps running; use `docker compose down` to stop Postgres/MinIO).

**Manual start** (separate terminals):

```bash
# Terminal 1 — API (from repo root)
DATABASE_URL="postgres://app:app@localhost:5433/go_connect?sslmode=disable" make run

# Terminal 2 — Website (port 3000)
cd web/website && npm install && npm run dev

# Terminal 3 — Admin (port 3001)
cd web/admin && npm install && npm run dev

# Terminal 4 — Customer portal (port 3002)
cd web/customer && npm install && npm run dev

# Terminal 5 — Employee/provider portal (port 3003)
cd web/employee && npm install && npm run dev
```

Apps default to `NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1` if unset.

## Demo seed data

Populate the database with customers, employees, categories, services, bookings, reviews, and more:

```bash
make seed
```

This writes **`docs/SEED_DATA.xlsx`** — a multi-sheet reference with every login, entity ID, and seeded record. All demo accounts use password **`Demo123!`** and emails ending in `@demo.go-connect.local`.

| Role | Example login | Portal |
|------|---------------|--------|
| Admin | `admin@demo.go-connect.local` | http://localhost:3001/login |
| Customer | `alice@demo.go-connect.local` | http://localhost:3002/login |
| Employee | `karim@demo.go-connect.local` | http://localhost:3003/login |

Re-running `make seed` clears previous `@demo.go-connect.local` data and re-inserts fresh records.

## Auth per app

Each app stores its JWT in its own cookie + localStorage key — tokens are never shared across apps.

| App | Port | Login endpoint | Required role | Cookie |
|-----|------|----------------|---------------|--------|
| Website | 3000 | — (public) | — | — |
| Admin | 3001 | `POST /auth/login/admin` | `admin` | `admin_access_token` |
| Customer | 3002 | `POST /auth/login/customer` | `customer` | `customer_access_token` |
| Employee | 3003 | `POST /auth/login/employee` | `employee` | `employee_access_token` |

The customer and employee portals are scaffolded from `web/admin` and share the same conventions (`lib/api-client.ts`, `lib/auth.ts`, `middleware.ts`, `services/*`, `hooks/use-*`, `components/ui/*`, `components/shared/*`). See [`docs/BUILD_PLAN.md`](../docs/BUILD_PLAN.md) for the screen build plan.
