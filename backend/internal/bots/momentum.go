// Package bots contains the momentum bot implementation
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
	"eth-amm-sim/internal/metrics"
	"eth-amm-sim/internal/store"
	"github.com/ethereum/go-ethereum/crypto"
)

// MomentumBot chases trends
type MomentumBot struct {
	*BaseBot
	privateKey   *ecdsa.PrivateKey
	priceProvider metrics.PriceProvider
	lastSignalPrice float64 // Track last price when signal fired to prevent duplicate trades
	lastCheckedPrice float64 // Track last price we checked for signals
}

// NewMomentumBot creates a new momentum bot
func NewMomentumBot(cfg *config.AccountConfig, executor *engine.Executor, priceProvider metrics.PriceProvider, store *store.MemoryStore) *MomentumBot {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key for " + cfg.Nickname)
	}

	return &MomentumBot{
		BaseBot:      NewBaseBot(cfg, executor, store),
		privateKey:   privateKey,
		priceProvider: priceProvider,
	}
}

// Run starts the momentum bot trading loop
func (m *MomentumBot) Run(ctx context.Context) {
	log.Printf("[%s] Momentum bot started (lookback: %d, trigger: %.1f%%)", 
		m.Nickname(), m.config.LookbackBlocks, m.config.TriggerPercent*100)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Momentum bot stopped (context cancelled)", m.Nickname())
			return
		case <-m.stopCh:
			log.Printf("[%s] Momentum bot stopped", m.Nickname())
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

			// Check if we should trade based on momentum signal
			side, size := m.checkMomentumSignal()
			if side != nil && size != nil && size.Sign() > 0 {
				// Only execute if price has changed since last signal (prevent duplicate trades)
				if currentPrice != m.lastSignalPrice {
					// Check balance before attempting trade
					if m.hasSufficientBalance(ctx, *side, size) {
						m.executeTrade(ctx, *side, size)
						m.lastSignalPrice = currentPrice // Mark that we've traded on this price
					}
				}
			}
		}
	}
}

// checkMomentumSignal checks if price has moved significantly (trend detection)
// Returns (side, size) if signal detected, (nil, nil) otherwise
func (m *MomentumBot) checkMomentumSignal() (*engine.TradeSide, *big.Int) {
	// Get price history - require at least lookback+1, but allow trading with less
	// Minimum 3 prices needed (current, previous, lookback)
	prices := m.priceProvider.GetRecentPrices(m.config.LookbackBlocks + 1)
	minRequired := 3
	if len(prices) < minRequired {
		return nil, nil
	}
	
	// If we don't have full lookback, use what we have (makes bots start trading sooner)
	lookback := m.config.LookbackBlocks
	if len(prices) < lookback+1 {
		lookback = len(prices) - 1
	}

	currentPrice := prices[len(prices)-1]
	lookbackPrice := prices[len(prices)-1-lookback]

	// Calculate percent change
	priceChange := (currentPrice - lookbackPrice) / lookbackPrice

	// Trigger if price moved more than TriggerPercent
	if math.Abs(priceChange) >= m.config.TriggerPercent {
		var side engine.TradeSide
		if priceChange > 0 {
			// Price going up - buy (chase the trend)
			side = engine.BUY
		} else {
			// Price going down - sell (chase the trend)
			side = engine.SELL
		}

		// Trade size based on momentum strength (stronger move = larger trade)
		momentumStrength := math.Min(math.Abs(priceChange) / m.config.TriggerPercent, 3.0) // Cap at 3x
		size := new(big.Int).Mul(m.config.MaxTradeSize, big.NewInt(int64(momentumStrength * 100)))
		size.Div(size, big.NewInt(100))

		log.Printf("[%s] Momentum signal: change=%.2f%%, price=%.6f→%.6f, side=%s", 
			m.Nickname(), priceChange*100, lookbackPrice, currentPrice, side)

		return &side, size
	}

	return nil, nil
}

// hasSufficientBalance checks if the bot has enough balance for the trade
// Note: For SELL trades, we don't check APPL balance - bots can build short positions over time
// The AMM contract will reject trades if APPL balance is insufficient
func (m *MomentumBot) hasSufficientBalance(ctx context.Context, side engine.TradeSide, size *big.Int) bool {
	addr := crypto.PubkeyToAddress(m.privateKey.PublicKey)
	
	// Reserve some ETH for gas (0.01 ETH)
	gasReserve := new(big.Int).Mul(big.NewInt(1e16), big.NewInt(1)) // 0.01 ETH
	
	if side == engine.BUY {
		// Need ETH to buy (trade size + gas)
		ethBalance, err := m.executor.GetETHBalance(ctx, addr)
		if err != nil {
			log.Printf("[%s] Failed to get ETH balance: %v", m.Nickname(), err)
			return false
		}
		required := new(big.Int).Add(size, gasReserve)
		return ethBalance.Cmp(required) >= 0
	} else {
		// For SELL: Only check ETH for gas (allow short positions)
		// The AMM contract will handle APPL balance validation
		ethBalance, err := m.executor.GetETHBalance(ctx, addr)
		if err != nil {
			log.Printf("[%s] Failed to get ETH balance: %v", m.Nickname(), err)
			return false
		}
		return ethBalance.Cmp(gasReserve) >= 0
	}
}

// executeTrade performs a single trade
func (m *MomentumBot) executeTrade(ctx context.Context, side engine.TradeSide, size *big.Int) {
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

	log.Printf("[%s] ✓ Momentum trade executed: %s (%s, size: %s)", 
		m.Nickname(), txHash[:10]+"...", side, formatEther(size))
}
