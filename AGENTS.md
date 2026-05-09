# AGENTS.md

Canonical agent instructions for this repository.

Do not add separate root-level `.cursorrules`, `CLAUDE.md`, `GEMINI.md`, or tool-specific instruction files unless a tool absolutely requires a shim. If a shim is required, it must only point back to this file and must not contain independent rules.

## Related documentation

- **[README.md](./README.md)** — local setup, Makefile targets, HTTP/WebSocket endpoints, and current system behavior. Keep it aligned when you change bots, metrics, or routes.
- **`docs/SESSION_BOT_LIFECYCLE.md`** — session vs process lifetime, bot shutdown, and finalize ordering.
- **[docs/IMPLEMENTATION_PLAN.md](./docs/IMPLEMENTATION_PLAN.md)** — stable baseline roadmap, toolchain contract, and recovery phases when dependency/tooling drift causes confusion.

## Project purpose

`eth-amm-sim` is a portfolio-grade AMM execution simulation. It demonstrates a local EVM-compatible chain, Foundry/Anvil deployment, immutable Solidity contracts, Go/geth execution infrastructure, bot behavior, WebSocket streaming, and a Vite + React SPA dashboard for AMM observability.

Demo focus: AMM execution lab.

## Demo principle

This should feel like a natural project demo, not a scripted gimmick.
The whale trade is the visual hook, but the project story is the full engineering loop:
local chain → contracts → Go/geth execution → real-time state → risk/market interpretation.

## Engineering principles

- Prefer clear, explicit code over clever abstractions.
- Demo value matters, but do not fake core behavior.
- Keep data flow visible through logs, events, and readable state transitions.
- Avoid new frameworks, storage layers, queues, or auth unless requested.
- Prefer small, focused changes over large rewrites.
- README and `DEMO_GUIDE.md` must match current behavior.

## Scope (current system)

- Local Anvil chain with AppleToken + AppleAMM.
- Go backend polls chain every 2s and streams over WebSocket.
- Bots run per session and stop on context cancel.
- Frontend is Vite + React SPA + Tailwind + lightweight-charts + React Router.

## Go backend rules

- Use context for RPC calls and long-lived work; add timeouts for external calls.
- Use errgroup when coordinating long-lived goroutines.
- Tie goroutines to session/server lifecycle; no orphaned workers.
- Return errors to callers; log at process or boundary layers.
- Keep bot behavior deterministic enough for repeatable demos.

## Solidity rules

- Keep contracts minimal and readable.
- No proxy/upgradeability patterns for this demo unless explicitly requested.
- Use ReentrancyGuard for state-changing functions where applicable.
- Use `Address.sendValue` for ETH transfers where the codebase pattern expects it.
- Preserve the AMM invariant and make state transitions easy to inspect.

## Frontend rules

- Hooks own data fetching/state; components are mostly presentational.
- Do not compute financial metrics in the frontend.
- Keep the dashboard readable on Loom recordings.
- Preserve the `DEMO_GUIDE.md` narrative flow and keep the dashboard causality-focused for recordings.
- Components should clarify causality: trade → reserves → price/TWAP → LP exposure.

## Out of scope unless requested

- Databases, auth, background job systems.
- Leverage, margin, liquidation mechanics.

## Documentation

- README must match current behavior, parameters, and endpoints.
- If you change bots, metrics, or routes, update README and frontend types.

## Verification

Before reporting success after code changes:

```bash
cd frontend && npm run verify
```

Also validate the Go backend when it changed:

```bash
cd backend && GOTOOLCHAIN=auto GOFLAGS=-buildvcs=false go build ./...
```

---

## Language stack rationale (Go vs TypeScript vs Python)

The backend is a concurrent trading simulator with 21+ bot goroutines, real-time WebSocket streaming, constant RPC polling, big-integer AMM math, and an entirely in-memory data store. Go is a strong fit for this workload.

### Would TypeScript (Node.js) or Python be noticeably worse?

#### Memory

| | Go | Node.js (TS) | Python |
|---|---|---|---|
| Base process | ~10–20 MB | ~50–80 MB | ~30–50 MB |
| Per-object overhead | Low (structs, value types) | High (everything is a heap object) | Higher (everything is an object + dict) |
| Big integers | `math/big` (efficient) | Native `BigInt` (decent) | Native `int` (decent) |
| GC pressure | Low (value types, goroutines are cheap) | Moderate (V8 GC handles it but more allocations) | Higher (reference counting + cycle collector) |

For the unbounded price history and trade blotter arrays, Node.js would use roughly **2–4×** more memory per entry and Python roughly **3–5×** more, due to object overhead.

#### Concurrency (the biggest difference)

- **Go**: 21 goroutines are trivial (~2–8 KB stack each). Goroutines are designed exactly for this workload — many lightweight concurrent tasks doing I/O and periodic computation.
- **Node.js**: Single-threaded event loop. The bot logic would need to be restructured as async/await. It would *work* since the bottleneck is I/O (RPC calls), but CPU-bound work (EWMA, volatility, impact curves) would block the event loop. You would need `worker_threads` for the heavier math, adding complexity.
- **Python**: The GIL means true parallelism requires `multiprocessing` or `asyncio` for I/O. Running 21 bots concurrently with real-time WebSocket broadcasting would be significantly more complex and slower. `asyncio` works for I/O but CPU-bound stats calculations would stall the loop.

#### Computation (EWMA, volatility, TWAP, impact curves)

Go is roughly **5–20×** faster than Python and **2–5×** faster than Node.js for tight numerical loops. The impact curve calculation (50+ data points with big-integer math) and per-tick volatility updates would be meaningfully slower in both alternatives.

### Bottom line

- **TypeScript/Node.js**: Would work but use ~2–3× more memory overall and require architectural changes for CPU-bound work. The concurrency model is less natural for this kind of multi-agent simulation.
- **Python**: Would be noticeably worse — higher memory, slower numerics, and the GIL makes the concurrent bot orchestration genuinely painful. You would likely need C extensions (NumPy) or multiprocessing to keep it responsive.

Go is a natural fit for this project. The combination of cheap goroutines, value-type structs, efficient big-integer math, and low GC overhead aligns well with a real-time multi-bot trading simulator. The alternatives would work for a simpler version but would require more engineering to achieve the same responsiveness.
