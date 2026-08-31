# Deep Analysis — eth-amm-sim

- **Generated:** 2026-08-31T16:43:28Z
- **Commit SHA:** `ec8e3d9e0c3bf1b3e836539716b7671408c76154`
  ("Even out file-level comments across frontend, backend, and contracts.",
  2026-08-26; checked out as detached HEAD)
- **Working tree:** dirty — modified `README.md`, `frontend/src/components/Toast.tsx`;
  untracked `scripts/start-staging-anvil.sh` and `public/` media assets. None
  affect the structural analysis below.
- **Analyzer:** oss-project-analyzer skill (structural pass; dynamic profiling not run)

---

## 1. Language / LOC summary

| Language     | Files | Code   | Role |
|--------------|------:|-------:|------|
| Go           |    37 |  7,143 | Concurrent simulator: chain I/O, bots, metrics, REST/WS server |
| TypeScript   |    27 |  2,907 | Vite + React 18 observability SPA |
| Solidity     |     5 |    599 | `AppleToken` (ERC20) + `AppleAMM` (constant-product) + deploy script |
| Shell        |     6 |    466 | Deploy, bindings generation, dev orchestration |
| JSON/TOML/MD/etc. | 20 | ~3,700 | Lockfiles, ABI fixtures, Foundry/Vite config, docs |

Application code total ≈ **11,100 LOC** (Go + TS + Solidity + shell).
Test code: 696 Go LOC (24 funcs) + 396 Solidity LOC (22 funcs).
Full table in [`metrics.md`](./metrics.md). Third-party (`contracts/lib/`,
`frontend/node_modules/`) excluded.

---

## 2. Top-level module map

### Backend (`backend/`, module `eth-amm-sim`, Go 1.25)

| Package | Files | Responsibility | Key types |
|---------|-------|----------------|-----------|
| `cmd/simulator` | `main.go` (~570 LOC) | Process entrypoint: load config, resolve contract addresses (env → broadcast JSON → hardcoded Anvil defaults), connect to Anvil, verify pool, build bots, register `OnTrade`, run price-poll + server + shutdown via `errgroup` | — |
| `internal/chain` | `client.go`, `accounts.go`, `nonce.go` (+1 test) | go-ethereum RPC wrapper, Anvil deterministic keys, per-account nonce manager | `Client`, `NonceManager` |
| `internal/contracts` | `apple_amm.go`, `apple_token.go` | `abigen`-style generated bindings | — |
| `internal/config` | `config.go`, `accounts.go`, `amm.go`, `accounts.go` (+2 tests) | **Single source of truth** for the 30-account roster, bot type/params, AMM seeding constants, thresholds | `Account`, `BotType`, `Config` |
| `internal/engine` | `executor.go`, `orchestrator.go`, `session.go`, `types.go` (+2 tests) | `Executor`: ABI-encode/sign/send swaps, derive amountOut & price, emit `Trade`. `Orchestrator`: bot goroutine lifecycle via `errgroup`. `Session`: idle→running→completed state machine + timer | `Executor`, `Orchestrator`, `Session`, `Trade` |
| `internal/bots` | `base.go`, `bot.go`, `whale.go`, `retail.go`, `meanrev.go` (+1 test) | Strategy goroutines. Retail = random noise (2-16); Whale = large infrequent impact (17-19); MeanRev = EWMA z-score reversion (20-22) | `Bot` iface, `BaseBot`, `WhaleBot`, `RetailBot`, `MeanRevBot` |
| `internal/metrics` | `price.go`, `lp.go`, `ewma.go`, `impact.go`, `account.go`, `trade_flow.go`, `price_provider.go` (+1 test) | Price/TWAP/volatility candles, LP impermanent-loss/fee/PnL, EWMA return tracker, price-impact curve, per-account mark-to-market, trade-flow pub/sub | `PriceMetrics`, `LPMetrics`, `ImpactCurve`, `AccountMetricsManager`, `TradeFlowTracker` |
| `internal/store` | `memory.go` | `MemoryStore`: composes all metrics + trade blotter + key events; implements `metrics.PriceDataStore` | `MemoryStore`, `KeyEvent` |
| `internal/server` | `server.go`, `handlers.go`, `broadcast.go`, `security.go` (+2 tests) | gorilla/mux REST (19 routes) + gorilla/websocket `/stream`; CORS + rate limiting; fan-out broadcaster with drop logging | `Server` |

