#!/usr/bin/env bash
# [Optional scale-out] Separate private VM for Postgres + Next.js.
# For a single VM in one VNet, use setup-vm-vpc.sh instead.
# Bootstrap the PRIVATE subnet VM: Postgres, MinIO, Next.js apps.
# No public IP on this VM. Run as root on the private VM.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo bash deploy/azure/setup-private-vm.sh)"
  exit 1
fi

echo "==> Installing system packages..."
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates curl git gettext-base ufw

if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable docker
systemctl start docker

if ! command -v node >/dev/null 2>&1 || [[ "$(node -v)" != v20* ]]; then
  curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
  DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs
fi

if ! command -v pm2 >/dev/null 2>&1; then
  npm install -g pm2
fi

if ! command -v migrate >/dev/null 2>&1; then
  curl -fsSL https://go.dev/dl/go1.26.3.linux-amd64.tar.gz -o /tmp/go.tar.gz
  tar -C /usr/local -xzf /tmp/go.tar.gz
  export PATH="/usr/local/go/bin:$PATH"
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ln -sf "$(go env GOPATH)/bin/migrate" /usr/local/bin/migrate
fi

# Private VM: SSH only from bastion/VPN (adjust source as needed)
ufw allow OpenSSH
ufw --force enable

if [[ ! -f .env ]]; then
  cp deploy/azure/env.vpc.example .env
  echo ""
  echo "Created .env — set POSTGRES_PASSWORD, MINIO_ROOT_PASSWORD, JWT secrets, PRIVATE_HOST, then re-run:"
  echo "  nano $ROOT_DIR/.env"
  echo "  bash deploy/azure/setup-private-vm.sh"
  exit 1
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

export VPC_MODE=true

if [[ "${JWT_ACCESS_SECRET}" == "change-me-access-secret-min-32-chars" ]]; then
  echo "Set JWT_ACCESS_SECRET and JWT_REFRESH_SECRET in .env (openssl rand -hex 32)"
  exit 1
fi

if [[ "${POSTGRES_PASSWORD}" == "change-me-strong-postgres-password" ]]; then
  echo "Set POSTGRES_PASSWORD in .env"
  exit 1
fi

if [[ -z "${PRIVATE_HOST:-}" || "${PRIVATE_HOST}" == "10.0.2.4" ]]; then
  detected="$(hostname -I | awk '{print $1}')"
  if [[ -n "$detected" ]]; then
    echo "==> PRIVATE_HOST not set — using detected private IP: ${detected}"
    export PRIVATE_HOST="$detected"
    grep -v '^PRIVATE_HOST=' .env > .env.tmp || true
    { cat .env.tmp; echo "PRIVATE_HOST=${PRIVATE_HOST}"; } > .env
    rm -f .env.tmp
    set -a
    # shellcheck disable=SC1091
    source .env
    set +a
  fi
fi

# shellcheck disable=SC1091
source deploy/azure/resolve-domain.sh
resolve_domain_config

echo "==> Starting Postgres + MinIO (private subnet only)..."
docker compose -f docker-compose.azure.private.yml up -d

echo "==> Waiting for Postgres..."
for _ in $(seq 1 30); do
  if docker exec go-connect-postgres pg_isready -U app -d go_connect >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "==> Running migrations..."
migrate -path migrations -database "${DATABASE_URL}" up

echo "==> Seeding demo data..."
export PATH="/usr/local/go/bin:$(go env GOPATH 2>/dev/null)/bin:$PATH"
go run ./cmd/seed

private_ip="${PRIVATE_HOST:-$(hostname -I | awk '{print $1}')}"

if [[ -n "${API_URL:-}" ]]; then
  echo "==> Building and starting Next.js apps (VPC-only, no public Nginx)..."
  bash deploy/azure/deploy-private.sh
  frontends_msg="Next.js apps are running (see URLs below)."
else
  frontends_msg="Next.js not built yet — after the public API VM is up, set API_URL in .env and run:
  bash deploy/azure/deploy-private.sh"
fi

cat <<EOF

Private VM setup complete.
${frontends_msg}

Private IP: ${private_ip}
Postgres:     ${private_ip}:5432  (VNet only — block in NSG from API subnet)
MinIO:        ${private_ip}:9000  (VNet only)
Next.js apps (reachable via VPN / VNet peering / Bastion tunnel):

  Website:  http://${private_ip}:3000
  Admin:    http://${private_ip}:3001/login
  Customer: http://${private_ip}:3002/login
  Employee: http://${private_ip}:3003/login

On the PUBLIC API VM, set PRIVATE_HOST=${private_ip} in .env, then run:
  bash deploy/azure/setup-api-vm.sh

Demo logins (password Demo123!):
  admin@yopmail.com, alice@yopmail.com, karim@yopmail.com

EOF
