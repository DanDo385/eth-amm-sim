#!/bin/bash
# ==============================================================================
# deploy.sh - Contract Deployment to Local Blockchain
# ==============================================================================
#
# PURPOSE:
# Deploys the compiled smart contracts to the running Anvil blockchain and
# sets up the initial simulation state (mints tokens, seeds liquidity pool).
#
# PREREQUISITES:
# 1. Anvil must be running: ./scripts/start-anvil.sh
# 2. Contracts should be compiled (forge build runs automatically in Deploy.s.sol)
#
# ==============================================================================
# HOW THIS SCRIPT FITS IN THE SYSTEM
# ==============================================================================
#
#   1. start-anvil.sh          Creates local blockchain with funded accounts
#           |
#           v
#   2. deploy.sh (THIS)        Deploys contracts, mints tokens, seeds pool
#           |                  Addresses saved to broadcast/ JSON files
#           v
#   3. Backend reads           Contract addresses from broadcast/ JSON
#           |                  (or env vars, or defaults)
#           v
#   4. Simulator runs          Bots trade against the deployed contracts
#
# ==============================================================================
# WHAT THE DEPLOY SCRIPT DOES
# ==============================================================================
#
# The Forge script (contracts/script/Deploy.s.sol) performs these actions:
#
# 1. DEPLOY CONTRACTS
#    - AppleToken: ERC20 token contract (the "APPL" token)
#    - AppleAMM: The AMM contract for swapping ETH <-> APPL
#      * Uses ReentrancyGuard and Address.sendValue() for security
#
# 2. REDISTRIBUTE ETH via vm.deal:
#    - Account 0 (LP):     15,000 ETH (10k liquidity + 5k for gas)
#    - Accounts 1-29:      1,000 ETH each
#
# 3. MINT TOKENS to all Anvil accounts:
#    - Account 0 (LP):         10,000 APPL (for pool seeding)
#    - Accounts 1-29:          1,000 APPL each
#
# 4. SEED LIQUIDITY POOL
#    - LP (account 0) deposits 10,000 APPL + 10,000 ETH
#    - This creates the initial trading pool with 1:1 price ratio
#
# NOTE:
# Deploy.s.sol is the authoritative on-chain pool state.
# The backend also reads the actual on-chain reserves on startup to initialize LP metrics,
# so the UI should reflect 10,000 APPL + 10,000 ETH immediately after deploy.
#
# ==============================================================================
# OUTPUT FILES
# ==============================================================================
#
# After deployment, contract addresses are saved to:
#   contracts/broadcast/Deploy.s.sol/31337/run-latest.json
#
# This JSON contains:
# - AppleToken address (e.g., 0x5FbDB2315678afecb367f032d93F642f64180aa3)
# - AppleAMM address (e.g., 0xe7f1725E7734CE288F8367e1bb143E90bb3F0512)
#
# The backend automatically reads these addresses from the broadcast JSON file.
# No .env file needed - addresses are extracted directly from deployment output.
# See backend/internal/config/config.go for details
#
# ==============================================================================
# REDEPLOYMENT SCENARIOS
# ==============================================================================
#
# When do you need to redeploy?
#
# 1. CONTRACT CODE CHANGED
#    - You modified AppleAMM.sol or AppleToken.sol
#    - First regenerate bindings: ./scripts/generate-bindings.sh
#    - Then redeploy: ./scripts/deploy.sh
#
# 2. WANT FRESH STATE
#    - Pool reserves have drifted from trades
#    - Want to reset all account balances
#    - Restart Anvil first: pkill anvil && ./scripts/start-anvil.sh
#    - Then redeploy: ./scripts/deploy.sh
#
# 3. ANVIL RESTARTED
#    - Anvil state is in-memory only
#    - Restarting Anvil wipes all deployed contracts
#    - Must redeploy after every Anvil restart
#
# ==============================================================================

set -e # Exit immediately if any command fails

echo "Deploying contracts to Anvil..."

# Navigate to contracts directory where foundry.toml lives
cd "$(dirname "$0")/../contracts"

# ==============================================================================
# Run the Forge deployment script
# ==============================================================================
#
# forge script: Foundry's script runner for deployment/interaction scripts
#
# Arguments:
#   script/Deploy.s.sol   : The Solidity script to run
#   --rpc-url             : Where to send transactions (Anvil's default port)
#   --broadcast           : Actually send transactions (vs dry-run)
#
# The script uses account 0's private key (from Anvil's deterministic mnemonic)
# to sign all transactions. Account 0 becomes the contract deployer/owner.

