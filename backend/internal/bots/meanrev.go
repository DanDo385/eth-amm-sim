// Package bots contains the mean reversion bot implementation
package bots

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/metrics"
	"eth-amm-sim/internal/store"
	"github.com/ethereum/go-ethereum/crypto"
)

// MeanRevBot fades extreme price moves
type MeanRevBot struct {
	*BaseBot
	privateKey   *ecdsa.PrivateKey
	priceProvider metrics.PriceProvider
	lastSignalPrice float64 // Track last price when signal fired
	lastCheckedPrice float64 // Track last price we checked for signals
	tradedLevels map[float64]bool // Track which sigma levels we've traded at (key is abs(zScore))
}

// NewMeanRevBot creates a new mean reversion bot
func NewMeanRevBot(cfg *config.AccountConfig, executor *engine.Executor, priceProvider metrics.PriceProvider, store *store.MemoryStore) *MeanRevBot {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key for " + cfg.Nickname)
	}

	return &MeanRevBot{
		BaseBot:      NewBaseBot(cfg, executor, store),
		privateKey:   privateKey,
		priceProvider: priceProvider,
		tradedLevels: make(map[float64]bool),
	}
}

// Run starts the mean reversion bot trading loop
func (m *MeanRevBot) Run(ctx context.Context) {
	// Log trigger levels if available, otherwise fallback to TriggerSigma
	var levelStr string
	if len(m.config.TriggerLevels) > 0 {
		levelStr = fmt.Sprintf("levels: %v", m.config.TriggerLevels)
	} else {
		levelStr = fmt.Sprintf("sigma: %.2f", m.config.TriggerSigma)
	}
	log.Printf("[%s] MeanRev bot started (lookback: %d, %s)", 
		m.Nickname(), m.config.LookbackBlocks, levelStr)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] MeanRev bot stopped (context cancelled)", m.Nickname())
			return
		case <-m.stopCh:
			log.Printf("[%s] MeanRev bot stopped", m.Nickname())
			return
		default:
			// Check if stopped out
			if m.IsStoppedOut() {
				select {
				case <-ctx.Done():
					return
				case <-m.stopCh:
					return
				case <-time.After(10 * time.Second):
				}
				continue
			}
			
			// Check price every few seconds
			delay := m.RandomDelay()
			select {
			case <-ctx.Done():
				return
			case <-m.stopCh:
				return
			case <-time.After(delay):
			}

			// Get current price first
			currentPrice := m.priceProvider.GetCurrentPrice()
			if currentPrice == 0 {
				continue // No price data yet
			}

			// Skip signal checking if price hasn't changed since last check
			// This prevents flooding logs with repeated signal detections
			if currentPrice == m.lastCheckedPrice {
				continue
			}
			m.lastCheckedPrice = currentPrice

			// Check if we should trade based on mean reversion signal
			side, size := m.checkMeanReversionSignal()
			if side != nil && size != nil && size.Sign() > 0 {
				// Only execute if price has changed since last signal
				if currentPrice != m.lastSignalPrice {
					// Check balance before attempting trade
					if m.hasSufficientBalance(ctx, *side, size) {
						m.executeTrade(ctx, *side, size)
						m.lastSignalPrice = currentPrice
					}
				}
			} else {
				// If no signal, check if price has moved back toward mean
				// Reset traded levels if price crosses back through mean (allows re-trading)
				// This allows the bot to trade again if price moves away from mean again
				prices := m.priceProvider.GetRecentPrices(m.config.LookbackBlocks)
				if len(prices) >= 5 {
					mean := 0.0
					for _, p := range prices {
						mean += p
					}
					mean /= float64(len(prices))
					
					// If price crosses back through mean (changes sign of deviation), reset traded levels
					// This allows the bot to trade at levels again if price moves away in opposite direction
					prevDeviation := m.lastSignalPrice - mean
					currentDeviation := currentPrice - mean
					if (prevDeviation > 0 && currentDeviation <= 0) || (prevDeviation < 0 && currentDeviation >= 0) {
						// Price crossed mean - reset traded levels to allow trading again
						m.tradedLevels = make(map[float64]bool)
					}
				}
			}
		}
	}
}

