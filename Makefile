.PHONY: anvil deploy bindings backend frontend clean test-contracts up down

# Start local Anvil chain with 30 accounts
# LP account (0) needs 15k ETH, others need 1k ETH each
anvil:
	anvil --accounts 30 --balance 10000

# Deploy contracts to Anvil (uses scripts/deploy.sh to keep run-latest.json only)
deploy:
	./scripts/deploy.sh

# Generate Go bindings from contract ABIs
bindings:
	./scripts/generate-bindings.sh

# Run Go backend
backend:
	cd backend && go run cmd/simulator/main.go

# Run Next.js frontend
frontend:
	cd frontend && npm run dev

# Run Foundry tests
test-contracts:
	cd contracts && forge test -vvv

# Clean build artifacts
clean:
	cd contracts && forge clean
	cd frontend && rm -rf .next node_modules
	cd backend && go clean

# Install frontend dependencies
frontend-install:
	cd frontend && npm install

# Install Foundry dependencies
contracts-install:
	cd contracts && forge install OpenZeppelin/openzeppelin-contracts --no-commit

# Full setup
setup: contracts-install frontend-install bindings

# Launch Anvil + deploy + backend + frontend in separate Terminal windows (macOS)
up:
	./scripts/dev-up.sh

# Stop tmux session and all running services
down:
	-@tmux kill-session -t eth-amm-sim 2>/dev/null || true
