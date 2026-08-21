#!/usr/bin/env bash
# Build/restart Next.js apps. Run from repo root after .env exists.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ ! -f .env ]]; then
  echo "Missing .env — copy deploy/azure/env.example to .env first."
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

# shellcheck disable=SC1091
source deploy/azure/resolve-domain.sh
resolve_domain_config

API_URL="${API_URL:-${URL_SCHEME}://api.${DOMAIN}/api/v1}"
CUSTOMER_URL="${CUSTOMER_URL:-${URL_SCHEME}://app.${DOMAIN}}"
EMPLOYEE_URL="${EMPLOYEE_URL:-${URL_SCHEME}://provider.${DOMAIN}}"

build_app() {
  local name="$1"
  local dir="$2"
  local port="$3"
  local env_file="$dir/.env.production.local"

  echo "==> Building ${name}..."
  cat >"$env_file" <<EOF
NEXT_PUBLIC_API_URL=${API_URL}
EOF

  if [[ "$name" == "website" ]]; then
    cat >>"$env_file" <<EOF
NEXT_PUBLIC_CUSTOMER_PORTAL_URL=${CUSTOMER_URL}
NEXT_PUBLIC_EMPLOYEE_PORTAL_URL=${EMPLOYEE_URL}
EOF
  fi

  local bind_host="${BIND_HOST:-127.0.0.1}"

  (cd "$dir" && npm install && npm run build)
  pm2 delete "$name" 2>/dev/null || true
  pm2 start npm --name "$name" --cwd "$dir" -- start -- -p "$port" -H "$bind_host"
}

build_app website web/website 3000
build_app admin web/admin 3001
build_app customer web/customer 3002
build_app employee web/employee 3003

pm2 save
echo "Frontends running under PM2."
