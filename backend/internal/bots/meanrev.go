// Package bots contains the mean reversion bot implementation
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

// MeanRevBot fades extreme price moves
type MeanRevBot struct {
	*BaseBot
	privateKey   *ecdsa.PrivateKey
	priceProvider PriceProvider
	lastSignalPrice float64 // Track last price when signal fired
}

// NewMeanRevBot creates a new mean reversion bot
func NewMeanRevBot(cfg *config.AccountConfig, executor *engine.Executor, priceProvider PriceProvider) *MeanRevBot {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("invalid private key for " + cfg.Nickname)
	}

	return &MeanRevBot{
		BaseBot:      NewBaseBot(cfg, executor),
		privateKey:   privateKey,
		priceProvider: priceProvider,
	}
}

// Run starts the mean reversion bot trading loop
func (m *MeanRevBot) Run(ctx context.Context) {
	log.Printf("[%s] MeanRev bot started (lookback: %d, sigma: %.2f)", 
		m.Nickname(), m.config.LookbackBlocks, m.config.TriggerSigma)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] MeanRev bot stopped (context cancelled)", m.Nickname())
			return
		case <-m.stopCh:
			log.Printf("[%s] MeanRev bot stopped", m.Nickname())
			return
		default:
			// Check price every few seconds
			delay := m.RandomDelay()
			select {
			case <-ctx.Done():
				return
			case <-m.stopCh:
				return
			case <-time.After(delay):
			}

			// Check if we should trade based on mean reversion signal
			side, size := m.checkMeanReversionSignal()
			if side != nil && size != nil && size.Sign() > 0 {
				// Get current price to check if we've already traded on this signal
				currentPrice := m.priceProvider.GetCurrentPrice()
				
				// Only execute if price has changed since last signal
				if currentPrice != m.lastSignalPrice {
					// Check balance before attempting trade
					if m.hasSufficientBalance(ctx, *side, size) {
						m.executeTrade(ctx, *side, size)
						m.lastSignalPrice = currentPrice
					}
				}
			}
		}
	}
}

// checkMeanReversionSignal checks if price has deviated significantly from mean
// Returns (side, size) if signal detected, (nil, nil) otherwise
func (m *MeanRevBot) checkMeanReversionSignal() (*engine.TradeSide, *big.Int) {
	// Get price history
	prices := m.priceProvider.GetRecentPrices(m.config.LookbackBlocks)
	if len(prices) < m.config.LookbackBlocks {
		// Not enough history yet
		return nil, nil
	}

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

	// Trigger if price is more than TriggerSigma standard deviations from mean
	if math.Abs(zScore) >= m.config.TriggerSigma {
		var side engine.TradeSide
		if zScore > 0 {
			// Price is above mean - sell (fade the move)
			side = engine.SELL
		} else {
			// Price is below mean - buy (fade the move)
			side = engine.BUY
		}

		// Trade size based on deviation magnitude (stronger signal = larger trade)
		sizeMultiplier := math.Min(math.Abs(zScore) / m.config.TriggerSigma, 2.0) // Cap at 2x
		size := new(big.Int).Mul(m.config.MaxTradeSize, big.NewInt(int64(sizeMultiplier * 100)))
		size.Div(size, big.NewInt(100))

		log.Printf("[%s] MeanRev signal: z-score=%.2f, price=%.6f, mean=%.6f, side=%s", 
			m.Nickname(), zScore, currentPrice, mean, side)

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
