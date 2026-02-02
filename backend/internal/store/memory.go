// Package store provides in-memory storage for all simulation metrics and events.
//
// SYSTEM ROLE:
// MemoryStore is the single source of truth for all backend state. It composes
// all metrics objects (price, LP, account, impact curve, trade flow) and holds
// the trade blotter and key events list. Every write comes from either the
// trade callback (main.go) or the price poll loop (main.go), and every read
// goes to the server's REST/WebSocket handlers.
//
// MemoryStore also implements the metrics.PriceDataStore interface, which lets
// metrics/price_provider.go wrap it as a PriceProvider for bots — breaking the
// import cycle between store and metrics packages.
//
// CONNECTIONS:
//   - Writers: main.go (RecordTrade, RecordPrice, RecordTradeFlow, RecordEvent)
//   - Readers: server/handlers.go (all GET endpoints read from MemoryStore)
//   - Interface: implements metrics.PriceDataStore (GetCandles, GetTWAP, GetVolatility)
//   - Reset: server/handlers.go handleSessionReset clears price, LP, trades, events
package store

import (
	"log"
	"math"
	"math/big"
	"sync"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/metrics"

	"github.com/ethereum/go-ethereum/common"
)

// MemoryStore holds all simulation state in memory
type MemoryStore struct {
	mu sync.RWMutex

	// Price metrics
	priceMetrics *metrics.PriceMetrics

	// LP metrics
	lpMetrics *metrics.LPMetrics

	// Account metrics
	accountMetrics *metrics.AccountMetricsManager

	// Impact curve
	impactCurve *metrics.ImpactCurve

	// Trade flow tracker (for EWMA mean reversion)
	tradeFlowTracker *metrics.TradeFlowTracker

	// Trade blotter
	trades []engine.Trade

	// Key events
	events []KeyEvent
}

// KeyEvent represents a significant event
type KeyEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // trade, strategy_trigger
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // info, warning, critical
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		priceMetrics:     metrics.NewPriceMetrics(5*time.Second, 60*time.Second),
		lpMetrics:        metrics.NewLPMetrics(),
		accountMetrics:   metrics.NewAccountMetricsManager(),
		impactCurve:      metrics.NewImpactCurve(),
		tradeFlowTracker: metrics.NewTradeFlowTracker(),
		trades:           make([]engine.Trade, 0),
		events:           make([]KeyEvent, 0),
	}
}

// Price methods

func (s *MemoryStore) RecordPrice(price float64) {
	s.priceMetrics.RecordPrice(price)
}

func (s *MemoryStore) GetPriceMetrics() *metrics.PriceMetrics {
	return s.priceMetrics
}

func (s *MemoryStore) GetCandles() []metrics.Candle {
	return s.priceMetrics.GetCandles()
}

func (s *MemoryStore) GetTWAP() float64 {
	return s.priceMetrics.GetTWAP()
}

func (s *MemoryStore) GetVolatility() float64 {
	return s.priceMetrics.GetVolatility()
}

// LP methods

func (s *MemoryStore) GetLPMetrics() *metrics.LPMetrics {
	return s.lpMetrics
}

func (s *MemoryStore) GetLPData() metrics.LPMetricsData {
	return s.lpMetrics.GetMetrics()
}

// Account methods

func (s *MemoryStore) GetAccountMetrics(nickname string) *metrics.AccountMetrics {
	return s.accountMetrics.Get(nickname)
}

func (s *MemoryStore) GetOrCreateAccountMetrics(nickname string, address common.Address, initialEquity float64) *metrics.AccountMetrics {
	return s.accountMetrics.GetOrCreate(nickname, address, initialEquity)
}

func (s *MemoryStore) GetAllAccountPerformance() []metrics.PerformanceData {
	vol := s.currentSessionVol()
	return s.accountMetrics.GetAllPerformanceWithVol(vol)
}

func (s *MemoryStore) GetAccountPerformance(nickname string) *metrics.PerformanceData {
	am := s.accountMetrics.Get(nickname)
	if am == nil {
		return nil
	}
	vol := s.currentSessionVol()
	perf := am.GetPerformance(vol)
	return &perf
}

