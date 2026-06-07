#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

PID_DIR="$ROOT_DIR/.dev/pids"
NGINX_CONF="$ROOT_DIR/deploy/azure/nginx/go-connect-vpc-local.conf"

if [[ -f "$PID_DIR/nginx.pid" ]]; then
  nginx -c "$NGINX_CONF" -p "$ROOT_DIR" -s stop 2>/dev/null || true
  rm -f "$PID_DIR/nginx.pid"
fi

for name in api website admin customer employee; do
  pidfile="$PID_DIR/$name.pid"
  if [[ -f "$pidfile" ]]; then
    pid="$(cat "$pidfile")"
    if kill -0 "$pid" 2>/dev/null; then
      pkill -P "$pid" 2>/dev/null || true
      kill "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  fi
done

echo "VPC dev stack stopped."
