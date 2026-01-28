// bot.go — Defines the Bot interface contract for all trading bots.
//
// SYSTEM ROLE:
// This file defines the minimal Bot interface that all trading bot implementations
// must satisfy. The interface is used by engine/orchestrator.go to manage bot
// lifecycle (start/stop) without needing to know specific bot types.
//
// FILE ORGANIZATION:
// This file is intentionally separate from base.go for the following reasons:
//  1. Separation of concerns: Interface definition (contract) vs implementation (base.go)
//  2. Go convention: Interfaces often live in their own file, especially when used
//     by other packages
//  3. Discoverability: Having the interface in bot.go makes it easy to find the
//     contract that all bots must implement
//
// NOTE ON DUPLICATE INTERFACES:
// engine/orchestrator.go defines its own Bot interface (without Type()) to avoid
// import cycles. Both interfaces are intentionally similar — all bot implementations
// satisfy both interfaces. This is a common Go pattern for breaking circular
// dependencies.
//
// The Bot interface is minimal by design — it only exposes lifecycle methods needed
// by the orchestrator. Bot-specific functionality (trading logic, EWMA state, etc.)
// is encapsulated in each bot's implementation (whale.go, retail.go, meanrev.go, etc.).
//
// CONNECTIONS:
//   - Used by: engine/orchestrator.go (manages bot lifecycle via its own Bot interface)
//   - Implemented by: All bot types (WhaleBot, RetailBot, MeanRevBot, LeverageBot, LiquidatorBot)
//   - Base implementation: base.go provides BaseBot struct with common functionality
package bots

import (
	"context"

	"eth-amm-sim/internal/config"
)

// Bot is the interface all trading bots must implement.
//
// This interface defines the minimal contract for bot lifecycle management.
// The orchestrator uses this interface to start/stop bots without needing to
// know specific bot types or trading strategies.
//
// All bots embed BaseBot (from base.go) which provides default implementations
// for Stop(), Nickname(), and Type(). Each bot implements Run() with its own
// trading strategy.
type Bot interface {
	// Run starts the bot's trading loop. It should run until ctx is cancelled.
	// The orchestrator calls this in a goroutine and cancels ctx when stopping.
	Run(ctx context.Context)

	// Stop signals the bot to stop gracefully. The bot should exit its Run() loop
	// on the next iteration after receiving the stop signal.
	Stop()

	// Nickname returns the bot's human-readable name (e.g., "Whale1", "Retail5").
	// Used for logging and frontend display.
	Nickname() string

	// Type returns the bot's type (BotTypeWhale, BotTypeRetail, etc.).
	// Used for filtering and categorization.
	Type() config.BotType
}