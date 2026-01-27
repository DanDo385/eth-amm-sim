// Package bots contains the leverage bot implementation
package bots

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/metrics"
	"eth-amm-sim/internal/store"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// LeverageBot trades with leverage (simplified for Phase 1)
// For now, this bot trades more aggressively based on leverage multiplier
type LeverageBot struct {
	*BaseBot
	privateKey   *ecdsa.PrivateKey
	priceProvider metrics.PriceProvider
	lastSignalPrice float64 // Track last price when signal fired
	lastCheckedPrice float64 // Track last price we checked for signals
}

// NewLeverageBot creates a new leverage bot
func NewLeverageBot(cfg *config.AccountConfig, executor *engine.Executor, priceProvider metrics.PriceProvider, store *store.MemoryStore) *LeverageBot {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key for " + cfg.Nickname)
	}

	return &LeverageBot{
		BaseBot:      NewBaseBot(cfg, executor, store),
		privateKey:   privateKey,
		priceProvider: priceProvider,
	}
}

// Run starts the leverage bot trading loop
func (l *LeverageBot) Run(ctx context.Context) {
	log.Printf("[%s] Leverage bot started (%dx leverage)", l.Nickname(), l.config.Leverage)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Leverage bot stopped (context cancelled)", l.Nickname())
			return
		case <-l.stopCh:
			log.Printf("[%s] Leverage bot stopped", l.Nickname())
			return
		default:
			// Trade more frequently with leverage
			// Higher leverage = more aggressive trading
			baseDelay := 5 // Base delay in seconds
			leverageDelay := baseDelay / l.config.Leverage
			if leverageDelay < 1 {
				leverageDelay = 1
			}
			
			delay := time.Duration(leverageDelay) * time.Second
			select {
			case <-ctx.Done():
				return
			case <-l.stopCh:
				return
			case <-time.After(delay):
			}

			// Get current price first
			currentPrice := l.priceProvider.GetCurrentPrice()
			if currentPrice == 0 {
				continue // No price data yet
			}

			// Skip signal checking if price hasn't changed since last check
			// This prevents flooding logs with repeated signal detections
			if currentPrice == l.lastCheckedPrice {
				continue
			}
			l.lastCheckedPrice = currentPrice

			// Check if we already have a position
			addr := crypto.PubkeyToAddress(l.privateKey.PublicKey)
			hasPosition, err := l.hasOpenPosition(ctx, addr)
			if err == nil && hasPosition {
				// Already have a position, skip opening new one
				// In future, could add logic to close/rebalance positions
				continue
			}
			
			// Simple momentum-based strategy for leverage traders
			// They chase trends more aggressively
			side, collateral := l.checkLeverageSignal()
			if side != nil && collateral != nil && collateral.Sign() > 0 {
				// Only execute if price has changed since last signal
				if currentPrice != l.lastSignalPrice {
					// Check balance before attempting trade
					if l.hasSufficientBalance(ctx, *side, collateral) {
						// Open leveraged position through contract
						l.openLeveragedPosition(ctx, collateral)
						l.lastSignalPrice = currentPrice
					}
				}
			}
		}
	}
}

// checkLeverageSignal uses simple momentum to decide trades
// Returns side (always BUY for long positions) and collateral amount
func (l *LeverageBot) checkLeverageSignal() (*engine.TradeSide, *big.Int) {
	// Get recent prices (shorter lookback for leverage traders)
	lookback := 5 // Look back 5 candles
	prices := l.priceProvider.GetRecentPrices(lookback + 1)
	
	// Minimum 3 prices needed (current, previous, lookback)
	minRequired := 3
	if len(prices) < minRequired {
		return nil, nil
	}
	
	// Adjust lookback if we don't have enough data
	if len(prices) < lookback+1 {
		lookback = len(prices) - 1
	}

	currentPrice := prices[len(prices)-1]
	previousPrice := prices[len(prices)-1-lookback]

	// Simple momentum: if price moved up, open long position
	// Leverage traders are more sensitive to smaller moves
	priceChange := (currentPrice - previousPrice) / previousPrice
	threshold := 0.01 // 1% move triggers (lower threshold for leverage)

	if priceChange >= threshold {
		// Use collateral from config
		collateral := l.config.Collateral
		if collateral == nil {
			// Default based on leverage
			collateral = new(big.Int).Mul(big.NewInt(50), big.NewInt(1e18))
		}

		log.Printf("[%s] Leverage signal: change=%.2f%%, opening long position with collateral=%s", 
			l.Nickname(), priceChange*100, formatEther(collateral))

		side := engine.BUY // Always BUY for long positions
		return &side, collateral
	}

	return nil, nil
}

// hasSufficientBalance checks if the bot has enough balance for the trade
// Note: For SELL trades, we don't check APPL balance - bots can build short positions over time
func (l *LeverageBot) hasSufficientBalance(ctx context.Context, side engine.TradeSide, size *big.Int) bool {
	addr := crypto.PubkeyToAddress(l.privateKey.PublicKey)
	gasReserve := new(big.Int).Mul(big.NewInt(1e16), big.NewInt(1)) // 0.01 ETH
	
	if side == engine.BUY {
		ethBalance, err := l.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		required := new(big.Int).Add(size, gasReserve)
		return ethBalance.Cmp(required) >= 0
	} else {
		// For SELL: Only check ETH for gas (allow short positions)
		ethBalance, err := l.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		return ethBalance.Cmp(gasReserve) >= 0
	}
}

// openLeveragedPosition opens a leveraged position through the contract
func (l *LeverageBot) openLeveragedPosition(ctx context.Context, collateral *big.Int) {
	leverage := int64(l.config.Leverage)
	if leverage == 0 {
		leverage = 5 // Default
	}
	
	txHash, err := l.executor.OpenLeveragedPosition(ctx, l.privateKey, collateral, leverage)
	if err != nil {
		log.Printf("[%s] Failed to open leveraged position: %v", l.Nickname(), err)
		return
	}

	log.Printf("[%s] ✓ Leveraged position opened: %s (collateral: %s, leverage: %dx)", 
		l.Nickname(), txHash[:10]+"...", formatEther(collateral), leverage)
}

// hasOpenPosition checks if the bot already has an open position
func (l *LeverageBot) hasOpenPosition(_ context.Context, _ common.Address) (bool, error) {
	// Check if position exists by trying to read it
	// For now, we'll use a simple approach - check if we can liquidate
	// (if we can't liquidate, position might not exist or might not be liquidatable)
	// Actually, better to add a getPosition method to executor
	// For now, return false (no position check) - positions will be tracked on-chain
	return false, nil
}
