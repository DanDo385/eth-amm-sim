# Session and bot lifecycle (how to think about it)

This note explains **how a simulation session starts and stops**, why a second session used to show **no trades**, and what changed.

## Big picture: two different “processes”

1. **The Go backend process** (`make backend`) — long-lived. It serves HTTP/WebSocket and owns in-memory state.
2. **A simulation session** — short-lived. It is a **timer + bot goroutines** started by `POST /session/start` and ended by timeout, `POST /session/stop`, or natural completion.

You should **not** need to restart the backend between sessions. If you did before, that was a bug, not the intended design.

## Who owns what

| Piece | Role |
|--------|------|
| `engine.Session` | State machine: `idle` → `running` → `completed`. Creates a **timeout context** for the run. |
| `engine.Orchestrator` | Starts one `Run(ctx)` goroutine per bot under an **errgroup** tied to that session context. |
| Each bot’s `Run(ctx)` | Should loop until **`ctx` is cancelled** (session over). |
| `cmd/simulator/main.go` | Creates bots **once** at startup and registers them with the orchestrator. |

## How stop is supposed to work

When the session ends, `Session.run` waits on `<-ctx.Done()`, then calls `orchestrator.Stop()`.

`Orchestrator.Stop()`:

1. Calls **`cancel()`** on the session’s context — every bot’s `Run` should see `<-ctx.Done()` and exit.
2. Calls **`bot.Stop()`** on each bot (interface hook).

So the **authoritative** shutdown signal is **context cancellation**, not a separate permanent flag on the bot object.

After **`orchestrator.Stop()` returns**, the server runs **`finalizeSession`** once via **`Session.SetOnSessionEnded`** (account close-out / PnL). Only then does status become **`completed`** and broadcast. That ordering avoids finalizing while bots can still submit swaps, and avoids duplicate finalization from HTTP handlers vs async WebSocket callbacks.

## The bug (second session silent)

Bots embedded `BaseBot` with a **`stopCh`** channel. `Stop()` **closed** that channel once.

In Go, **receive from a closed channel is always ready** in a `select`. The bot loops had:

```go
select {
case <-ctx.Done(): return
case <-stopCh: return
default: // trade...
}
```

After the first session, `stopCh` stayed **closed forever**. On session 2, `ctx` was new and still alive, but **`case <-stopCh` won immediately**, so `Run` exited without trading.

Restarting the backend “fixed” it only because **new bot structs** got fresh channels.

## The fix

- Removed **`stopCh`** from `BaseBot`.
- Bot loops exit only on **`<-ctx.Done()`** (and normal return paths).
- `BaseBot.Stop()` remains a **no-op** so the `Bot` interface and orchestrator still compile; shutdown is **context-only**.

## How to verify yourself

1. Run stack (`make up` or manual terminals).
2. Open UI, **Start** session, wait for **completed** (or Stop).
3. Wait 30s, **Start** again without restarting Go.
4. You should see **new trades** in the blotter and logs like `Retail bot started` without an immediate `stopped` from a dead channel.

Optional: watch backend logs for your bot nicknames; session 2 should show sustained trading, not instant exit.

## Mental model to keep

Treat each session as: **new context in, cancel out**.  
Do not store **one-way** shutdown state on long-lived bot objects unless you **reset** it at the start of every session.
