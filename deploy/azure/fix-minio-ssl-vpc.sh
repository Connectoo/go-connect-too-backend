#!/usr/bin/env bash
# Add minio.${DOMAIN} to the Let's Encrypt cert and enable HTTPS on the MinIO nginx site.
# Run when https://minio.<domain> shows "Not secure" / certificate name mismatch.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo bash deploy/azure/fix-minio-ssl-vpc.sh)"
  exit 1
fi

if [[ ! -f .env ]]; then
  echo "Missing .env"
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

if [[ -z "${DOMAIN:-}" || -z "${LETSENCRYPT_EMAIL:-}" ]]; then
  echo "Set DOMAIN and LETSENCRYPT_EMAIL in .env"
  exit 1
fi

CONSOLE_HOST="minio.${DOMAIN}"
CONSOLE_URL="https://${CONSOLE_HOST}"

cert_covers_minio() {
  echo | openssl s_client -servername "${CONSOLE_HOST}" -connect 127.0.0.1:443 2>/dev/null \
    | openssl x509 -noout -text 2>/dev/null \
    | grep -q "DNS:${CONSOLE_HOST}"
}

echo "==> Ensuring MinIO nginx site and redirect URL..."
grep -v '^MINIO_BROWSER_REDIRECT_URL=' .env \
  | grep -v '^MINIO_SERVER_URL=' > .env.tmp || true
{
  cat .env.tmp
  echo "MINIO_BROWSER_REDIRECT_URL=${CONSOLE_URL}"
  echo "MINIO_SERVER_URL=${CONSOLE_URL}"
} > .env
rm -f .env.tmp

docker compose -f docker-compose.azure.yml up -d minio

export DOMAIN
envsubst '$DOMAIN' < deploy/azure/nginx/minio-console-vpc.conf > /etc/nginx/sites-available/minio-console
ln -sf /etc/nginx/sites-available/minio-console /etc/nginx/sites-enabled/minio-console
nginx -t
systemctl reload nginx

if cert_covers_minio; then
  echo "Certificate already valid for ${CONSOLE_HOST}."
  exit 0
fi

echo "==> Expanding Let's Encrypt certificate to include ${CONSOLE_HOST}..."
certbot --nginx --non-interactive --agree-tos --expand \
  --email "${LETSENCRYPT_EMAIL}" \
  -d "www.${DOMAIN}" \
  -d "admin.${DOMAIN}" \
  -d "app.${DOMAIN}" \
  -d "provider.${DOMAIN}" \
  -d "api.${DOMAIN}" \
  -d "${CONSOLE_HOST}" \
  --redirect

nginx -t
systemctl reload nginx

if cert_covers_minio; then
  echo "OK: https://${CONSOLE_HOST} now has a valid certificate."
else
  echo "Certbot finished but ${CONSOLE_HOST} is still missing from the certificate."
  echo "Check: certbot certificates"
  echo "       cat /etc/nginx/sites-enabled/minio-console"
  exit 1
fi
