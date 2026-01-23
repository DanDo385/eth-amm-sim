#!/bin/bash
# Generate Go bindings from contract ABIs

set -e

CONTRACTS_DIR="$(dirname "$0")/../contracts"
BACKEND_DIR="$(dirname "$0")/../backend"
BINDINGS_DIR="$BACKEND_DIR/internal/contracts"

echo "Compiling contracts..."
cd "$CONTRACTS_DIR"
forge build

echo "Generating Go bindings..."

# AppleToken
abigen --abi out/AppleToken.sol/AppleToken.json \
       --pkg contracts \
       --type AppleToken \
       --out "$BINDINGS_DIR/apple_token.go"

# AppleAMM
abigen --abi out/AppleAMM.sol/AppleAMM.json \
       --pkg contracts \
       --type AppleAMM \
       --out "$BINDINGS_DIR/apple_amm.go"

echo "Go bindings generated in $BINDINGS_DIR"
