# ETH-AMM-SIM

> Portfolio-grade DeFi AMM simulation showcasing market microstructure, LP dynamics, and TradFi-style performance analytics.

## Overview

**ETH-AMM-SIM** is a comprehensive demonstration project that simulates an Automated Market Maker (AMM) ecosystem with heterogeneous trading agents. The system showcases:

- **Market Microstructure**: Real-time price discovery through constant product AMM mechanics
- **Heterogeneous Agents**: Multiple bot types (Whales, Retail, Mean Reversion) with distinct trading behaviors
- **Liquidity Provider Economics**: Impermanent loss tracking, fee accumulation, and net PnL calculations
- **TradFi Analytics**: Institutional-grade risk metrics (Sharpe ratio, volatility, max drawdown, equity curves)
- **Real-Time Visualization**: Live price charts, TWAP/volatility overlays, impact curves, and trade blotters

This project demonstrates fluency in both **DeFi** (smart contracts, AMM mechanics, on-chain execution) and **TradFi** (risk metrics, performance analytics, market microstructure). It's designed as a portfolio demonstration that hiring managers can understand in a 2-3 minute demo video.

### Key Features

- **Constant Product AMM** (x*y=k) with 0.30% fee on all swaps
- **22 Simulation Accounts**: 1 LP, 3 Whales, 3 Mean Reversion bots, 14 Retail bots, 1 User account
- **Real-Time LP Metrics**: Impermanent loss, fees earned, net PnL with historical tracking
- **TradFi Performance Analytics**: Sharpe ratio, volatility, max drawdown, win rate, equity curves
- **Live Price Charts**: OHLC candlesticks with TradingView Lightweight Charts
- **TWAP & Volatility**: Time-weighted average price and rolling standard deviation overlays
- **Price Impact Curves**: Visualization of execution prices vs trade size
- **Trade Blotter**: Real-time trade feed with trader identification
- **Key Events**: Large trades, price movements, and strategy triggers
- **WebSocket Streaming**: Real-time data updates to frontend without polling

## Architecture

The system is architected as a three-tier application:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Next.js Frontend                            │
│                    (localhost:3000)                             │
│                                                                 │
│  Components:                                                    │
│  - PriceChart: OHLC candlestick chart                           │
│  - TWAPChart: TWAP/Std Dev toggle chart                         │
│  - ImpactCurve: Price impact visualization                      │
│  - Blotter: Real-time trade feed                                │
│  - LPStats: Impermanent loss, fees, net PnL                     │
│  - AccountMetrics: Per-account performance                      │
│  - KeyEvents: Large trades, price moves                         │
│  - SessionControls: Start/stop/reset simulation                 │
│  - TradingPanel: Manual trading for User account                │
│                                                                 │
│  Hooks:                                                         │
│  - useSession: Session lifecycle management                     │
│  - usePriceData: Candle history and LP metrics                  │
│  - useWebSocket: Real-time data stream                          │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTP REST + WebSocket
                            │ (localhost:8080)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Go Backend                                │
│                    (localhost:8080)                             │
│                                                                 │
│  Core Components:                                               │
│  - cmd/simulator/main.go: Entry point, orchestrates all         │
│  - engine/orchestrator.go: Bot lifecycle management             │
│  - engine/session.go: Session state (start/stop/reset)          │
│  - engine/executor.go: Bridge to on-chain contracts             │
│  - chain/client.go: Ethereum RPC client (go-ethereum)           │
│  - chain/nonce.go: Transaction nonce management                 │
│  - store/memory.go: In-memory data store (candles, trades)      │
│  - server/server.go: HTTP + WebSocket server                    │
│  - server/handlers.go: REST API endpoints                       │
│  - server/broadcast.go: WebSocket message broadcasting          │
│                                                                 │
│  Trading Bots:                                                  │
│  - bots/whale.go: Large random trades (100-500 APPL)            │
│  - bots/retail.go: Small noise trades (1-10 APPL)               │
│  - bots/meanrev.go: EWMA-based mean reversion                   │
│                                                                 │
│  Metrics:                                                       │
│  - metrics/price.go: OHLC candles, TWAP, volatility             │
│  - metrics/lp.go: Impermanent loss, fees, net PnL               │
│  - metrics/account.go: Per-account performance metrics          │
│  - metrics/ewma.go: Exponential weighted moving average         │
│  - metrics/impact.go: Price impact curve calculations           │
│  - metrics/trade_flow.go: Trade flow events for mean rev        │
│                                                                 │
│  Configuration:                                                 │
│  - config/accounts.go: All 22 accounts and trading params       │
│  - config/amm.go: AMM-specific constants                        │
│  - config/config.go: System-wide configuration                  │
└───────────────────────────┬─────────────────────────────────────┘
                            │ JSON-RPC (localhost:8545)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Anvil                                   │
