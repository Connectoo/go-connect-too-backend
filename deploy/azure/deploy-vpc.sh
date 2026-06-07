#!/usr/bin/env bash
# Build/restart Next.js apps for VPC mode — public frontends, API via same-origin /api/ proxy.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ ! -f .env ]]; then
  echo "Missing .env — copy deploy/azure/env.vpc.example to .env first."
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

export VPC_MODE=true

# shellcheck disable=SC1091
source deploy/azure/resolve-domain.sh
resolve_domain_config

CUSTOMER_URL="${CUSTOMER_URL:-${URL_SCHEME}://app.${DOMAIN}}"
EMPLOYEE_URL="${EMPLOYEE_URL:-${URL_SCHEME}://provider.${DOMAIN}}"
BIND_HOST="${BIND_HOST:-127.0.0.1}"

build_app() {
  local name="$1"
  local dir="$2"
  local port="$3"
  local host_label="$4"
  local env_file="$dir/.env.production.local"
  local api_url="${URL_SCHEME}://${host_label}.${DOMAIN}/api/v1"

  echo "==> Building ${name} (API via ${api_url})..."
  cat >"$env_file" <<EOF
NEXT_PUBLIC_API_URL=${api_url}
EOF

  if [[ "$name" == "website" ]]; then
    cat >>"$env_file" <<EOF
NEXT_PUBLIC_CUSTOMER_PORTAL_URL=${CUSTOMER_URL}
NEXT_PUBLIC_EMPLOYEE_PORTAL_URL=${EMPLOYEE_URL}
EOF
  fi

  (cd "$dir" && npm install && npm run build)
  pm2 delete "$name" 2>/dev/null || true
  pm2 start npm --name "$name" --cwd "$dir" -- start -- -p "$port" -H "$BIND_HOST"
}

build_app website web/website 3000 www
build_app admin web/admin 3001 admin
build_app customer web/customer 3002 app
build_app employee web/employee 3003 provider

pm2 save
echo "Frontends public via Nginx; API private on 127.0.0.1:8080 (proxied as /api/)."
