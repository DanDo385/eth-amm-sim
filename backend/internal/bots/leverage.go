// leverage.go — Leveraged momentum trader bot (accounts 25-27).
//
// Strategy: Uses EWMA-based momentum detection similar to meanrev bots, but trades
// WITH momentum instead of against it. Subscribes to trade flow events and maintains
// EWMA of liquidity-normalized log returns. When z-score crosses thresholds, the
// bot trades in the direction of momentum — buying strength (z > 0) and selling
// weakness (z < 0).
//
// Three variants with different leverage levels:
//
//	Lev5x:  ~12 trade half-life (from MeanRev1's 50/4)
//	Lev10x: ~25 trade half-life (from MeanRev2's 100/4)
//	Lev25x: ~44 trade half-life (from MeanRev3's 175/4)
//
// Trade size scales with volatility: higher sigma = larger positions
//
// CONNECTIONS:
//   - Subscribes to: metrics/trade_flow.go TradeFlowTracker for trade events
//   - Uses: metrics/ewma.go EWMACalculator for z-score computation
//   - Executes: executor.OpenLeveragedPosition → AppleAMM.openLeveragedPosition()
//   - Monitored by: liquidator.go checks isLiquidatable on these accounts
//   - Config: config/accounts.go defines Leverage, Collateral, LiquidationThreshold
package bots

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math"
	"math/big"
	"sync"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/metrics"
	"eth-amm-sim/internal/store"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// LeverageBot implements momentum trading using EWMA on liquidity-normalized log returns
// Key differences from mean reversion:
// - Uses same EWMA methodology but trades WITH momentum (not against it)
// - Uses 1/4th the half-life of meanrev bots for faster adaptation
// - Scales position size with volatility (higher sigma = larger positions)
type LeverageBot struct {
	*BaseBot
	privateKey *ecdsa.PrivateKey
	store      *store.MemoryStore

	// EWMA state for liquidity-normalized returns (1/4th half-life of meanrev)
	ewmaState *metrics.EWMAState

	// Trading state
	tradedLevels     map[float64]bool // Track which sigma levels we've traded at
	lastZ            float64          // Last z-score (for zero-crossing detection)
	lastRTilde       float64          // Most recent rTilde (for recalculating z-score)
	tradesSinceTrade int              // Number of trades observed since last executed trade
	hasTradedOnce    bool             // Track if bot has executed at least one trade (for warm-up bypass)

	mu sync.RWMutex // Mutex for thread-safe access to trading state
}

// NewLeverageBot creates a new leverage bot with EWMA momentum strategy
func NewLeverageBot(cfg *config.AccountConfig, executor *engine.Executor, priceProvider metrics.PriceProvider, store *store.MemoryStore) (*LeverageBot, error) {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key for %s: %w", cfg.Nickname, err)
	}

	// Calculate half-life as 1/4th of meanrev bots
	// MeanRev1: 50 → Lev5x: ~12
	// MeanRev2: 100 → Lev10x: ~25
	// MeanRev3: 175 → Lev25x: ~44
	halfLifeTrades := calculateLeverageHalfLife(cfg.Leverage)
	ewmaState := metrics.NewEWMAState(halfLifeTrades)

	bot := &LeverageBot{
		BaseBot:          NewBaseBot(cfg, executor, store),
		privateKey:       privateKey,
		store:            store,
		ewmaState:        ewmaState,
		tradedLevels:     make(map[float64]bool),
		lastZ:            0,
		tradesSinceTrade: 0, // Start at 0, but first trade bypasses half-life requirement
		hasTradedOnce:    false,
	}

	// Subscribe to trade flow events (will receive all trades except own)
	tradeFlowTracker := store.GetTradeFlowTracker()
	tradeFlowTracker.Subscribe(bot)

	return bot, nil
}

// calculateLeverageHalfLife returns 1/4th of the corresponding meanrev half-life
// Updated to match halved meanrev half-lives: MeanRev1=25, MeanRev2=50, MeanRev3=87
func calculateLeverageHalfLife(leverage int) int {
	// Map leverage to meanrev half-life, then divide by 4
	switch leverage {
	case 5:
		return 6 // MeanRev1: 25 / 4 = 6.25 → 6
	case 10:
		return 12 // MeanRev2: 50 / 4 = 12.5 → 12
	case 25:
		return 22 // MeanRev3: 87 / 4 = 21.75 → 22
	default:
		return 6 // Default to Lev5x half-life
	}
}