│                    (localhost:8545)                             │
│  Local Ethereum node (Foundry)                                  │
│                                                                 │
│  Smart Contracts:                                               │
│  - AppleToken.sol: ERC20 token (APPL)                           │
│  - AppleAMM.sol: Constant product AMM with:                     │
│    • Liquidity provision (add/remove)                           │
│    • Swaps (ETH ↔ APPL) with 0.30% fee                          │
│    • TWAP and volatility calculations                           │
│                                                                 │
│  Accounts: 22 pre-funded accounts (Anvil determinstic)          │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Price Polling Loop** (every 2 seconds):
   - `main.go` → `executor.GetSpotPrice()` → reads `AppleAMM.getSpotPrice()`
   - Updates `MemoryStore` candles, TWAP, volatility
   - Reads reserves and fees → updates LP metrics
   - Broadcasts price and LP data via WebSocket

2. **Trade Execution**:
   - Bot decides to trade → calls `executor.SwapETHForApples()` or `SwapApplesForETH()`
   - `NonceManager` assigns nonce → transaction signed with account private key
   - Transaction sent to Anvil → `AppleAMM` contract executes swap
   - `executor` parses events → creates `Trade` struct
   - `OnTrade` callback → `MemoryStore.RecordTrade()` + `WebSocket.BroadcastTrade()`
   - Frontend receives trade → updates `Blotter` component

3. **Mean Reversion Strategy**:
   - `TradeFlow` events emitted on each trade (price before/after, reserves)
   - `EWMA` calculates liquidity-normalized log returns
   - Z-score computed vs EWMA mean → triggers trades at threshold levels
   - Each `MeanRevBot` maintains independent EWMA state

## Tech Stack

### Smart Contracts
- **Solidity** ^0.8.20
- **Foundry** (forge, anvil, cast)
- **OpenZeppelin Contracts** (ReentrancyGuard, SafeERC20, Address)
- **ABI Generation**: `forge build` → Go bindings via `abigen`

### Backend
- **Go** 1.21+
- **go-ethereum** (Ethereum client library)
- **gorilla/websocket** (WebSocket server)
- **golang.org/x/sync/errgroup** (goroutine coordination)

### Frontend
- **Next.js 14** (App Router)
- **TypeScript**
- **Tailwind CSS** (styling)
- **TradingView Lightweight Charts** (price charts)
- **React Hooks** (state management, no Redux/Zustand)

### Development Tools
- **Anvil** (local Ethereum node)
- **Makefile** (build automation)
- **Shell scripts** (deployment, binding generation)

## Project Structure