### Contracts (`contracts/`, Foundry, Solidity ^0.8.20)

- `AppleToken.sol` (30 LOC) — OpenZeppelin `ERC20` + `Ownable`, owner-mint only.
- `AppleAMM.sol` (394 LOC) — constant-product `x*y=k`, 30 bps fee retained in
  pool, LP-token accounting, `ReentrancyGuard`, `SafeERC20`, custom errors,
  slippage-protected `swapETHForApples` / `swapApplesForETH`, `getSpotPrice`,
  `getReserves`, `getTotalFees`.
- `script/Deploy.s.sol` (279 LOC) — deploy both contracts, mint APPL to all 30
  Anvil accounts, normalize ETH balances (LP 25k+gas, User 5k, bots 1k), seed
  pool 25,000 APPL + 25,000 ETH (initial price 1.0), write broadcast JSON.

### Frontend (`frontend/src/`, Vite 5 + React 18.3 + TS 5.8)

- Entry `main.tsx` → `App.tsx` (`react-router-dom` v6): `ShellLayout` wraps
  `HomePage` (live dashboard) and `PerformancePage` (per-account analytics).
- **Hooks own all I/O** (per `AGENTS.md`): `useWebSocket` (persistent `/stream`
  with reconnect), `useSession`, `usePriceData`. `lib/backend.ts` resolves
  REST/WS URLs (env override → localhost → staging tunnel).
- 14 components, mostly presentational; charts via `lightweight-charts`.
- `types/index.ts` mirrors backend JSON wire types (manual sync).

### Build / ops

`Makefile` targets (`anvil`, `deploy`, `bindings`, `backend`, `frontend`, `up`,
`down`, `test-contracts`) + `scripts/*.sh`. `dev-up.sh` launches the full stack;
`generate-bindings.sh` regenerates Go bindings from ABIs. Frontend Node pinned
via `.nvmrc` and wrapped by `with-frontend-node.sh`.

---

## 3. Critical execution path

**Primary demo loop — bot trade to pixel:**

```
bot.Run(ctx)  [internal/bots/*.go]
  └► Executor.SwapETHForApples / SwapApplesForETH   [engine/executor.go]
       └► NonceManager.assign → types.SignTx(privkey) → Anvil eth_sendRawTransaction
       └► read LogSwap + constant-product math → build engine.Trade
  └► OnTrade callback (registered in main.go:163)
       ├► MemoryStore.RecordTrade
       ├► MemoryStore.RecordTradeFlow  → notifies MeanRev bots (EWMA z-score)
       ├► MemoryStore.RecordEvent      → large/critical trade thresholds
       └► server.BroadcastTrade / BroadcastEvent / BroadcastAccountUpdate
             └► WS /stream fan-out → SPA handleWSMessage
                   └► Blotter / KeyEvents / AccountMetrics / PriceChart re-render
```

**Parallel market-data loop (`main.pollPrices`, 2s ticker):**

```
Executor.GetSpotPrice / GetReserves / GetTotalFees   (RPC eth_call)
  └► MemoryStore.RecordPrice        → candle/TWAP/volatility update
  └► LPMetrics.UpdateState          → impermanent loss, fees earned, net PnL
  └► ImpactCurve.UpdateReserves     → 50-point slippage curve
  └► server.BroadcastPrice / BroadcastLPMetrics → WS → PriceChart/TWAPChart/LPStats/ImpactCurve
```

**Session lifecycle (control plane):**

```
SPA SessionControls → POST /session/start → handlers.handleSessionStart
  → Session.SetDuration? → Session.Start(ctx, duration)
    → Orchestrator.Start(ctx): errgroup spawns bot.Run for every non-LP/non-User account
  → status broadcast; timer or POST /session/stop cancels ctx → all bots exit → StatusCompleted
```

**Concurrency model:** one `errgroup` in `main` coordinates {price-poll,
HTTP server, shutdown watcher}; a second `errgroup` in `Orchestrator` owns the
~21 bot goroutines, all tied to the session context. Graceful shutdown on
SIGINT/SIGTERM: stop session → 500ms bot drain → 5s server shutdown timeout.
This directly satisfies the `AGENTS.md` "no orphaned workers" rule.

