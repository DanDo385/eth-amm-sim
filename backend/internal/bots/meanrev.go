// meanrev.go — EWMA-based mean reversion bot (accounts 4-6).
//
// Strategy: Subscribes to trade flow events (via metrics/trade_flow.go) and
// maintains an EWMA of liquidity-normalized log returns. When the z-score
// crosses configured trigger levels (e.g., ±0.75σ, ±1.0σ), the bot trades
// against the deviation — buying when price is below the moving average and
// selling when above.
//
// Three variants with different sensitivities:
//
//	MeanRev1: Fast (50-trade half-life), low thresholds → trades frequently
//	MeanRev2: Medium (100-trade half-life), moderate thresholds
//	MeanRev3: Slow (175-trade half-life), high thresholds → trades rarely
//
// CONNECTIONS:
//   - Subscribes to: metrics/trade_flow.go TradeFlowTracker for trade events
//   - Uses: metrics/ewma.go EWMACalculator for z-score computation
//   - Reads: metrics/price_provider.go PriceProvider for current price (to size sells)
//   - Executes: via engine/executor.go → contracts/src/AppleAMM.sol
//   - Config: config/accounts.go TriggerLevels and HalfLifeTrades parameters
package bots

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/metrics"
	"eth-amm-sim/internal/store"

	"github.com/ethereum/go-ethereum/crypto"
)

// MeanRevBot implements AMM-correct mean reversion using EWMA on liquidity-normalized log returns
//
// Key differences from price-level mean reversion:
// - Uses log returns (r = ln(priceAfter/priceBefore)) for stationarity
// - Normalizes by trade flow (q = dX/reserveX) to account for AMM price impact
// - Uses EWMA instead of rolling windows for adaptive statistics
// - Observes all trades except own trades (external market flow only)
type MeanRevBot struct {
	*BaseBot
	privateKey    *ecdsa.PrivateKey
	priceProvider metrics.PriceProvider
	store         *store.MemoryStore

	// EWMA state for liquidity-normalized returns
	ewmaState *metrics.EWMAState

	// Trading state
	tradedLevels     map[float64]bool // Track which sigma levels we've traded at (key is abs(zScore))
	lastZ            float64          // Last z-score (for zero-crossing detection)
	lastRTilde       float64          // Most recent rTilde (for recalculating z-score)
	tradesSinceTrade int              // Number of trades observed since last executed trade

	mu sync.RWMutex // Mutex for thread-safe access to trading state
}

// NewMeanRevBot creates a new mean reversion bot with EWMA
func NewMeanRevBot(cfg *config.AccountConfig, executor *engine.Executor, priceProvider metrics.PriceProvider, store *store.MemoryStore) (*MeanRevBot, error) {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key for %s: %w", cfg.Nickname, err)
	}

	// Initialize EWMA state with configured half-life
	halfLifeTrades := cfg.HalfLifeTrades
	if halfLifeTrades <= 0 {
		halfLifeTrades = 50 // Default half-life
	}
	ewmaState := metrics.NewEWMAState(halfLifeTrades)

	bot := &MeanRevBot{
		BaseBot:          NewBaseBot(cfg, executor, store),
		privateKey:       privateKey,
		priceProvider:    priceProvider,
		store:            store,
		ewmaState:        ewmaState,
		tradedLevels:     make(map[float64]bool),
		lastZ:            0,
		tradesSinceTrade: 0, // Start at 0, need half-life trades before first trade
	}

	// Subscribe to trade flow events (will receive all trades except own)
	tradeFlowTracker := store.GetTradeFlowTracker()
	tradeFlowTracker.Subscribe(bot)

	return bot, nil
}

// GetNickname returns the bot's nickname (required for TradeFlowSubscriber interface)
func (m *MeanRevBot) GetNickname() string {
	return m.Nickname()
}

// OnTradeFlow is called when an external trade occurs (not this bot's own trade)
// This is where we update EWMA statistics and check for trading signals
func (m *MeanRevBot) OnTradeFlow(event metrics.TradeFlowEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Skip events with zero or invalid normalized flow (rTilde would be 0 or invalid)
	// Normalized flow should be > 0 for meaningful liquidity-normalized returns
	if event.NormalizedFlow <= 0 || math.IsNaN(event.RTilde) || math.IsInf(event.RTilde, 0) {
		return
	}

	// Update EWMA statistics with new rTilde observation
	m.ewmaState.UpdateEWMA(event.RTilde)

	// Increment trades since last trade (for half-life requirement)
	m.tradesSinceTrade++

	// Store most recent rTilde for z-score recalculation
	m.lastRTilde = event.RTilde

	// Calculate z-score: z = (rTilde - mu) / sigma
	zScore := m.ewmaState.GetZScore(event.RTilde)

	// Check for zero-crossing (for MeanRev2/3 reset logic)
	// Reset traded levels when z-score crosses zero (sign change)
	if m.lastZ != 0 {
		if (m.lastZ > 0 && zScore <= 0) || (m.lastZ < 0 && zScore >= 0) {
			// Zero crossing detected - reset traded levels
			m.tradedLevels = make(map[float64]bool)
		}
	}
	m.lastZ = zScore

	// Check if we should trade based on z-score thresholds
	// This will be checked in the Run loop
}

