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
	"eth-amm-sim/internal/store"
	"github.com/ethereum/go-ethereum/crypto"
)

// RetailBot executes small, frequent random trades
type RetailBot struct {
	*BaseBot
	privateKey *ecdsa.PrivateKey
}

// NewRetailBot creates a new retail bot
func NewRetailBot(cfg *config.AccountConfig, executor *engine.Executor, store *store.MemoryStore) *RetailBot {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key for " + cfg.Nickname)
	}

	return &RetailBot{
		BaseBot:    NewBaseBot(cfg, executor, store),
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
			// Check if stopped out
			if r.IsStoppedOut() {
				select {
				case <-ctx.Done():
					return
				case <-r.stopCh:
					return
				case <-time.After(10 * time.Second):
				}
				continue
			}
			
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
				// Check balance before attempting trade
				if r.hasSufficientBalance(ctx, side, size) {
					r.executeTrade(ctx, side, size)
				}
			}
		}
	}
}

// hasSufficientBalance checks if the bot has enough balance for the trade
// Note: For SELL trades, we don't check APPL balance - bots can build short positions over time
func (r *RetailBot) hasSufficientBalance(ctx context.Context, side engine.TradeSide, size *big.Int) bool {
	addr := crypto.PubkeyToAddress(r.privateKey.PublicKey)
	gasReserve := new(big.Int).Mul(big.NewInt(1e16), big.NewInt(1)) // 0.01 ETH
	
	if side == engine.BUY {
		ethBalance, err := r.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		required := new(big.Int).Add(size, gasReserve)
		return ethBalance.Cmp(required) >= 0
	} else {
		// For SELL: Only check ETH for gas (allow short positions)
		ethBalance, err := r.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		return ethBalance.Cmp(gasReserve) >= 0
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

	log.Printf("[%s] ✓ Trade executed: %s (%s, size: %s)", r.Nickname(), txHash[:10]+"...", side, formatEther(size))
}
