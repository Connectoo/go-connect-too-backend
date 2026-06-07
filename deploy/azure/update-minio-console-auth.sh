#!/usr/bin/env bash
# Sync Nginx basic auth for MinIO console from .env (MINIO_CONSOLE_AUTH_*).
# Run as root after changing MINIO_CONSOLE_AUTH_PASSWORD in .env:
#   sudo bash deploy/azure/update-minio-console-auth.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo bash deploy/azure/update-minio-console-auth.sh)"
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

AUTH_USER="${MINIO_CONSOLE_AUTH_USER:-admin}"

if [[ -z "${MINIO_CONSOLE_AUTH_PASSWORD:-}" ]]; then
  echo "Set MINIO_CONSOLE_AUTH_PASSWORD in .env first."
  exit 1
fi

DEBIAN_FRONTEND=noninteractive apt-get install -y apache2-utils

echo "==> Updating Nginx basic auth for MinIO console (${AUTH_USER})..."
htpasswd -bc /etc/nginx/.htpasswd-minio "${AUTH_USER}" "${MINIO_CONSOLE_AUTH_PASSWORD}"
chmod 640 /etc/nginx/.htpasswd-minio
chown root:www-data /etc/nginx/.htpasswd-minio

nginx -t
systemctl reload nginx

echo "MinIO console Nginx auth updated for user: ${AUTH_USER}"