// Run starts the mean reversion bot trading loop
func (m *MeanRevBot) Run(ctx context.Context) {
	// Log configuration
	var levelStr string
	if len(m.config.TriggerLevels) > 0 {
		levelStr = fmt.Sprintf("levels: %v", m.config.TriggerLevels)
	} else {
		levelStr = "no trigger levels configured"
	}
	log.Printf("[%s] MeanRev bot started (EWMA half-life: %d trades, %s)",
		m.Nickname(), m.config.HalfLifeTrades, levelStr)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] MeanRev bot stopped (context cancelled)", m.Nickname())
			return
		case <-m.stopCh:
			log.Printf("[%s] MeanRev bot stopped", m.Nickname())
			return
		default:
			// Check for trading signals more frequently (reduced delay for faster response)
			// Use shorter delay: 0.5-1.5 seconds instead of configured TradeFreqMin/Max
			delayMs := 500 + rand.Intn(1000) // 500-1500ms
			delay := time.Duration(delayMs) * time.Millisecond
			select {
			case <-ctx.Done():
				return
			case <-m.stopCh:
				return
			case <-time.After(delay):
			}

			// Check if we should trade based on current EWMA z-score
			side, size := m.checkMeanReversionSignal()
			if side != nil && size != nil && size.Sign() > 0 {
				// Check balance before attempting trade
				if m.hasSufficientBalance(ctx, *side, size) {
					m.executeTrade(ctx, *side, size)

					// MeanRev1: Reset EWMA after each executed trade (fast reset)
					// MeanRev2/3: Keep EWMA, only reset traded levels (already done on zero-crossing)
					if m.config.HalfLifeTrades <= 30 { // MeanRev1 threshold (halved from 60 since half-life is now 25)
						m.ewmaState.ResetEWMA()
						m.mu.Lock()
						m.tradedLevels = make(map[float64]bool)
						m.lastZ = 0
						m.mu.Unlock()
					}
					// Note: tradesSinceTrade is reset in executeTrade()
				}
			}
		}
	}
}

// checkMeanReversionSignal checks if z-score has crossed a trigger level
// Returns (side, size) if signal detected, (nil, nil) otherwise
func (m *MeanRevBot) checkMeanReversionSignal() (*engine.TradeSide, *big.Int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Require fewer trades observed since last trade (fraction of half-life for more frequent trading)
	halfLife := m.config.HalfLifeTrades
	if halfLife <= 0 {
		halfLife = 50 // Default
	}
	// Use ~20% of half-life as a minimum requirement (reduced from 35% to allow faster trading).
	// This ensures some observation period but allows bots to trade sooner.
	minTradesRequired := int(float64(halfLife) * 0.20)
	if minTradesRequired < 3 {
		minTradesRequired = 3 // Minimum of 3 trades to ensure some observation
	}

	if m.tradesSinceTrade < minTradesRequired {
		return nil, nil
	}

	// Require that we've observed at least one trade (lastRTilde should be set)
	// If lastRTilde is still 0, we haven't received any trade flow events yet
	if m.lastRTilde == 0 && m.ewmaState.GetTradeCount() == 0 {
		return nil, nil
	}

	// Require non-zero sigma for meaningful z-score
	// Sigma becomes non-zero after variance builds up (typically after 2+ observations)
	sigma := m.ewmaState.GetSigma()
	if sigma == 0 {
		// Log debug info if we have trades but no variance yet
		if m.ewmaState.GetTradeCount() > 0 {
			log.Printf("[%s] MeanRev waiting for variance: trades=%d, tradesSinceTrade=%d/%d",
				m.Nickname(), m.ewmaState.GetTradeCount(), m.tradesSinceTrade, minTradesRequired)
		}
		return nil, nil
	}

	// Recalculate z-score using most recent rTilde and current EWMA state
	// This ensures z-score is always relative to current statistics
	mu := m.ewmaState.GetMu()
	zScore := m.ewmaState.GetZScore(m.lastRTilde)
	absZScore := math.Abs(zScore)
	
	// Debug logging every 10 trades to monitor bot state
	if m.tradesSinceTrade%10 == 0 {
		log.Printf("[%s] MeanRev state: trades=%d, tradesSinceTrade=%d/%d, z=%.3f, mu=%.6f, sigma=%.6f, lastRTilde=%.6f",
			m.Nickname(), m.ewmaState.GetTradeCount(), m.tradesSinceTrade, minTradesRequired, zScore, mu, sigma, m.lastRTilde)
	}

	// Get trigger levels
	var triggerLevels []float64
	if len(m.config.TriggerLevels) > 0 {
		triggerLevels = m.config.TriggerLevels
	} else {
		return nil, nil
	}

	// Check each trigger level to see if we should trade
	// Trade at the highest level that's been crossed and hasn't been traded yet
	var bestLevel float64 = -1
	for _, level := range triggerLevels {
		if absZScore >= level {
			// Check if we've already traded at this level
			levelKey := math.Floor(level*10) / 10 // Round to 0.1 precision
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
			// Positive z-score: price moved up (rTilde > mu) - sell (fade the move)
			side = engine.SELL
		} else {
			// Negative z-score: price moved down (rTilde < mu) - buy (fade the move)
			side = engine.BUY
		}

		// Trade size: use MaxTradeSize (fixed size, not percentage-based)
		// For BUY: size is in ETH
		// For SELL: we'll convert to APPL in executeTrade
		size := new(big.Int).Set(m.config.MaxTradeSize)

		// Mark this level as traded
		levelKey := math.Floor(bestLevel*10) / 10
		m.tradedLevels[levelKey] = true

		mu := m.ewmaState.GetMu()
		sigma := m.ewmaState.GetSigma()
		log.Printf("[%s] MeanRev signal: z-score=%.2f, mu=%.6f, sigma=%.6f, level=%.2f std, side=%s, tradesSinceTrade=%d/%d",
			m.Nickname(), zScore, mu, sigma, bestLevel, side, m.tradesSinceTrade, halfLife)

		return &side, size
	}

	return nil, nil
}