```
eth-amm-sim/
├── contracts/                    # Foundry project
│   ├── src/
│   │   ├── AppleToken.sol         # ERC20 token with owner-only mint
│   │   └── AppleAMM.sol          # Constant product AMM
│   ├── script/
│   │   └── Deploy.s.sol          # Deployment script (mints tokens, seeds pool)
│   ├── test/                     # Foundry tests
│   │   ├── AppleToken.t.sol
│   │   └── AppleAMM.t.sol
│   ├── out/                      # Compiled contracts (ABIs)
│   ├── broadcast/                # Deployment addresses (JSON)
│   └── foundry.toml              # Foundry configuration
│
├── backend/                      # Go simulation engine
│   ├── cmd/simulator/
│   │   └── main.go               # Entry point, orchestrates all components
│   ├── internal/
│   │   ├── bots/                 # Trading bot implementations
│   │   │   ├── base.go           # Shared bot foundation
│   │   │   ├── bot.go            # Bot interface
│   │   │   ├── whale.go          # Large random trades
│   │   │   ├── retail.go         # Small noise trades
│   │   │   └── meanrev.go        # EWMA mean reversion
│   │   ├── chain/                # Ethereum client layer
│   │   │   ├── client.go         # RPC client wrapper
│   │   │   ├── accounts.go       # Account management
│   │   │   └── nonce.go          # Nonce manager (prevents conflicts)
│   │   ├── config/               # Configuration
│   │   │   ├── accounts.go        # All 22 accounts and trading params
│   │   │   ├── amm.go            # AMM constants
│   │   │   └── config.go         # System config
│   │   ├── contracts/            # Generated Go bindings
│   │   │   ├── apple_amm.go      # AppleAMM contract binding
│   │   │   └── apple_token.go    # AppleToken contract binding
│   │   ├── engine/               # Core simulation engine
│   │   │   ├── orchestrator.go   # Bot lifecycle management
│   │   │   ├── session.go        # Session state (start/stop/reset)
│   │   │   ├── executor.go       # Bridge to on-chain contracts
│   │   │   └── types.go          # Shared types
│   │   ├── metrics/              # TradFi analytics
│   │   │   ├── price.go          # OHLC candles, TWAP, volatility
│   │   │   ├── lp.go             # Impermanent loss, fees, net PnL
│   │   │   ├── account.go        # Per-account performance
│   │   │   ├── ewma.go           # Exponential weighted moving average
│   │   │   ├── impact.go         # Price impact curves
│   │   │   ├── trade_flow.go     # Trade flow events
│   │   │   └── price_provider.go # Price data interface
│   │   ├── server/                # HTTP + WebSocket server
│   │   │   ├── server.go         # Server setup
│   │   │   ├── handlers.go       # REST API endpoints
│   │   │   └── broadcast.go      # WebSocket broadcasting
│   │   └── store/                # In-memory data store
│   │       └── memory.go         # Candles, trades, LP metrics, events
│   ├── go.mod
│   └── go.sum
│
├── frontend/                      # Next.js app
│   ├── app/
│   │   ├── page.tsx              # Main dashboard
│   │   ├── performance/
│   │   │   └── page.tsx         # Performance analytics page
│   │   ├── layout.tsx            # Root layout
│   │   └── globals.css           # Global styles
│   ├── components/               # React components
│   │   ├── PriceChart.tsx        # OHLC candlestick chart
│   │   ├── TWAPChart.tsx         # TWAP/Std Dev toggle chart
│   │   ├── ImpactCurve.tsx       # Price impact visualization
│   │   ├── Blotter.tsx           # Trade feed table
│   │   ├── LPStats.tsx           # LP metrics display
│   │   ├── AccountMetrics.tsx    # Per-account performance
│   │   ├── KeyEvents.tsx         # Event feed
│   │   ├── SessionControls.tsx   # Start/stop/reset buttons
│   │   └── TradingPanel.tsx     # Manual trading UI
│   ├── hooks/                    # Custom React hooks
│   │   ├── useSession.ts         # Session lifecycle
│   │   ├── usePriceData.ts       # Price data management
│   │   └── useWebSocket.ts       # WebSocket connection
│   ├── lib/
│   │   └── api.ts                # REST API client
│   ├── types/
│   │   └── index.ts              # TypeScript types
│   ├── package.json
│   └── tsconfig.json
│
├── scripts/                       # Helper scripts
│   ├── deploy.sh                 # Contract deployment
│   ├── generate-bindings.sh       # Go bindings generation
│   ├── start-anvil.sh            # Anvil launcher
│   └── post-deploy.sh            # Post-deployment setup
│
├── Makefile                       # Build automation
└── README.md                      # This file
```

## Quick Start

### Prerequisites

