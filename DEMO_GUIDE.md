# Loom Demo Guide: eth-amm-sim

Status: optimized for a natural 180-second project walkthrough.
Portfolio tag: sim
Primary repo URL: https://github.com/DanDo385/eth-amm-sim

## Core story

Problem:
AMMs are usually explained with static x*y=k diagrams, but real trading systems are not static. Trades arrive over time, bots react, liquidity shifts, prices move, and LP risk changes as state changes.

Solution:
This project turns AMM mechanics into a running local protocol lab: a local EVM-compatible chain, immutable Solidity contracts, a Go/geth execution layer, bot activity, WebSocket streaming, and a live dashboard that makes market state observable.

One-liner:
A production-shaped DeFi simulation where Solidity contracts, Go execution infrastructure, and a real-time dashboard expose how AMM trades move price, liquidity, TWAP, and LP risk.

Positioning:
This is the `sim` lane. It is not a cartoon explainer and it is not a mainnet trading bot. It is a controlled simulation that proves the software shape of DeFi execution infrastructure.

## What the Loom should feel like

Natural, not over-produced.
Record it as if you are walking an interviewer through the actual system you built:
1. Here is the problem.
2. Here is the architecture.
3. Here is the app running.
4. Here is the causal event.
5. Here is how it maps to real trading systems.
6. Here is what I would harden for production.

The whale trade is the visual moment, not the whole thesis. The stronger thesis is: local chain -> contracts -> backend execution -> real-time observability -> risk framing.

## What to run before recording

1. Start the Loom launcher:

   cd /Users/openclaw/eth-amm-sim
   make demo-120

2. Open the app in a clean browser window at `http://localhost:3000`.
3. Zoom browser to 90-100% so the Loom Demo Director, Show Architecture panel, chart, blotter, and LP panels fit.
4. Reset the session before recording.
5. Click Start. The dashboard now defaults to a 120-second session for recording.
6. Let a few normal bot/retail trades appear first.
7. Around 1:25, click **Shock Pool: sell 500 APPL** for the deterministic price-impact moment.
8. Keep the trade blotter, price chart, TWAP, reserves, and LP metrics visible when the shock lands.
9. Close by comparing simulation boundaries to real market infrastructure.

## Step-by-step 180-second Loom

### 0:00-0:15 -- Hook: why this exists

Show the dashboard, not the terminal.
Say:
"This is an AMM execution lab. I built it because most AMM demos stop at the formula, but real DeFi systems are about state changing over time: trades, bots, liquidity, price impact, TWAP, and LP exposure."

Point at:
- Main chart.
- Blotter.
- Pool metrics.

### 0:15-0:40 -- Architecture in one breath

Show the Loom Demo Director / architecture text.
Say:
"The system spins up a local EVM-compatible chain, deploys immutable Solidity contracts with Foundry, drives execution through Go/geth, then streams the resulting state into the frontend. The point is not fake money. The point is the production-shaped integration pattern."

Point at:
- Local EVM / Foundry / Solidity label.
- Go/geth backend label.
- WebSocket / dashboard flow.

### 0:40-1:05 -- Normal market behavior

Start or resume the simulation.
Say:
"Before stress, this is normal flow. Smaller trades enter the pool. The constant-product curve updates price, reserves shift, and the UI shows the state change immediately."

Point at:
- Trade rows appearing.
- Price moving modestly.
- Reserve/liquidity panels updating.

### 1:05-1:35 -- The visual event

Trigger or wait for the whale trade.
Say:
"Now the visual moment: a large trade hits the pool. Watch the blotter row, price chart, TWAP, and LP metrics all move together. I designed the UI to show causality, not just numbers."

Point at:
- Purple whale row pulse.
- Price impact.
- TWAP movement.
- LP risk / pool metrics.

### 1:35-2:15 -- Compare to real trades / real infrastructure

Stay on the dashboard. Do not go full terminal unless needed.
Say:
"In a real trade, the risk boundary changes, but the shape is familiar: signed transactions, RPC or node access, nonce and gas handling, logs, event indexing, bot policy, observability, and risk monitoring. Production would add audited contracts, proper key management, circuit breakers, alerting, and testnet/mainnet deployment controls. Same software shape, different risk boundary."

If useful, mention:
- This is comparable to a controlled version of seeing a large swap affect pool price and LP exposure.
- The demo isolates the mechanism so it can be explained without mainnet noise.

### 2:15-2:45 -- Design decisions and tech stack

Say:
"I used Solidity for the invariant and contract boundary, Foundry/Anvil for fast local deployment, Go because concurrent bot execution and RPC polling are natural with goroutines, and Next.js because the dashboard needs to teach the system while the system is running."

Point at:
- Solidity / Foundry / Anvil.
- Go/geth backend.
- Next.js dashboard.

### 2:45-3:00 -- Close with hiring signal

Say:
"The hiring signal is full-stack protocol fluency: contracts, execution infrastructure, event/state handling, frontend observability, and market intuition in one small but realistic lab."

End on:
- Dashboard after the shock has stabilized.
- Whale row still visible if possible.

## Short 30-second cut

0:00-0:05
"AMMs are not just formulas. They are state machines under live trading pressure."

0:05-0:15
Show normal trades, chart, reserves.
"This local EVM lab runs Solidity contracts, Go/geth execution, bots, and a live dashboard."

0:15-0:25
Show whale trade.
"A large trade hits the pool, price moves, TWAP shifts, and LP risk updates in real time."

0:25-0:30
"Controlled simulation, production-shaped architecture."

## GIF / MP4 preview plan

Length: 8-12 seconds.
Loop:
1. Start simulation button.
2. A few normal trades appear.
3. Whale row pulses purple.
4. Price/TWAP/LP panels move.
5. End on the dashboard stabilized after the shock.

Caption baked into preview:
"AMM sim: local EVM trades -> price impact -> LP risk"

Prefer MP4/WebM over GIF if magro.dev supports it.

## Thumbnail plan

Title:
"AMM EXECUTION LAB"

Subtitle:
"Solidity + Go/geth + real-time DeFi simulation"

Visual composition:
Dark trading dashboard, purple whale-row pulse, price chart movement, small badges for Solidity, Go, Foundry, Anvil.

## Speaking guidance

Sound like a builder, not a tour guide.
Do not over-apologize for it being local.
Say: "Same software shape, different risk boundary."
Do not say: "This is just a simulation."
Say: "This is a controlled protocol lab."

Avoid long terminal time. Use terminal only briefly to prove local chain/backend if needed, then return to the dashboard.
