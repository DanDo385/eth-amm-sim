#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
NVMRC_PATH="$FRONTEND_DIR/.nvmrc"

if [ "$#" -eq 0 ]; then
  echo "[frontend] Usage: scripts/with-frontend-node.sh <command...>"
  exit 1
fi

node_major() {
  node -p "process.versions.node.split('.')[0]" 2>/dev/null || true
}

is_supported_node() {
  local major="$1"
  [[ "$major" =~ ^[0-9]+$ ]] && [ "$major" -ge 20 ] && [ "$major" -lt 25 ]
}

run_in_frontend() {
  cd "$FRONTEND_DIR"
  "$@"
}

major="$(node_major)"
if is_supported_node "$major"; then
  run_in_frontend "$@"
  exit 0
fi

if command -v fnm >/dev/null 2>&1; then
  cd "$FRONTEND_DIR"
  exec fnm exec --using-file "$NVMRC_PATH" -- "$@"
fi

export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
if [ -s "$NVM_DIR/nvm.sh" ]; then
  # shellcheck disable=SC1090
  . "$NVM_DIR/nvm.sh"
  cd "$FRONTEND_DIR"
  nvm install >/dev/null
  nvm use >/dev/null
  exec "$@"
fi

current_version="$(node --version 2>/dev/null || echo "not found")"
echo ""
echo "[frontend] Unsupported Node.js version: $current_version"
echo "[frontend] This project requires Node 20.x or 22.x LTS."
echo "[frontend] Install fnm or nvm to auto-switch from frontend/.nvmrc, then retry."
echo "[frontend] Manual fix: nvm use 20 && cd frontend && npm install"
echo ""
exit 1
