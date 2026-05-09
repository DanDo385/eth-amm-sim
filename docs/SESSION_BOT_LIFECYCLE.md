# Session And Bot Lifecycle

This document explains how session state, bot execution, and resets interact in the current backend.

## Lifecycle model

- The Go process can stay up across many sessions.
- A session is a bounded run of bot activity (`idle -> running -> paused/completed`).
- Bots are started/stopped per session by the orchestrator, not by restarting the process.

## Session states

- `idle`: no active session.
- `running`: bots are active and trading.
- `paused`: timer/session is paused; resume continues from remaining duration.
- `completed`: timer elapsed and bots stopped.
- `error`: panic/failure path.

Transitions are managed in `backend/internal/engine/session.go`.

## Start / stop / pause / resume

- `POST /session/start`
  - Reinitializes LP baseline from current on-chain reserves.
  - Resets account metrics for session tracking.
  - Starts orchestrator with a session context timeout.
- `POST /session/pause`
  - Cancels current run in pause mode.
  - Bots stop; session becomes `paused`.
- `POST /session/resume`
  - Resets trading account positions to configured starting balances.
  - Re-anchors account metrics and resumes for remaining duration.
- `POST /session/stop`
  - Stops a running session, or normalizes paused/completed back to `idle`.

## Reset modes

- `soft` (`POST /session/reset?mode=soft`)
  - Resets in-memory store/session view.
  - Reinitializes LP metrics from current chain state.
- `hard` (`mode=hard`)
  - Includes soft reset behavior.
  - Resets account metrics and user balances.
- `reseed` (`mode=reseed`)
  - Includes hard reset behavior.
  - Executes `anvil_reset`, redeploys contracts, clears nonce cache, and reinitializes LP/account baselines.

## Event behavior between sessions

- Price-move strategy-trigger events are emitted only during `running` sessions.
- During reset/redeploy windows between sessions, price baseline is re-anchored to avoid synthetic key events.
- User trades are always surfaced as key events; non-user trades still follow size thresholds.

## Finalization ordering

- On normal completion, bots stop first.
- Session finalization then computes/stores close-dependent account metrics.
- Final account updates are broadcast after finalization.

This ordering avoids finalizing while bot activity is still in-flight.
