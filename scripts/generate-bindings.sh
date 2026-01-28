#!/bin/bash
# ==============================================================================
# generate-bindings.sh - Contract-to-Go Bridge Generator
# ==============================================================================
#
# PURPOSE:
# This script is the critical link between Solidity smart contracts and the Go
# backend. It compiles contracts and generates type-safe Go bindings that allow
# the backend to interact with on-chain contracts.
#
# WHEN TO RUN:
# Run this script whenever you modify a Solidity contract (.sol file). The
# changes won't be visible to the Go backend until you regenerate bindings.
#
# ==============================================================================
# ARCHITECTURE OVERVIEW
# ==============================================================================
#
#   contracts/src/*.sol          (Solidity source code)
#           |
#           | forge build
#           v
#   contracts/out/*.json         (Compiled ABI + bytecode)
#           |
#           | abigen (this script)
#           v
#   backend/internal/contracts/  (Go bindings - type-safe contract interface)
#           |
#           | imported by
#           v
#   backend/internal/engine/     (executor.go uses bindings to call contracts)
#
# ==============================================================================
# DIRECTORY STRUCTURE EXPLAINED
# ==============================================================================
#
# CONTRACTS_DIR: Where Solidity source lives
#   contracts/
#   ├── src/
#   │   ├── AppleToken.sol      <- ERC20 token contract
#   │   └── AppleAMM.sol        <- AMM swap/liquidity contract
#   ├── out/                    <- forge build output (ABI JSON files)
#   │   ├── AppleToken.sol/
#   │   │   └── AppleToken.json <- Contains ABI (function signatures)
#   │   └── AppleAMM.sol/
#   │       └── AppleAMM.json
#   └── script/
#       └── Deploy.s.sol        <- Deployment script
#
# BINDINGS_DIR: Where Go bindings are generated
#   backend/internal/contracts/
#   ├── apple_token.go          <- Generated: Go interface to AppleToken
#   └── apple_amm.go            <- Generated: Go interface to AppleAMM
#
# ==============================================================================
# WHAT HAPPENS WHEN YOU CHANGE A CONTRACT
# ==============================================================================
#
# Example: You add a new function `getVolume()` to AppleAMM.sol
#
# 1. Edit contracts/src/AppleAMM.sol
#    + function getVolume() public view returns (uint256) { ... }
#
# 2. Run this script: ./scripts/generate-bindings.sh
#    - forge build: Recompiles AppleAMM.sol -> updates out/AppleAMM.sol/AppleAMM.json
#    - abigen: Reads new ABI -> generates new apple_amm.go with GetVolume() method
#
# 3. Go backend now has access to the new function:
#    result, err := ammContract.GetVolume(&bind.CallOpts{})
#
# 4. If you deployed to Anvil before this change, you must REDEPLOY:
#    ./scripts/deploy.sh
#    (The old on-chain contract doesn't have the new function)
#
# ==============================================================================

set -e # Exit immediately if any command fails

# Path resolution - these are relative to wherever this script lives
# $(dirname "$0") = the scripts/ directory
# ../contracts    = go up one level, then into contracts/
CONTRACTS_DIR="$(dirname "$0")/../contracts"
BACKEND_DIR="$(dirname "$0")/../backend"
BINDINGS_DIR="$BACKEND_DIR/internal/contracts"

# ==============================================================================
# STEP 1: Compile Solidity contracts with Forge
# ==============================================================================
# This runs `forge build` which:
# - Reads all .sol files in contracts/src/
# - Compiles them to EVM bytecode
# - Outputs ABI JSON files to contracts/out/
#
# The ABI (Application Binary Interface) describes:
# - Function names and their parameters
# - Event signatures
# - Return types
# This is what abigen needs to generate Go code.

echo "Compiling contracts..."
cd "$CONTRACTS_DIR"
forge build

# ==============================================================================
# STEP 2: Generate Go bindings using abigen
# ==============================================================================
# abigen is a tool from go-ethereum that reads ABI JSON and generates Go code.
#
# Arguments explained:
#   --abi    : Path to the compiled ABI JSON file
#   --pkg    : Go package name for generated code (contracts)
#   --type   : Go struct name for the contract binding
#   --out    : Output file path
#
# The generated Go file contains:
# - A struct representing the contract (e.g., AppleAMM)
# - Methods for each contract function (e.g., SwapETHForApples, GetReserves)
# - Event types and parsing utilities
# - Constructor for connecting to a deployed contract address

echo "Generating Go bindings..."

# AppleToken binding - ERC20 token with mint capability
# Used by: backend to check balances, approve spending
abigen --abi out/AppleToken.sol/AppleToken.json \
       --pkg contracts \
       --type AppleToken \
       --out "$BINDINGS_DIR/apple_token.go"

# AppleAMM binding - The core AMM contract
# Used by: executor.go to execute swaps, read reserves, get prices
abigen --abi out/AppleAMM.sol/AppleAMM.json \
       --pkg contracts \
       --type AppleAMM \
       --out "$BINDINGS_DIR/apple_amm.go"

echo "Go bindings generated in $BINDINGS_DIR"

# ==============================================================================
# AFTER RUNNING THIS SCRIPT
# ==============================================================================
#
# The Go backend can now:
#
# 1. Import the bindings:
#    import "eth-amm-sim/backend/internal/contracts"
#
# 2. Connect to a deployed contract:
#    amm, _ := contracts.NewAppleAMM(contractAddress, ethClient)
#
# 3. Call contract methods:
#    reserves, _ := amm.GetReserves(&bind.CallOpts{})  // Read-only call
#    tx, _ := amm.SwapETHForApples(auth, minOut)       // State-changing tx
#
# See backend/internal/engine/executor.go for usage examples.
# ==============================================================================
