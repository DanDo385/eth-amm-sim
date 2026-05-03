# ETH-AMM-SIM

Local AMM simulation designed as a portfolio-grade demo for a Solutions Engineer in crypto/blockchain. It wires Solidity contracts to a Go execution engine and a Next.js dashboard so you can explain protocol mechanics, real-time data pipelines, and system design tradeoffs in one cohesive project.

## Portfolio demo package

- Loom guide: [`DEMO_GUIDE.md`](./DEMO_GUIDE.md)
- Recommended launcher: `make demo-120`
- Recommended video length: 120 seconds primary, 180 seconds deep cut
- Portfolio tag: `sim`
- Best GIF loop: start bots → Shock Pool → price/TWAP/LP panels move
- Thumbnail hook: `AMM EXECUTION LAB`
- Loom link: TBD
- Preview MP4/GIF: TBD
- Architecture image: dashboard includes a built-in Show Architecture panel


## Why this project exists
This repo is intentionally built for demoability:
- Clear, inspectable end-to-end flow (contracts -> backend -> UI).
- Real-time WebSocket streaming and REST snapshots.
- Bots that stress the pool and generate meaningful metrics.
- A small but realistic system that is easy to run in interviews or live demos.

## What this showcases (Solutions Engineer focus)
- **Protocol mechanics:** Constant-product AMM, fees, LP economics, price impact.
- **System integration:** go-ethereum bindings, on-chain execution, backend orchestration.
- **Data pipeline:** on-chain reads -> metrics -> WebSocket -> visualization.
- **Operational thinking:** session lifecycle, safe defaults, deterministic local infra.
- **Customer-facing clarity:** explorable UI and explicit data provenance.

## Architecture (Current)

```
Next.js (localhost:3000)
  - Dashboard (charts + controls)
  - REST fetch for snapshots
  - WebSocket for live updates
          |
          v
Go backend (localhost:8080)
  - Session + bot orchestration
  - On-chain executor (go-ethereum)
  - In-memory store + metrics
          v
Anvil (localhost:8545)
  - AppleToken + AppleAMM contracts
```

### Interfaces
- **REST**: snapshot endpoints for candles, trades, account performance, and session state.
- **WebSocket**: `ws://localhost:8080/stream` for real-time trades, prices, and events.
- **Chain RPC**: `http://localhost:8545` (Anvil local chain).

## Simulation Flow
1. Deploy contracts via Foundry (Deploy.s.sol writes broadcast JSON).
2. Backend starts and loads contract addresses from broadcast output.
3. Bots are created based on `backend/internal/config/accounts.go`.
4. On-chain trades emit updates -> in-memory store -> metrics pipeline.
5. WebSocket broadcasts push live updates to the dashboard.

## Feature Highlights
- AppleToken (ERC20) + AppleAMM (constant product, 0.30% fee) on Anvil.
- Go executor that submits swaps and computes trade details.
- Trading bots:
  - Retail (15) small random trades
  - Whale (3) large random trades
  - MeanRev (3) EWMA mean reversion on trade-flow events
- Session control (start/stop/reset) with per-session bot lifecycle.
- Metrics computed in Go:
  - 5s OHLC candles, 60s TWAP, volatility from observed returns
  - LP metrics (IL, fees earned, net PnL)
  - Account performance (equity curve, Sharpe, drawdown, win rate)
- Dashboard components: Price, TWAP, Impact Curve, Blotter, LP Stats, Key Events, Account Metrics.
- Performance analytics page at `/performance`.

## Liquidity Pool Calculations (LP Metrics)
The LP metrics are calculated in `backend/internal/metrics/lp.go` from on-chain reserves and fee totals. The goal is to clearly separate **price-only IL** from **fees** and the LP’s **realized vs HODL** performance.

Core inputs:
- **Reserves**: current APPL + ETH reserves from the AMM.
- **Initial state**: reserves and price at session start.
- **Fees**: cumulative fees from the contract, tracked as “current minus initial”.
- **Spot price**: `ETH reserve / APPL reserve` (ETH per APPL).

Key outputs:
- **LP Value**: `currentApples * price + currentETH`.
- **HODL Value**: value of the initial deposit at current price  
  `initialApples * price + initialETH`.
- **Theoretical IL (price-only)**:  
  `IL% = (2 * sqrt(r) / (1 + r)) - 1`, where `r = currentPrice / initialPrice`.  
  Converted to ETH terms: `HODLValue * IL%`. This is always ≤ 0 and **ignores fees**.
- **LP vs HODL PnL**: `LPValue - HODLValue` (can be +/-).
- **Fees Earned**:  
  `feesApple = currentFeesApple - initialFeesApple`  
  `feesETH = currentFeesETH - initialFeesETH`  
  Convert apples to ETH at the **current price** and sum.
- **Net PnL**: `LPvsHODL + feesEarned`.
- **Net PnL %**: `NetPnL / HODLValue`.

These values are snapshotted over time for charting (history includes reserves, price, LP value, HODL value, IL, fees, and net PnL).

## Concurrency Model (Go + Goroutines)
This project is intentionally concurrent: many bots trade simultaneously, the backend polls on-chain state, and the UI consumes real-time updates. Go’s goroutines are a **superior fit** for this type of system because they are lightweight, cheap to spawn, and provide clear coordination primitives for many parallel workflows.

Where goroutines are used:
- **Bot execution**: each bot runs in its own goroutine with context cancellation (start/stop/reset).
- **Price polling**: a dedicated goroutine polls reserves every few seconds and updates metrics.
- **WebSocket broadcasting**: non-blocking fan-out of updates to clients.
- **Trade callbacks + analytics**: async callbacks and trade-flow notifications to avoid blocking execution.

