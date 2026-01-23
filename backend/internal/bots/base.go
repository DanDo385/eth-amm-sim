// Package bots contains base bot implementation
package bots

import (
	"math/big"
	"math/rand"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
)

// BaseBot contains common functionality for all bots
type BaseBot struct {
	config   *config.AccountConfig
	executor *engine.Executor
	stopCh   chan struct{}
	rng      *rand.Rand
}

// NewBaseBot creates a new base bot
func NewBaseBot(cfg *config.AccountConfig, executor *engine.Executor) *BaseBot {
	return &BaseBot{
		config:   cfg,
		executor: executor,
		stopCh:   make(chan struct{}),
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

// Stop signals the bot to stop
func (b *BaseBot) Stop() {
	select {
	case <-b.stopCh:
		// Already closed
	default:
		close(b.stopCh)
	}
}

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
