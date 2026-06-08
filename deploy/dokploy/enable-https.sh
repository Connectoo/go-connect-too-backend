#!/usr/bin/env bash
# Dokploy + Traefik: issue ONE Let's Encrypt cert, adding subdomains one at a time.
#
# Why: Dokploy creates separate Traefik ACME orders per domain (races / "Cannot retrieve
# challenge"). This script uses host certbot like nginx --expand, then installs a single
# static cert for Traefik.
#
# Run on the Dokploy VM as root (SSH):
#   DOMAIN=connectoo.online LETSENCRYPT_EMAIL=you@example.com bash deploy/dokploy/enable-https.sh
#
# Before running:
#   1. Dokploy → Compose → Domains: turn OFF HTTPS / Let's Encrypt on every subdomain.
#   2. DNS A records for all subdomains → this server's public IP.
#   3. In /etc/dokploy/traefik/traefik.yml remove (if present):
#        entryPoints.websecure.http.tls.certResolver: letsencrypt
#
# After success:
#   - Dokploy → Domains: HTTPS On, Certificate = None (NOT Let's Encrypt).
#   - Redeploy compose. Traefik serves the static cert from connectoo-static-tls.yml.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root: sudo DOMAIN=connectoo.online LETSENCRYPT_EMAIL=you@example.com bash deploy/dokploy/enable-https.sh"
  exit 1
fi

DOMAIN="${DOMAIN:-}"
LETSENCRYPT_EMAIL="${LETSENCRYPT_EMAIL:-}"
CERT_DELAY_SEC="${CERT_DELAY_SEC:-90}"
TRAEFIK_CONTAINER="${TRAEFIK_CONTAINER:-}"
DOKPLOY_CERT_DIR="${DOKPLOY_CERT_DIR:-/etc/dokploy/traefik/certs/connectoo}"
DOKPLOY_DYNAMIC_DIR="${DOKPLOY_DYNAMIC_DIR:-/etc/dokploy/traefik/dynamic}"
STATIC_TLS_FILE="${DOKPLOY_DYNAMIC_DIR}/connectoo-static-tls.yml"

# Order: api first (health), then frontends, then MinIO.
SUBDOMAINS=(
  api
  www
  admin
  app
  provider
  minio
  storage
)

usage() {
  cat <<'EOF'
Usage (on Dokploy server as root):
  DOMAIN=connectoo.online LETSENCRYPT_EMAIL=you@example.com bash deploy/dokploy/enable-https.sh

Optional env:
  CERT_DELAY_SEC=90          seconds between each --expand step
  TRAEFIK_CONTAINER=name     auto-detected if empty
  DOKPLOY_CERT_DIR=...       copy fullchain.pem / privkey.pem here
  DOKPLOY_DYNAMIC_DIR=...    Traefik file provider directory

Adds one subdomain per certbot step into the same certificate (like nginx --expand).
EOF
}

if [[ -z "${DOMAIN}" || -z "${LETSENCRYPT_EMAIL}" ]]; then
  usage
  echo ""
  echo "Set DOMAIN and LETSENCRYPT_EMAIL."
  exit 1
fi

if [[ "${DOMAIN}" == *"nip.io" ]]; then
  echo "Let's Encrypt does not issue certificates for nip.io."
  exit 1
fi

if ! command -v certbot >/dev/null 2>&1; then
  echo "Installing certbot..."
  apt-get update -qq
  apt-get install -y certbot
fi

find_traefik_container() {
  if [[ -n "${TRAEFIK_CONTAINER}" ]]; then
    echo "${TRAEFIK_CONTAINER}"
    return
  fi
  docker ps --format '{{.Names}}' | grep -i traefik | head -1 || true
}

TRAEFIK_CONTAINER="$(find_traefik_container)"
if [[ -z "${TRAEFIK_CONTAINER}" ]]; then
  echo "Warning: no Traefik container found — certbot will bind port 80 directly."
else
  echo "Traefik container: ${TRAEFIK_CONTAINER}"
fi

stop_traefik() {
  if [[ -n "${TRAEFIK_CONTAINER}" ]]; then
    echo "==> Stopping Traefik (free port 80 for certbot)..."
    docker stop "${TRAEFIK_CONTAINER}" >/dev/null
  fi
}

start_traefik() {
  if [[ -n "${TRAEFIK_CONTAINER}" ]]; then
    echo "==> Starting Traefik..."
    docker start "${TRAEFIK_CONTAINER}" >/dev/null
  fi
}

check_http_reachable() {
  local host="$1"
  local path="${2:-/}"
  echo "==> Checking http://${host}${path}"
  if ! curl -fsS --max-time 15 "http://${host}${path}" -o /dev/null; then
    echo "Cannot reach http://${host}${path}"
    echo "Fix DNS (A → this server) and Dokploy domain routing first."
    exit 1
  fi
}

