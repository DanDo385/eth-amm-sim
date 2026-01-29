#!/bin/bash
# ==============================================================================
# dev-up.sh - Launch Anvil + deploy + bindings + backend + frontend in tmux
# ==============================================================================
#
# Uses tmux as the sole orchestrator (no Terminal.app, no backgrounding).
#
# ==============================================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SESSION_NAME="eth-amm-sim"

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required for make up."
  exit 1
fi

window_exists() {
  tmux list-windows -t "$SESSION_NAME" -F '#{window_name}' 2>/dev/null | grep -qx "$1"
}

ensure_window() {
  local name="$1"
  local cmd="$2"
  local remain_on_exit="${3:-off}"
  if ! window_exists "$name"; then
    tmux new-window -t "$SESSION_NAME" -n "$name" -c "$ROOT_DIR" "$cmd"
    if [ "$remain_on_exit" = "on" ]; then
      tmux set-window-option -t "$SESSION_NAME:$name" remain-on-exit on
    fi
  fi
}

if ! tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
  tmux new-session -d -s "$SESSION_NAME" -n anvil -c "$ROOT_DIR" "make anvil"
else
  ensure_window "anvil" "make anvil"
fi

ANVIL_READY_CMD="until curl -s http://localhost:8545 >/dev/null; do sleep 1; done;"
DEPLOY_READY_CMD="until [ -f \"$ROOT_DIR/contracts/broadcast/Deploy.s.sol/31337/run-latest.json\" ]; do sleep 1; done;"

ensure_window "deploy" "$ANVIL_READY_CMD make deploy" "on"
ensure_window "bindings" "$ANVIL_READY_CMD make bindings" "on"
ensure_window "backend" "$DEPLOY_READY_CMD make backend" "on"
ensure_window "frontend" "make frontend" "on"

if [ -n "${TMUX:-}" ]; then
  tmux switch-client -t "$SESSION_NAME"
else
  tmux attach -t "$SESSION_NAME"
fi
