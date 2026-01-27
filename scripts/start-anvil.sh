#!/bin/bash
# ==============================================================================
# start-anvil.sh - Local Ethereum Blockchain Launcher
# ==============================================================================
#
# PURPOSE:
# Starts Anvil, a fast local Ethereum node from the Foundry toolkit. This
# provides the blockchain environment where smart contracts are deployed
# and where all trading simulation occurs.
#
# ==============================================================================
# WHAT IS ANVIL?
# ==============================================================================
#
# Anvil is NOT geth (go-ethereum). It's a lightweight, fast Ethereum simulator
# designed for local development. Key differences from geth:
#
# - In-memory state (resets on restart)
# - Instant block mining (no waiting for confirmations)
# - Deterministic accounts (same addresses every time)
# - Fast execution (no real proof-of-work)
#
# ==============================================================================
# SYSTEM STARTUP SEQUENCE
# ==============================================================================
#
#   1. start-anvil.sh (THIS)   Creates blockchain + 30 funded accounts
#           |
#           | Anvil listens on localhost:8545
#           v
#   2. deploy.sh               Deploys AppleToken + AppleAMM contracts
#           |
#           | Contracts deployed, pool seeded
#           v
#   3. Backend connects        ethclient.Dial("http://localhost:8545")
#           |
#           | RPC connection established
#           v
#   4. Bots start trading      Send transactions to Anvil
#           |
#           | Anvil mines blocks instantly
#           v
#   5. Frontend reads state    Via backend WebSocket + REST API
#
# ==============================================================================
# ANVIL CONFIGURATION
# ==============================================================================
#
# --accounts 30
#   Creates 30 accounts from a deterministic mnemonic:
#   "test test test test test test test test test test test junk"
#
#   These addresses are ALWAYS the same on every machine:
#   Account 0: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 (LP/Deployer)
#   Account 1: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 (Whale1)
#   Account 2: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC (Whale2)
#   ... etc (see backend/internal/config/accounts.go for full list)
#
# --balance 10000
#   Each account starts with 10,000 ETH
#   This is enough for trading simulation but can be increased
#
# ==============================================================================
# NETWORK DETAILS
# ==============================================================================
#
# RPC URL:     http://localhost:8545
# WebSocket:   ws://localhost:8545
# Chain ID:    31337 (Anvil's default)
#
# The backend connects to these endpoints via go-ethereum's ethclient:
#   client, _ := ethclient.Dial("http://localhost:8545")
#
# ==============================================================================
# IMPORTANT: EPHEMERAL STATE
# ==============================================================================
#
# Anvil stores ALL state in memory. This means:
#
# - Restarting Anvil = ALL contracts and state are GONE
# - You must redeploy contracts after every Anvil restart
# - Account balances reset to initial values
# - Pool reserves reset (no more price drift)
#
# This is actually useful for testing - you get a clean slate every time.
#
# To preserve state between restarts, you could use Anvil's --state flag,
# but for this simulation, fresh state is typically preferred.
#
# ==============================================================================
# HOW THE BACKEND FINDS ANVIL
# ==============================================================================
#
# The backend's default RPC URL is configured in:
#   backend/internal/config/config.go
#
#   func DefaultConfig() *Config {
#       return &Config{
#           RPCURL:  "http://localhost:8545",  <- Anvil's default
#           ChainID: big.NewInt(31337),        <- Anvil's chain ID
#       }
#   }
#
# The backend then:
# 1. Dials Anvil: ethclient.Dial(config.RPCURL)
# 2. Verifies chain ID matches
# 3. Loads contract addresses from .env
# 4. Starts sending transactions
#
# ==============================================================================

echo "Starting Anvil local Ethereum node..."
echo "Accounts: 30"
echo "Balance per account: 10,000 ETH"
echo ""

# ==============================================================================
# Start Anvil
# ==============================================================================
#
# This command runs in the foreground (blocks the terminal).
# Open a new terminal for deploy.sh and running the backend.
#
# To stop Anvil: Ctrl+C in this terminal, or `pkill anvil` from another terminal
#
# Anvil output shows:
# - All 30 account addresses and private keys
# - Listening port (8545)
# - Chain ID (31337)
# - Block mining logs as transactions come in

anvil --accounts 30 --balance 10000

# ==============================================================================
# TROUBLESHOOTING
# ==============================================================================
#
# "Address already in use" error:
#   Another Anvil or process is using port 8545
#   Fix: pkill anvil  (or: lsof -i :8545 | kill)
#
# "Command not found: anvil"
#   Foundry is not installed
#   Fix: curl -L https://foundry.paradigm.xyz | bash && foundryup
#
# Backend can't connect:
#   - Make sure Anvil is running (this script)
#   - Check backend RPCURL config matches localhost:8545
#   - Ensure contracts are deployed (./scripts/deploy.sh)
#
# ==============================================================================