install_static_tls_config() {
  local cert_dir_on_host="$1"
  mkdir -p "${DOKPLOY_CERT_DIR}"
  cp "${cert_dir_on_host}/fullchain.pem" "${DOKPLOY_CERT_DIR}/fullchain.pem"
  cp "${cert_dir_on_host}/privkey.pem" "${DOKPLOY_CERT_DIR}/privkey.pem"
  chmod 644 "${DOKPLOY_CERT_DIR}/fullchain.pem"
  chmod 600 "${DOKPLOY_CERT_DIR}/privkey.pem"

  cat >"${STATIC_TLS_FILE}" <<EOF
# Installed by deploy/dokploy/enable-https.sh — one cert, all subdomains.
# Dokploy Domains: HTTPS On, Certificate = None (do not use Let's Encrypt here).
tls:
  stores:
    default:
      defaultCertificate:
        certFile: ${DOKPLOY_CERT_DIR}/fullchain.pem
        keyFile: ${DOKPLOY_CERT_DIR}/privkey.pem
  certificates:
    - certFile: ${DOKPLOY_CERT_DIR}/fullchain.pem
      keyFile: ${DOKPLOY_CERT_DIR}/privkey.pem
EOF
  echo "==> Wrote ${STATIC_TLS_FILE}"
}

install_renew_hook() {
  local hook="/etc/letsencrypt/renewal-hooks/deploy/dokploy-connectoo.sh"
  mkdir -p "$(dirname "${hook}")"
  cat >"${hook}" <<EOF
#!/usr/bin/env bash
set -euo pipefail
cp "${CERT_DIR}/fullchain.pem" "${DOKPLOY_CERT_DIR}/fullchain.pem"
cp "${CERT_DIR}/privkey.pem" "${DOKPLOY_CERT_DIR}/privkey.pem"
chmod 644 "${DOKPLOY_CERT_DIR}/fullchain.pem"
chmod 600 "${DOKPLOY_CERT_DIR}/privkey.pem"
TRAEFIK="\$(docker ps --format '{{.Names}}' | grep -i traefik | head -1 || true)"
if [[ -n "\${TRAEFIK}" ]]; then
  docker restart "\${TRAEFIK}" >/dev/null || true
fi
EOF
  chmod +x "${hook}"
  echo "==> Renew hook: ${hook}"
}

echo "==> Pre-flight HTTP checks (Traefik routing on port 80)..."
start_traefik
check_http_reachable "api.${DOMAIN}" "/api/v1/health"
for sub in www admin app provider minio storage; do
  check_http_reachable "${sub}.${DOMAIN}" "/"
done

CERT_ACCUM=()
step=0
total="${#SUBDOMAINS[@]}"

for sub in "${SUBDOMAINS[@]}"; do
  host="${sub}.${DOMAIN}"
  CERT_ACCUM+=("-d" "${host}")
  step=$((step + 1))

  echo ""
  echo "========================================"
  echo "==> Certbot step ${step}/${total}: ${host}"
  echo "========================================"

  stop_traefik

  if [[ "${step}" -eq 1 ]]; then
    certbot certonly \
      --standalone \
      --preferred-challenges http \
      --non-interactive \
      --agree-tos \
      --email "${LETSENCRYPT_EMAIL}" \
      --keep-until-expiring \
      "${CERT_ACCUM[@]}"
  else
    certbot certonly \
      --standalone \
      --preferred-challenges http \
      --non-interactive \
      --agree-tos \
      --email "${LETSENCRYPT_EMAIL}" \
      --expand \
      --keep-until-expiring \
      "${CERT_ACCUM[@]}"
  fi

  start_traefik

  if [[ "${step}" -lt "${total}" ]]; then
    echo "==> Waiting ${CERT_DELAY_SEC}s before next subdomain..."
    sleep "${CERT_DELAY_SEC}"
  fi
done

CERT_DIR="/etc/letsencrypt/live/api.${DOMAIN}"
if [[ ! -f "${CERT_DIR}/fullchain.pem" ]]; then
  echo "Certificate not found at ${CERT_DIR}"
  certbot certificates
  exit 1
fi

echo ""
echo "==> Installing static TLS cert for Traefik..."
install_static_tls_config "${CERT_DIR}"
install_renew_hook

if [[ -n "${TRAEFIK_CONTAINER}" ]]; then
  echo "==> Restarting Traefik..."
  docker restart "${TRAEFIK_CONTAINER}" >/dev/null
fi

echo ""
echo "Done — one certificate covers:"
for sub in "${SUBDOMAINS[@]}"; do
  echo "  - ${sub}.${DOMAIN}"
done
echo ""
echo "Files:"
echo "  ${DOKPLOY_CERT_DIR}/fullchain.pem"
echo "  ${DOKPLOY_CERT_DIR}/privkey.pem"
echo "  ${STATIC_TLS_FILE}"
echo ""
echo "Next steps:"
echo "  1. Remove websecure certResolver from /etc/dokploy/traefik/traefik.yml if still present."
echo "  2. Dokploy → Domains: HTTPS On, Certificate = None (not Let's Encrypt)."
echo "  3. Redeploy compose."
echo "  4. Test: curl -vI https://api.${DOMAIN}/api/v1/health"
echo ""
echo "Renewal: systemctl status certbot.timer"
