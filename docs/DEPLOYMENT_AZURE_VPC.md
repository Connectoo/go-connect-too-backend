# Deploy on Azure VNet (frontends public, API + Postgres private)

One **Linux VM** in a **single VNet** runs the full stack:

| Component | Internet | Notes |
|-----------|----------|-------|
| Website, admin, customer, employee | **Public** (Nginx → nip.io) | Share with clients |
| Go API | **Private** (`127.0.0.1:8080`) | Proxied as `/api/` on each frontend host |
| PostgreSQL, MinIO | **Private** (localhost only) | NSG blocks 5432, 9000 |

Browsers call `https://admin.<domain>/api/v1/...` — Nginx forwards to the private API. No public `api.*` subdomain.

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
| Website | `http://www.<nip.io>` |
| Admin | `http://admin.<nip.io>/login` |
| Customer | `http://app.<nip.io>/login` |
| Employee | `http://provider.<nip.io>/login` |

**API** (via frontend proxy, not a separate public host):

```bash
curl http://www.<nip.io>/api/v1/health
```

**Private** (only on VM):

```bash
docker ps
pm2 list
curl http://127.0.0.1:8080/api/v1/health   # direct — not reachable from internet
```

Demo logins: `admin@yopmail.com` / `alice@yopmail.com` / `karim@yopmail.com` — password **`Demo123!`**

---

## 5. HTTPS with Let's Encrypt

**Requires a real domain** — Let's Encrypt does **not** work with nip.io.

### DNS (before certbot)

Point A records to your VM **public IP**:

| Host | Example |
|------|---------|
| `www` | `www.goconnect.in` |
| `admin` | `admin.goconnect.in` |
| `app` | `app.goconnect.in` |
| `provider` | `provider.goconnect.in` |

Wait until all resolve: `dig +short www.goconnect.in`

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
curl https://www.goconnect.in/api/v1/health
```

Certs auto-renew: `systemctl status certbot.timer`

**Do not** re-run `setup-vm-vpc.sh` after HTTPS — it overwrites Nginx SSL config. Use `deploy-vpc.sh` for frontend-only redeploys.

---

## 6. Redeploy

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
│  Nginx (public)                                        │
│    www.*   ──► Next.js :3000                           │
│    admin.* ──► Next.js :3001                           │
│    app.*   ──► Next.js :3002                           │
│    provider.* ──► Next.js :3003                        │
│    */api/* ──► Go API :8080 (127.0.0.1, private)       │
│    */files/* ──► MinIO :9000 (127.0.0.1, private)      │
│                                                        │
│  Docker (localhost only): Postgres :5432               │
└────────────────────────────────────────────────────────┘
```

---

## Files

| Path | Purpose |
|------|---------|
| `deploy/azure/setup-vm-vpc.sh` | Bootstrap VM |
| `deploy/azure/deploy-vpc.sh` | Rebuild Next.js (same-origin API URLs) |
| `deploy/azure/nginx/go-connect-vpc.conf` | Nginx — frontends + `/api/` proxy |
| `deploy/azure/env.vpc.example` | Env template |
| `docker-compose.azure.yml` | Postgres + MinIO + API (localhost ports) |
