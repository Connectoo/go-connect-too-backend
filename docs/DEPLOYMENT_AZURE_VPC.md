# Deploy on Azure VNet (frontends public, API + Postgres private)

One **Linux VM** in a **single VNet** runs the full stack:

| Component | Public URL | Backend |
|-----------|------------|---------|
| Website | `https://www.<domain>` | Next.js `:3000` |
| Admin | `https://admin.<domain>` | Next.js `:3001` |
| Customer | `https://app.<domain>` | Next.js `:3002` |
| Employee | `https://provider.<domain>` | Next.js `:3003` |
| **API** | `https://api.<domain>/api/v1` | Go API `:8080` |
| Files | `https://www.<domain>/files/` or `api.<domain>/files/` | MinIO `:9000` |
| **PostgreSQL** | **Never public** | SSH tunnel only (see below) |

Frontends also proxy `/api/` on their own host. Mobile apps and Swagger can use `https://api.<domain>` directly.

For everything public including `api.*`, see [DEPLOYMENT_AZURE.md](./DEPLOYMENT_AZURE.md).

### Test locally first (Mac/Linux)

Mirrors the Azure VPC layout on your machine:

```bash
brew install nginx   # macOS (once)
cp .env.example .env
docker compose up -d
make migrate-up
make seed            # optional demo data
make dev-vpc
```

Then open:

| App | URL |
|-----|-----|
| Website | http://www.localhost:8880 |
| Admin | http://admin.localhost:8880/login |
| API (via proxy) | http://www.localhost:8880/api/v1/health |

Stop: `make dev-vpc-down`

Normal dev (API on `:8080` directly, no Nginx): `make dev-up`

---

## 1. Create the VNet and VM

### Virtual network

1. **Create a resource** → **Virtual network**
   - Name: `go-connect-vnet`
   - Region: **Central India**
   - IPv4: `10.0.0.0/16`
   - Subnet: `snet-default` → `10.0.1.0/24`

### Network Security Group

Allow **only**:

| Rule | Port | Source |
|------|------|--------|
| SSH | 22 | Your IP |
| HTTP | 80 | Internet |
| HTTPS | 443 | Internet |

Do **not** open 5432, 8080, 9000, or 3000–3003.

### Virtual machine

| Field | Value |
|-------|-------|
| VNet / subnet | `go-connect-vnet` / `snet-default` |
| Public IP | **Static** |
| Size | **Standard_B2s** |
| Image | **Ubuntu Server 24.04 LTS** |
| NSG | as above |
| Auth | SSH public key |

```bash
chmod 400 ~/Downloads/go-connect.pem
ssh -i ~/Downloads/go-connect.pem azureuser@<PUBLIC_IP>
```

---

## 2. Clone and configure

```bash
sudo -i
apt-get update && apt-get install -y git
git clone https://github.com/YOUR_USER/go-connect-too-backend.git /opt/go-connect
cd /opt/go-connect

cp deploy/azure/env.vpc.example .env
nano .env
```

**Edit `.env`:**

1. `POSTGRES_PASSWORD` — strong password (match in `DATABASE_URL`)
2. `MINIO_ROOT_PASSWORD` / `S3_SECRET_KEY`
3. `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` — `openssl rand -hex 32` twice
4. Leave `DOMAIN` empty for nip.io

---

## 3. One-command setup

```bash
chmod +x deploy/azure/setup-vm-vpc.sh deploy/azure/deploy-vpc.sh
bash deploy/azure/setup-vm-vpc.sh
```

---

## 4. Verify

**Public frontends** (share with client):

| App | URL |
|-----|-----|
| Website | `https://www.connectoo.online` |
| Admin | `https://admin.connectoo.online/login` |
| Customer | `https://app.connectoo.online/login` |
| Employee | `https://provider.connectoo.online/login` |
| API | `https://api.connectoo.online/api/v1/health` |
| Swagger | `https://api.connectoo.online/api/v1/docs/` |

```bash
curl https://api.connectoo.online/api/v1/health
```

**Private** (only on VM):

```bash
docker ps
pm2 list
curl http://127.0.0.1:8080/api/v1/health   # direct — not reachable from internet
```

Demo logins: `admin@yopmail.com` / `alice@yopmail.com` / `karim@yopmail.com` — password **`Demo123!`**

**API docs (Swagger):** disabled by default in production. For the demo VM, set `ENABLE_API_DOCS=true` in `.env`, then `docker compose -f docker-compose.azure.yml up -d --build`. Open `https://www.connectoo.online/api/v1/docs/` (or your domain). Turn off after testing.

---

## 5. HTTPS with Let's Encrypt

**Requires a real domain** — Let's Encrypt does **not** work with nip.io.

### DNS (before certbot)

Point A records to your VM **public IP**:

| Host | Example |
|------|---------|
| `www` | `www.connectoo.online` |
| `admin` | `admin.connectoo.online` |
| `app` | `app.connectoo.online` |
| `provider` | `provider.connectoo.online` |
| `api` | `api.connectoo.online` |
| `minio` | `minio.connectoo.online` (optional — MinIO console UI) |

Wait until all resolve: `dig +short www.connectoo.online`

### Enable HTTPS (one command)

On the VM, after HTTP setup works:

```bash
cd /opt/go-connect
nano .env
```

Set:

```env
DOMAIN=goconnect.in
LETSENCRYPT_EMAIL=you@yourdomain.com
```

Then:

```bash
chmod +x deploy/azure/enable-https-vpc.sh
sudo bash deploy/azure/enable-https-vpc.sh
```

