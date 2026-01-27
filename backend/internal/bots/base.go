// Package bots contains base bot implementation
package bots

import (
	"log"
	"math/big"
	"math/rand"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/store"
)

// BaseBot contains common functionality for all bots
type BaseBot struct {
	config     *config.AccountConfig
	executor   *engine.Executor
	store      *store.MemoryStore
	stopCh     chan struct{}
	rng        *rand.Rand
	stoppedOut bool // Track if we've been stopped out
}

// NewBaseBot creates a new base bot
func NewBaseBot(cfg *config.AccountConfig, executor *engine.Executor, store *store.MemoryStore) *BaseBot {
	return &BaseBot{
		config:     cfg,
		executor:   executor,
		store:      store,
		stopCh:     make(chan struct{}),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano() + int64(cfg.Index))),
		stoppedOut: false,
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

// IsStoppedOut checks if the bot should stop trading due to losses
func (b *BaseBot) IsStoppedOut() bool {
	if b.stoppedOut {
		return true // Already stopped out
	}
	
	if b.store == nil || b.config.StopOutPercent >= 0 {
		return false // No stop-out configured or invalid value (only negative makes sense)
	}
	
	metrics := b.store.GetAccountMetrics(b.Nickname())
	if metrics == nil {
		return false // No metrics yet, can't check
	}
	
	perf := metrics.GetPerformance()
	
	// Calculate return percentage
	// TotalReturn is already a percentage, so divide by 100 to get decimal
	returnPercent := perf.TotalReturn / 100.0
	
	// Stop out if loss exceeds threshold (e.g., -20% when stopOutPercent = -0.20)
	if returnPercent <= b.config.StopOutPercent {
		b.stoppedOut = true
		log.Printf("[%s] ⚠️ STOPPED OUT: Loss %.2f%% exceeded threshold %.1f%%", 
			b.Nickname(), perf.TotalReturn, b.config.StopOutPercent*100)
		return true
	}
	
	return false
}
