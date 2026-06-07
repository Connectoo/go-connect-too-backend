#!/usr/bin/env bash
# Expose MinIO Console at https://minio.${DOMAIN} via Nginx (basic auth + Let's Encrypt).
# Run as root AFTER enable-https-vpc.sh and DNS A record for minio.${DOMAIN}.
#
# Optional in .env before running:
#   MINIO_CONSOLE_AUTH_USER=admin
#   MINIO_CONSOLE_AUTH_PASSWORD=your-strong-password
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo bash deploy/azure/enable-minio-console-vpc.sh)"
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

if [[ -z "${DOMAIN:-}" || "${DOMAIN}" == *"nip.io" ]]; then
  echo "Set DOMAIN in .env to your real domain (not nip.io)."
  exit 1
fi

if [[ -z "${LETSENCRYPT_EMAIL:-}" ]]; then
  echo "Set LETSENCRYPT_EMAIL in .env (same as enable-https-vpc.sh)."
  exit 1
fi

if [[ "${USE_HTTPS:-}" != "true" ]]; then
  echo "Run enable-https-vpc.sh first (USE_HTTPS=true required)."
  exit 1
fi

CONSOLE_HOST="minio.${DOMAIN}"
CONSOLE_URL="https://${CONSOLE_HOST}"
AUTH_USER="${MINIO_CONSOLE_AUTH_USER:-admin}"

if [[ -z "${MINIO_CONSOLE_AUTH_PASSWORD:-}" ]]; then
  MINIO_CONSOLE_AUTH_PASSWORD="$(openssl rand -base64 18 | tr -d '/+=' | head -c 20)"
  echo "Generated MinIO console basic-auth password (save this): ${MINIO_CONSOLE_AUTH_PASSWORD}"
fi

echo "==> Installing htpasswd helper..."
DEBIAN_FRONTEND=noninteractive apt-get install -y apache2-utils

echo "==> Creating Nginx basic auth (${AUTH_USER})..."
htpasswd -bc /etc/nginx/.htpasswd-minio "${AUTH_USER}" "${MINIO_CONSOLE_AUTH_PASSWORD}"
chmod 640 /etc/nginx/.htpasswd-minio
chown root:www-data /etc/nginx/.htpasswd-minio

echo "==> Installing Nginx site..."
export DOMAIN
envsubst '$DOMAIN' < deploy/azure/nginx/minio-console-vpc.conf > /etc/nginx/sites-available/minio-console
ln -sf /etc/nginx/sites-available/minio-console /etc/nginx/sites-enabled/minio-console
nginx -t
systemctl reload nginx

echo "==> Checking HTTP for ${CONSOLE_HOST}..."
if ! curl -fsS --max-time 10 "http://${CONSOLE_HOST}/" -o /dev/null -u "${AUTH_USER}:${MINIO_CONSOLE_AUTH_PASSWORD}"; then
  echo "Cannot reach http://${CONSOLE_HOST}/ — add DNS A record → VM public IP, wait a few minutes, retry."
  exit 1
fi

echo "==> Expanding Let's Encrypt certificate for ${CONSOLE_HOST}..."
certbot --nginx --non-interactive --agree-tos --expand \
  --email "${LETSENCRYPT_EMAIL}" \
  -d "www.${DOMAIN}" \
  -d "admin.${DOMAIN}" \
  -d "app.${DOMAIN}" \
  -d "provider.${DOMAIN}" \
  -d "api.${DOMAIN}" \
  -d "${CONSOLE_HOST}" \
  --redirect

MINIO_BROWSER_REDIRECT_URL="${CONSOLE_URL}"

grep -v '^MINIO_BROWSER_REDIRECT_URL=' .env \
  | grep -v '^MINIO_CONSOLE_AUTH_USER=' \
  | grep -v '^MINIO_CONSOLE_AUTH_PASSWORD=' > .env.tmp || true
{
  cat .env.tmp
  echo "MINIO_BROWSER_REDIRECT_URL=${MINIO_BROWSER_REDIRECT_URL}"
  echo "MINIO_CONSOLE_AUTH_USER=${AUTH_USER}"
  echo "MINIO_CONSOLE_AUTH_PASSWORD=${MINIO_CONSOLE_AUTH_PASSWORD}"
} > .env
rm -f .env.tmp

echo "==> Restarting MinIO with console redirect URL..."
docker compose -f docker-compose.azure.yml up -d minio

nginx -t
systemctl reload nginx

cat <<EOF

MinIO Console enabled.

  URL:      ${CONSOLE_URL}
  Nginx:    user ${AUTH_USER} / password (see .env MINIO_CONSOLE_AUTH_PASSWORD)
  MinIO:    user \${MINIO_ROOT_USER} / \${MINIO_ROOT_PASSWORD} from .env
  Bucket:   go-connect-uploads

Do not open ports 9000/9001 in Azure NSG — Nginx handles public access.

EOF
