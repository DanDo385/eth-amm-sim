# ETH-AMM-SIM

> Portfolio-grade DeFi AMM simulation showcasing market microstructure, LP dynamics, and TradFi-style performance analytics.

## Overview

This project simulates market microstructure in an Automated Market Maker (AMM) using heterogeneous agents and evaluates trading strategy performance using institutional risk metrics.

**Key Features:**
- Constant product AMM (x*y=k) with 0.30% fee
- Heterogeneous trading agents (Whales, Retail, Strategy bots)
- Real-time LP metrics including impermanent loss
- TradFi performance analytics (Sharpe ratio, volatility, max drawdown)
- Live price charts and trade blotter

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Next.js Frontend                             │
│                    (localhost:3000)                              │
│  - Dashboard with price charts, trade blotter, LP stats         │
│  - Performance page with Sharpe, drawdown, equity curves        │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTP + WebSocket
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Go Backend                                 │
│                    (localhost:8080)                              │
│  - Session management (start/stop/reset)                        │
│  - Bot orchestration (goroutines)                               │
│  - Trade execution (NonceManager)                               │
│  - Metrics calculation                                          │
│  - WebSocket broadcast                                          │
└───────────────────────────┬─────────────────────────────────────┘
                            │ JSON-RPC
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Anvil                                    │
│                    (localhost:8545)                              │
│  - AppleToken.sol (ERC20)                                       │
│  - AppleAMM.sol (Constant Product AMM)                          │
└─────────────────────────────────────────────────────────────────┘
```

## Tech Stack

- **Smart Contracts**: Solidity + Foundry
- **Local Chain**: Anvil (Foundry's local Ethereum node)
- **Backend**: Go + go-ethereum
- **Frontend**: Next.js 14 + TypeScript + Tailwind CSS
- **Charts**: TradingView Lightweight Charts

## Quick Start

### Prerequisites

- [Foundry](https://book.getfoundry.sh/getting-started/installation) (for Solidity compilation and Anvil)
- [Go 1.21+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)

### 1. Start Anvil (Local Ethereum)

```bash
make anvil
# or: anvil --accounts 30 --balance 10000
```

### 2. Deploy Contracts

In a new terminal:

```bash
make deploy
```

Note the deployed contract addresses from the output.

### 3. Start Go Backend

Set environment variables and start:

```bash
export TOKEN_ADDRESS=<deployed_token_address>
export AMM_ADDRESS=<deployed_amm_address>
make backend
```

### 4. Start Frontend

In another terminal:

```bash
cd frontend && npm install
npm run dev
```

### 5. Open Dashboard

Navigate to [http://localhost:3000](http://localhost:3000)

## Demo Walkthrough

1. **Start a Session**: Click "Start" on the dashboard. This spawns trading bots.
2. **Watch Price Movement**: Whale bots make large trades that visibly move the market.
3. **Observe LP Metrics**: See impermanent loss and fee accumulation in real-time.
4. **View Performance**: Navigate to `/performance` to see TradFi analytics.

### What to Look For

- **Price Impact**: Large whale trades cause significant price slippage
- **LP PnL**: Track how fees offset impermanent loss
- **Sharpe Ratio**: Measure risk-adjusted returns for each bot
- **Max Drawdown**: Observe worst-case losses during volatile periods

## Project Structure

```
eth-amm-sim/
├── contracts/                    # Foundry project
│   ├── src/
│   │   ├── AppleToken.sol       # ERC20 token
│   │   └── AppleAMM.sol         # Constant product AMM
│   ├── script/
│   │   └── Deploy.s.sol         # Deployment script
│   └── test/                    # Foundry tests
│
├── backend/                      # Go simulation engine
│   ├── cmd/simulator/           # Entry point
│   └── internal/
│       ├── bots/                # Trading bot implementations
│       ├── chain/               # Ethereum client
│       ├── engine/              # Session & executor
│       ├── metrics/             # TradFi analytics
│       └── server/              # HTTP + WebSocket
│
├── frontend/                     # Next.js app
│   ├── app/                     # Pages (App Router)
│   ├── components/              # React components
│   └── hooks/                   # Custom hooks
│
└── scripts/                     # Helper scripts
```

## Smart Contracts

### AppleToken.sol
Standard ERC20 with owner-only mint function.

### AppleAMM.sol
Constant product AMM with:
- 0.30% fee on all swaps
- LP token accounting
- Events for swaps and liquidity changes

Key functions:
- `addLiquidity(uint256 apples) payable`
- `removeLiquidity(uint256 lpTokens)`
- `swapETHForApples(uint256 minOut) payable`
- `swapApplesForETH(uint256 apples, uint256 minOut)`
- `getReserves() view`
- `getSpotPrice() view`

## Trading Bots

| Bot Type | Behavior | Trade Size | Frequency |
|----------|----------|------------|-----------|
| Whale | Large random trades | 100-500 APPL | 15-30s |
| Retail | Small noise trades | 1-10 APPL | 2-5s |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /session/start | Start simulation |
| POST | /session/stop | Stop simulation |
| POST | /session/reset | Reset state |
| GET | /session/state | Get session state |
| GET | /accounts | List accounts |
| GET | /accounts/:nickname/performance | Get performance metrics |
| GET | /lp/metrics | Get LP metrics |
| GET | /candles | Get OHLC data |
| GET | /trades | Get trade history |
| WS | /stream | Real-time updates |

## Metrics Calculated

### Price Metrics
- Spot price
- OHLC candles
- TWAP (Time-Weighted Average Price)
- Volatility

### LP Metrics
- Pool reserves
- Fees earned
- Impermanent loss
- Net PnL

### Account Performance (TradFi-style)
- Total return
- Sharpe ratio (rf = 0)
- Volatility (std dev of returns)
- Max drawdown
- Win rate
- Trade count
- Equity curve

## Development

### Run Contract Tests

```bash
cd contracts && forge test -vvv
```

### Build Backend

```bash
cd backend && go build ./...
```

### Type Check Frontend

```bash
cd frontend && npx tsc --noEmit
```

## License

MIT

---

*This is a portfolio demonstration project. Not intended for production use.*
