#!/usr/bin/env bash
# Enable Let's Encrypt HTTPS for VPC deployment (www, admin, app, provider, api, minio).
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
  echo "Point a real domain (A records for www, admin, app, provider, minio) to this VM, then set DOMAIN=goconnect.in"
  exit 1
fi

if [[ -z "${LETSENCRYPT_EMAIL:-}" ]]; then
  echo "Set LETSENCRYPT_EMAIL in .env (used for cert expiry notices)."
  exit 1
fi

# shellcheck disable=SC1091
source deploy/azure/resolve-domain.sh
resolve_domain_config

MINIO_CONSOLE_HOST="minio.${DOMAIN}"
MINIO_CONSOLE_URL="https://${MINIO_CONSOLE_HOST}"

echo "==> Checking HTTP is reachable before certbot..."
for host in www admin app provider; do
  if ! curl -fsS --max-time 10 "http://${host}.${DOMAIN}/" -o /dev/null; then
    echo "Cannot reach http://${host}.${DOMAIN}/ — fix DNS (A → VM public IP) and Nginx first."
    exit 1
  fi
done
if ! curl -fsS --max-time 10 "http://api.${DOMAIN}/api/v1/health" -o /dev/null; then
  echo "Cannot reach http://api.${DOMAIN}/api/v1/health — fix DNS (A → VM public IP) and Nginx first."
  exit 1
fi

echo "==> Installing MinIO console Nginx site (for minio.${DOMAIN} certificate)..."
grep -v '^MINIO_BROWSER_REDIRECT_URL=' .env \
  | grep -v '^MINIO_CONSOLE_AUTH_USER=' \
  | grep -v '^MINIO_CONSOLE_AUTH_PASSWORD=' > .env.tmp || true
{
  cat .env.tmp
  echo "MINIO_BROWSER_REDIRECT_URL=${MINIO_CONSOLE_URL}"
} > .env
rm -f .env.tmp

set -a
# shellcheck disable=SC1091
source .env
set +a

docker compose -f docker-compose.azure.yml up -d minio

echo "==> Waiting for MinIO console on 127.0.0.1:9001..."
for _ in $(seq 1 30); do
  if curl -fsS --max-time 2 "http://127.0.0.1:9001/" -o /dev/null; then
    break
  fi
  sleep 2
done
if ! curl -fsS --max-time 5 "http://127.0.0.1:9001/" -o /dev/null; then
  echo "MinIO console is not responding on 127.0.0.1:9001 — check: docker logs go-connect-minio"
  exit 1
fi

export DOMAIN
bash deploy/azure/render-minio-console-nginx.sh http "${DOMAIN}"
nginx -t
systemctl reload nginx

if ! curl -fsS --max-time 10 -H "Host: ${MINIO_CONSOLE_HOST}" "http://127.0.0.1/" -o /dev/null; then
  echo "Nginx is not proxying ${MINIO_CONSOLE_HOST} — check sites-enabled/minio-console"
  exit 1
fi

echo "==> Requesting Let's Encrypt certificate..."
certbot --nginx --non-interactive --agree-tos --expand \
  --email "${LETSENCRYPT_EMAIL}" \
  -d "www.${DOMAIN}" \
  -d "admin.${DOMAIN}" \
  -d "app.${DOMAIN}" \
  -d "provider.${DOMAIN}" \
  -d "api.${DOMAIN}" \
  -d "${MINIO_CONSOLE_HOST}" \
  --redirect

echo "==> Installing MinIO HTTPS nginx (listen 443 ssl)..."
bash deploy/azure/render-minio-console-nginx.sh ssl "${DOMAIN}"

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

echo "==> Restarting API (pick up CORS / S3_BASE_URL)..."
docker compose -f docker-compose.azure.yml up -d

if [[ "${REBUILD_FRONTENDS_AFTER_HTTPS:-}" == "1" ]]; then
  echo "==> Rebuilding frontends with HTTPS URLs..."
  bash deploy/azure/deploy-vpc.sh
fi

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
  MinIO:    https://minio.${DOMAIN}

Cert auto-renews via certbot systemd timer. Check: systemctl status certbot.timer

If frontends still use http:// API URLs: REBUILD_FRONTENDS_AFTER_HTTPS=1 bash deploy/azure/enable-https-vpc.sh
Or: bash deploy/azure/deploy-vpc.sh

EOF
