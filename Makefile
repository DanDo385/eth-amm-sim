.PHONY: anvil deploy bindings backend frontend frontend-fresh clean test-contracts up demo-120 down kill-anvil kill-backend kill-all

# backend/go.mod pins Go 1.25.x (patch may bump after dependency upgrades). Shells that set GOTOOLCHAIN=local with an older `go` binary
# break `go run`/`go clean`; auto lets the toolchain switch fetch the matching Go for this module.
# GOFLAGS=-buildvcs=false avoids `go build` failing when `git status` errors (agents/sandboxes/broken .git).
BACKEND_TOOLCHAIN := GOTOOLCHAIN=auto GOFLAGS=-buildvcs=false

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
	cd backend && $(BACKEND_TOOLCHAIN) go run cmd/simulator/main.go

# Run Vite + React frontend (see http://localhost:3000 when Vite prints Local URL)
frontend:
	@echo "→ http://localhost:3000  (after Vite prints Local URL). If odd proxy/chunk errors: make frontend-fresh"
	./scripts/with-frontend-node.sh npm run dev

# Clean Vite cache/output then dev — fixes stale chunk / proxy issues after failed builds
frontend-fresh:
	cd frontend && rm -rf dist .vite .next
	./scripts/with-frontend-node.sh npm run dev

# Run Foundry tests
test-contracts:
	cd contracts && forge test -vvv

# Clean build artifacts
clean:
	cd contracts && forge clean
	cd frontend && rm -rf dist .vite .next node_modules
	cd backend && $(BACKEND_TOOLCHAIN) go clean

# Install frontend dependencies
frontend-install:
	./scripts/with-frontend-node.sh npm install

# Install Foundry dependencies
contracts-install:
	cd contracts && forge install OpenZeppelin/openzeppelin-contracts --no-commit

# Full setup
setup: contracts-install frontend-install bindings

# Launch Anvil + deploy + backend + frontend in separate Terminal windows (macOS)
up:
	./scripts/dev-up.sh

# Loom-ready demo launcher. Use the UI session control for exact duration (e.g., 120s).
demo-120: kill-all
	@echo "Starting Loom demo mode: Anvil + deploy + Go backend + Vite frontend"
	@echo "Open http://localhost:3000, click Start, then showcase a large key-event trade around 1:25."
	./scripts/dev-up.sh

# Stop tmux session and all running services
down:
	-@tmux kill-session -t eth-amm-sim 2>/dev/null || true
