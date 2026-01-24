// Package bots contains the whale bot implementation
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

// WhaleBot executes large, infrequent trades
type WhaleBot struct {
	*BaseBot
	privateKey *ecdsa.PrivateKey
}

// NewWhaleBot creates a new whale bot
func NewWhaleBot(cfg *config.AccountConfig, executor *engine.Executor) *WhaleBot {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key for " + cfg.Nickname)
	}

	return &WhaleBot{
		BaseBot:    NewBaseBot(cfg, executor),
		privateKey: privateKey,
	}
}

// Run starts the whale bot trading loop
func (w *WhaleBot) Run(ctx context.Context) {
	log.Printf("[%s] Whale bot started", w.Nickname())

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Whale bot stopped (context cancelled)", w.Nickname())
			return
		case <-w.stopCh:
			log.Printf("[%s] Whale bot stopped", w.Nickname())
			return
		default:
			// Wait random interval from config
			delay := w.RandomDelay()
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-time.After(delay):
			}

			// Decide trade direction based on starting position
			side := w.decideSide()
			size := w.RandomSize()

			if size.Sign() > 0 {
				// Check balance before attempting trade
				if w.hasSufficientBalance(ctx, side, size) {
					w.executeTrade(ctx, side, size)
				}
			}
		}
	}
}

// decideSide determines trade direction based on starting position
func (w *WhaleBot) decideSide() engine.TradeSide {
	startingApples := w.config.StartingApples

	// If started with lots of APPLES, 70% chance to sell
	eth500 := new(big.Int).Mul(big.NewInt(500), big.NewInt(1e18))
	if startingApples != nil && startingApples.Cmp(eth500) > 0 {
		if w.rng.Float64() < 0.7 {
			return engine.SELL
		}
		return engine.BUY
	}

	// If started with no APPLES, 70% chance to buy
	if startingApples == nil || startingApples.Sign() == 0 {
		if w.rng.Float64() < 0.7 {
			return engine.BUY
		}
		return engine.SELL
	}

	// Otherwise 50/50
	if w.rng.Float64() < 0.5 {
		return engine.BUY
	}
	return engine.SELL
}

// hasSufficientBalance checks if the bot has enough balance for the trade
// Note: For SELL trades, we don't check APPL balance - bots can build short positions over time
func (w *WhaleBot) hasSufficientBalance(ctx context.Context, side engine.TradeSide, size *big.Int) bool {
	addr := crypto.PubkeyToAddress(w.privateKey.PublicKey)
	gasReserve := new(big.Int).Mul(big.NewInt(1e16), big.NewInt(1)) // 0.01 ETH
	
	if side == engine.BUY {
		ethBalance, err := w.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		required := new(big.Int).Add(size, gasReserve)
		return ethBalance.Cmp(required) >= 0
	} else {
		// For SELL: Only check ETH for gas (allow short positions)
		ethBalance, err := w.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		return ethBalance.Cmp(gasReserve) >= 0
	}
}

// executeTrade performs a single trade
func (w *WhaleBot) executeTrade(ctx context.Context, side engine.TradeSide, size *big.Int) {
	var err error
	var txHash string

	if side == engine.BUY {
		// Buy: swap ETH for APPLES
		txHash, err = w.executor.SwapETHForApples(ctx, w.privateKey, size)
	} else {
		// Sell: swap APPLES for ETH
		txHash, err = w.executor.SwapApplesForETH(ctx, w.privateKey, size)
	}

	if err != nil {
		log.Printf("[%s] Trade failed: %v", w.Nickname(), err)
		return
	}

	log.Printf("[%s] ✓ Trade executed: %s (%s, size: %s)", w.Nickname(), txHash[:10]+"...", side, formatEther(size))
}

// formatEther formats a big.Int as ETH with 4 decimal places
func formatEther(wei *big.Int) string {
	eth := new(big.Float).SetInt(wei)
	eth.Quo(eth, big.NewFloat(1e18))
	return eth.Text('f', 4) + " ETH"
}
