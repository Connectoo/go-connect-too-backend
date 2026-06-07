#!/usr/bin/env bash
# SSH tunnel to Postgres on the VM — run on your LAPTOP, not on the server.
# Postgres must NEVER be opened to the public internet.
#
# Usage:
#   VM_HOST=4.240.109.134 SSH_KEY=~/Downloads/go-connect.pem bash deploy/azure/postgres-tunnel.sh
# Then connect: postgres://app:PASSWORD@127.0.0.1:15432/go_connect?sslmode=disable
set -euo pipefail

VM_HOST="${VM_HOST:?set VM_HOST to the VM public IP}"
SSH_KEY="${SSH_KEY:-$HOME/Downloads/go-connect.pem}"
LOCAL_PORT="${LOCAL_PORT:-15432}"
REMOTE_PORT="${REMOTE_PORT:-5432}"

if [[ ! -f "$SSH_KEY" ]]; then
  echo "SSH key not found: $SSH_KEY"
  exit 1
fi

chmod 400 "$SSH_KEY" 2>/dev/null || true

echo "Tunnel: 127.0.0.1:${LOCAL_PORT} → ${VM_HOST}:${REMOTE_PORT}"
echo "Connect (TablePlus, psql, DBeaver):"
echo "  Host: 127.0.0.1  Port: ${LOCAL_PORT}  User: app  DB: go_connect"
echo "Press Ctrl+C to close the tunnel."
echo ""

exec ssh -i "$SSH_KEY" -o ExitOnForwardFailure=yes \
  -L "127.0.0.1:${LOCAL_PORT}:127.0.0.1:${REMOTE_PORT}" \
  "azureuser@${VM_HOST}" -N
