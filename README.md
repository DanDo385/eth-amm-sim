# ETH-AMM-SIM

> Portfolio-grade DeFi AMM simulation showcasing market microstructure, LP dynamics, and TradFi-style performance analytics.

## Overview

This project simulates market microstructure in an Automated Market Maker (AMM) using heterogeneous agents and evaluates trading strategy performance using institutional risk metrics.

**Key Features:**
- Constant product AMM (x*y=k) with 0.30% fee
- Leveraged positions with liquidation mechanics
- Heterogeneous trading agents (Whales, Retail, Strategy bots, Leverage traders, Liquidators)
- Real-time LP metrics including impermanent loss
- TradFi performance analytics (Sharpe ratio, volatility, max drawdown)
- Live price charts with TWAP and rolling standard deviation
- Price impact curve visualization
- Trade blotter and key events tracking

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Next.js Frontend                             │
│                    (localhost:3000)                              │
│  - Dashboard with price charts, TWAP/Std Dev, impact curves    │
│  - Trade blotter, LP stats, account performance                │
│  - Performance page with Sharpe, drawdown, equity curves       │
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

The backend automatically reads contract addresses from the broadcast JSON output. No `.env` file needed!

### 3. Start Go Backend

The backend automatically reads contract addresses from the deployment broadcast JSON. Simply start:

```bash
make backend
```

The backend will find addresses in `contracts/broadcast/Deploy.s.sol/31337/run-latest.json` automatically.

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
2. **Watch Price Movement**: 
   - Whale bots make large trades that visibly move the market
   - Mean reversion bots trade when price deviates from TWAP
   - Leverage bots open positions and may get liquidated
3. **Observe Charts**:
   - Price chart shows OHLC candles with real-time updates
   - TWAP/Std Dev chart toggles between time-weighted average and rolling volatility
   - Impact curve shows execution prices vs trade size
4. **Observe LP Metrics**: See impermanent loss and fee accumulation in real-time
5. **View Performance**: Navigate to `/performance` to see TradFi analytics

### What to Look For

- **Price Impact**: Large whale trades cause significant price slippage (visible in impact curve)
- **TWAP Deviation**: Mean reversion bots trade when spot price deviates from TWAP
- **Liquidation Events**: Leverage positions get liquidated when margin ratio falls below 10%
- **LP PnL**: Track how fees offset impermanent loss
- **Sharpe Ratio**: Measure risk-adjusted returns for each bot
- **Max Drawdown**: Observe worst-case losses during volatile periods
- **Rolling Volatility**: Standard deviation chart only appears after 100 trades

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
│   │   ├── PriceChart.tsx       # OHLC candlestick chart
│   │   ├── TWAPChart.tsx        # TWAP/Std Dev toggle chart
│   │   ├── ImpactCurve.tsx      # Price impact visualization
│   │   └── ...                  # Other components
│   └── hooks/                   # Custom hooks
│
└── scripts/                     # Helper scripts
    ├── deploy.sh               # Contract deployment
    ├── generate-bindings.sh   # Go bindings generation
    └── start-anvil.sh          # Anvil launcher
```

## Smart Contracts

### AppleToken.sol
Standard ERC20 with owner-only mint function.

### AppleAMM.sol
Constant product AMM with:
- 0.30% fee on all swaps
- LP token accounting
- Leveraged positions (up to 25x) with liquidation
- Reentrancy protection using OpenZeppelin's ReentrancyGuard
- Safe ETH transfers using Address.sendValue()
- Events for swaps, liquidity changes, and position events

Key functions:
- `addLiquidity(uint256 apples) payable`
- `removeLiquidity(uint256 lpTokens)`
- `swapETHForApples(uint256 minOut) payable`
- `swapApplesForETH(uint256 apples, uint256 minOut)`
- `openLeveragedPosition(uint256 leverage, uint256 minApples) payable`
- `closePosition()`
- `liquidate(address trader)`
- `getReserves() view`
- `getSpotPrice() view`
- `getTWAP() view`
- `getVolatility() view`

## Trading Bots

| Bot Type | Behavior | Trade Size | Frequency |
|----------|----------|------------|-----------|
| Whale | Large random trades | 100-500 APPL | 15-30s |
| Retail | Small noise trades | 1-10 APPL | 2-5s |
| Mean Reversion | Trades when price deviates from TWAP | Variable | 10-20s |
| Leverage | Opens leveraged long positions | 5-25x leverage | 30-60s |
| Liquidator | Liquidates underwater positions | N/A | On-demand |

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
| GET | /impact-curve | Get price impact curve data |
| GET | /events | Get key events |
| WS | /stream | Real-time updates (trades, prices, metrics, events) |

## Metrics Calculated

### Price Metrics
- Spot price
- OHLC candles (5-second intervals)
- TWAP (Time-Weighted Average Price, 60-second window)
- Volatility (standard deviation of returns)
- Rolling standard deviation (100-trade window)
- Price impact curves (buy/sell execution prices)

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

## Security Features

The smart contracts implement several security best practices:

- **Reentrancy Protection**: All state-changing functions use OpenZeppelin's `ReentrancyGuard`
- **Safe ETH Transfers**: Uses `Address.sendValue()` instead of low-level `.call{value:}`
- **Slippage Protection**: All swap functions require minimum output amounts
- **Liquidation Safety**: Hybrid TWAP/spot price checks prevent manipulation-based liquidations

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

### Regenerate Go Bindings

After modifying Solidity contracts:

```bash
make bindings
# or: ./scripts/generate-bindings.sh
```

## License

MIT

---

*This is a portfolio demonstration project. Not intended for production use.*