Why Go is especially well-suited here:
- **Low overhead concurrency**: dozens of bots and background tasks without heavy threads.
- **Simple coordination**: `context` + `errgroup` for lifecycle management and clean shutdown.
- **Safe parallelism**: mutex-protected metrics + a nonce manager to prevent tx collisions.
- **Great IO fit**: RPC calls, timers, and WebSocket writes are naturally concurrent.

## Repository Layout
- `contracts/` Solidity contracts + Foundry scripts
- `backend/` Go engine, bots, metrics, REST + WebSocket server
- `frontend/` Next.js dashboard
- `scripts/` deployment, bindings, and dev tooling
- `Makefile` orchestration targets

## Prerequisites
- **Foundry** (forge, anvil)
- **Go** 1.21+
- **Node.js** 18+ (npm)
- **tmux** (optional but recommended for `make up`)
- **abigen** (from go-ethereum) for Go contract bindings
- **jq** or **python3** for ABI extraction in `scripts/generate-bindings.sh`

### tmux installation
macOS (Homebrew):
```bash
brew install tmux
```

Ubuntu/Debian:
```bash
sudo apt-get update
sudo apt-get install -y tmux
```

Fedora:
```bash
sudo dnf install -y tmux
```

Arch:
```bash
sudo pacman -S tmux
```

Windows:
- Use WSL and install with `sudo apt-get install -y tmux`.

## Quickstart (Recommended - tmux)
This launches Anvil, deploys contracts, generates bindings, and starts backend + frontend.

```bash
make setup   # installs deps (Foundry/solidity deps + frontend deps + bindings)
make up      # tmux session: anvil, deploy, bindings, backend, frontend
```

To stop everything:
```bash
make down
```

Open `http://localhost:3000`, click **Start**, and watch metrics stream in.

## Manual Run (4 terminals)
If you prefer to run each service yourself:

```bash
make anvil
make deploy
make bindings
make backend
make frontend
```

## Useful Make Targets
- `make setup` - install deps and generate bindings
- `make anvil` - start local chain on :8545
- `make deploy` - deploy contracts and write broadcast JSON
- `make bindings` - generate Go bindings from contract ABIs
- `make backend` - run Go simulator on :8080
- `make frontend` - run Next.js on :3000
- `make kill-all` - free ports 8545, 8080, 3000-3004
- `make test-contracts` - run Foundry tests

## Configuration & Tuning
- **Accounts & bot params**: `backend/internal/config/accounts.go`
- **AMM constants & thresholds**: `backend/internal/config/amm.go`
- **Chain + session defaults**: `backend/internal/config/config.go`

Common knobs:
- Session duration
- Bot frequency and max trade size
- Pool seed reserves
- Volatility and LP metrics behavior

## API Endpoints (REST)
- POST `/session/start` (optional body: `{ "duration": 300 }`)
- POST `/session/stop`
- POST `/session/reset` (query: `?hard=true` for hard reset)
- GET  `/session/state`
- GET  `/candles`
- GET  `/trades` (query: `?limit=...`)
- GET  `/impact-curve`
- GET  `/lp/metrics`
- GET  `/events` (query: `?limit=...`)
- GET  `/accounts`
- GET  `/accounts/{nickname}/performance`
- POST `/trade/buy` (body: `{ "ethAmount": "1.0" }`)
- POST `/trade/sell` (body: `{ "appleAmount": "1.0" }`)
- GET  `/user/balance`

## WebSocket
- `ws://localhost:8080/stream`
- Message types: `trade`, `price`, `lp_metrics`, `key_event`, `session_state`, `account_update`, `trades`, `candles`, `events`.

## Demo Walkthrough (5 minutes)
1. `make up` and open `http://localhost:3000`.
2. Click **Start** and highlight the session timer + live trade blotter.
3. Point out price impact from whale trades on the chart.
4. Open **LP Stats** to explain impermanent loss vs. fees earned.
5. Open `/performance` to show account equity curves and Sharpe.

## Troubleshooting
- **Ports in use**: run `make kill-all` then retry.
- **No data in UI**: confirm backend is running on `:8080` and contracts were deployed.
- **Deploy failed**: delete old `contracts/broadcast/Deploy.s.sol/31337/run-latest.json` and redeploy.
- **Bindings errors**: ensure `abigen` is installed and `jq` or `python3` is available.
- **tmux not found**: install tmux or run the manual workflow.

## Current Parameters (Defaults)
- Accounts: 30 Anvil addresses; active 23
  - 1 LP (index 0)
  - 1 User (index 1)
  - 15 Retail (indices 2-16)
  - 3 Whale (indices 17-19)
  - 3 MeanRev (indices 20-22)
  - 7 Reserved (indices 23-29)
- Session duration: 180s default (UI-configurable)
- Retail: 10-100% of 15 ETH, every 1-4s
- Whales: up to 600 ETH, every 45-90s
- MeanRev: 50/75/150 ETH notional, EWMA half-life 25/50/75 trades

## In Progress
- Expose trade-flow diagnostics (rTilde/flow) in the UI for debugging.

## Future Ideas
- Multi-pool or multi-token simulations.
- Persist metrics to disk for replay.
- Compare against an external reference price feed (read-only).

## Disclaimer
This project is for local simulation and educational/demo purposes. It is not audited and is not intended for production or mainnet usage.
