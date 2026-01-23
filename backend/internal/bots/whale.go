// Package bots contains the whale bot implementation
package bots

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"math/rand"
	"time"
)

// TradeExecutor interface for executing trades (breaks import cycle)
type TradeExecutor interface {
	SwapETHForApples(ctx context.Context, privateKey *ecdsa.PrivateKey, ethAmount *big.Int) (string, error)
	SwapApplesForETH(ctx context.Context, privateKey *ecdsa.PrivateKey, appleAmount *big.Int) (string, error)
}

// WhaleBot makes large, infrequent trades that visibly move the market
// Whale1: Starts long (has APPL), tends to sell
// Whale2: Starts empty, tends to buy
// Whale3: Partial position, trades both ways
type WhaleBot struct {
	*BaseBot
	executor   TradeExecutor
	privateKey *ecdsa.PrivateKey
	
	// Trading parameters
	minSize    *big.Int // Minimum trade size
	maxSize    *big.Int // Maximum trade size
	minDelay   time.Duration
	maxDelay   time.Duration
	buyBias    float64  // 0.0 = always sell, 1.0 = always buy, 0.5 = neutral
	
	rng *rand.Rand
}

// WhaleConfig configures a whale bot
type WhaleConfig struct {
	Nickname   string
	PrivateKey *ecdsa.PrivateKey
	MinSize    *big.Int
	MaxSize    *big.Int
	MinDelay   time.Duration
	MaxDelay   time.Duration
	BuyBias    float64
	Seed       int64
}

// NewWhaleBot creates a new whale bot
func NewWhaleBot(executor TradeExecutor, config WhaleConfig) *WhaleBot {
	return &WhaleBot{
		BaseBot:    NewBaseBot(config.Nickname),
		executor:   executor,
		privateKey: config.PrivateKey,
		minSize:    config.MinSize,
		maxSize:    config.MaxSize,
		minDelay:   config.MinDelay,
		maxDelay:   config.MaxDelay,
		buyBias:    config.BuyBias,
		rng:        rand.New(rand.NewSource(config.Seed)),
	}
}

// Run starts the whale bot trading loop
func (w *WhaleBot) Run(ctx context.Context) {
	w.SetRunning()
	log.Printf("[%s] Whale bot started", w.Nickname())
	
	for {
		// Random delay between trades
		delay := w.randomDelay()
		
		select {
		case <-ctx.Done():
			log.Printf("[%s] Whale bot stopped (context cancelled)", w.Nickname())
			return
		case <-w.StopCh():
			log.Printf("[%s] Whale bot stopped", w.Nickname())
			return
		case <-time.After(delay):
			w.executeTrade(ctx)
		}
	}
}

// executeTrade performs a single trade
func (w *WhaleBot) executeTrade(ctx context.Context) {
	// Determine direction based on bias
	direction := Sell
	if w.rng.Float64() < w.buyBias {
		direction = Buy
	}
	
	// Random size within range
	size := w.randomSize()
	
	log.Printf("[%s] Executing %s trade, size: %s", w.Nickname(), direction, formatEther(size))
	
	var err error
	var txHash string
	
	if direction == Buy {
		txHash, err = w.executor.SwapETHForApples(ctx, w.privateKey, size)
	} else {
		txHash, err = w.executor.SwapApplesForETH(ctx, w.privateKey, size)
	}
	
	if err != nil {
		log.Printf("[%s] Trade failed: %v", w.Nickname(), err)
		return
	}
	
	log.Printf("[%s] Trade submitted: %s", w.Nickname(), txHash)
}

// randomDelay returns a random delay between minDelay and maxDelay
func (w *WhaleBot) randomDelay() time.Duration {
	diff := w.maxDelay - w.minDelay
	return w.minDelay + time.Duration(w.rng.Int63n(int64(diff)))
}

// randomSize returns a random size between minSize and maxSize
func (w *WhaleBot) randomSize() *big.Int {
	diff := new(big.Int).Sub(w.maxSize, w.minSize)
	
	// Generate random value up to diff
	randomOffset := new(big.Int).Rand(w.rng, diff)
	
	return new(big.Int).Add(w.minSize, randomOffset)
}

// formatEther formats a big.Int as ETH with 4 decimal places
func formatEther(wei *big.Int) string {
	eth := new(big.Float).SetInt(wei)
	eth.Quo(eth, big.NewFloat(1e18))
	return eth.Text('f', 4) + " ETH"
}

// DefaultWhaleConfigs returns default configurations for whale bots
func DefaultWhaleConfigs() []WhaleConfig {
	ether := func(n int64) *big.Int {
		return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18))
	}
	
	return []WhaleConfig{
		{
			Nickname:   "Whale1",
			MinSize:    ether(100),
			MaxSize:    ether(500),
			MinDelay:   15 * time.Second,
			MaxDelay:   30 * time.Second,
			BuyBias:    0.3, // Tends to sell (starts with APPL)
			Seed:       1,
		},
		{
			Nickname:   "Whale2",
			MinSize:    ether(100),
			MaxSize:    ether(500),
			MinDelay:   15 * time.Second,
			MaxDelay:   30 * time.Second,
			BuyBias:    0.7, // Tends to buy (starts empty)
			Seed:       2,
		},
		{
			Nickname:   "Whale3",
			MinSize:    ether(50),
			MaxSize:    ether(300),
			MinDelay:   20 * time.Second,
			MaxDelay:   40 * time.Second,
			BuyBias:    0.5, // Neutral
			Seed:       3,
		},
	}
}