---

## 4. Dependency & coupling observations

1. **`config` is a deliberate hub.** `internal/config` is imported by `engine`,
   `bots`, `server`, `store`, `cmd`, and its Anvil addresses are re-derived
   independently by `Deploy.s.sol` and displayed by the frontend. Changing the
   account roster is a cross-language, cross-process edit with no compiler to
   catch drift between Go and Solidity. *(supported — verified imports; drift
   risk is an inference.)*

2. **Inverted `store ↔ engine` dependency.** `store` imports `engine` (stores
   `engine.Trade`), but `engine` never imports `store`. The write path is closed
   by a closure in `main.go` (`executor.OnTrade(func(t *engine.Trade){ ... })`).
   Clean, but it means `main.go` carries ~120 LOC of business logic (event
   severity classification, trade-flow gating) that arguably belongs in a
   package. *(verified — `grep` of imports; `main.go:163-208`.)*

3. **`store ↔ metrics` cycle broken by interface.** `MemoryStore` implements
   `metrics.PriceDataStore` (`GetCandles`/`GetTWAP`/`GetVolatility`), and
   `metrics.NewPriceProvider(memStore)` hands bots a read-only view. Idiomatic
   Go cycle-break. *(verified — `store/memory.go` doc comment + `main.go:135`.)*

4. **Two representations of the AMM ABI.** `executor.go` hand-packs calldata
   (`packSwapETHForApples`, …) *and* `internal/contracts/apple_amm.go` holds
   generated bindings. If the Solidity signature changes, both must be updated;
   only the generated one is refreshed by `generate-bindings.sh`. *(supported —
   both files present; executor doc comment says "Manually packed".)*

5. **Frontend/backend wire contract is hand-synced.** `frontend/src/types/index.ts`
   mirrors Go JSON tags; `AGENTS.md` explicitly instructs updating both together.
   No codegen or schema. *(verified — `AGENTS.md` "update README and frontend
   types".)*

6. **go-ethereum is a heavy transitive dependency.** `backend/go.mod` has 4
   direct requires but pulls ~35 indirect modules (gnark-crypto, blst, c-kzg,
   otel, gopsutil, and notably `ProjectZKM/Ziren` zkvm runtime). Build/CI cost
   and supply-chain surface are dominated by go-ethereum, not project code.
   *(verified — `go.mod`.)*

7. **Server God-object.** `server.Server` holds router, session, store,
   executor, WS client set, broadcast channel, and several mutexes; it is passed
   into the `OnTrade` closure and the poll loop. Central chokepoint but keeps the
   fan-out logic in one place. *(supported — `server.go` struct definition.)*

8. **Frontend runtime deps are minimal** (`react`, `react-dom`,
   `react-router-dom`, `lightweight-charts`) — low bundle/security surface;
   everything else is dev-only. *(verified — `frontend/package.json`.)*

---

## 5. Test evidence

| Suite | Result | Detail |
|-------|--------|--------|
| `go test ./...` (backend) | **PASS** | `chain` 2.57s, `config` 1.40s, `engine` 3.90s, `metrics` 3.43s, `server` 1.18s. `bots`, `store`, `contracts`, `cmd` have **no test files**. |
| `forge test` (contracts) | **PASS** | 22/22 — `AppleAMM.t.sol` 15 (invariant preservation, price impact, slippage protection, fee accumulation, multi-swap invariant, revert-on-empty-pool), `AppleToken.t.sol` 7 (mint auth, transfer, ownership). ~40ms. |
| Frontend | **not run** | `npm run verify` (tsc + build) per `AGENTS.md`; no unit tests present. |

**Coverage shape:**
- Well covered: AMM math/invariants (Solidity), EWMA math, session/orchestrator
  lifecycle, HTTP handlers, server security (CORS/rate-limit), config parsing,
  chain client.
- **Untested:** all bot strategy logic (`internal/bots/` — 0 test files despite
  ~850 LOC of trading heuristics), `MemoryStore` aggregation, `Executor`
  calldata packing / price derivation, `main.go` wiring & `OnTrade` classification,
  the entire frontend.
- No integration test spins up Anvil + backend together; contract/Go boundary is
  only exercised manually via `make up`.

---

## 6. Optimization hypotheses

Confidence legend: **verified** (confirmed from code/tests/runs) ·
**supported** (strong code evidence, not dynamically measured) ·
**unverified** (plausible, needs profiling/measurement).

### H1 — Bot strategy logic is the highest-value place to add tests · **verified**
`internal/bots/` has 0 test files but contains the demo's core behavior (whale
impact sizing, retail frequency, MeanRev z-score triggers with three tuned
half-lives). `AGENTS.md` demands "deterministic enough for repeatable demos" —
seam the RNG (`RandomDelay`/`RandomSize` in `base.go`) behind an interface and
table-test the MeanRev trigger boundaries. Low effort, directly de-risks the
demo narrative.

