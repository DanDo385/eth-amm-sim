#!/bin/bash
# ==============================================================================
# dev-up.sh - Launch Anvil + deploy + bindings + backend + frontend in tmux
# ==============================================================================
#
# Sequential execution with proper waiting:
# 1. Start anvil and wait for it to be ready
# 2. Run deploy (waits for anvil)
# 3. Run bindings (waits for anvil)
# 4. Run backend (waits for deploy to complete)
# 5. Run frontend (can start independently)
#
# All processes persist in the same tmux session with separate windows.
#
# ==============================================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SESSION_NAME="eth-amm-sim"

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required for make up."
  exit 1
fi

# Kill existing Anvil process if running
if lsof -ti:8545 >/dev/null 2>&1; then
  echo "Killing existing Anvil process on port 8545..."
  lsof -ti:8545 | xargs kill -9 2>/dev/null || true
  sleep 2
fi

# Kill existing session if it exists
if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
  echo "Killing existing tmux session: $SESSION_NAME"
  tmux kill-session -t "$SESSION_NAME" || true
fi

# Clean up old broadcast files to ensure fresh deployment
BROADCAST_DIR="$ROOT_DIR/contracts/broadcast/Deploy.s.sol/31337"
if [ -d "$BROADCAST_DIR" ]; then
  echo "Cleaning up old broadcast files for fresh deployment..."
  find "$BROADCAST_DIR" -name "run-*.json" ! -name "run-latest.json" -delete 2>/dev/null || true
  # Remove run-latest.json so deploy creates a fresh one
  rm -f "$BROADCAST_DIR/run-latest.json" 2>/dev/null || true
fi

# Wait for Anvil to be ready
wait_for_anvil() {
  echo "Waiting for Anvil to be ready on port 8545..."
  local max_attempts=30
  local attempt=0
  while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:8545 >/dev/null 2>&1; then
      echo "✓ Anvil is ready"
      return 0
    fi
    attempt=$((attempt + 1))
    sleep 1
  done
  echo "✗ Anvil failed to start after $max_attempts seconds"
  return 1
}

# Wait for deploy to complete
wait_for_deploy() {
  echo "Waiting for deployment to complete..."
  local deploy_file="$ROOT_DIR/contracts/broadcast/Deploy.s.sol/31337/run-latest.json"
  local max_attempts=180  # 3 minutes (deployment can take time)
  local attempt=0
  
  while [ $attempt -lt $max_attempts ]; do
    # Check for run-latest.json first (deploy script creates this directly)
    if [ -f "$deploy_file" ] && [ -s "$deploy_file" ]; then
      # Verify it contains contract addresses (not just empty or error)
      if grep -q "AppleToken\|AppleAMM\|transactions" "$deploy_file" 2>/dev/null; then
        # Give it a moment to ensure file is fully written
        sleep 1
        echo "✓ Deployment complete"
        return 0
      fi
    fi
    
    # Also check for any timestamped run-*.json files (forge creates these)
    local newest_file=$(find "$ROOT_DIR/contracts/broadcast/Deploy.s.sol/31337" -name "run-*.json" -type f 2>/dev/null | sort -r | head -1)
    if [ -n "$newest_file" ] && [ -f "$newest_file" ] && [ -s "$newest_file" ]; then
      # Found a broadcast file, check if it has contract data
      if grep -q "AppleToken\|AppleAMM\|transactions" "$newest_file" 2>/dev/null; then
        # Wait for deploy script to create run-latest.json
        sleep 2
        if [ -f "$deploy_file" ] && [ -s "$deploy_file" ]; then
          echo "✓ Deployment complete"
          return 0
        fi
      fi
    fi
    
    # Check for success message in deploy window output
    local deploy_output=$(tmux capture-pane -t "$SESSION_NAME:deploy" -p 2>/dev/null | tail -10)
    if echo "$deploy_output" | grep -q "ONCHAIN EXECUTION COMPLETE\|Deployment complete\|Transactions saved"; then
      # Deployment completed, wait for file to be written
      sleep 2
      if [ -f "$deploy_file" ] && [ -s "$deploy_file" ]; then
        echo "✓ Deployment complete"
        return 0
      fi
    fi
    
    # Check for error indicators
    if echo "$deploy_output" | grep -q "Error:\|Failed\|Insufficient funds\|make.*Error"; then
      echo "✗ Deployment failed - check deploy window for details"
      echo "  Last output:"
      echo "$deploy_output" | tail -5 | sed 's/^/    /'
      return 1
    fi
    
    attempt=$((attempt + 1))
    # Show progress every 15 seconds
    if [ $((attempt % 15)) -eq 0 ]; then
      echo "  Still waiting... (${attempt}s / ${max_attempts}s)"
    fi
    sleep 1
  done
  
  echo "✗ Deployment timed out after $max_attempts seconds"
  echo "  Check the deploy window (tmux window 'deploy') for errors"
  echo "  Last output from deploy window:"
  tmux capture-pane -t "$SESSION_NAME:deploy" -p 2>/dev/null | tail -10 | sed 's/^/    /'
  return 1
}

