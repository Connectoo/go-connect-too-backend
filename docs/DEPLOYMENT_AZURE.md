# Deploy full stack on Azure (testing)

> **VPC / private API + Postgres?** Use [DEPLOYMENT_AZURE_VPC.md](./DEPLOYMENT_AZURE_VPC.md) — single VM in one VNet (frontends public, API + DB private behind Nginx `/api/` proxy).

One **Azure Linux VM** in **Central India** runs everything:

| Component | How |
|-----------|-----|
| Go API | Docker (`docker-compose.azure.yml`) |
| PostgreSQL | Docker |
| MinIO | Docker |
| Website, admin, customer, employee | PM2 (`npm start`) |
| HTTPS | Nginx + Certbot |

**VM size:** `Standard_B2s` (2 vCPU, 4 GB RAM) — enough for a 1-month client demo.

**No domain?** Setup auto-uses **nip.io** from your VM public IP (e.g. `http://admin.20-12-34-56.nip.io`). No DNS purchase required.

---

## 1. Create the Azure VM

1. [portal.azure.com](https://portal.azure.com) → **Create a resource** → **Virtual machine**
2. Settings:

| Field | Value |
|-------|-------|
| Region | **Central India** |
| Image | **Ubuntu Server 24.04 LTS** |
| Size | **Standard_B2s** |
| Authentication | SSH public key |
| Public inbound ports | **22, 80, 443** |

3. Create. Note the **public IP**.

---

## 2. No domain (default for testing)

Leave `DOMAIN=` empty in `.env`. The setup script detects your VM public IP and uses:

| App | Example URL (IP = 20.12.34.56) |
|-----|--------------------------------|
| Website | `http://www.20-12-34-56.nip.io` |
| Admin | `http://admin.20-12-34-56.nip.io` |
| Customer | `http://app.20-12-34-56.nip.io` |
| Employee | `http://provider.20-12-34-56.nip.io` |
| API | `http://api.20-12-34-56.nip.io/api/v1/health` |

nip.io resolves these hostnames to your VM IP automatically. Share the links with your client.

### Optional: custom domain later

Add A records for `api`, `www`, `admin`, `app`, `provider` → VM IP, set `DOMAIN=goconnect.in` in `.env`, re-run `bash deploy/azure/deploy.sh`, then certbot.

---

## 3. Clone and configure

SSH into the VM:

```bash
ssh azureuser@<VM_PUBLIC_IP>
sudo -i
```

```bash
apt-get update && apt-get install -y git
git clone https://github.com/YOUR_USER/go-connect-too-backend.git /opt/go-connect
cd /opt/go-connect

cp deploy/azure/env.example .env
nano .env
```

**Edit `.env` (only these — leave `DOMAIN` empty if you have no domain):**

1. `POSTGRES_PASSWORD` — strong password (same value in `DATABASE_URL`)
2. `MINIO_ROOT_PASSWORD` and `S3_SECRET_KEY` — same strong password
3. `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` — `openssl rand -hex 32` twice

`DOMAIN`, `API_URL`, and `CORS_ALLOWED_ORIGINS` are filled automatically when `DOMAIN` is empty.

---

## 4. One-command setup

```bash
chmod +x deploy/azure/setup-vm.sh deploy/azure/deploy.sh
bash deploy/azure/setup-vm.sh
```

This installs Docker, Node 20, Go, Nginx, PM2, starts the API stack, runs migrations + seed, builds all four Next.js apps, and configures Nginx.

---

## 5. HTTPS (optional)

Only needed if you want `https://` instead of `http://`. Works with nip.io or a custom domain:

```bash
certbot --nginx \
  -d api.YOUR_NIP_OR_DOMAIN \
  -d www.YOUR_NIP_OR_DOMAIN \
  -d admin.YOUR_NIP_OR_DOMAIN \
  -d app.YOUR_NIP_OR_DOMAIN \
  -d provider.YOUR_NIP_OR_DOMAIN
```

Then set `USE_HTTPS=true` in `.env` and run `bash deploy/azure/deploy.sh`.

---

## 6. Verify

```bash
curl http://api.20-12-34-56.nip.io/api/v1/health   # use your nip.io host
pm2 list
docker ps
```

Setup prints the exact URLs when it finishes. Demo logins: `admin@yopmail.com` / `alice@yopmail.com` / `karim@yopmail.com` — password **`Demo123!`**

---

## 7. Redeploy after code changes

```bash
cd /opt/go-connect
git pull
docker compose -f docker-compose.azure.yml up -d --build
bash deploy/azure/deploy.sh
```

---

## 8. Cost (approx.)

| Resource | ~Monthly (INR) |
|----------|----------------|
| Standard_B2s VM | ₹2,500–3,500 |
| 64 GB disk | included |
| Bandwidth | minimal for demo |

New Azure accounts often have **$200 free credit** for 30 days.

**After testing:** delete the VM resource group to stop billing.

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `502 Bad Gateway` | `docker logs go-connect-api` and `pm2 logs` |
| CORS errors | Match `CORS_ALLOWED_ORIGINS` exactly to browser URLs |
| Migrate fails | Check `DATABASE_URL` uses `127.0.0.1:5432` |
| Out of memory | `free -h` — restart: `pm2 restart all` |
| API can't reach MinIO | API container uses `http://minio:9000` (set in compose) |

---

## Files

| Path | Purpose |
|------|---------|
| `Dockerfile` | API image |
| `docker-compose.azure.yml` | Postgres + MinIO + API |
| `deploy/azure/setup-vm.sh` | First-time bootstrap |
| `deploy/azure/deploy.sh` | Rebuild frontends |
| `deploy/azure/env.example` | Env template |
| `.azure/deployment-plan.md` | Architecture notes |
