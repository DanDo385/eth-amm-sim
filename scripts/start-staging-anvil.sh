#!/usr/bin/env bash
# Durable staging Anvil for api-staging-eth-amm-sim (MBP tunnel origin).
# LaunchAgent KeepAlive runs this in the foreground. On each start:
#   1) start Anvil on 127.0.0.1:11545
#   2) deploy contracts once it is ready
#   3) wait on Anvil (exit tears down the chain)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="${HOME}/.foundry/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:${PATH}"
export ETH_AMM_SIM_RPC_URL="${ETH_AMM_SIM_RPC_URL:-http://127.0.0.1:11545}"
ANVIL_HOST="${ANVIL_HOST:-127.0.0.1}"
ANVIL_PORT="${ANVIL_PORT:-11545}"

anvil --host "$ANVIL_HOST" --port "$ANVIL_PORT" --accounts 30 --balance 30000 &
ANVIL_PID=$!

cleanup() {
  kill "$ANVIL_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# Wait for RPC
for _ in $(seq 1 60); do
  if cast block-number --rpc-url "$ETH_AMM_SIM_RPC_URL" >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

if ! cast block-number --rpc-url "$ETH_AMM_SIM_RPC_URL" >/dev/null 2>&1; then
  echo "error: Anvil did not become ready on ${ANVIL_HOST}:${ANVIL_PORT}" >&2
  exit 1
fi

echo "Anvil ready on ${ANVIL_HOST}:${ANVIL_PORT}; deploying contracts..."
"$ROOT/scripts/deploy.sh"

echo "Deploy complete; Anvil pid=${ANVIL_PID} holding foreground."
wait "$ANVIL_PID"
