// whale.go - Large opportunistic trader bot (accounts 17-19).
//
// Strategy: Whales randomly buy or sell with equal probability (no directional bias).
// Trade sizes are large (up to 600 ETH) with long intervals (45-90s, ~2-4 trades per session).
// This creates periodic large price impacts visible on the frontend PriceChart.
//
// Parameters come from config/accounts.go (MaxTradeSize, TradeFreqMin/Max).
// Trades execute via executor.SwapETHForApples / SwapApplesForETH → AppleAMM contract.
package bots

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/store"
	"github.com/ethereum/go-ethereum/crypto"
)

// WhaleBot executes large, infrequent trades
type WhaleBot struct {
	*BaseBot
	privateKey *ecdsa.PrivateKey
}

// NewWhaleBot creates a new whale bot
func NewWhaleBot(cfg *config.AccountConfig, executor *engine.Executor, store *store.MemoryStore) (*WhaleBot, error) {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key for %s: %w", cfg.Nickname, err)
	}

	return &WhaleBot{
		BaseBot:    NewBaseBot(cfg, executor, store),
		privateKey: privateKey,
	}, nil
}

// Run starts the whale bot trading loop
func (w *WhaleBot) Run(ctx context.Context) {
	log.Printf("[%s] Whale bot started", w.Nickname())

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Whale bot stopped (context cancelled)", w.Nickname())
			return
		default:
			// Wait random interval from config
			delay := w.RandomDelay()
			select {
			case <-ctx.Done():
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

// decideSide randomly picks buy or sell with equal probability (no directional bias)
func (w *WhaleBot) decideSide() engine.TradeSide {
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
		// Convert ETH notional to APPL amount
		spotPrice, err := w.executor.GetSpotPrice(ctx)
		if err != nil || spotPrice == nil || spotPrice.Sign() == 0 {
			log.Printf("[%s] Cannot execute SELL: failed to get spot price: %v", w.Nickname(), err)
			return
		}
		
		// Convert ETH size to APPL: appleAmount = ethAmount / price
		// spotPrice is ETH per APPL (scaled by 1e18), so APPL = (ETH * 1e18) / spotPrice
		sizeFloat := new(big.Float).SetInt(size)
		priceFloat := new(big.Float).SetInt(spotPrice)
		priceFloat.Quo(priceFloat, big.NewFloat(1e18)) // Convert from wei to ETH
		appleAmount := new(big.Float).Quo(sizeFloat, priceFloat)
		appleAmountInt, _ := appleAmount.Int(nil)
		
		txHash, err = w.executor.SwapApplesForETH(ctx, w.privateKey, appleAmountInt)
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
