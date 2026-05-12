# ETH-AMM-SIM

Local AMM simulation designed as a portfolio-grade demo for a Solutions Engineer in crypto/blockchain. It wires Solidity contracts to a Go execution engine and a Vite + React SPA dashboard so you can explain protocol mechanics, real-time data pipelines, and system design tradeoffs in one cohesive project.

## Portfolio demo package

- Loom guide: [`DEMO_GUIDE.md`](./DEMO_GUIDE.md)
- Recommended launcher: `make demo-120`
- Recommended video length: 120 seconds primary, 180 seconds deep cut
- Demo focus: AMM execution lab
- Best GIF loop: start bots -> whale trade -> price/TWAP/LP panels move
- Thumbnail hook: `AMM EXECUTION LAB`
- Loom link: TBD
- Preview MP4/GIF: TBD
- Demo notes script: [`.hermes/workspace/projects/eth-amm-sim/DEMO.md`](file:///Users/danmagro/.hermes/workspace/projects/eth-amm-sim/DEMO.md)


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
Vite + React SPA (localhost:3000)
  - Dashboard (charts + controls)
  - REST fetch for snapshots (via `/api/*` → Go in dev/preview)
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
- **WebSocket**: [ws://localhost:8080/stream](ws://localhost:8080/stream) for real-time trades, prices, and events.
- **Chain RPC**: [http://localhost:8545](http://localhost:8545) (Anvil local chain).

## Simulation Flow
1. Deploy contracts via Foundry (Deploy.s.sol writes broadcast JSON).
2. Backend starts and loads contract addresses from broadcast output.
3. Bots are created based on `backend/internal/config/accounts.go`.
4. On-chain trades emit updates -> in-memory store -> metrics pipeline.
5. WebSocket broadcasts push live updates to the dashboard.

**Session / bot lifecycle (multi-session without restarting the Go process):** [docs/SESSION_BOT_LIFECYCLE.md](docs/SESSION_BOT_LIFECYCLE.md).

## Feature Highlights
- AppleToken (ERC20) + AppleAMM (constant product, 0.30% fee) on Anvil.
- Go executor that submits swaps and computes trade details.
- Trading bots:
  - Retail (15) small random trades
  - Whale (3) large random trades
  - MeanRev (3) EWMA mean reversion on trade-flow events
- Session control (start/pause/resume/stop/reset) with per-session bot lifecycle.
- Metrics computed in Go:
  - 5s OHLC candles, 60s TWAP, volatility from observed returns
  - LP metrics (IL, fees earned, net PnL)
  - Account performance (equity curve, Sharpe, drawdown, win rate)
- Dashboard components: Price, TWAP, Impact Curve (hover readout), Blotter, LP Stats, Key Events (clickable trade rows), AMM Details, User Trading (buy/sell).
- Account analytics live on the dedicated `/performance` page.

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
- `frontend/` Vite + React dashboard
- `scripts/` deployment, bindings, and dev tooling
- `Makefile` orchestration targets

## Prerequisites
- **Foundry** (forge, anvil)
- **Go** 1.25.x — **recommended:** install the patch release matching `backend/go.mod` (see `go` line), or keep **`GOTOOLCHAIN=auto`** so the toolchain can upgrade itself. The repo root **`.vscode/settings.json`** sets `GOTOOLCHAIN=auto` for the Go extension / `gopls`, which avoids IDE **packages.Load** failures when the module requires a newer patch than your default `go` binary.
- **Node.js** **20.x or 22.x LTS** (npm). **Avoid Node 25+ for the frontend** until Vite/Rollup fully support it — Node 25 can throw **`ERR_INVALID_PACKAGE_CONFIG`** on `rollup/dist/es/package.json`. Use **`frontend/.nvmrc`** (`nvm use` / `fnm use`).
- **Vite + React** — fast local dev server and static production build under `frontend/dist/` (see `frontend/package.json`).
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

If `make up` would block waiting for `tmux attach` (e.g. from a script or agent), use **`ETH_AMM_SIM_SKIP_ATTACH=1 make up`**, then attach yourself: **`tmux attach -t eth-amm-sim`**.

To stop everything:
```bash
make down
```

Open [http://localhost:3000](http://localhost:3000), click **Start**, and watch metrics stream in.

## Web / LAN / Vercel
- **Same machine:** REST goes through Vite’s **`/api/*` dev/preview proxy** to Go on `:8080` (no CORS issues). WebSocket uses **`ws://<host>:8080/stream`**, so it tracks the hostname you used in the browser (works for `localhost` or your LAN IP).
- **Another device on your Wi‑Fi:** Start stack as usual, then open `http://<your-computer-LAN-IP>:3000`. Keep the Go backend on the same machine (`:8080` is already reachable on the LAN).
- **Vercel (UI only):** Link the repo with **Root Directory = `frontend`**, or from `frontend/` run `vercel`. The build is a **static SPA** (`dist/`); configure **`VITE_API_URL`** (and **`VITE_WS_URL`** as **`wss://…/stream`**) for your public Go backend so the browser can reach REST/WebSocket from an HTTPS origin (or put both UI and API behind a gateway that preserves same-origin `/api`). See `frontend/.env.example`.

## Manual Run (4 terminals)
If you prefer to run each service yourself:

```bash
make anvil
make deploy
make bindings
make backend
make frontend
```

## Frontend checks

- **`npm run build`** — runs **`vite build`** only (fast production bundle to `dist/`).
- **`npm run lint`** — TypeScript check via **`node ./node_modules/typescript/lib/tsc.js --noEmit`** (avoids flaky `node_modules/.bin/tsc` shims that can error with **`Operation timed out`** on some macOS setups).
- **`npm run verify`** — **`lint` then `build`** (use before commits / when you want both).

## Useful Make Targets
- `make setup` - install deps and generate bindings
- `make anvil` - start local chain on :8545
- `make deploy` - deploy contracts and write broadcast JSON
- `make bindings` - generate Go bindings from contract ABIs
- `make backend` - run Go simulator on :8080
- `make frontend` - run Vite dev server on **0.0.0.0:3000** (open [http://localhost:3000](http://localhost:3000) after the Local URL prints; Makefile prints a reminder)
- `make frontend-fresh` - **`rm -rf frontend/dist frontend/.vite`** then dev (use after stale chunk/proxy glitches)
- `make kill-all` - free ports 8545, 8080, 3000-3004
- `make test-contracts` - run Foundry tests

## Configuration & Tuning
- **Accounts & bot params**: `backend/internal/config/accounts.go`
- **AMM constants & thresholds**: `backend/internal/config/amm.go`
- **Chain + session defaults**: `backend/internal/config/config.go`

### Backend security (optional, for non-local demos)

By default the Go server uses **permissive CORS and WebSocket origin checks** so local `localhost` / LAN demos work without extra setup.

To restrict browser origins, set a comma-separated allowlist:

```bash
export ETH_AMM_SIM_ALLOWED_ORIGINS="http://localhost:3000,http://127.0.0.1:3000"
```

When this variable is set, requests with an `Origin` header not in the list receive **403** (REST and WebSocket upgrade).

### Health endpoints

- `GET /healthz` — process liveness: `uptime_seconds`, `ws_clients`, `broadcast_queue_len`
- `GET /readyz` — same payload once the HTTP server has started (503 only if called before startup, which is unusual in normal operation)

Common knobs:
- Session duration
- Bot frequency and max trade size
- Pool seed reserves
- Volatility and LP metrics behavior

## API Endpoints (REST)
- GET  `/healthz` — liveness JSON (`uptime_seconds`, `ws_clients`, `broadcast_queue_len`)
- GET  `/readyz` — readiness JSON (same shape once the server has started)
- POST `/session/start` (optional body: `{ "duration": 300 }`)
- POST `/session/pause`
- POST `/session/resume`
- POST `/session/stop`
- POST `/session/reset` (query: `?mode=soft|hard|reseed`; legacy `?hard=true` and `?reseed=true` still supported)
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
- [ws://localhost:8080/stream](ws://localhost:8080/stream)
- Message types: `trade`, `price`, `lp_metrics`, `key_event`, `session_state`, `account_update`, `trades`, `candles`, `events`.

If logs show **broadcast queue full** or the UI stops updating while the backend still runs, you usually have a **stale or slow WebSocket client** (background browser tab, full TCP buffers). Close extra tabs, hard-refresh the dashboard, or set `ETH_AMM_SIM_ALLOWED_ORIGINS` correctly. See [docs/SESSION_BOT_LIFECYCLE.md](docs/SESSION_BOT_LIFECYCLE.md) for session vs process lifetime.

## Demo Walkthrough (5 minutes)
1. `make up` and open [http://localhost:3000](http://localhost:3000).
2. Click **Start** and highlight the session timer + live trade blotter.
3. Point out price impact from whale trades on the chart, then hover the AMM reserve curve to show local price differences.
4. Click a trade in **Key Events** and show **AMM Details** (before/after reserves + spot impact).
5. Open **LP Stats** to explain impermanent loss vs. fees earned.
6. Open [http://localhost:3000/performance](http://localhost:3000/performance) to show account equity curves and Sharpe.

## Troubleshooting
- **`go build` fails with `error obtaining VCS status: exit status 128`**: Go tries to stamp binaries with Git metadata; if `git` errors (sandbox, partial clone), build with **`GOFLAGS=-buildvcs=false`** — **`make backend`** / **`make clean`** already set this via the Makefile’s **`BACKEND_TOOLCHAIN`**.
- **`Invalid package config` / `ERR_INVALID_PACKAGE_CONFIG` under `rollup/dist/es` when running `npm run dev`**: You are almost certainly on **Node 25+**. Switch to **Node 20 or 22** (see **`frontend/.nvmrc`**), then `cd frontend && rm -rf node_modules && npm install`.
- **`npm run build` stalls / `Operation timed out` on `node_modules/.bin/tsc`**: `npm run build` no longer invokes `tsc` (only **`vite build`**). Run types separately with **`npm run lint`**, or both with **`npm run verify`**. If `npm run lint` still times out, use the system Node binary (not an IDE‑bundled helper): `PATH="/opt/homebrew/bin:$PATH" npm run lint`.
- **“It built as HTML” / opening the app from Finder doesn’t work**: `vite build` writes static assets under **`frontend/dist/`**. Use **`npm run dev`** (development), or **`npm run build` then `npm run preview`** (production-like static server with `/api` proxy). Open [http://localhost:3000](http://localhost:3000) in a browser — client routing and `/api` proxy expect that origin plus the Go backend on `:8080`.
- **Ports in use**: run `make kill-all` then retry.
- **No data in UI**: confirm backend is running on `:8080` and contracts were deployed.
- **Reset does not reseed / price not back near 1.0**: session must be idle (not running). Use **Stop/Pause** first, then **Reseed** reset.
- **Deploy failed**: if manual redeploy runs on a dirty chain, LP account funds can be exhausted; use reseed reset (which runs `anvil_reset` before deploy) or restart from fresh Anvil.
- **Deploy fails with `max fee per gas less than block base fee`**: update scripts and retry; repo `scripts/deploy.sh` now pins local gas price for Anvil deploys.
- **Deploy stuck** (`Waiting for pending transactions`, receipt count not increasing, huge “seconds” in the progress bar): Anvil and Forge got out of sync. In tmux: **Ctrl+C** in the `deploy` window, switch to `anvil` and stop it (**Ctrl+C**), then run `make kill-all`, `make down`, and `make up` again. `scripts/deploy.sh` uses **`forge script --slow`** so each transaction confirms before the next (reduces this class of hang).
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
- User execution baseline after startup/reset/reseed: 5,000 ETH + 5,000 APPL
- Session analytics normalization baseline: 1,000 ETH + 1,000 APPL (for cross-account return comparability)
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