// checkMeanReversionSignal checks if price has deviated significantly from mean
// Returns (side, size) if signal detected, (nil, nil) otherwise
func (m *MeanRevBot) checkMeanReversionSignal() (*engine.TradeSide, *big.Int) {
	// Get price history - use requested lookback or available data
	prices := m.priceProvider.GetRecentPrices(m.config.LookbackBlocks)
	
	// Require minimum 5 price points to calculate meaningful statistics
	// This allows bots to start trading much sooner (~10 seconds instead of 40+)
	minRequired := 5
	if len(prices) < minRequired {
		return nil, nil
	}
	
	// Use available prices (may be less than requested lookback)
	// This makes bots more responsive and allows earlier trading

	currentPrice := prices[len(prices)-1]
	
	// Calculate mean and standard deviation
	mean := 0.0
	for _, p := range prices {
		mean += p
	}
	mean /= float64(len(prices))

	// Calculate standard deviation
	variance := 0.0
	for _, p := range prices {
		variance += math.Pow(p-mean, 2)
	}
	variance /= float64(len(prices))
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return nil, nil // No volatility
	}

	// Calculate z-score (how many standard deviations from mean)
	zScore := (currentPrice - mean) / stdDev
	absZScore := math.Abs(zScore)

	// Get trigger levels - use TriggerLevels if available, otherwise fallback to TriggerSigma
	var triggerLevels []float64
	if len(m.config.TriggerLevels) > 0 {
		triggerLevels = m.config.TriggerLevels
	} else if m.config.TriggerSigma > 0 {
		// Fallback to single TriggerSigma for backward compatibility
		triggerLevels = []float64{m.config.TriggerSigma}
	} else {
		return nil, nil // No trigger levels configured
	}

	// Check each trigger level to see if we should trade
	// Trade at the highest level that's been crossed and hasn't been traded yet
	var bestLevel float64 = -1
	for _, level := range triggerLevels {
		if absZScore >= level {
			// Check if we've already traded at this level
			// Use a small tolerance (0.1 std) to account for price movement
			levelKey := math.Floor(level * 10) / 10 // Round to 0.1 precision
			if !m.tradedLevels[levelKey] {
				if level > bestLevel {
					bestLevel = level
				}
			}
		}
	}

	// If we found a level to trade at
	if bestLevel > 0 {
		var side engine.TradeSide
		if zScore > 0 {
			// Price is above mean - sell (fade the move)
			side = engine.SELL
		} else {
			// Price is below mean - buy (fade the move)
			side = engine.BUY
		}

		// Trade size scales with level - more extreme levels get larger trades
		// Base size at lowest level, scales up for higher levels
		baseLevel := triggerLevels[0]
		levelMultiplier := 1.0 + (bestLevel - baseLevel) * 0.3 // 30% increase per level
		sizeMultiplier := math.Min(levelMultiplier, 2.5) // Cap at 2.5x
		size := new(big.Int).Mul(m.config.MaxTradeSize, big.NewInt(int64(sizeMultiplier * 100)))
		size.Div(size, big.NewInt(100))

		// Mark this level as traded
		levelKey := math.Floor(bestLevel * 10) / 10
		m.tradedLevels[levelKey] = true

		log.Printf("[%s] MeanRev signal: z-score=%.2f, price=%.6f, mean=%.6f, level=%.2f std, side=%s", 
			m.Nickname(), zScore, currentPrice, mean, bestLevel, side)

		return &side, size
	}

	return nil, nil
}

// hasSufficientBalance checks if the bot has enough balance for the trade
// Note: For SELL trades, we don't check APPL balance - bots can build short positions over time
func (m *MeanRevBot) hasSufficientBalance(ctx context.Context, side engine.TradeSide, size *big.Int) bool {
	addr := crypto.PubkeyToAddress(m.privateKey.PublicKey)
	gasReserve := new(big.Int).Mul(big.NewInt(1e16), big.NewInt(1)) // 0.01 ETH
	
	if side == engine.BUY {
		ethBalance, err := m.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		required := new(big.Int).Add(size, gasReserve)
		return ethBalance.Cmp(required) >= 0
	} else {
		// For SELL: Only check ETH for gas (allow short positions)
		ethBalance, err := m.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		return ethBalance.Cmp(gasReserve) >= 0
	}
}

// executeTrade performs a single trade
func (m *MeanRevBot) executeTrade(ctx context.Context, side engine.TradeSide, size *big.Int) {
	var err error
	var txHash string

	if side == engine.BUY {
		txHash, err = m.executor.SwapETHForApples(ctx, m.privateKey, size)
	} else {
		txHash, err = m.executor.SwapApplesForETH(ctx, m.privateKey, size)
	}

	if err != nil {
		log.Printf("[%s] Trade failed: %v", m.Nickname(), err)
		return
	}

	log.Printf("[%s] ✓ MeanRev trade executed: %s (%s, size: %s)", 
		m.Nickname(), txHash[:10]+"...", side, formatEther(size))
}
