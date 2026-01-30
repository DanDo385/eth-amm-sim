# ETH-AMM-SIM

Local AMM simulation that runs on a local Anvil chain. A Go backend drives bots and metrics, and a Next.js dashboard visualizes price, LP economics, and account performance in real time.

## Project Goals
- Demonstrate constant-product AMM mechanics end-to-end (Solidity -> Go -> UI).
- Visualize price impact, LP economics, and per-account performance in a short demo.
- Keep the system explicit and inspectable (no hidden data pipelines).

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
  - REST + WebSocket server
          |
          v
Anvil (localhost:8545)
  - AppleToken + AppleAMM contracts
```

### Simulation Flow
1. Deploy contracts via Foundry (Deploy.s.sol writes broadcast JSON).
2. Backend starts, loads contract addresses, verifies pool reserves.
3. Session start launches bots under a shared context.
4. Price polling loop (every 2s) reads on-chain spot price/reserves and updates metrics.
5. Trades emit callbacks -> MemoryStore -> WebSocket -> UI.

## Implemented Features (✅)
- AppleToken (ERC20) + AppleAMM (constant product, 0.30% fee) on Anvil.
- Go executor that submits swaps and computes trade details.
- Trading bots:
  - Retail (15) small random trades
  - Whale (3) large random trades
  - MeanRev (3) EWMA mean reversion on trade-flow events
- Session control (start/stop/reset) with per-session bot lifecycle.
- Metrics computed in Go:
  - 5s OHLC candles, 60s TWAP, simple volatility from observed returns
  - LP metrics (IL, fees earned, net PnL) from on-chain reserves/fees
  - Account performance (equity curve, Sharpe, drawdown, win rate)
- WebSocket streaming for live updates (trades, prices, LP metrics, events).
- Dashboard components: Price, TWAP, Impact Curve, Blotter, LP Stats, Key Events, Account Metrics.
- Performance analytics page at `/performance` with navigation link from main dashboard.
- Account performance baselines aligned to actual on-chain starting balances.

## In Progress (🚧)
- Expose trade-flow diagnostics (rTilde/flow) in the UI for debugging.

## Future Ideas (🧠)
- Multi-pool or multi-token simulations.
- Persist metrics to disk for replay.
- Compare against an external reference price feed (read-only).

## Tech Stack
- Contracts: Solidity ^0.8.20, Foundry, OpenZeppelin
- Backend: Go, go-ethereum, gorilla/websocket, errgroup
- Frontend: Next.js App Router, TypeScript, Tailwind CSS, lightweight-charts

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

## Running Locally

Prereqs: Foundry, Go 1.21+, Node 18+

```bash
make anvil       # start local chain (30 accounts)
make deploy      # deploy contracts and seed liquidity
make backend     # run Go backend
make frontend    # run Next.js frontend
```

Open `http://localhost:3000` and click Start.

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
- Message types include: `trade`, `price`, `lp_metrics`, `key_event`, `session_state`, `account_update`, `trades`, `candles`, `events`.
