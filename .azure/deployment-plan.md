# Azure deployment plan — Go Connect (testing)

## Goal

Host the **full stack** on one Azure Linux VM for a short client demo:

- Go API (Docker)
- PostgreSQL + MinIO (Docker)
- 4 Next.js apps (PM2)
- Nginx reverse proxy + optional Let's Encrypt SSL

## Architecture

### Demo (single VM)

```text
Internet → Azure VM (Central India, Standard_B2s)
  ├── Nginx :80/:443
  ├── PM2: website :3000, admin :3001, customer :3002, employee :3003
  └── Docker: api :8080, postgres, minio
```

### VPC (single VM — frontends public, API + DB private)

```text
Internet → Nginx :80/:443 → www/admin/app/provider.*
  */api/* → Go API :8080 (127.0.0.1, not public)
Same VM: Postgres + MinIO on localhost only
```

See `docs/DEPLOYMENT_AZURE_VPC.md`.

## Azure resources

| Resource | SKU | Region |
|----------|-----|--------|
| Linux VM | **Standard_B2s** (2 vCPU, 4 GB) | **Central India** |
| OS disk | 64 GB Premium SSD | — |
| Public IP | Static | — |
| NSG | SSH 22, HTTP 80, HTTPS 443 | — |

Optional: custom domain A records → VM public IP.

## Repo artifacts

| Path | Purpose |
|------|---------|
| `Dockerfile` | Go API container image |
| `docker-compose.azure.yml` | Postgres, MinIO, API |
| `deploy/azure/env.example` | Production env template |
| `deploy/azure/setup-vm.sh` | First-time VM bootstrap |
| `deploy/azure/deploy.sh` | Build frontends + reload services |
| `deploy/azure/nginx/go-connect.conf` | Nginx site config |
| `docs/DEPLOYMENT_AZURE.md` | Step-by-step operator guide |

## Environment

- `APP_ENV=production`
- `DATABASE_URL` → Docker postgres on VM
- `CORS_ALLOWED_ORIGINS` → four frontend HTTPS origins
- `NEXT_PUBLIC_API_URL` → `https://api.<domain>/api/v1`
- JWT secrets → openssl-generated

## Out of scope (testing)

- Azure Database for PostgreSQL (using Docker Postgres on VM)
- Azure Blob (using MinIO on VM)
- WebSocket scaling / Container Apps
- CI/CD pipeline

## Validation

1. `curl https://api.<domain>/api/v1/health`
2. Login on admin / customer / employee portals
3. Public website loads categories from API
