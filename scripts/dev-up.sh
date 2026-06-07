#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

PID_DIR="$ROOT_DIR/.dev/pids"
LOG_DIR="$ROOT_DIR/.dev/logs"
DATABASE_URL="${DATABASE_URL:-postgres://app:app@localhost:5433/go_connect?sslmode=disable}"
API_URL="${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}"

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

stop_dev_processes() {
  for name in api website admin customer employee; do
    kill_pid_file "$name"
  done
}

cleanup() {
  echo ""
  echo "Stopping dev stack..."
  stop_dev_processes
  exit 0
}

trap cleanup SIGINT SIGTERM

if [[ ! -f .env ]]; then
  echo "Creating .env from .env.example"
  cp .env.example .env
fi

# Do not `source .env` — values like FCM_CREDENTIALS_JSON break bash parsing.
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

echo "Starting API on :8080..."
make run >"$LOG_DIR/api.log" 2>&1 &
echo $! >"$PID_DIR/api.pid"

echo "Waiting for API health check..."
for _ in $(seq 1 60); do
  if curl -sf "http://localhost:8080/api/v1/health" >/dev/null; then
    break
  fi
  sleep 1
done

if ! curl -sf "http://localhost:8080/api/v1/health" >/dev/null; then
  echo "API failed to start. See $LOG_DIR/api.log"
  tail -n 20 "$LOG_DIR/api.log" || true
  cleanup
fi

start_web_app() {
  local name="$1"
  local port="$2"
  local dir="$ROOT_DIR/web/$name"

  if [[ ! -d "$dir/node_modules" ]]; then
    echo "Installing npm dependencies for $name..."
    (cd "$dir" && npm install)
  fi

  echo "Starting $name on :$port..."
  (
    cd "$dir"
    export NEXT_PUBLIC_API_URL="$API_URL"
    npm run dev
  ) >"$LOG_DIR/$name.log" 2>&1 &
  echo $! >"$PID_DIR/$name.pid"
}

start_web_app website 3000
start_web_app admin 3001
start_web_app customer 3002
start_web_app employee 3003

cat <<EOF

Dev stack is running:

  API:       http://localhost:8080/api/v1  (docs: /api/v1/docs/)
  Website:   http://localhost:3000
  Admin:     http://localhost:3001
  Customer:  http://localhost:3002
  Employee:  http://localhost:3003

Logs: $LOG_DIR/
Stop: make dev-down  (or Ctrl+C in this terminal)

EOF

wait
