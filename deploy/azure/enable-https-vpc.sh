#!/usr/bin/env bash
# Enable Let's Encrypt HTTPS for VPC deployment (www, admin, app, provider).
# Requires a real domain — nip.io will NOT work with Let's Encrypt.
# Run as root AFTER setup-vm-vpc.sh: bash deploy/azure/enable-https-vpc.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo bash deploy/azure/enable-https-vpc.sh)"
  exit 1
fi

if [[ ! -f .env ]]; then
  echo "Missing .env — run setup-vm-vpc.sh first."
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

export VPC_MODE=true

if [[ -z "${DOMAIN:-}" ]]; then
  echo "Set DOMAIN in .env to your real domain (e.g. goconnect.in)."
  echo "Let's Encrypt does not issue certificates for nip.io."
  exit 1
fi

if [[ "${DOMAIN}" == *"nip.io" ]]; then
  echo "DOMAIN is nip.io (${DOMAIN}) — Let's Encrypt cannot issue certs for nip.io."
  echo "Point a real domain (A records for www, admin, app, provider) to this VM, then set DOMAIN=goconnect.in"
  exit 1
fi

if [[ -z "${LETSENCRYPT_EMAIL:-}" ]]; then
  echo "Set LETSENCRYPT_EMAIL in .env (used for cert expiry notices)."
  exit 1
fi

# shellcheck disable=SC1091
source deploy/azure/resolve-domain.sh
resolve_domain_config

echo "==> Checking HTTP is reachable before certbot..."
for host in www admin app provider api; do
  if ! curl -fsS --max-time 10 "http://${host}.${DOMAIN}/" -o /dev/null; then
    echo "Cannot reach http://${host}.${DOMAIN}/ — fix DNS (A → VM public IP) and Nginx first."
    exit 1
  fi
done

echo "==> Requesting Let's Encrypt certificate..."
certbot --nginx --non-interactive --agree-tos \
  --email "${LETSENCRYPT_EMAIL}" \
  -d "www.${DOMAIN}" \
  -d "admin.${DOMAIN}" \
  -d "app.${DOMAIN}" \
  -d "provider.${DOMAIN}" \
  -d "api.${DOMAIN}" \
  --redirect

export USE_HTTPS=true
S3_BASE_URL="https://www.${DOMAIN}/files"

grep -v '^USE_HTTPS=' .env | grep -v '^DOMAIN=' | grep -v '^CORS_ALLOWED_ORIGINS=' \
  | grep -v '^API_URL=' | grep -v '^S3_BASE_URL=' > .env.tmp || true
{
  cat .env.tmp
  echo "USE_HTTPS=true"
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

# shellcheck disable=SC1091
source deploy/azure/resolve-domain.sh
resolve_domain_config

echo "==> Rebuilding frontends with HTTPS URLs..."
bash deploy/azure/deploy-vpc.sh

echo "==> Restarting API (pick up CORS / S3_BASE_URL)..."
docker compose -f docker-compose.azure.yml up -d

nginx -t
systemctl reload nginx

cat <<EOF

HTTPS enabled (Let's Encrypt).

  Website:  https://www.${DOMAIN}
  Admin:    https://admin.${DOMAIN}/login
  Customer: https://app.${DOMAIN}/login
  Employee: https://provider.${DOMAIN}/login
  API:      https://api.${DOMAIN}/api/v1/health
  Docs:     https://api.${DOMAIN}/api/v1/docs/

Cert auto-renews via certbot systemd timer. Check: systemctl status certbot.timer

EOF