This runs certbot, enables HTTP→HTTPS redirect, rebuilds frontends with `https://` URLs, and restarts the API.

### Verify

```bash
curl https://api.connectoo.online/api/v1/health
```

Certs auto-renew: `systemctl status certbot.timer`

**Do not** re-run `setup-vm-vpc.sh` after HTTPS — it overwrites Nginx SSL config. Use `deploy-vpc.sh` for frontend-only redeploys.

---

## 5b. CORS (browser API calls)

VPC frontends should call the API **same-origin** via Nginx (`https://www.<domain>/api/v1`, etc.). `deploy-vpc.sh` sets `NEXT_PUBLIC_API_URL` per app — **no CORS needed** for normal web UI.

Direct `https://api.<domain>` is for mobile apps, Swagger, and curl. If a browser page on another subdomain calls `api.*` directly, set `CORS_ALLOWED_ORIGINS` in `.env` (HTTPS URLs, comma-separated):

```env
CORS_ALLOWED_ORIGINS=https://www.connectoo.online,https://admin.connectoo.online,https://app.connectoo.online,https://provider.connectoo.online,https://api.connectoo.online
```

After any change:

```bash
cd /opt/go-connect
set -a && source .env && source deploy/azure/resolve-domain.sh && resolve_domain_config && set +a
bash deploy/azure/deploy-vpc.sh
docker compose -f docker-compose.azure.yml up -d --build
```

Verify from Mac:

```bash
curl -s -H "Origin: https://www.connectoo.online" -I \
  https://api.connectoo.online/api/v1/public/categories | grep -i access-control
```

Should show `Access-Control-Allow-Origin: https://www.connectoo.online`.

---

## 5c. MinIO Console UI (optional)

Browse uploaded files in a web UI at `https://minio.<domain>` (Nginx + basic auth). **Do not** open ports 9000/9001 in Azure NSG.

### DNS

| Host | Example |
|------|---------|
| `minio` | `minio.connectoo.online` |

### Enable (after HTTPS)

```bash
cd /opt/go-connect
git pull
chmod +x deploy/azure/enable-minio-console-vpc.sh

# Set basic-auth password in .env before running:
# MINIO_CONSOLE_AUTH_USER=admin
# MINIO_CONSOLE_AUTH_PASSWORD=your-password

sudo bash deploy/azure/enable-minio-console-vpc.sh
```

To change the Nginx basic-auth password later:

```bash
nano .env   # update MINIO_CONSOLE_AUTH_PASSWORD
sudo bash deploy/azure/update-minio-console-auth.sh
```

Open **https://minio.connectoo.online** — two logins:

1. **Nginx basic auth** — `MINIO_CONSOLE_AUTH_USER` / `MINIO_CONSOLE_AUTH_PASSWORD` (saved in `.env`)
2. **MinIO** — `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` from `.env`

Bucket: `go-connect-uploads`

---

## 6. Access PostgreSQL (SSH tunnel — never public)

**Do not** open port 5432 in Azure NSG. Use an SSH tunnel from your laptop:

```bash
# On your Mac (new terminal, keep it open)
chmod +x deploy/azure/postgres-tunnel.sh
VM_HOST=4.240.109.134 SSH_KEY=~/Downloads/go-connect.pem \
  bash deploy/azure/postgres-tunnel.sh
```

Then connect with TablePlus, DBeaver, or `psql`:

| Field | Value |
|-------|-------|
| Host | `127.0.0.1` |
| Port | `15432` |
| User | `app` |
| Password | from VM `.env` `POSTGRES_PASSWORD` |
| Database | `go_connect` |

```bash
psql "postgres://app:YOUR_PASSWORD@127.0.0.1:15432/go_connect?sslmode=disable"
```

**On the VM directly** (SSH session):

```bash
docker exec -it go-connect-postgres psql -U app -d go_connect
```

---

## 7. Redeploy

```bash
cd /opt/go-connect
git pull
docker compose -f docker-compose.azure.yml up -d --build
bash deploy/azure/deploy-vpc.sh
```

---

## Architecture

```text
Internet
    │  :80/:443
    ▼
┌────────────────────────────────────────────────────────┐
│  VNet 10.0.0.0/16 — single VM                          │
│                                                        │
│  Nginx (public :443)                                   │
│    www.* / admin.* / app.* / provider.* ──► Next.js    │
│    api.*  ──► Go API :8080                             │
│    */api/* and api.* ──► Go API :8080 (127.0.0.1)       │
│    */files/* ──► MinIO :9000                           │
│    minio.* ──► MinIO Console :9001 (basic auth)        │
│                                                        │
│  Postgres :5432 — SSH tunnel only, never internet      │
└────────────────────────────────────────────────────────┘
```

---

## Files

| Path | Purpose |
|------|---------|
| `deploy/azure/setup-vm-vpc.sh` | Bootstrap VM |
| `deploy/azure/deploy-vpc.sh` | Rebuild Next.js (same-origin API URLs) |
| `deploy/azure/nginx/go-connect-vpc.conf` | Nginx — frontends + `/api/` proxy |
| `deploy/azure/enable-https-vpc.sh` | Let's Encrypt HTTPS |
| `deploy/azure/enable-minio-console-vpc.sh` | MinIO web UI at minio.* |
| `deploy/azure/update-minio-console-auth.sh` | Change MinIO console Nginx password |
| `deploy/azure/nginx/minio-console-vpc.conf` | Nginx — MinIO console proxy |
| `deploy/azure/env.vpc.example` | Env template |
| `docker-compose.azure.yml` | Postgres + MinIO + API (localhost ports) |
