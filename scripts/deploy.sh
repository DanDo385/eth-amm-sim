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
#           |
#           v
#   3. Backend reads           TOKEN_ADDRESS and AMM_ADDRESS from .env
#           |                  or uses addresses from broadcast/ output
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
#
# 2. MINT TOKENS to all Anvil accounts based on their role:
#    - Account 0 (LP):      2000 APPL (1000 for pool, 1000 reserve)
#    - Accounts 1-3 (Whales): 0-1000 APPL (varies by whale)
#    - Accounts 4-9 (Bots):   100 APPL each
#    - Accounts 10-24 (Retail): 20 APPL each
#    - Accounts 25-28:        0 APPL (use ETH only)
#
# 3. SEED LIQUIDITY POOL
#    - LP (account 0) deposits 1000 APPL + 1000 ETH
#    - This creates the initial trading pool with 1:1 price ratio
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
# The backend reads these addresses from environment variables:
#   TOKEN_ADDRESS=0x5FbDB...
#   AMM_ADDRESS=0xe7f17...
#
# Or uses default addresses if not set (see backend/cmd/simulator/main.go)
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

forge script script/Deploy.s.sol --rpc-url http://localhost:8545 --broadcast

echo ""
echo "Deployment complete!"
echo "Contract addresses are saved in contracts/broadcast/"

# ==============================================================================
# NEXT STEPS
# ==============================================================================
#
# After successful deployment:
#
# 1. Note the contract addresses from the output (or check broadcast/run-latest.json)
#
# 2. Update .env if needed:
#    TOKEN_ADDRESS=<AppleToken address>
#    AMM_ADDRESS=<AppleAMM address>
#
# 3. Start the simulator:
#    cd backend && go run cmd/simulator/main.go
#    Or: make run
#
# 4. Open the frontend to visualize trades:
#    cd frontend && npm run dev
#
# ==============================================================================