### H2 — Collapse the dual AMM ABI representation · **supported**
`executor.go` hand-packs calldata while `internal/contracts/*.go` holds generated
bindings for the same functions. Switching `Executor` to the generated bindings
(or deleting the unused bindings) removes a silent drift class and ~100+ LOC of
manual `abi.Arguments` packing. Verify which is actually on the call path first
(`packSwap*` appears to be — bindings may be dead code).

### H3 — Unbounded in-memory history growth · **supported**
`MemoryStore` doc comment references "unbounded price history and trade blotter
arrays" and `AGENTS.md`'s language rationale explicitly discusses per-entry
memory cost. For a 2-3 minute demo this is fine, but a long-running staging
instance (there is a `start-staging-anvil.sh` and a staging WS tunnel) will grow
without bound. Add a ring buffer / retention cap on `trades`, `events`, and
candle history. Effort: low; impact: prevents slow OOM on the always-on staging box.

### H4 — Price-poll does 3 sequential RPC round-trips every 2s · **unverified**
`pollPrices` calls `GetSpotPrice`, then `GetReserves`, then `GetTotalFees`
serially each tick (`main.go:458-532`); `GetSpotPrice` is derivable from
reserves. Collapsing to a single `getReserves` + `getTotalFees` call (or a
multicall) roughly halves poll-loop RPC traffic. Against local Anvil the latency
is negligible; matters only if pointed at a remote RPC. Needs measurement to
justify.

### H5 — WebSocket broadcast fan-out under many clients · **unverified**
`server.Server` has a single `broadcast chan interface{}` with "channel full"
drop logging and a `lastBroadcastDropLog` rate limiter — implying observed
backpressure. Each trade currently triggers up to 3 separate broadcasts
(`BroadcastTrade` + `BroadcastEvent` + `BroadcastAccountUpdate`). Coalescing
per-tick state into one framed message would cut serialization and lock
contention. Profile with N simulated WS clients before investing.

### H6 — `main.go` carries package-worthy logic · **supported**
The `OnTrade` closure (event severity thresholds, trade-flow gating, key-event
construction) and the `pollPrices` price-move classifier are ~200 LOC of
business rules living in `package main`, untestable without a full wiring
harness. Extracting an `internal/events` (or `internal/pipeline`) package would
make H1-style testing possible and shrink `main` to pure composition.

### H7 — go-ethereum dominates build/dependency surface · **verified**
4 direct deps pull ~35 indirect modules including a zkVM runtime
(`ProjectZKM/Ziren`) not obviously needed for a local-Anvil simulator. Audit
whether a lighter RPC client (e.g. `ethclient` alone, or a minimal JSON-RPC
wrapper — the project already hand-packs ABIs) could drop most of the transitive
tree. High effort, meaningful CI-time and supply-chain payoff; verify feature
usage first.

---

## 7. Structural risk register (non-optimization)

| Risk | Severity | Evidence |
|------|----------|----------|
| Go↔Solidity account-roster drift (no shared schema) | Medium | H1/obs. 1 |
| Frontend wire types hand-synced to Go JSON | Medium | obs. 5, `AGENTS.md` |
| No end-to-end (Anvil+backend) integration test | Medium | §5 |
| Hardcoded Anvil private keys in `Deploy.s.sol` / `chain/accounts.go` | Low (by design, local only) | `Deploy.s.sol:25` |
| Detached-HEAD working state, dirty tree | Low | `git status` |
