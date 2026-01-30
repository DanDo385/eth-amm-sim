.PHONY: anvil deploy bindings backend frontend clean test-contracts up down kill-anvil kill-backend kill-all

# Kill any existing Anvil process
kill-anvil:
	@lsof -ti:8545 | xargs kill -9 2>/dev/null || echo "No Anvil process found on port 8545"

# Kill any existing backend process
kill-backend:
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No backend process found on port 8080"

# Kill all processes (Anvil, backend, frontend ports)
kill-all:
	@lsof -ti:8545 | xargs kill -9 2>/dev/null || true
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@lsof -ti:3000,3001,3002,3003,3004 | xargs kill -9 2>/dev/null || true
	@echo "Killed processes on ports 8545, 8080, and 3000-3004"

# Start local Anvil chain with 30 accounts
# LP account (0) needs 30k ETH, others need 1k ETH each
# Note: If port 8545 is in use, run 'make kill-anvil' first
# Starting with 30000 ETH ensures account 0 has enough for liquidity + gas
anvil:
	anvil --accounts 30 --balance 30000

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
