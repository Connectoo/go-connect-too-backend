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

persist_minio_env() {
  grep -v '^MINIO_BROWSER_REDIRECT_URL=' .env \
    | grep -v '^MINIO_CONSOLE_AUTH_USER=' \
    | grep -v '^MINIO_CONSOLE_AUTH_PASSWORD=' > .env.tmp || true
  {
    cat .env.tmp
    echo "MINIO_BROWSER_REDIRECT_URL=${CONSOLE_URL}"
    echo "MINIO_CONSOLE_AUTH_USER=${AUTH_USER}"
    echo "MINIO_CONSOLE_AUTH_PASSWORD=${MINIO_CONSOLE_AUTH_PASSWORD}"
  } > .env
  rm -f .env.tmp
}

curl_console() {
  curl -fsS --max-time 10 -u "${AUTH_USER}:${MINIO_CONSOLE_AUTH_PASSWORD}" "$@" -o /dev/null
}

echo "==> Installing htpasswd helper..."
DEBIAN_FRONTEND=noninteractive apt-get install -y apache2-utils

echo "==> Creating Nginx basic auth (${AUTH_USER})..."
htpasswd -bc /etc/nginx/.htpasswd-minio "${AUTH_USER}" "${MINIO_CONSOLE_AUTH_PASSWORD}"
chmod 640 /etc/nginx/.htpasswd-minio
chown root:www-data /etc/nginx/.htpasswd-minio

echo "==> Configuring MinIO console redirect URL (required before reverse proxy works)..."
persist_minio_env

echo "==> Restarting MinIO..."
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

echo "==> Installing Nginx site..."
export DOMAIN
envsubst '$DOMAIN' < deploy/azure/nginx/minio-console-vpc.conf > /etc/nginx/sites-available/minio-console
if ! grep -q "server_name ${CONSOLE_HOST};" /etc/nginx/sites-available/minio-console; then
  echo "Nginx config render failed — expected server_name ${CONSOLE_HOST};"
  cat /etc/nginx/sites-available/minio-console
  exit 1
fi
ln -sf /etc/nginx/sites-available/minio-console /etc/nginx/sites-enabled/minio-console
nginx -t
systemctl reload nginx

echo "==> Checking Nginx proxy locally..."
if ! curl_console -H "Host: ${CONSOLE_HOST}" "http://127.0.0.1/"; then
  echo "Nginx is not proxying ${CONSOLE_HOST} correctly."
  echo "Debug:"
  echo "  curl -I -u ${AUTH_USER}:**** -H \"Host: ${CONSOLE_HOST}\" http://127.0.0.1/"
  echo "  nginx -T | grep -A5 minio"
  exit 1
fi

echo "==> Checking HTTP for ${CONSOLE_HOST} (public DNS, optional)..."
if ! curl_console "http://${CONSOLE_HOST}/"; then
  vm_ip="$(curl -fsS --max-time 5 ifconfig.me 2>/dev/null || true)"
  dns_ip="$(dig +short "${CONSOLE_HOST}" | tail -n1)"
  echo "Note: cannot reach http://${CONSOLE_HOST}/ from this VM (common on Azure hairpin NAT)."
  echo "  DNS ${CONSOLE_HOST} → ${dns_ip:-<no record>}"
  echo "  VM public IP       → ${vm_ip:-unknown}"
  if [[ -n "${vm_ip}" && -n "${dns_ip}" && "${dns_ip}" != "${vm_ip}" ]]; then
    echo "Warning: DNS may not point to this VM — certbot may fail."
  else
    echo "Local Nginx check passed — continuing (verify from your laptop if needed)."
  fi
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
