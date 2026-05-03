# AGENTS.md

Canonical agent instructions for this repository.

Do not add separate root-level `.cursorrules`, `CLAUDE.md`, `GEMINI.md`, or tool-specific instruction files unless a tool absolutely requires a shim. If a shim is required, it must only point back to this file and must not contain independent rules.

## Project purpose

`eth-amm-sim` is a portfolio-grade AMM execution simulation. It demonstrates a local EVM-compatible chain, Foundry/Anvil deployment, immutable Solidity contracts, Go/geth execution infrastructure, bot behavior, WebSocket streaming, and a Next.js dashboard for AMM observability.

Demo focus: AMM execution lab.

## Demo principle

This should feel like a natural project demo, not a scripted gimmick.
The whale trade is the visual hook, but the project story is the full engineering loop:
local chain -> contracts -> Go/geth execution -> real-time state -> risk/market interpretation.

## Engineering principles

- Prefer clear, explicit code over clever abstractions.
- Demo value matters, but do not fake core behavior.
- Keep data flow visible through logs, events, and readable state transitions.
- Avoid new frameworks, storage layers, queues, or auth unless requested.
- README and `DEMO_GUIDE.md` must match current behavior.

## Go backend rules

- Use context for RPC calls and long-lived work.
- Tie goroutines to session/server lifecycle; no orphaned workers.
- Return errors to callers; log at process or boundary layers.
- Keep bot behavior deterministic enough for repeatable demos.

## Solidity rules

- Keep contracts minimal and readable.
- No proxy/upgradeability patterns for this demo unless explicitly requested.
- Preserve the AMM invariant and make state transitions easy to inspect.

## Frontend rules

- Keep the dashboard readable on Loom recordings.
- Preserve the `LoomDemoDirector` and `DEMO_GUIDE.md` narrative flow.
- Components should clarify causality: trade -> reserves -> price/TWAP -> LP exposure.

## Verification

Before reporting success after code changes:

```bash
cd /Users/openclaw/Code/eth-amm-sim/frontend
npm run build
```
