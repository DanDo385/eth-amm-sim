#!/bin/bash
# Start Anvil with 30 accounts for simulation
# Each account starts with 10,000 ETH

echo "Starting Anvil local Ethereum node..."
echo "Accounts: 30"
echo "Balance per account: 10,000 ETH"
echo ""

anvil --accounts 30 --balance 10000