- **[Foundry](https://book.getfoundry.sh/getting-started/installation)** (for Solidity compilation and Anvil)
- **[Go 1.21+](https://golang.org/dl/)**
- **[Node.js 18+](https://nodejs.org/)**

### 1. Install Dependencies

```bash
# Install Foundry dependencies (OpenZeppelin)
make contracts-install

# Install frontend dependencies
make frontend-install

# Generate Go bindings from contract ABIs
make bindings
```

Or run all at once:
```bash
make setup
```

### 2. Start Anvil (Local Ethereum Chain)

In the first terminal:

```bash
make anvil
# or: anvil --accounts 30 --balance 15000
```

This starts Anvil with 30 pre-funded accounts. Account 0 (LP) gets 15,000 ETH; others get 1,000 ETH each.

### 3. Deploy Contracts

In a new terminal:

```bash
make deploy
```

This:
- Compiles `AppleToken.sol` and `AppleAMM.sol`
- Deploys both contracts to Anvil
- Mints 10,000 APPL tokens to the LP account
- Seeds the AMM pool with 10,000 APPL + 10,000 ETH (initial price: 1.0 ETH/APPL)
- Distributes APPL and ETH to all trading accounts
- Writes deployment addresses to `contracts/broadcast/Deploy.s.sol/31337/run-latest.json`

The backend automatically reads contract addresses from this broadcast JSON file.

### 4. Start Go Backend

In a new terminal:

```bash
make backend
```

The backend will:
- Load contract addresses from broadcast JSON (or environment variables)
- Connect to Anvil at `localhost:8545`
- Verify contracts are deployed and pool has liquidity
- Create all trading bots from `config/accounts.go`
- Start price polling loop (every 2 seconds)
- Start HTTP + WebSocket server on `:8080`

### 5. Start Frontend

In another terminal:

```bash
make frontend
# or: cd frontend && npm run dev
```

The frontend will start on [http://localhost:3000](http://localhost:3000)

### 6. Open Dashboard

Navigate to [http://localhost:3000](http://localhost:3000) and click **"Start"** to begin the simulation.

## Demo Walkthrough

### Starting a Session

1. Click **"Start"** on the dashboard. This:
   - Starts all bot goroutines (Whale, Retail, MeanRev)
   - Begins a new session timer
   - Initializes LP metrics baseline (for impermanent loss calculation)

2. Watch the **Price Chart**: OHLC candlesticks update every 2 seconds as bots trade.

3. Observe **Trade Blotter**: Real-time feed of all trades with trader nicknames.

### What to Look For

#### Price Impact
- **Whale trades** (100-500 APPL) cause significant price slippage
- The **Impact Curve** chart shows execution prices vs trade size
- Large buys push price up; large sells push price down

#### Mean Reversion
- **MeanRev bots** trade when spot price deviates from TWAP
- TWAP chart shows 60-second time-weighted average price
- When price moves >1-2 standard deviations from TWAP, mean rev bots trade

#### LP Metrics
- **Impermanent Loss**: Loss from price divergence (always ≤ 0)
- **Fees Earned**: 0.30% of all swap volume accumulates in pool
- **Net PnL**: Fees + IL (fees offset impermanent loss)

#### Performance Analytics
- Navigate to `/performance` to see:
  - **Sharpe Ratio**: Risk-adjusted returns (rf = 0)
  - **Volatility**: Standard deviation of returns
  - **Max Drawdown**: Worst peak-to-trough loss
  - **Equity Curves**: Cumulative PnL over time per account

### Manual Trading

- Use the **Trading Panel** to manually trade as the "User" account
- Buy APPL with ETH or sell APPL for ETH
- Trades execute on-chain and appear in the Blotter

## Smart Contracts

### AppleToken.sol

Standard ERC20 token with:
- Owner-only `mint()` function
- 18 decimals
- Used as the base asset in the AMM pair (APPL/ETH)

### AppleAMM.sol

Constant product AMM (x * y = k) with the following features:

#### Core Mechanics
- **Liquidity Provision**: `addLiquidity()` and `removeLiquidity()`
  - First deposit uses geometric mean: `LP tokens = sqrt(appleAmount * ethAmount)`
  - Subsequent deposits maintain reserve ratio
  - LP tokens represent pro-rata share of pool

- **Swaps**: `swapETHForApples()` and `swapApplesForETH()`
  - 0.30% fee (30 bps) on all swaps
  - Fee remains in pool (benefits all LPs)
  - Slippage protection via `minOut` parameters
  - Constant product formula: `amountOut = (amountIn * reserveOut) / (reserveIn + amountIn)`

#### View Functions
- `getReserves()`: Current pool reserves
- `getSpotPrice()`: Current price (ETH per APPL, scaled by 1e18)
- `getTWAP()`: Time-weighted average price (60-second window)
- `getVolatility()`: Standard deviation of returns
- `getTotalFees()`: Cumulative fees collected

#### Security Features
- **ReentrancyGuard**: All state-changing functions protected
- **Safe ETH Transfers**: Uses `Address.sendValue()` (never `.call{value:}`)
- **Slippage Protection**: All swaps require minimum output amounts

## Trading Bots

The simulation includes 22 accounts with distinct trading behaviors:

### Whale Bots (3 accounts: Whale1, Whale2, Whale3)

**Strategy**: Large random trades that visibly move the market

- **Trade Size**: 100-500 APPL (random 10-100% of max)
- **Frequency**: Every 12-25 seconds
- **Direction**: Random (buy or sell)
- **Starting Position**: Pre-funded with APPL (1000, 700, 500 APPL respectively)
- **Purpose**: Create market volatility and price impact

**Implementation**: `bots/whale.go`
- Random delay between trades
- Random trade size within configured bounds
- Random direction (weighted by current position)

### Retail Bots (15 accounts: Retail1-Retail15)

**Strategy**: Small, frequent noise trades

- **Trade Size**: 1-10 APPL (random 10-100% of max)
- **Frequency**: Every 1-4 seconds
- **Direction**: Random
- **Starting Position**: 40 APPL each
- **Purpose**: Add market noise and volume

**Implementation**: `bots/retail.go`
- Very short delays (1-4 seconds)
- Small trade sizes
- Random direction

### Mean Reversion Bots (3 accounts: MeanRev1, MeanRev2, MeanRev3)

**Strategy**: Trade when price deviates from TWAP using EWMA

- **Trade Size**: Fixed (50, 75, 100 APPL respectively)
- **Frequency**: On z-score triggers (10-20 seconds typical)
- **Direction**: 
  - Buy when price < TWAP (oversold)
  - Sell when price > TWAP (overbought)
- **Starting Position**: 100, 250, 500 APPL respectively

**EWMA Parameters**:
- **MeanRev1**: Fast (25 trades half-life), trades at z = 0.75, 1.0, 1.25
- **MeanRev2**: Medium (50 trades half-life), trades at z = 1.0, 1.5, 2.0
- **MeanRev3**: Slow (87 trades half-life), trades at z = 2.0, 2.5

**Implementation**: `bots/meanrev.go`
- Listens to `TradeFlow` events (price before/after, reserves)
- Calculates liquidity-normalized log returns
- Maintains EWMA state per bot
- Computes z-score: `(current_return - ewma_mean) / ewma_std`
- Trades when z-score crosses threshold levels
- Resets EWMA after trade (MeanRev1) or when z crosses zero (MeanRev2, MeanRev3)

### User Account (1 account: User)

**Strategy**: Manual trading via frontend

- **Purpose**: Allows user to trade manually from the UI
- **Starting Position**: 750 APPL, 1,000 ETH
- **Implementation**: Frontend `TradingPanel` → REST API → `executor` → on-chain transaction

## Metrics & Analytics

### Price Metrics

All price metrics are computed in the Go backend (frontend only displays):

- **Spot Price**: Current AMM price (ETH per APPL)
- **OHLC Candles**: 5-second intervals (Open, High, Low, Close)
- **TWAP**: Time-weighted average price (60-second window)
- **Volatility**: Standard deviation of returns (rolling window)
- **Rolling Std Dev**: 100-trade window (only appears after 100 trades)

**Implementation**: `metrics/price.go`

### LP Metrics

Liquidity provider economics tracked in real-time:

- **Current Reserves**: APPL and ETH in pool
- **LP Value**: Current value of LP position (in ETH)
- **HODL Value**: Value if LP had just held initial tokens
- **Impermanent Loss**: `LP Value - HODL Value` (always ≤ 0)
- **Fees Earned**: Cumulative fees from swaps (read from contract)
- **Net PnL**: `Fees + IL` (fees offset impermanent loss)

**Implementation**: `metrics/lp.go`
- Initial state set when session starts
- Updated every 2 seconds from contract reserves and fees
- Historical snapshots for charts

### Account Performance (TradFi-style)

Per-account metrics computed from trade history:

- **Total Return**: Cumulative PnL (realized + unrealized)
- **Sharpe Ratio**: `(mean_return - rf) / std_dev` (rf = 0)
- **Volatility**: Standard deviation of returns
- **Max Drawdown**: Worst peak-to-trough loss
- **Win Rate**: Percentage of profitable trades
- **Trade Count**: Total number of trades
- **Equity Curve**: Cumulative PnL over time

**Implementation**: `metrics/account.go`
- Tracks equity after each trade
- Computes returns (log returns for volatility)
- Calculates Sharpe using rolling window
- Max drawdown from equity curve

### Price Impact

Execution prices vs trade size:

- **Buy Impact**: Price when buying different sizes
- **Sell Impact**: Price when selling different sizes
- **Visualization**: `ImpactCurve` component shows curves

**Implementation**: `metrics/impact.go`
- Simulates trades of different sizes
- Calculates execution prices using constant product formula
- Updates when reserves change

## API Endpoints

### REST API

All endpoints return JSON and are served on `localhost:8080`:

#### Session Management

- **POST** `/session/start`
  - Start simulation session
  - Optional body: `{"duration": 300}` (seconds)
  - Returns: `{"status": "started"}`

- **POST** `/session/stop`
  - Stop simulation (cancels all bot goroutines)
  - Returns: `{"status": "stopped"}`

- **POST** `/session/reset`
  - Reset session state (clears trades, events)
  - Returns: `{"status": "reset"}`

- **GET** `/session/state`
  - Get current session state
  - Returns: `{"status": "running", "startedAt": "...", "duration": 300}`

#### Account Data

- **GET** `/accounts`
  - List all accounts
  - Returns: Array of account objects with nickname, address, type

- **GET** `/accounts/:nickname/performance`
  - Get performance metrics for an account
  - Returns: `{totalReturn, sharpe, volatility, maxDrawdown, winRate, tradeCount, equityCurve}`

#### Market Data

- **GET** `/candles`
  - Get OHLC candle history
  - Query params: `?limit=100` (default: 1000)
  - Returns: Array of candle objects

- **GET** `/trades`
  - Get trade history
  - Query params: `?limit=100` (default: 1000)
  - Returns: Array of trade objects

- **GET** `/lp/metrics`
  - Get LP metrics
  - Returns: `{currentApples, currentETH, lpValue, hodlValue, impermanentLoss, feesEarned, netPnL, ...}`

- **GET** `/impact-curve`
  - Get price impact curve data
  - Returns: `{buy: [...], sell: [...]}`

- **GET** `/events`
  - Get key events
  - Query params: `?limit=20` (default: 50)
  - Returns: Array of event objects

#### User Trading

- **POST** `/user/buy`
  - Buy APPL with ETH (User account)
  - Body: `{"ethAmount": "1.0"}` (ETH amount as string)
  - Returns: `{txHash, applesOut, price, ...}`

- **POST** `/user/sell`
  - Sell APPL for ETH (User account)
  - Body: `{"appleAmount": "100.0"}` (APPL amount as string)
  - Returns: `{txHash, ethOut, price, ...}`

### WebSocket API

**Endpoint**: `ws://localhost:8080/stream`

Connects to real-time data stream. Messages are JSON with `type` and `data` fields:

- **`session_state`**: Session state updates
  ```json
  {"type": "session_state", "data": {"status": "running", "startedAt": "...", "duration": 300}}
  ```

- **`trade`**: New trade executed
  ```json
  {"type": "trade", "data": {"txHash": "...", "nickname": "Whale1", "isBuy": true, "amountIn": "...", ...}}
  ```

- **`trades`**: Bulk trade update (on connection)
  ```json
  {"type": "trades", "data": [...]}
  ```

- **`price`**: New candle/price update
  ```json
  {"type": "price", "data": {"timestamp": "...", "open": 1.0, "high": 1.05, "low": 0.95, "close": 1.02}}
  ```

- **`lp_metrics`**: LP metrics update
  ```json
  {"type": "lp_metrics", "data": {"currentApples": 10000, "currentETH": 10000, "impermanentLoss": -0.5, ...}}
  ```

- **`key_event`**: Key event (large trade, price move)
  ```json
  {"type": "key_event", "data": {"timestamp": "...", "type": "strategy_trigger", "description": "...", "severity": "warning"}}
  ```

- **`events`**: Bulk event update (on connection)
  ```json
  {"type": "events", "data": [...]}
  ```

## Development

### Running Contract Tests

```bash
cd contracts && forge test -vvv
```

### Building Backend

```bash
cd backend && go build ./...
```

### Type Checking Frontend

```bash
cd frontend && npx tsc --noEmit
```

### Regenerating Go Bindings

After modifying Solidity contracts:

```bash
make bindings
# or: ./scripts/generate-bindings.sh
```

This:
1. Compiles contracts with `forge build`
2. Extracts ABIs from `out/`
3. Generates Go bindings with `abigen`
4. Writes to `backend/internal/contracts/`

### Project Philosophy

This is a **DEMO-FIRST** project. The code prioritizes:

- **Visual Clarity**: What happens should be immediately visible in the UI
- **Conceptual Correctness**: Metrics and mechanics should be mathematically sound
- **Readability**: Code should be clear and well-commented
- **Simplicity**: Avoid unnecessary abstractions or frameworks

**What this project is NOT**:
- A production-ready system (no auth, no persistence, no error recovery)
- A long-lived open-source project (no CI/CD, no Docker, no Kubernetes)
- An enterprise application (no repositories, no DI containers, no factories)

## Key Concepts

### Constant Product AMM

The AMM uses the formula: `x * y = k`

- `x` = APPL reserve
- `y` = ETH reserve
- `k` = constant (changes when liquidity added/removed)

For a swap:
- Input: `amountIn` (after fee)
- Output: `amountOut = (amountIn * reserveOut) / (reserveIn + amountIn)`
- New reserves: `reserveIn += amountIn`, `reserveOut -= amountOut`

### Impermanent Loss

Impermanent loss occurs when the price of the pair diverges from the initial price:

```
IL = LP_Value - HODL_Value
```

Where:
- `LP_Value` = Current value of LP position (in ETH)
- `HODL_Value` = Value if LP had just held initial tokens

IL is always ≤ 0 (LP always loses vs HODL when price moves). Fees offset this loss.

### Mean Reversion Strategy

Mean reversion bots use EWMA (Exponential Weighted Moving Average) on liquidity-normalized log returns:

1. **Trade Flow Events**: Each trade emits price before/after and reserves
2. **Log Returns**: `log(price_after / price_before)`
3. **Liquidity Normalization**: Adjust for pool depth (larger pools = smaller impact)
4. **EWMA**: Exponential weighted moving average of returns
5. **Z-Score**: `(current_return - ewma_mean) / ewma_std`
6. **Trading**: Buy when z < -threshold, sell when z > threshold

## Security Features

The smart contracts implement several security best practices:

- **Reentrancy Protection**: All state-changing functions use OpenZeppelin's `ReentrancyGuard`
- **Safe ETH Transfers**: Uses `Address.sendValue()` instead of low-level `.call{value:}`
- **Slippage Protection**: All swap functions require minimum output amounts
- **Input Validation**: All functions validate inputs (non-zero amounts, etc.)

## License

MIT

---

*This is a portfolio demonstration project. Not intended for production use.*
