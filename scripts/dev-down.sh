#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="$ROOT_DIR/.dev/pids"

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
    echo "Stopped $name"
  fi
}

if [[ -d "$PID_DIR" ]]; then
  for name in api website admin customer employee; do
    kill_pid_file "$name"
  done
fi

# Fallback: free known dev ports if processes were started outside dev-up.sh
for port in 8080 3000 3001 3002 3003; do
  pids="$(lsof -ti ":$port" 2>/dev/null || true)"
  if [[ -n "$pids" ]]; then
    echo "Stopping process(es) on port $port"
    # shellcheck disable=SC2086
    kill $pids 2>/dev/null || true
  fi
done

echo "Dev stack stopped. Docker services are still running (use: docker compose down)"