// GetNickname returns the bot's nickname (required for TradeFlowSubscriber interface)
func (l *LeverageBot) GetNickname() string {
	return l.Nickname()
}

// OnTradeFlow is called when an external trade occurs (not this bot's own trade)
// This is where we update EWMA statistics
func (l *LeverageBot) OnTradeFlow(event metrics.TradeFlowEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Update EWMA statistics with new rTilde observation
	l.ewmaState.UpdateEWMA(event.RTilde)

	// Increment trades since last trade (for half-life requirement)
	l.tradesSinceTrade++

	// Store most recent rTilde for z-score recalculation
	l.lastRTilde = event.RTilde

	// Calculate z-score: z = (rTilde - mu) / sigma
	zScore := l.ewmaState.GetZScore(event.RTilde)

	// Check for zero-crossing (reset traded levels when z-score crosses zero)
	if l.lastZ != 0 {
		if (l.lastZ > 0 && zScore <= 0) || (l.lastZ < 0 && zScore >= 0) {
			// Zero crossing detected - reset traded levels
			l.tradedLevels = make(map[float64]bool)
		}
	}
	l.lastZ = zScore
}

// Run starts the leverage bot trading loop
func (l *LeverageBot) Run(ctx context.Context) {
	log.Printf("[%s] Leverage bot started (%dx leverage, EWMA half-life: %d trades)",
		l.Nickname(), l.config.Leverage, calculateLeverageHalfLife(l.config.Leverage))

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Leverage bot stopped (context cancelled)", l.Nickname())
			return
		case <-l.stopCh:
			log.Printf("[%s] Leverage bot stopped", l.Nickname())
			return
		default:
			// Check for trading signals periodically
			// Trade more frequently with higher leverage
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

			// Check for momentum signal using EWMA z-score
			side, size := l.checkMomentumSignal()
			if side != nil && size != nil && size.Sign() > 0 {
				// For BUY (leveraged positions), check if we already have a position
				if *side == engine.BUY {
					addr := crypto.PubkeyToAddress(l.privateKey.PublicKey)
					hasPosition, err := l.hasOpenPosition(ctx, addr)
					if err == nil && hasPosition {
						// Already have a leveraged position, skip opening new one
						log.Printf("[%s] Skipping leveraged BUY: existing position detected for %s", l.Nickname(), addr.Hex())
						continue
					}
				}

				// Check balance before attempting trade
				if l.hasSufficientBalance(ctx, *side, size) {
					if *side == engine.BUY {
						// Open leveraged position through contract
						l.openLeveragedPosition(ctx, size)
					} else {
						// Sell APPL for ETH (sell into weakness)
						l.executeSell(ctx, size)
					}
				} else {
					log.Printf("[%s] Signal present but insufficient balance for %s trade (size=%s)",
						l.Nickname(), tradeSideToString(*side), formatEther(size))
				}
			}
		}
	}
}

