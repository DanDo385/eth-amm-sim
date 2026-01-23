// Package bots contains the retail bot implementation
package bots

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"math/rand"
	"time"
)

// RetailBot makes small, frequent trades adding noise to the market
// Simulates retail traders with random behavior
type RetailBot struct {
	*BaseBot
	executor   TradeExecutor
	privateKey *ecdsa.PrivateKey
	
	// Trading parameters
	minSize  *big.Int
	maxSize  *big.Int
	minDelay time.Duration
	maxDelay time.Duration
	
	rng *rand.Rand
}

// RetailConfig configures a retail bot
type RetailConfig struct {
	Nickname   string
	PrivateKey *ecdsa.PrivateKey
	MinSize    *big.Int
	MaxSize    *big.Int
	MinDelay   time.Duration
	MaxDelay   time.Duration
	Seed       int64
}

// NewRetailBot creates a new retail bot
func NewRetailBot(executor TradeExecutor, config RetailConfig) *RetailBot {
	return &RetailBot{
		BaseBot:    NewBaseBot(config.Nickname),
		executor:   executor,
		privateKey: config.PrivateKey,
		minSize:    config.MinSize,
		maxSize:    config.MaxSize,
		minDelay:   config.MinDelay,
		maxDelay:   config.MaxDelay,
		rng:        rand.New(rand.NewSource(config.Seed)),
	}
}

// Run starts the retail bot trading loop
func (r *RetailBot) Run(ctx context.Context) {
	r.SetRunning()
	log.Printf("[%s] Retail bot started", r.Nickname())
	
	for {
		delay := r.randomDelay()
		
		select {
		case <-ctx.Done():
			log.Printf("[%s] Retail bot stopped (context cancelled)", r.Nickname())
			return
		case <-r.StopCh():
			log.Printf("[%s] Retail bot stopped", r.Nickname())
			return
		case <-time.After(delay):
			r.executeTrade(ctx)
		}
	}
}

// executeTrade performs a single random trade
func (r *RetailBot) executeTrade(ctx context.Context) {
	// 50/50 buy or sell
	direction := Buy
	if r.rng.Float64() < 0.5 {
		direction = Sell
	}
	
	size := r.randomSize()
	
	log.Printf("[%s] Executing %s trade, size: %s", r.Nickname(), direction, formatEther(size))
	
	var err error
	var txHash string
	
	if direction == Buy {
		txHash, err = r.executor.SwapETHForApples(ctx, r.privateKey, size)
	} else {
		txHash, err = r.executor.SwapApplesForETH(ctx, r.privateKey, size)
	}
	
	if err != nil {
		log.Printf("[%s] Trade failed: %v", r.Nickname(), err)
		return
	}
	
	log.Printf("[%s] Trade submitted: %s", r.Nickname(), txHash)
}

// randomDelay returns a random delay
func (r *RetailBot) randomDelay() time.Duration {
	diff := r.maxDelay - r.minDelay
	return r.minDelay + time.Duration(r.rng.Int63n(int64(diff)))
}

// randomSize returns a random size
func (r *RetailBot) randomSize() *big.Int {
	diff := new(big.Int).Sub(r.maxSize, r.minSize)
	randomOffset := new(big.Int).Rand(r.rng, diff)
	return new(big.Int).Add(r.minSize, randomOffset)
}

// DefaultRetailConfig returns a default retail bot configuration
func DefaultRetailConfig(nickname string, seed int64) RetailConfig {
	ether := func(n int64) *big.Int {
		return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18))
	}
	
	return RetailConfig{
		Nickname: nickname,
		MinSize:  ether(1),
		MaxSize:  ether(10),
		MinDelay: 2 * time.Second,
		MaxDelay: 5 * time.Second,
		Seed:     seed,
	}
}