// currentSessionVol returns the session volatility. If a finalized value has
// been stored it is used; otherwise it is computed live from current candles.
func (s *MemoryStore) currentSessionVol() float64 {
	if v := s.accountMetrics.GetSessionVolatility(); v > 0 {
		return v
	}
	return computeSessionDailyVol(s.priceMetrics.GetCandles())
}

func (s *MemoryStore) GetAccountMetricsManager() *metrics.AccountMetricsManager {
	return s.accountMetrics
}

// ResetAccountsForSession resets all accounts to track session-only performance
// Requires a function to get current balances for each account and current spot price
func (s *MemoryStore) ResetAccountsForSession(getBalance func(nickname string) (ethBalance, applBalance float64, err error), currentSpotPrice float64) {
	s.accountMetrics.ResetForSession(getBalance, currentSpotPrice)
}

// FinalizeAccountsForSession finalizes all accounts using the final spot price at session end.
// It also computes and stores the session-level daily volatility from price candles
// so that the Sharpe ratio uses a consistent market-level denominator.
func (s *MemoryStore) FinalizeAccountsForSession(finalSpotPrice float64) {
	// Compute session volatility from candle data before finalizing accounts
	candles := s.priceMetrics.GetCandles()
	sessionVol := computeSessionDailyVol(candles)
	s.accountMetrics.SetSessionVolatility(sessionVol)

	s.accountMetrics.FinalizeSession(finalSpotPrice)
}

// Impact curve methods

func (s *MemoryStore) GetImpactCurve() *metrics.ImpactCurve {
	return s.impactCurve
}

func (s *MemoryStore) GetBuyImpact() []metrics.ImpactPoint {
	return s.impactCurve.CalculateBuyCurve(metrics.GetDefaultSizes())
}

func (s *MemoryStore) GetSellImpact() []metrics.ImpactPoint {
	return s.impactCurve.CalculateSellCurve(metrics.GetDefaultSizes())
}

// Trade methods

func (s *MemoryStore) RecordTrade(trade engine.Trade) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trades = append(s.trades, trade)

	// Cap trades slice to prevent unbounded memory growth
	if len(s.trades) > config.MaxTrades {
		s.trades = s.trades[len(s.trades)-config.MaxTrades:]
	}

	// Update account metrics for this trade
	// Get or create account metrics
	am := s.accountMetrics.GetOrCreate(trade.Nickname, trade.Trader, config.InitialAccountEquityETH)
	
	// Only record trades during an active session
	if !am.IsSessionActive() {
		// Not in a session, skip recording this trade for performance metrics
		return
	}

	// Convert trade amounts to float64
	// For BUY: amountIn is ETH (wei), amountOut is APPL (wei)
	// For SELL: amountIn is APPL (wei), amountOut is ETH (wei)
	amountInFloat := new(big.Float).SetInt(trade.AmountIn)
	amountInFloat.Quo(amountInFloat, big.NewFloat(1e18))
	amountIn, _ := amountInFloat.Float64()

	// Check for nil AmountOut before converting
	var amountOut float64
	if trade.AmountOut != nil {
		amountOutFloat := new(big.Float).SetInt(trade.AmountOut)
		amountOutFloat.Quo(amountOutFloat, big.NewFloat(1e18))
		amountOut, _ = amountOutFloat.Float64()
	} else {
		// Log warning when trade data is incomplete
		log.Printf("Warning: Trade %s has nil AmountOut, using 0", trade.TxHash)
		amountOut = 0
	}

	// Get current spot price for mark-to-market valuation
	// Use the most recent price from price metrics
	var currentSpotPrice float64
	candles := s.priceMetrics.GetCandles()
	if len(candles) > 0 {
		currentSpotPrice = candles[len(candles)-1].Close
	} else {
		// Fallback: use execution price if no price history yet
		if trade.Price != nil {
			priceFloat := new(big.Float).SetInt(trade.Price)
			priceFloat.Quo(priceFloat, big.NewFloat(1e18))
			currentSpotPrice, _ = priceFloat.Float64()
		} else {
			// If both price history and trade price are unavailable, use 0
			log.Printf("Warning: Trade %s has nil Price and no price history, using 0", trade.TxHash)
			currentSpotPrice = 0
		}
	}

	// Record trade with actual amounts and current spot price
	// RecordTrade will update balances and calculate equity using mark-to-market
	if trade.IsBuy {
		// Buy: ethAmount = amountIn (ETH spent), appleAmount = amountOut (APPL received)
		am.RecordTrade(true, amountIn, amountOut, currentSpotPrice)
	} else {
		// Sell: ethAmount = amountOut (ETH received), appleAmount = amountIn (APPL spent)
		am.RecordTrade(false, amountOut, amountIn, currentSpotPrice)
	}
}

