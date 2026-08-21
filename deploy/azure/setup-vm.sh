#!/usr/bin/env bash
# First-time bootstrap on a fresh Ubuntu 24.04 Azure VM.
# Run as root: bash deploy/azure/setup-vm.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo bash deploy/azure/setup-vm.sh)"
  exit 1
fi

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
  cp deploy/azure/env.example .env
  echo ""
  echo "Created .env — set POSTGRES_PASSWORD, MINIO_ROOT_PASSWORD, JWT secrets, then re-run:"
  echo "  nano $ROOT_DIR/.env"
  echo "  bash deploy/azure/setup-vm.sh"
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

# Persist auto-resolved domain URLs for redeploys
grep -v '^DOMAIN=' .env | grep -v '^CORS_ALLOWED_ORIGINS=' | grep -v '^API_URL=' | grep -v '^S3_BASE_URL=' > .env.tmp || true
{
  cat .env.tmp
  echo "DOMAIN=${DOMAIN}"
  echo "API_URL=${API_URL}"
  echo "CORS_ALLOWED_ORIGINS=${CORS_ALLOWED_ORIGINS}"
  echo "S3_BASE_URL=${URL_SCHEME}://api.${DOMAIN}/files"
} > .env
rm -f .env.tmp

set -a
# shellcheck disable=SC1091
source .env
set +a

echo "==> Using DOMAIN=${DOMAIN} (VM IP ${PUBLIC_IP})"

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
export PATH="/usr/local/go/bin:$(go env GOPATH 2>/dev/null)/bin:$PATH"
go run ./cmd/seed

echo "==> Building and starting frontends..."
bash deploy/azure/deploy.sh

echo "==> Configuring Nginx..."
export DOMAIN
envsubst '$DOMAIN' < deploy/azure/nginx/go-connect.conf > /etc/nginx/sites-available/go-connect
ln -sf /etc/nginx/sites-available/go-connect /etc/nginx/sites-enabled/go-connect
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

cat <<EOF

Setup complete (no custom domain required).

VM IP: ${PUBLIC_IP}
Base host: ${DOMAIN}

Open in browser:
  Website:  ${URL_SCHEME}://www.${DOMAIN}
  Admin:    ${URL_SCHEME}://admin.${DOMAIN}/login
  Customer: ${URL_SCHEME}://app.${DOMAIN}/login
  Employee: ${URL_SCHEME}://provider.${DOMAIN}/login
  API:      ${API_URL}/health

Demo logins (password Demo123!):
  admin@yopmail.com, alice@yopmail.com, karim@yopmail.com

Share these links with your client — nip.io resolves automatically, no DNS setup.

Optional HTTPS (if certbot accepts nip.io for your IP):
  certbot --nginx -d api.${DOMAIN} -d www.${DOMAIN} -d admin.${DOMAIN} \\
    -d app.${DOMAIN} -d provider.${DOMAIN}
  # then set USE_HTTPS=true in .env and run: bash deploy/azure/deploy.sh

EOF
