#!/usr/bin/env bash
# Local VPC simulation — same as Azure VPC: frontends public via Nginx, API private on :8080.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

PID_DIR="$ROOT_DIR/.dev/pids"
LOG_DIR="$ROOT_DIR/.dev/logs"
NGINX_CONF="$ROOT_DIR/deploy/azure/nginx/go-connect-vpc-local.conf"
DATABASE_URL="${DATABASE_URL:-postgres://app:app@localhost:5433/go_connect?sslmode=disable}"
GATEWAY="http://localhost:8880"

mkdir -p "$PID_DIR" "$LOG_DIR"

kill_pid_file() {
  local name="$1"
  local pidfile="$PID_DIR/$name.pid"
  if [[ -f "$pidfile" ]]; then
    local pid
    pid="$(cat "$pidfile")"
    if kill -0 "$pid" 2>/dev/null; then
      pkill -P "$pid" 2>/dev/null || true
      kill "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  fi
}

stop_nginx() {
  if [[ -f "$PID_DIR/nginx.pid" ]]; then
    nginx -c "$NGINX_CONF" -p "$ROOT_DIR" -s stop 2>/dev/null || true
    rm -f "$PID_DIR/nginx.pid"
  fi
}

stop_dev_processes() {
  stop_nginx
  for name in api website admin customer employee; do
    kill_pid_file "$name"
  done
}

cleanup() {
  echo ""
  echo "Stopping VPC dev stack..."
  stop_dev_processes
  exit 0
}

trap cleanup SIGINT SIGTERM

if ! command -v nginx >/dev/null 2>&1; then
  echo "nginx is required for local VPC simulation."
  echo "  macOS: brew install nginx"
  echo "  Ubuntu: sudo apt install nginx"
  exit 1
fi

if [[ ! -f .env ]]; then
  echo "Creating .env from .env.example"
  cp .env.example .env
fi

if [[ -f .env ]]; then
  db_url="$(grep -E '^[[:space:]]*DATABASE_URL=' .env | grep -v '^[[:space:]]*#' | tail -1 | cut -d= -f2-)"
  if [[ -n "$db_url" ]]; then
    DATABASE_URL="$db_url"
  fi
fi
export DATABASE_URL

echo "Starting Docker services..."
docker compose up -d

echo "Waiting for PostgreSQL..."
for _ in $(seq 1 60); do
  if docker compose exec -T postgres pg_isready -U app -d go_connect >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "Running migrations..."
make migrate-up

stop_dev_processes

echo "Starting API on 127.0.0.1:8080 (private — not exposed directly)..."
make run >"$LOG_DIR/api.log" 2>&1 &
echo $! >"$PID_DIR/api.pid"

echo "Waiting for API..."
for _ in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:8080/api/v1/health" >/dev/null; then
    break
  fi
  sleep 1
done

if ! curl -sf "http://127.0.0.1:8080/api/v1/health" >/dev/null; then
  echo "API failed to start. See $LOG_DIR/api.log"
  tail -n 20 "$LOG_DIR/api.log" || true
  cleanup
fi

start_web_app() {
  local name="$1"
  local port="$2"
  local host_label="$3"
  local dir="$ROOT_DIR/web/$name"
  local api_url="http://${host_label}.localhost:8880/api/v1"

  if [[ ! -d "$dir/node_modules" ]]; then
    echo "Installing npm dependencies for $name..."
    (cd "$dir" && npm install)
  fi

  echo "Starting $name on :$port (API via $api_url)..."
  (
    cd "$dir"
    export NEXT_PUBLIC_API_URL="$api_url"
    if [[ "$name" == "website" ]]; then
      export NEXT_PUBLIC_CUSTOMER_PORTAL_URL="http://app.localhost:8880"
      export NEXT_PUBLIC_EMPLOYEE_PORTAL_URL="http://provider.localhost:8880"
    fi
    npm run dev -- -p "$port"
  ) >"$LOG_DIR/$name.log" 2>&1 &
  echo $! >"$PID_DIR/$name.pid"
}

start_web_app website 3000 www
start_web_app admin 3001 admin
start_web_app customer 3002 app
start_web_app employee 3003 provider

echo "Starting Nginx gateway on :8880..."
nginx -c "$NGINX_CONF" -p "$ROOT_DIR"

sleep 1

cat <<EOF

VPC dev stack running (mirrors Azure VPC layout):

  Gateway (use these in browser):
    Website:  http://www.localhost:8880
    Admin:    http://admin.localhost:8880/login
    Customer: http://app.localhost:8880/login
    Employee: http://provider.localhost:8880/login

  API via proxy (public path):
    curl http://www.localhost:8880/api/v1/health

  API direct (private — should work locally only):
    curl http://127.0.0.1:8080/api/v1/health

  Demo logins: admin@yopmail.com / alice@yopmail.com / karim@yopmail.com
  Password: Demo123!  (run: make seed  if not seeded yet)

Logs: $LOG_DIR/
Stop: make dev-vpc-down

EOF

wait