func (s *MemoryStore) GetTrades() []engine.Trade {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]engine.Trade, len(s.trades))
	copy(result, s.trades)
	return result
}

func (s *MemoryStore) GetRecentTrades(n int) []engine.Trade {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n >= len(s.trades) {
		result := make([]engine.Trade, len(s.trades))
		copy(result, s.trades)
		return result
	}

	result := make([]engine.Trade, n)
	copy(result, s.trades[len(s.trades)-n:])
	return result
}

// Event methods

func (s *MemoryStore) RecordEvent(eventType, description, severity string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, KeyEvent{
		Timestamp:   time.Now(),
		Type:        eventType,
		Description: description,
		Severity:    severity,
	})

	// Cap events slice to prevent unbounded memory growth
	if len(s.events) > config.MaxEvents {
		s.events = s.events[len(s.events)-config.MaxEvents:]
	}
}

func (s *MemoryStore) GetEvents() []KeyEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]KeyEvent, len(s.events))
	copy(result, s.events)
	return result
}

func (s *MemoryStore) GetRecentEvents(n int) []KeyEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n >= len(s.events) {
		result := make([]KeyEvent, len(s.events))
		copy(result, s.events)
		return result
	}

	result := make([]KeyEvent, n)
	copy(result, s.events[len(s.events)-n:])
	return result
}

// Trade flow methods

func (s *MemoryStore) GetTradeFlowTracker() *metrics.TradeFlowTracker {
	return s.tradeFlowTracker
}

// RecordTradeFlow records a trade flow event (called after trade execution)
// This captures price changes and calculates log returns and normalized flow
func (s *MemoryStore) RecordTradeFlow(
	priceBefore, priceAfter float64,
	trade *engine.Trade,
	reservesBeforeETH, reservesBeforeAPPL float64,
) {
	s.tradeFlowTracker.RecordTrade(
		priceBefore, priceAfter,
		trade,
		&reservesBeforeETH, &reservesBeforeAPPL,
	)
}

// Reset clears stored data but preserves account metrics
// Account metrics persist across sessions to show performance history
func (s *MemoryStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.priceMetrics.Reset()
	s.lpMetrics.Reset()
	// Note: We don't reset accountMetrics here - they persist across sessions
	s.trades = make([]engine.Trade, 0)
	s.events = make([]KeyEvent, 0) // Clear Key Events on reset
}

// computeSessionDailyVol calculates the realized session volatility (%) from
// candle close prices observed during the session.
//
//	sessionVol = stdDev(returns) * sqrt(N) * 100
//
// where N is the number of return observations in the session. This measures
// the actual dispersion of the session's total return rather than extrapolating
// to a full 24-hour day (which would be misleading for short simulation runs).
func computeSessionDailyVol(candles []metrics.Candle) float64 {
	if len(candles) < 3 {
		return 0
	}

	// Calculate simple returns between consecutive candle closes
	returns := make([]float64, 0, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		prev := candles[i-1].Close
		curr := candles[i].Close
		if prev > 0 {
			returns = append(returns, (curr-prev)/prev)
		}
	}
	if len(returns) < 2 {
		return 0
	}

	// Standard deviation of returns
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))
	var variance float64
	for _, r := range returns {
		d := r - mean
		variance += d * d
	}
	variance /= float64(len(returns))
	stdDev := math.Sqrt(variance)

	// Scale by sqrt(N) to get the session's realized total-return volatility.
	// This represents the expected dispersion of the cumulative return over
	// the session, which is intuitive: a session where price ranges 12%
	// peak-to-trough will report roughly that magnitude.
	return stdDev * math.Sqrt(float64(len(returns))) * 100
}

// Compile-time assertion that MemoryStore implements metrics.PriceDataStore
var _ metrics.PriceDataStore = (*MemoryStore)(nil)