# --slow: wait for each tx to confirm before sending the next. Avoids Anvil hanging on
# batched broadcasts (e.g. "25/34 receipts" forever) on some macOS / Foundry combos.
# --non-interactive: safe for tmux/CI (no size-limit prompts).
forge script script/Deploy.s.sol \
  --rpc-url http://127.0.0.1:8545 \
  --with-gas-price 2000000000 \
  --broadcast \
  --slow \
  --non-interactive

echo ""
echo "Deployment complete!"
echo "Contract addresses are saved in contracts/broadcast/"

# ==============================================================================
# Cleanup Foundry cache outputs (keep only run-latest.json)
# ==============================================================================

CACHE_DIR="cache/Deploy.s.sol/31337"
if [ -d "$CACHE_DIR" ]; then
    LATEST_CACHE=""
    for f in "$CACHE_DIR"/run-*.json; do
        # Skip if glob didn't match
        [ -e "$f" ] || continue
        if [ "$(basename "$f")" = "run-latest.json" ]; then
            continue
        fi
        if [ -z "$LATEST_CACHE" ] || [ "$f" -nt "$LATEST_CACHE" ]; then
            LATEST_CACHE="$f"
        fi
    done

    if [ -n "$LATEST_CACHE" ]; then
        cp "$LATEST_CACHE" "$CACHE_DIR/run-latest.json"
    fi

for f in "$CACHE_DIR"/run-*.json; do
        [ -e "$f" ] || continue
        if [ "$(basename "$f")" != "run-latest.json" ]; then
            rm -f "$f"
        fi
    done
fi

# ==============================================================================
# Cleanup Foundry broadcast outputs (keep only run-latest.json)
# ==============================================================================

BROADCAST_DIR="broadcast/Deploy.s.sol/31337"
if [ -d "$BROADCAST_DIR" ]; then
    LATEST_BROADCAST=""
    for f in "$BROADCAST_DIR"/run-*.json; do
        [ -e "$f" ] || continue
        if [ "$(basename "$f")" = "run-latest.json" ]; then
            continue
        fi
        if [ -z "$LATEST_BROADCAST" ] || [ "$f" -nt "$LATEST_BROADCAST" ]; then
            LATEST_BROADCAST="$f"
        fi
    done

    if [ -n "$LATEST_BROADCAST" ]; then
        cp "$LATEST_BROADCAST" "$BROADCAST_DIR/run-latest.json"
    fi

    for f in "$BROADCAST_DIR"/*; do
        [ -e "$f" ] || continue
        base="$(basename "$f")"
        if [ "$base" != "run-latest.json" ] && [ "$base" != ".gitignore" ] && [ "$base" != ".gitkeep" ] && [ "$base" != "README.md" ]; then
            rm -f "$f"
        fi
    done
fi

# ==============================================================================
# Show deployed contract addresses
# ==============================================================================

# Find the latest broadcast file
LATEST_BROADCAST=""

# Check for run-latest.json first (symlink to most recent)
if [ -f "$BROADCAST_DIR/run-latest.json" ]; then
    LATEST_BROADCAST="$BROADCAST_DIR/run-latest.json"
else
    # Find the most recent run-*.json file
    LATEST_BROADCAST=$(ls -t "$BROADCAST_DIR"/run-*.json 2>/dev/null | head -1)
fi

if [ -n "$LATEST_BROADCAST" ] && [ -f "$LATEST_BROADCAST" ]; then
    echo ""
    echo "Deployed contract addresses:"
    
    # Check if jq is available for pretty output
    if command -v jq &> /dev/null; then
        TOKEN_ADDR=$(jq -r '.transactions[] | select(.contractName=="AppleToken") | .contractAddress' "$LATEST_BROADCAST" | head -1)
        AMM_ADDR=$(jq -r '.transactions[] | select(.contractName=="AppleAMM") | .contractAddress' "$LATEST_BROADCAST" | head -1)
        
        if [ -n "$TOKEN_ADDR" ] && [ -n "$AMM_ADDR" ] && [ "$TOKEN_ADDR" != "null" ] && [ "$AMM_ADDR" != "null" ]; then
            echo "  AppleToken: $TOKEN_ADDR"
            echo "  AppleAMM:   $AMM_ADDR"
            echo ""
            echo "✓ Backend will automatically read these addresses from broadcast output"
        fi
    else
        echo "  (Install jq for address extraction: brew install jq)"
        echo "  Addresses saved in: $LATEST_BROADCAST"
    fi
fi

# ==============================================================================
# NEXT STEPS
# ==============================================================================
#
# After successful deployment:
#
# 1. Contract addresses are saved in broadcast/ JSON files
#    The backend automatically reads them - no .env file needed!
#
# 2. Start the simulator:
#    cd backend && go run cmd/simulator/main.go
#    Or: make backend
#
# 3. Open the frontend to visualize trades:
#    cd frontend && npm run dev
#    Or: make frontend
#
# ==============================================================================
