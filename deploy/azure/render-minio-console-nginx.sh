#!/usr/bin/env bash
# Render MinIO console nginx site. Mode: http (port 80) or ssl (443 + redirect).
set -euo pipefail

mode="${1:-ssl}"
domain="${2:-}"

if [[ -z "${domain}" ]]; then
  echo "usage: render-minio-console-nginx.sh <http|ssl> <domain>"
  exit 1
fi

case "${mode}" in
  http)
    template="deploy/azure/nginx/minio-console-vpc-http.conf"
    ;;
  ssl)
    template="deploy/azure/nginx/minio-console-vpc.conf"
    if [[ ! -f "/etc/letsencrypt/live/www.${domain}/fullchain.pem" ]]; then
      echo "Missing /etc/letsencrypt/live/www.${domain}/fullchain.pem — run certbot first."
      exit 1
    fi
    ;;
  *)
    echo "usage: render-minio-console-nginx.sh <http|ssl> <domain>"
    exit 1
    ;;
esac

export DOMAIN="${domain}"
envsubst '$DOMAIN' < "${template}" > /etc/nginx/sites-available/minio-console
ln -sf /etc/nginx/sites-available/minio-console /etc/nginx/sites-enabled/minio-console
