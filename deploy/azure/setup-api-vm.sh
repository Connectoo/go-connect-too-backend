#!/usr/bin/env bash
# [Optional scale-out] Separate public API VM.
# For a single VM in one VNet, use setup-vm-vpc.sh instead.
# Bootstrap the PUBLIC subnet VM: Go API + Nginx (API only).
# Postgres and Next.js run on the private VM (PRIVATE_HOST).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo bash deploy/azure/setup-api-vm.sh)"
  exit 1
fi

echo "==> Installing system packages..."
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
DEBIAN_FRONTEND=noninteractive apt-get install -y \
  ca-certificates curl git nginx certbot python3-certbot-nginx gettext-base ufw

if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable docker
systemctl start docker

ufw allow OpenSSH
ufw allow 'Nginx Full'
ufw --force enable

if [[ ! -f .env ]]; then
  cp deploy/azure/env.vpc.example .env
  echo ""
  echo "Created .env — set secrets and PRIVATE_HOST (private VM IP), then re-run:"
  echo "  nano $ROOT_DIR/.env"
  echo "  bash deploy/azure/setup-api-vm.sh"
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

export VPC_MODE=true

if [[ -z "${PRIVATE_HOST:-}" || "${PRIVATE_HOST}" == "10.0.2.4" ]]; then
  echo "Set PRIVATE_HOST in .env to the private VM's VNet IP (e.g. 10.0.2.4)"
  exit 1
fi

if [[ "${JWT_ACCESS_SECRET}" == "change-me-access-secret-min-32-chars" ]]; then
  echo "Set JWT_ACCESS_SECRET and JWT_REFRESH_SECRET in .env (must match private VM)"
  exit 1
fi

# shellcheck disable=SC1091
source deploy/azure/resolve-domain.sh
resolve_domain_config

# Persist auto-resolved URLs
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

echo "==> Using DOMAIN=${DOMAIN} (public IP ${PUBLIC_IP})"
echo "==> Private backend at PRIVATE_HOST=${PRIVATE_HOST}"

echo "==> Testing connectivity to private Postgres..."
if ! timeout 5 bash -c "echo >/dev/tcp/${PRIVATE_HOST}/5432" 2>/dev/null; then
  echo "WARNING: Cannot reach ${PRIVATE_HOST}:5432 — check VNet, NSG, and that private VM is running."
fi

echo "==> Starting API (docker-compose.azure.api.yml)..."
docker compose -f docker-compose.azure.api.yml up -d --build

echo "==> Configuring Nginx (API only)..."
export DOMAIN
envsubst '$DOMAIN' < deploy/azure/nginx/go-connect-api-only.conf > /etc/nginx/sites-available/go-connect
ln -sf /etc/nginx/sites-available/go-connect /etc/nginx/sites-enabled/go-connect
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

cat <<EOF

Public API VM setup complete.

API (internet-facing):
  ${API_URL}/health

Frontends (private VM only — use VPN/Bastion):
  http://${PRIVATE_HOST}:3000  (website)
  http://${PRIVATE_HOST}:3001  (admin)
  http://${PRIVATE_HOST}:3002  (customer)
  http://${PRIVATE_HOST}:3003  (employee)

Optional HTTPS on API only:
  certbot --nginx -d api.${DOMAIN}
  # then set USE_HTTPS=true in .env and re-run this script

EOF