// checkMomentumSignal checks if z-score indicates momentum to trade WITH
// Returns (side, size) if signal detected, (nil, nil) otherwise
func (l *LeverageBot) checkMomentumSignal() (*engine.TradeSide, *big.Int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	halfLife := calculateLeverageHalfLife(l.config.Leverage)

	// Warm-up / gating logic:
	// - First trade: require only a small fraction of half-life (ensures at least one trade early)
	// - Subsequent trades: require ~half of half-life between trades (still conservative but active)
	requiredTrades := halfLife / 2
	if requiredTrades < 2 {
		requiredTrades = 2
	}
	if !l.hasTradedOnce {
		// First trade: be more permissive – use quarter half-life but never less than 2 trades
		requiredTrades = halfLife / 4
		if requiredTrades < 2 {
			requiredTrades = 2
		}
	}

	if l.tradesSinceTrade < requiredTrades {
		log.Printf("[%s] Signal check skipped: tradesSinceTrade=%d < required=%d (halfLife=%d, hasTradedOnce=%v)",
			l.Nickname(), l.tradesSinceTrade, requiredTrades, halfLife, l.hasTradedOnce)
		return nil, nil
	}

	// Require non-zero sigma (variance) for meaningful z-score
	sigma := l.ewmaState.GetSigma()
	if sigma == 0 {
		log.Printf("[%s] Signal check skipped: sigma=0 (EWMA not warmed up yet, tradesSinceTrade=%d)",
			l.Nickname(), l.tradesSinceTrade)
		return nil, nil
	}

	// Recalculate z-score using most recent rTilde and current EWMA state
	// If lastRTilde is 0 (no trades received yet), skip
	if l.lastRTilde == 0 {
		log.Printf("[%s] Signal check skipped: lastRTilde=0 (no trade flow events received yet)",
			l.Nickname())
		return nil, nil
	}

	zScore := l.ewmaState.GetZScore(l.lastRTilde)
	absZScore := math.Abs(zScore)
	mu := l.ewmaState.GetMu()

	// Trigger levels: slightly looser than MeanRev1 so leverage bots see more signals
	// while still requiring meaningful deviations
	triggerLevels := []float64{0.5, 0.75, 1.0}

	// Check each trigger level to see if we should trade
	var bestLevel float64 = -1
	for _, level := range triggerLevels {
		if absZScore >= level {
			// Check if we've already traded at this level
			levelKey := math.Floor(level*10) / 10 // Round to 0.1 precision
			if !l.tradedLevels[levelKey] {
				if level > bestLevel {
					bestLevel = level
				}
			}
		}
	}

	// If we found a level to trade at
	if bestLevel > 0 {
		log.Printf("[%s] ✓ Signal detected: z-score=%.2f, level=%.2f std, mu=%.6f, sigma=%.6f, tradesSinceTrade=%d/%d",
			l.Nickname(), zScore, bestLevel, mu, sigma, l.tradesSinceTrade, requiredTrades)

		// Trade WITH momentum (opposite of meanrev):
		// - z-score > 0 (price moving up) → BUY (buy strength) - open leveraged long
		// - z-score < 0 (price moving down) → SELL (sell weakness) - sell APPL for ETH
		if zScore > 0 {
			// BUY: Open leveraged long position
			// Get base collateral from config
			baseCollateral := l.config.Collateral
			if baseCollateral == nil {
				// Default based on leverage
				baseCollateral = new(big.Int).Mul(big.NewInt(50), big.NewInt(1e18))
			}

			// Scale collateral with volatility (sigma)
			// Higher volatility = larger positions
			volatilityMultiplier := 1.0 + (sigma * 10.0) // Scale sigma by 10x for meaningful impact
			if volatilityMultiplier < 1.0 {
				volatilityMultiplier = 1.0 // Don't reduce below base
			}
			if volatilityMultiplier > 2.0 {
				volatilityMultiplier = 2.0 // Cap at 2x base
			}

			// Calculate scaled collateral
			scaledCollateral := new(big.Float).SetInt(baseCollateral)
			scaledCollateral.Mul(scaledCollateral, big.NewFloat(volatilityMultiplier))
			collateral, _ := scaledCollateral.Int(nil)

			log.Printf("[%s] Momentum BUY signal: z-score=%.2f, mu=%.6f, sigma=%.6f, level=%.2f std, collateral=%s (volatility scaled: %.2fx), tradesSinceTrade=%d/%d",
				l.Nickname(), zScore, mu, sigma, bestLevel, formatEther(collateral), volatilityMultiplier, l.tradesSinceTrade, halfLife)

			// Mark this level as traded
			levelKey := math.Floor(bestLevel*10) / 10
			l.tradedLevels[levelKey] = true

			side := engine.BUY
			return &side, collateral
		} else if zScore < 0 {
			// SELL: Sell APPL for ETH (sell into weakness)
			// Use MaxTradeSize as base size, scaled by volatility
			baseSize := l.config.MaxTradeSize
			if baseSize == nil || baseSize.Sign() == 0 {
				// Default size
				baseSize = new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)) // 100 APPL default
			}

			// Scale size with volatility (higher volatility = larger sells)
			volatilityMultiplier := 1.0 + (sigma * 10.0)
			if volatilityMultiplier < 1.0 {
				volatilityMultiplier = 1.0
			}
			if volatilityMultiplier > 2.0 {
				volatilityMultiplier = 2.0
			}

			// Calculate scaled size (in APPL)
			scaledSize := new(big.Float).SetInt(baseSize)
			scaledSize.Mul(scaledSize, big.NewFloat(volatilityMultiplier))
			size, _ := scaledSize.Int(nil)

			log.Printf("[%s] Momentum SELL signal: z-score=%.2f, mu=%.6f, sigma=%.6f, level=%.2f std, size=%s APPL (volatility scaled: %.2fx), tradesSinceTrade=%d/%d",
				l.Nickname(), zScore, mu, sigma, bestLevel, formatEther(size), volatilityMultiplier, l.tradesSinceTrade, halfLife)

			// Mark this level as traded
			levelKey := math.Floor(bestLevel*10) / 10
			l.tradedLevels[levelKey] = true

			side := engine.SELL
			return &side, size
		}
	}

	// No signal detected
	log.Printf("[%s] Signal check: no trigger level reached (z-score=%.2f, absZScore=%.2f, mu=%.6f, sigma=%.6f)",
		l.Nickname(), zScore, absZScore, mu, sigma)
	return nil, nil
}

