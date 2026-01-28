#!/bin/bash
# ==============================================================================
# post-deploy.sh - Post-Deployment Balance Adjustment
# ==============================================================================
#
# PURPOSE:
# Adjusts leverage account balances after deployment to ensure realistic
# liquidation risk. Anvil starts all accounts with 15,000 ETH, but leverage
# accounts should have constrained capital (130, 80, 60 ETH respectively).
#
# This script:
# 1. Transfers excess ETH from leverage accounts (25-27) to lending pool (account 7)
# 2. Deposits lending pool ETH into the AMM contract for leverage borrowing
#
# ==============================================================================
# ACCOUNT ALLOCATIONS
# ==============================================================================
#
# Account 7:  Lending Pool (receives excess ETH from leverage accounts and user)
# Account 19: User (keeps 1,000 ETH, sends excess to account 7)
# Account 25: Lev5x  (keeps 130 ETH, sends excess to account 7)
# Account 26: Lev10x (keeps 80 ETH, sends excess to account 7)
# Account 27: Lev25x (keeps 60 ETH, sends excess to account 7)
#
# ==============================================================================

set -e # Exit immediately if any command fails

echo "Adjusting leverage account balances..."

# Navigate to project root
cd "$(dirname "$0")/.."

# ==============================================================================
# Get AMM contract address from broadcast JSON
# ==============================================================================

BROADCAST_DIR="contracts/broadcast/Deploy.s.sol/31337"
LATEST_BROADCAST=""

# Check for run-latest.json first
if [ -f "$BROADCAST_DIR/run-latest.json" ]; then
    LATEST_BROADCAST="$BROADCAST_DIR/run-latest.json"
else
    # Find the most recent run-*.json file
    LATEST_BROADCAST=$(ls -t "$BROADCAST_DIR"/run-*.json 2>/dev/null | head -1)
fi

if [ -z "$LATEST_BROADCAST" ] || [ ! -f "$LATEST_BROADCAST" ]; then
    echo "Error: Could not find broadcast JSON file"
    echo "Make sure deploy.sh has been run first"
    exit 1
fi

# Extract AMM address using jq if available, otherwise use grep
if command -v jq &> /dev/null; then
    AMM_ADDR=$(jq -r '.transactions[] | select(.contractName=="AppleAMM") | .contractAddress' "$LATEST_BROADCAST" | head -1)
else
    # Fallback: use grep to find AMM address (less reliable)
    AMM_ADDR=$(grep -o '"contractName":"AppleAMM"' -A 5 "$LATEST_BROADCAST" | grep -o '"contractAddress":"[^"]*"' | head -1 | cut -d'"' -f4)
fi

if [ -z "$AMM_ADDR" ] || [ "$AMM_ADDR" == "null" ]; then
    echo "Error: Could not extract AMM contract address from broadcast file"
    exit 1
fi

echo "AMM Contract Address: $AMM_ADDR"
echo ""

# ==============================================================================
# Anvil account addresses and private keys
# ==============================================================================

# Account 7 (Lending Pool)
LENDING_POOL_ADDR="0x14dC79964da2C08b23698B3D3cc7Ca32193d9955"
LENDING_POOL_KEY="0x4bbbf85ce3377467afe5d46f804f221813b2bb87f24d81f60f1fcdbf7cbf4356"

# Account 25 (Lev5x) - should keep 130 ETH
LEV5X_ADDR="0xDf37F81dAAD2b0327A0A50003740e1C935C70913"
LEV5X_KEY="0x18743e59419b01d1d846d97ea070b5a3368a3e7f6f0242cf497e1baac6972427"
LEV5X_TARGET=130

# Account 26 (Lev10x) - should keep 80 ETH
LEV10X_ADDR="0x553BC17A05702530097c3677091C5BB47a3a7931"
LEV10X_KEY="0xe383b226df7c8282489889170b0f68f66af6459261f4833a781acd0804fafe7a"
LEV10X_TARGET=80

# Account 27 (Lev25x) - should keep 60 ETH
LEV25X_ADDR="0x87BdCE72c06C21cd96219BD8521bDF1F42C78b5e"
LEV25X_KEY="0xf3a6b71b94f5cd909fb2dbb287da47badaa6d8bcdc45d595e2884835d8749001"
LEV25X_TARGET=60

# Account 19 (User) - should keep 1,000 ETH (like other non-leverage accounts)
USER_ADDR="0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199"
USER_KEY="0xdf57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e"
USER_TARGET=1000

RPC_URL="http://localhost:8545"

# ==============================================================================
# Helper function to get ETH balance (in wei, convert to ETH)
# ==============================================================================

get_balance_eth() {
    local addr=$1
    # cast balance returns wei, we convert to ETH by dividing by 1e18
    # For simplicity, we'll work in wei and convert at the end
    cast balance "$addr" --rpc-url "$RPC_URL" 2>/dev/null || echo "0"
}

# ==============================================================================
# Transfer excess ETH from leverage accounts to lending pool
# ==============================================================================

