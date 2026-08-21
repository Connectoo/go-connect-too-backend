#!/usr/bin/env bash
# Resolve DOMAIN and public URLs for Azure (custom domain or nip.io fallback).
set -euo pipefail

public_ip() {
  curl -fsS -4 --max-time 5 ifconfig.me 2>/dev/null \
    || curl -fsS -4 --max-time 5 icanhazip.com 2>/dev/null \
    || hostname -I | awk '{print $1}'
}

ip_to_nip_domain() {
  echo "${1//./-}.nip.io"
}

resolve_domain_config() {
  local ip domain scheme

  ip="$(public_ip)"
  domain="${DOMAIN:-}"

  if [[ -z "$domain" || "$domain" == "YOUR_DOMAIN" ]]; then
    domain="$(ip_to_nip_domain "$ip")"
    scheme="http"
  elif [[ "${USE_HTTPS:-}" == "true" ]]; then
    scheme="https"
  elif [[ "${API_URL:-}" == https://* ]]; then
    scheme="https"
  else
    scheme="http"
  fi

  export PUBLIC_IP="$ip"
  export DOMAIN="$domain"
  export URL_SCHEME="$scheme"
  export API_URL="${API_URL:-${scheme}://api.${domain}/api/v1}"

  export CUSTOMER_URL="${CUSTOMER_URL:-${scheme}://app.${domain}}"
  export EMPLOYEE_URL="${EMPLOYEE_URL:-${scheme}://provider.${domain}}"
  export CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS:-${scheme}://www.${domain},${scheme}://admin.${domain},${scheme}://app.${domain},${scheme}://provider.${domain},${scheme}://api.${domain}}"

  if [[ "${VPC_MODE:-}" == "true" ]]; then
    export PRIVATE_HOST="${PRIVATE_HOST:-$(hostname -I | awk '{print $1}')}"
    export API_PUBLIC_URL="${API_PUBLIC_URL:-${scheme}://api.${domain}/api/v1}"
  else
    export API_PUBLIC_URL="${API_URL}"
  fi
}