# Create tmux session with anvil window
echo "Creating tmux session: $SESSION_NAME"
tmux new-session -d -s "$SESSION_NAME" -n anvil -c "$ROOT_DIR"

# Window 1: Start Anvil
echo "Starting Anvil..."
tmux send-keys -t "$SESSION_NAME:anvil" "make anvil" C-m
tmux set-window-option -t "$SESSION_NAME:anvil" remain-on-exit on

# Wait for Anvil to be ready
if ! wait_for_anvil; then
  echo "Failed to start Anvil. Check the anvil window for errors."
  tmux attach -t "$SESSION_NAME"
  exit 1
fi

# Window 2: Deploy (waits for anvil, which is now ready)
echo "Starting deployment..."
tmux new-window -t "$SESSION_NAME" -n deploy -c "$ROOT_DIR"
tmux send-keys -t "$SESSION_NAME:deploy" "make deploy" C-m
tmux set-window-option -t "$SESSION_NAME:deploy" remain-on-exit on

# Wait for deployment to complete
if ! wait_for_deploy; then
  echo "Deployment failed or timed out. Check the deploy window for errors."
  tmux attach -t "$SESSION_NAME"
  exit 1
fi

# Window 3: Bindings (can run after anvil is ready)
echo "Generating bindings..."
tmux new-window -t "$SESSION_NAME" -n bindings -c "$ROOT_DIR"
tmux send-keys -t "$SESSION_NAME:bindings" "make bindings" C-m
tmux set-window-option -t "$SESSION_NAME:bindings" remain-on-exit on

# Give bindings a moment to complete (it's usually fast)
sleep 2

# Window 4: Backend (waits for deploy, which is now complete)
echo "Starting backend..."
tmux new-window -t "$SESSION_NAME" -n backend -c "$ROOT_DIR"
tmux send-keys -t "$SESSION_NAME:backend" "make backend" C-m
tmux set-window-option -t "$SESSION_NAME:backend" remain-on-exit on

# Window 5: Frontend (can start independently)
echo "Starting frontend..."
tmux new-window -t "$SESSION_NAME" -n frontend -c "$ROOT_DIR"
tmux send-keys -t "$SESSION_NAME:frontend" "make frontend" C-m
tmux set-window-option -t "$SESSION_NAME:frontend" remain-on-exit on

echo ""
echo "✓ All services started in tmux session: $SESSION_NAME"
echo "  Windows: anvil, deploy, bindings, backend, frontend"
echo ""

# Attach to session
if [ -n "${TMUX:-}" ]; then
  tmux switch-client -t "$SESSION_NAME"
else
  tmux attach -t "$SESSION_NAME"
fi
