// base.go - Shared foundation for all trading bot types.
//
// SYSTEM ROLE:
// BaseBot holds the account config, executor reference, store reference, and
// private key that every bot needs. It provides utility methods (RandomDelay,
// RandomSize, Stop) used by whale, retail, and meanrev bots.
//
// Each bot's Run() method loops until the context is cancelled (session ends).
// Bots call executor.SwapETHForApples / SwapApplesForETH to submit transactions
// to the on-chain AppleAMM contract. Trade callbacks in main.go record results
// to the MemoryStore and broadcast them via WebSocket to the frontend.
//
// CONNECTIONS:
//  - Config: config/accounts.go defines each bot's parameters
//  - Executor: engine/executor.go submits transactions to contracts/src/AppleAMM.sol
//  - Store: store/memory.go for metrics access (price data, account metrics)
//  - Lifecycle: engine/orchestrator.go starts/stops bots via context cancellation
package bots

import (
	"math/big"
	"math/rand"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/store"
)

// BaseBot contains common functionality for all bots
type BaseBot struct {
	config   *config.AccountConfig
	executor *engine.Executor
	store    *store.MemoryStore
	rng      *rand.Rand
}

// NewBaseBot creates a new base bot
func NewBaseBot(cfg *config.AccountConfig, executor *engine.Executor, store *store.MemoryStore) *BaseBot {
	return &BaseBot{
		config:   cfg,
		executor: executor,
		store:    store,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano() + int64(cfg.Index))),
	}
}

// Nickname returns the bot's nickname
func (b *BaseBot) Nickname() string {
	return b.config.Nickname
}

// Type returns the bot's type
func (b *BaseBot) Type() config.BotType {
	return b.config.Type
}

// Stop is kept for the engine.Bot / orchestrator contract but does not carry
// per-session shutdown state. Each session uses a fresh context from
// engine.Orchestrator; Run(ctx) exits on <-ctx.Done() when the orchestrator
// cancels that context. A one-shot channel (previously closed here) made bots
// unusable after the first session because receiving from a closed channel is
// always ready in select.
func (b *BaseBot) Stop() {}

// RandomDelay returns a random duration between the configured min and max
func (b *BaseBot) RandomDelay() time.Duration {
	min := b.config.TradeFreqMin
	max := b.config.TradeFreqMax
	if max <= min {
		return time.Duration(min) * time.Second
	}
	return time.Duration(min+b.rng.Intn(max-min+1)) * time.Second
}

// RandomSize returns a random trade size up to MaxTradeSize
func (b *BaseBot) RandomSize() *big.Int {
	max := b.config.MaxTradeSize
	if max == nil || max.Sign() <= 0 {
		return big.NewInt(0)
	}

	// Random percentage of max (10% to 100%)
	pct := 10 + b.rng.Intn(91) // 10-100
	size := new(big.Int).Mul(max, big.NewInt(int64(pct)))
	size.Div(size, big.NewInt(100))
	return size
}

// Config returns the bot's account configuration
func (b *BaseBot) Config() *config.AccountConfig {
	return b.config
}

// Executor returns the bot's executor
func (b *BaseBot) Executor() *engine.Executor {
	return b.executor
}