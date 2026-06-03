# Go Connect — Web Frontends

| App | Path | Port (dev) | Purpose |
|-----|------|------------|---------|
| Public website | `website/` | 3000 | Marketing, discovery, customer auth |
| Admin dashboard | `admin/` | 3001 | Operations, approvals, management |

## Quick start

```bash
# Terminal 1 — API (from repo root)
DATABASE_URL="postgres://app:app@localhost:5433/go_connect?sslmode=disable" make run

# Terminal 2 — Website
cd web/website && cp .env.example .env.local && npm run dev

# Terminal 3 — Admin
cd web/admin && cp .env.example .env.local && npm run dev
```

Set `NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1` in each `.env.local`.
