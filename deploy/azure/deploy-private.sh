#!/usr/bin/env bash
# Build/restart Next.js apps bound to the VNet private IP (no public Nginx).
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

CUSTOMER_URL="${CUSTOMER_URL:-http://${PRIVATE_HOST}:3002}"
EMPLOYEE_URL="${EMPLOYEE_URL:-http://${PRIVATE_HOST}:3003}"
BIND_HOST="${BIND_HOST:-${PRIVATE_HOST}}"

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

  (cd "$dir" && npm install && npm run build)
  pm2 delete "$name" 2>/dev/null || true
  pm2 start npm --name "$name" --cwd "$dir" -- start -- -p "$port" -H "$BIND_HOST"
}

build_app website web/website 3000
build_app admin web/admin 3001
build_app customer web/customer 3002
build_app employee web/employee 3003

pm2 save
echo "Frontends running on ${BIND_HOST} (VPC-only)."