// hasSufficientBalance checks if the bot has enough balance for the trade
func (l *LeverageBot) hasSufficientBalance(ctx context.Context, side engine.TradeSide, size *big.Int) bool {
	addr := crypto.PubkeyToAddress(l.privateKey.PublicKey)
	gasReserve := new(big.Int).Mul(big.NewInt(1e16), big.NewInt(1)) // 0.01 ETH

	if side == engine.BUY {
		// BUY: Check ETH balance for collateral
		ethBalance, err := l.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		required := new(big.Int).Add(size, gasReserve)
		return ethBalance.Cmp(required) >= 0
	} else {
		// SELL: Check APPL balance (size is in APPL)
		appleBalance, err := l.executor.GetAPPLBalance(ctx, addr)
		if err != nil {
			return false
		}

		// Check ETH for gas
		ethBalance, err := l.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}

		return appleBalance.Cmp(size) >= 0 && ethBalance.Cmp(gasReserve) >= 0
	}
}

// tradeSideToString is a small helper for logging
func tradeSideToString(side engine.TradeSide) string {
	if side == engine.BUY {
		return "BUY"
	}
	return "SELL"
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

	// Reset trades since last trade counter after executing trade
	l.mu.Lock()
	l.tradesSinceTrade = 0 // Reset counter - need another half-life before next trade
	l.hasTradedOnce = true // Mark that we've executed at least one trade
	l.mu.Unlock()
}

// executeSell sells APPL for ETH (sell into weakness)
func (l *LeverageBot) executeSell(ctx context.Context, appleAmount *big.Int) {
	txHash, err := l.executor.SwapApplesForETH(ctx, l.privateKey, appleAmount)
	if err != nil {
		log.Printf("[%s] Failed to sell APPL: %v", l.Nickname(), err)
		return
	}

	log.Printf("[%s] ✓ Sold APPL: %s (amount: %s)",
		l.Nickname(), txHash[:10]+"...", formatEther(appleAmount))

	// Reset trades since last trade counter after executing trade
	l.mu.Lock()
	l.tradesSinceTrade = 0 // Reset counter - need another half-life before next trade
	l.hasTradedOnce = true // Mark that we've executed at least one trade
	l.mu.Unlock()
}

// hasOpenPosition checks if the bot already has an open position
func (l *LeverageBot) hasOpenPosition(ctx context.Context, addr common.Address) (bool, error) {
	pos, err := l.executor.GetTraderPosition(ctx, addr)
	if err != nil {
		return false, err
	}

	// Position exists if collateral is non-zero
	return pos != nil && pos.Collateral != nil && pos.Collateral.Sign() > 0, nil
}
