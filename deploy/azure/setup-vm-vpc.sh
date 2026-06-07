#!/usr/bin/env bash
# Single VM in Azure VNet — frontends public, API + Postgres private.
# Run as root: bash deploy/azure/setup-vm-vpc.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo bash deploy/azure/setup-vm-vpc.sh)"
  exit 1
fi

export VPC_MODE=true

echo "==> Installing system packages..."
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
DEBIAN_FRONTEND=noninteractive apt-get install -y \
  ca-certificates curl git nginx certbot python3-certbot-nginx \
  gettext-base ufw

if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable docker
systemctl start docker

if ! command -v node >/dev/null 2>&1 || [[ "$(node -v)" != v20* ]]; then
  curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
  DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs
fi

if ! command -v pm2 >/dev/null 2>&1; then
  npm install -g pm2
fi

if ! command -v go >/dev/null 2>&1 && [[ ! -x /usr/local/go/bin/go ]]; then
  echo "==> Installing Go..."
  curl -fsSL https://go.dev/dl/go1.26.3.linux-amd64.tar.gz -o /tmp/go.tar.gz
  tar -C /usr/local -xzf /tmp/go.tar.gz
fi
export PATH="/usr/local/go/bin:${PATH}"
if command -v go >/dev/null 2>&1; then
  export PATH="$(go env GOPATH)/bin:${PATH}"
fi

if ! command -v migrate >/dev/null 2>&1; then
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ln -sf "$(go env GOPATH)/bin/migrate" /usr/local/bin/migrate
fi

ufw allow OpenSSH
ufw allow 'Nginx Full'
ufw --force enable

if [[ ! -f .env ]]; then
  cp deploy/azure/env.vpc.example .env
  echo ""
  echo "Created .env — set POSTGRES_PASSWORD, MINIO_ROOT_PASSWORD, JWT secrets, then re-run:"
  echo "  nano $ROOT_DIR/.env"
  echo "  bash deploy/azure/setup-vm-vpc.sh"
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

# shellcheck disable=SC1091
source deploy/azure/resolve-domain.sh
resolve_domain_config

if [[ "${JWT_ACCESS_SECRET}" == "change-me-access-secret-min-32-chars" ]]; then
  echo "Set JWT_ACCESS_SECRET and JWT_REFRESH_SECRET in .env (openssl rand -hex 32)"
  exit 1
fi

if [[ "${POSTGRES_PASSWORD}" == "change-me-strong-postgres-password" ]]; then
  echo "Set POSTGRES_PASSWORD in .env and match it in DATABASE_URL"
  exit 1
fi

S3_BASE_URL="${URL_SCHEME}://www.${DOMAIN}/files"

# Persist auto-resolved domain URLs
grep -v '^DOMAIN=' .env | grep -v '^CORS_ALLOWED_ORIGINS=' | grep -v '^API_URL=' \
  | grep -v '^S3_BASE_URL=' | grep -v '^VPC_MODE=' > .env.tmp || true
{
  cat .env.tmp
  echo "VPC_MODE=true"
  echo "DOMAIN=${DOMAIN}"
  echo "API_URL=${API_PUBLIC_URL}"
  echo "CORS_ALLOWED_ORIGINS=${CORS_ALLOWED_ORIGINS}"
  echo "S3_BASE_URL=${S3_BASE_URL}"
} > .env
rm -f .env.tmp

set -a
# shellcheck disable=SC1091
source .env
set +a

echo "==> VPC mode — public frontends on ${DOMAIN}"
echo "==> API private on 127.0.0.1:8080 (proxied as /api/ on each frontend host)"
echo "==> Postgres + MinIO on localhost only"

echo "==> Starting Docker services (postgres, minio, api)..."
docker compose -f docker-compose.azure.yml up -d --build

echo "==> Waiting for Postgres..."
for _ in $(seq 1 30); do
  if docker exec go-connect-postgres pg_isready -U app -d go_connect >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "==> Running migrations..."
migrate -path migrations -database "${DATABASE_URL}" up

echo "==> Seeding demo data..."
go run ./cmd/seed

echo "==> Building public Next.js apps..."
bash deploy/azure/deploy-vpc.sh

echo "==> Configuring Nginx (frontends public, API via /api/ proxy)..."
export DOMAIN
envsubst '$DOMAIN' < deploy/azure/nginx/go-connect-vpc.conf > /etc/nginx/sites-available/go-connect
ln -sf /etc/nginx/sites-available/go-connect /etc/nginx/sites-enabled/go-connect
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

cat <<EOF

VPC setup complete (single VM in go-connect-vnet).

Public frontends:
  Website:  ${URL_SCHEME}://www.${DOMAIN}
  Admin:    ${URL_SCHEME}://admin.${DOMAIN}/login
  Customer: ${URL_SCHEME}://app.${DOMAIN}/login
  Employee: ${URL_SCHEME}://provider.${DOMAIN}/login

API (private — not on api.*; reached via frontend /api/ proxy):
  curl ${URL_SCHEME}://www.${DOMAIN}/api/v1/health

Demo logins (password Demo123!):
  admin@yopmail.com, alice@yopmail.com, karim@yopmail.com

HTTPS (real domain required — not nip.io):
  # 1. A records: www, admin, app, provider → VM public IP
  # 2. Set DOMAIN=goconnect.in and LETSENCRYPT_EMAIL in .env
  # 3. bash deploy/azure/enable-https-vpc.sh

EOF
