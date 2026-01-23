// Package bots contains the retail bot implementation
package bots

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"github.com/ethereum/go-ethereum/crypto"
)

// RetailBot executes small, frequent random trades
type RetailBot struct {
	*BaseBot
	privateKey *ecdsa.PrivateKey
}

// NewRetailBot creates a new retail bot
func NewRetailBot(cfg *config.AccountConfig, executor *engine.Executor) *RetailBot {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key for " + cfg.Nickname)
	}

	return &RetailBot{
		BaseBot:    NewBaseBot(cfg, executor),
		privateKey: privateKey,
	}
}

// Run starts the retail bot trading loop
func (r *RetailBot) Run(ctx context.Context) {
	log.Printf("[%s] Retail bot started", r.Nickname())

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Retail bot stopped (context cancelled)", r.Nickname())
			return
		case <-r.stopCh:
			log.Printf("[%s] Retail bot stopped", r.Nickname())
			return
		default:
			// Wait random interval from config (2-5 seconds)
			delay := r.RandomDelay()
			select {
			case <-ctx.Done():
				return
			case <-r.stopCh:
				return
			case <-time.After(delay):
			}

			// Random direction (pure noise)
			var side engine.TradeSide
			if r.rng.Float64() < 0.5 {
				side = engine.BUY
			} else {
				side = engine.SELL
			}

			// Random size from config
			size := r.RandomSize()

			if size.Sign() > 0 {
				r.executeTrade(ctx, side, size)
			}
		}
	}
}

// executeTrade performs a single random trade
func (r *RetailBot) executeTrade(ctx context.Context, side engine.TradeSide, size *big.Int) {
	var err error
	var txHash string

	if side == engine.BUY {
		// Buy: swap ETH for APPLES
		txHash, err = r.executor.SwapETHForApples(ctx, r.privateKey, size)
	} else {
		// Sell: swap APPLES for ETH
		txHash, err = r.executor.SwapApplesForETH(ctx, r.privateKey, size)
	}

	if err != nil {
		log.Printf("[%s] Trade failed: %v", r.Nickname(), err)
		return
	}

	log.Printf("[%s] Trade submitted: %s (%s, size: %s)", r.Nickname(), txHash, side, formatEther(size))
}
