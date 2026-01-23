#!/bin/bash
# Deploy contracts to Anvil

set -e

echo "Deploying contracts to Anvil..."
cd "$(dirname "$0")/../contracts"

forge script script/Deploy.s.sol --rpc-url http://localhost:8545 --broadcast

echo ""
echo "Deployment complete!"
echo "Contract addresses are saved in contracts/broadcast/"