// hasSufficientBalance checks if the bot has enough balance for the trade
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
		// For SELL: Check APPL balance
		appleBalance, err := m.executor.GetAPPLBalance(ctx, addr)
		if err != nil {
			return false
		}
		ethBalance, err := m.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		// Convert ETH size to APPL equivalent for balance check
		currentPrice := m.priceProvider.GetCurrentPrice()
		if currentPrice == 0 {
			return false
		}
		sizeFloat := new(big.Float).SetInt(size)
		priceFloat := big.NewFloat(currentPrice)
		appleNeeded := new(big.Float).Quo(sizeFloat, priceFloat)
		appleNeededInt, _ := appleNeeded.Int(nil)
		return appleBalance.Cmp(appleNeededInt) >= 0 && ethBalance.Cmp(gasReserve) >= 0
	}
}

// executeTrade performs a single trade
func (m *MeanRevBot) executeTrade(ctx context.Context, side engine.TradeSide, size *big.Int) {
	var err error
	var txHash string

	if side == engine.BUY {
		txHash, err = m.executor.SwapETHForApples(ctx, m.privateKey, size)
	} else {
		// For SELL: size is already in APPL (MaxTradeSize represents APPL for SELL)
		// But we need to convert MaxTradeSize (ETH) to APPL equivalent
		currentPrice := m.priceProvider.GetCurrentPrice()
		if currentPrice == 0 {
			log.Printf("[%s] Cannot execute SELL: no price data", m.Nickname())
			return
		}
		// Convert ETH size to APPL: appleAmount = ethAmount / price
		sizeFloat := new(big.Float).SetInt(size)
		priceFloat := big.NewFloat(currentPrice)
		appleAmount := new(big.Float).Quo(sizeFloat, priceFloat)
		appleAmountInt, _ := appleAmount.Int(nil)
		txHash, err = m.executor.SwapApplesForETH(ctx, m.privateKey, appleAmountInt)
	}

	if err != nil {
		log.Printf("[%s] Trade failed: %v", m.Nickname(), err)
		return
	}

	log.Printf("[%s] ✓ MeanRev trade executed: %s (%s, size: %s)",
		m.Nickname(), txHash[:10]+"...", side, formatEtherMeanRev(size))

	// Reset trades since last trade counter after executing trade
	m.mu.Lock()
	m.tradesSinceTrade = 0 // Reset counter - need another half-life before next trade
	m.mu.Unlock()
}

// formatEtherMeanRev formats a big.Int as ETH with 4 decimal places
func formatEtherMeanRev(wei *big.Int) string {
	if wei == nil {
		return "0.0000"
	}
	eth := new(big.Float).SetInt(wei)
	eth.Quo(eth, big.NewFloat(1e18))
	return fmt.Sprintf("%.4f", eth)
}
