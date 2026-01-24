// Package bots contains the leverage bot implementation
package bots

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math"
	"math/big"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"github.com/ethereum/go-ethereum/crypto"
)

// LeverageBot trades with leverage (simplified for Phase 1)
// For now, this bot trades more aggressively based on leverage multiplier
type LeverageBot struct {
	*BaseBot
	privateKey   *ecdsa.PrivateKey
	priceProvider PriceProvider
	lastSignalPrice float64 // Track last price when signal fired
}

// NewLeverageBot creates a new leverage bot
func NewLeverageBot(cfg *config.AccountConfig, executor *engine.Executor, priceProvider PriceProvider) *LeverageBot {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key for " + cfg.Nickname)
	}

	return &LeverageBot{
		BaseBot:      NewBaseBot(cfg, executor),
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

			// Simple momentum-based strategy for leverage traders
			// They chase trends more aggressively
			side, size := l.checkLeverageSignal()
			if side != nil && size != nil && size.Sign() > 0 {
				// Get current price to check if we've already traded on this signal
				currentPrice := l.priceProvider.GetCurrentPrice()
				
				// Only execute if price has changed since last signal
				if currentPrice != l.lastSignalPrice {
					// Check balance before attempting trade
					if l.hasSufficientBalance(ctx, *side, size) {
						l.executeTrade(ctx, *side, size)
						l.lastSignalPrice = currentPrice
					}
				}
			}
		}
	}
}

// checkLeverageSignal uses simple momentum to decide trades
// Leverage traders are more aggressive and trade larger sizes
func (l *LeverageBot) checkLeverageSignal() (*engine.TradeSide, *big.Int) {
	// Get recent prices (shorter lookback for leverage traders)
	lookback := 5 // Look back 5 candles
	prices := l.priceProvider.GetRecentPrices(lookback + 1)
	if len(prices) < lookback+1 {
		return nil, nil
	}

	currentPrice := prices[len(prices)-1]
	previousPrice := prices[len(prices)-1-lookback]

	// Simple momentum: if price moved, trade in that direction
	// Leverage traders are more sensitive to smaller moves
	priceChange := (currentPrice - previousPrice) / previousPrice
	threshold := 0.01 // 1% move triggers (lower threshold for leverage)

	if math.Abs(priceChange) >= threshold {
		var side engine.TradeSide
		if priceChange > 0 {
			side = engine.BUY
		} else {
			side = engine.SELL
		}

		// Trade size scaled by leverage (higher leverage = larger trades)
		// Use collateral as base, scaled by leverage
		baseSize := l.config.Collateral
		if baseSize == nil {
			// Default 100 ETH
			baseSize = new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
		}
		
		// Scale by leverage (but cap at MaxTradeSize if configured)
		size := new(big.Int).Mul(baseSize, big.NewInt(int64(l.config.Leverage)))
		size.Div(size, big.NewInt(10)) // Divide by 10 to keep sizes reasonable
		
		// Cap at MaxTradeSize if configured
		if l.config.MaxTradeSize != nil && size.Cmp(l.config.MaxTradeSize) > 0 {
			size = new(big.Int).Set(l.config.MaxTradeSize)
		}

		log.Printf("[%s] Leverage signal: change=%.2f%%, side=%s, size=%s", 
			l.Nickname(), priceChange*100, side, formatEther(size))

		return &side, size
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

// executeTrade performs a single trade
func (l *LeverageBot) executeTrade(ctx context.Context, side engine.TradeSide, size *big.Int) {
	var err error
	var txHash string

	if side == engine.BUY {
		txHash, err = l.executor.SwapETHForApples(ctx, l.privateKey, size)
	} else {
		txHash, err = l.executor.SwapApplesForETH(ctx, l.privateKey, size)
	}

	if err != nil {
		log.Printf("[%s] Trade failed: %v", l.Nickname(), err)
		return
	}

	log.Printf("[%s] ✓ Leverage trade executed: %s (%s, size: %s)", 
		l.Nickname(), txHash[:10]+"...", side, formatEther(size))
}