transfer_excess() {
    local account_addr=$1
    local account_key=$2
    local target_eth=$3
    local account_name=$4
    
    # Get current balance (in wei)
    current_balance_wei=$(get_balance_eth "$account_addr")
    target_balance_wei=$(echo "$target_eth * 1000000000000000000" | bc)
    
    # Calculate excess (in wei)
    excess_wei=$(echo "$current_balance_wei - $target_balance_wei" | bc)
    
    # Only transfer if there's excess (and it's significant, > 0.1 ETH to account for gas)
    min_excess=$(echo "0.1 * 1000000000000000000" | bc)
    if [ "$(echo "$excess_wei > $min_excess" | bc)" -eq 1 ]; then
        # Convert to hex for cast send
        excess_hex=$(printf "0x%x" "$excess_wei")
        echo "Transferring excess ETH from $account_name to lending pool..."
        echo "  Current: $(echo "scale=2; $current_balance_wei / 1000000000000000000" | bc) ETH"
        echo "  Target:  $target_eth ETH"
        echo "  Excess:  $(echo "scale=2; $excess_wei / 1000000000000000000" | bc) ETH"
        
        if cast send "$LENDING_POOL_ADDR" --value "$excess_hex" \
            --private-key "$account_key" \
            --rpc-url "$RPC_URL" > /dev/null 2>&1; then
            echo "  ✓ Transfer complete"
        else
            echo "  ⚠ Warning: Transfer failed (account may already have correct balance or insufficient gas)"
        fi
        
    else
        echo "$account_name already has correct balance (~$target_eth ETH)"
        echo ""
    fi
}

# Check if bc is available for calculations
if ! command -v bc &> /dev/null; then
    echo "Error: 'bc' command not found. Install it with: brew install bc"
    exit 1
fi

echo "Step 1: Transferring excess ETH from leverage accounts to lending pool"
echo "========================================================================"
echo ""

transfer_excess "$LEV5X_ADDR" "$LEV5X_KEY" "$LEV5X_TARGET" "Lev5x (Account 25)"
transfer_excess "$LEV10X_ADDR" "$LEV10X_KEY" "$LEV10X_TARGET" "Lev10x (Account 26)"
transfer_excess "$LEV25X_ADDR" "$LEV25X_KEY" "$LEV25X_TARGET" "Lev25x (Account 27)"
transfer_excess "$USER_ADDR" "$USER_KEY" "$USER_TARGET" "User (Account 19)"

# ==============================================================================
# Deposit lending pool ETH into AMM contract
# ==============================================================================

echo "Step 2: Depositing lending pool ETH into AMM contract"
echo "========================================================================"
echo ""

lending_pool_balance_wei=$(get_balance_eth "$LENDING_POOL_ADDR")
lending_pool_balance_eth=$(echo "scale=2; $lending_pool_balance_wei / 1000000000000000000" | bc)

if [ "$(echo "$lending_pool_balance_eth > 0.1" | bc)" -eq 1 ]; then
    echo "Lending pool balance: $lending_pool_balance_eth ETH"
    echo "Depositing into AMM contract..."
    
    # Send ETH to AMM contract (it has a receive() function)
    # We keep 0.1 ETH for gas
    deposit_amount_wei=$(echo "$lending_pool_balance_wei - 100000000000000000" | bc)
    deposit_amount_hex=$(printf "0x%x" "$deposit_amount_wei")
    
    if cast send "$AMM_ADDR" --value "$deposit_amount_hex" \
        --private-key "$LENDING_POOL_KEY" \
        --rpc-url "$RPC_URL" > /dev/null 2>&1; then
        echo "  ✓ Deposited $(echo "scale=2; $deposit_amount_wei / 1000000000000000000" | bc) ETH into AMM contract"
    else
        echo "  ⚠ Warning: Deposit failed (may already be deposited or insufficient gas)"
    fi
    
    echo ""
else
    echo "Lending pool has insufficient balance ($lending_pool_balance_eth ETH)"
    echo "Skipping deposit (leverage accounts may already be constrained)"
    echo ""
fi

# ==============================================================================
# Verify final balances
# ==============================================================================

echo "Step 3: Verifying final balances"
echo "========================================================================"
echo ""

echo "Leverage Accounts:"
lev5x_final=$(echo "scale=2; $(get_balance_eth "$LEV5X_ADDR") / 1000000000000000000" | bc)
lev10x_final=$(echo "scale=2; $(get_balance_eth "$LEV10X_ADDR") / 1000000000000000000" | bc)
lev25x_final=$(echo "scale=2; $(get_balance_eth "$LEV25X_ADDR") / 1000000000000000000" | bc)

echo "  Lev5x  (25): $lev5x_final ETH (target: $LEV5X_TARGET ETH)"
echo "  Lev10x (26): $lev10x_final ETH (target: $LEV10X_TARGET ETH)"
echo "  Lev25x (27): $lev25x_final ETH (target: $LEV25X_TARGET ETH)"
echo ""

user_final=$(echo "scale=2; $(get_balance_eth "$USER_ADDR") / 1000000000000000000" | bc)
echo "User Account:"
echo "  User (19): $user_final ETH (target: $USER_TARGET ETH)"
echo ""

lending_pool_final=$(echo "scale=2; $(get_balance_eth "$LENDING_POOL_ADDR") / 1000000000000000000" | bc)
amm_balance=$(echo "scale=2; $(get_balance_eth "$AMM_ADDR") / 1000000000000000000" | bc)

echo "Lending Pool:"
echo "  Account 7 balance: $lending_pool_final ETH (kept for gas)"
echo "  AMM contract balance: $amm_balance ETH (available for leverage loans)"
echo ""

echo "✓ Post-deployment balance adjustment complete!"
echo ""
echo "Leverage accounts now have constrained capital for realistic liquidation risk."
