// account.go - Per-account performance analytics (Sharpe ratio, drawdown, win rate).
//
// Tracks equity curves and trade history for every account in the simulation.
// Updated after each trade via the trade callback in main.go. Computes:
//  - Total return, PnL, volatility, Sharpe ratio, max drawdown, win rate
//  - Equity curve snapshots for charting
//
// CONNECTIONS:
//  - Updated by: main.go trade callback → store.RecordTrade → AccountMetrics
//  - Frontend: GET /accounts/{nickname}/performance → AccountMetrics component
//  - Persists across sessions (not cleared on reset unless hard reset)
package metrics

import (
	"log"
	"math"
	"sync"
	"time"

	"eth-amm-sim/internal/config"

	"github.com/ethereum/go-ethereum/common"
)

// AccountMetrics tracks performance metrics for a single account
type AccountMetrics struct {
	mu sync.RWMutex

	nickname string
	address  common.Address

	// Trade history
	trades []TradeRecord

	// Equity curve (value over time)
	equityCurve []EquityPoint

	// Running calculations
	initialEquity float64
	peakEquity    float64

	// Balance tracking (for accurate equity calculation)
	ethBalance   float64 // ETH balance in ETH
	appleBalance float64 // APPL balance in APPL tokens

	// Session tracking
	sessionStarted      bool
	sessionStartTime    *time.Time
	sessionEndTime      *time.Time
	sessionStartEquity  float64 // Total equity at session start (ETH + APPL*price)
	sessionStartETH     float64 // ETH-only starting capital (return denominator)
	sessionClosingPrice float64 // Closing price at session end (for PnL calculation)
}

// TradeRecord records a single trade
type TradeRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	IsBuy       bool      `json:"isBuy"`
	Size        float64   `json:"size"`
	Price       float64   `json:"price"`
	ClosePrice  float64   `json:"closePrice"`  // Closing price used for PnL calculation (set at session end)
	AppleAmount float64   `json:"appleAmount"` // Amount of APPL traded (for PnL calculation)
	PnL         float64   `json:"pnl"`
	Equity      float64   `json:"equity"`
}

// EquityPoint represents equity at a point in time
type EquityPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Equity    float64   `json:"equity"`
	Drawdown  float64   `json:"drawdown"`
}

// PerformanceData is the API response for account performance
type PerformanceData struct {
	Nickname    string        `json:"nickname"`
	Address     string        `json:"address"`
	TotalReturn float64       `json:"totalReturn"` // Portfolio equity change (%)
	TotalPnL    float64       `json:"totalPnL"`    // Portfolio equity change (ETH) = position + trading
	PositionPnL float64       `json:"positionPnL"` // APPL inventory mark-to-market (ETH)
	TradingPnL  float64       `json:"tradingPnL"`  // Sum of per-trade PnLs (ETH)
	Volatility  float64       `json:"volatility"`
	SharpeRatio float64       `json:"sharpeRatio"`
	MaxDrawdown float64       `json:"maxDrawdown"`
	WinRate     float64       `json:"winRate"`
	TradeCount  int           `json:"tradeCount"`
	EquityCurve []EquityPoint `json:"equityCurve"`
	Trades      []TradeRecord `json:"trades"`
}

// NewAccountMetrics creates a new account metrics tracker
func NewAccountMetrics(nickname string, address common.Address, initialEquity float64) *AccountMetrics {
	now := time.Now()
	return &AccountMetrics{
		nickname:      nickname,
		address:       address,
		trades:        make([]TradeRecord, 0),
		equityCurve:   []EquityPoint{{Timestamp: now, Equity: initialEquity, Drawdown: 0}},
		initialEquity: initialEquity,
		peakEquity:    initialEquity,
		ethBalance:    initialEquity, // Start with all ETH (no APPL)
		appleBalance:  0,
	}
}

// RecordTrade records a trade and updates metrics
// isBuy: true for buy, false for sell
// ethAmount: ETH amount (spent for buy, received for sell)
// appleAmount: APPL amount (received for buy, spent for sell)
// currentSpotPrice: current market price (ETH per APPL) for mark-to-market
func (am *AccountMetrics) RecordTrade(isBuy bool, ethAmount, appleAmount, currentSpotPrice float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()

	// Update balances based on trade
	if isBuy {
		// Buy: spend ETH, receive APPL
		am.ethBalance -= ethAmount
		am.appleBalance += appleAmount
	} else {
		// Sell: spend APPL, receive ETH
		am.ethBalance += ethAmount
		am.appleBalance -= appleAmount
	}

	// Calculate equity using current spot price (mark-to-market)
	// Equity = ETH_balance + APPL_balance * current_spot_price
	newEquity := am.ethBalance + (am.appleBalance * currentSpotPrice)

	// Calculate PnL from this trade (change in equity)
	var pnl float64
	if len(am.equityCurve) > 0 {
		lastEquity := am.equityCurve[len(am.equityCurve)-1].Equity
		pnl = newEquity - lastEquity
	}

	// Calculate trade size and execution price for display
	var tradeSize float64
	var executionPrice float64
	if isBuy {
		tradeSize = ethAmount // Size in ETH for buys
		if appleAmount > 0 {
			executionPrice = ethAmount / appleAmount // ETH per APPL
		}
	} else {
		tradeSize = appleAmount * currentSpotPrice // Size in ETH equivalent for sells
		if appleAmount > 0 {
			executionPrice = ethAmount / appleAmount // ETH per APPL
		}
	}

	// Record trade
	trade := TradeRecord{
		Timestamp:   now,
		IsBuy:       isBuy,
		Size:        tradeSize,
		Price:       executionPrice,
		AppleAmount: appleAmount, // Store APPL amount for PnL calculation
		PnL:         pnl,
		Equity:      newEquity,
	}
	am.trades = append(am.trades, trade)

	// Update peak
	if newEquity > am.peakEquity {
		am.peakEquity = newEquity
	}

	// Calculate drawdown
	drawdown := 0.0
	if am.peakEquity > 0 {
		drawdown = (am.peakEquity - newEquity) / am.peakEquity
	}

	// Record equity point
	am.equityCurve = append(am.equityCurve, EquityPoint{
		Timestamp: now,
		Equity:    newEquity,
		Drawdown:  drawdown,
	})
}

// UpdateEquity updates the equity without a trade (for mark-to-market)
func (am *AccountMetrics) UpdateEquity(newEquity float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()

	if newEquity > am.peakEquity {
		am.peakEquity = newEquity
	}

	drawdown := 0.0
	if am.peakEquity > 0 {
		drawdown = (am.peakEquity - newEquity) / am.peakEquity
	}

	am.equityCurve = append(am.equityCurve, EquityPoint{
		Timestamp: now,
		Equity:    newEquity,
		Drawdown:  drawdown,
	})
}

// GetPerformance returns the performance metrics.
// sessionVol is the market-level daily volatility (%) observed during the session,
// used as the Sharpe denominator for cross-account comparability.
func (am *AccountMetrics) GetPerformance(sessionVol float64) PerformanceData {
	am.mu.RLock()
	defer am.mu.RUnlock()

	tradingPnL := am.sumTradePnL()
	portfolioPnL := am.calculatePortfolioPnL()
	positionPnL := portfolioPnL - tradingPnL
	tradingReturn := am.calculateTradingReturn()
	sharpe := am.calculateSharpe(tradingReturn, sessionVol)
	maxDD := am.calculateMaxDrawdown()
	winRate := am.calculateWinRate()

	return PerformanceData{
		Nickname:    am.nickname,
		Address:     am.address.Hex(),
		TotalReturn: tradingReturn,
		TotalPnL:    portfolioPnL,
		PositionPnL: positionPnL,
		TradingPnL:  tradingPnL,
		Volatility:  sessionVol,
		SharpeRatio: sharpe,
		MaxDrawdown: maxDD,
		WinRate:     winRate,
		TradeCount:  len(am.trades),
		EquityCurve: am.copyEquityCurve(),
		Trades:      am.copyTrades(),
	}
}

// calculateTradingReturn returns the trading return percentage.
// This is the sum of per-trade PnLs divided by the starting portfolio value.
// It reflects only the active trading performance, not the initial position
// mark-to-market. Sharpe ratio and other return metrics use this value.
func (am *AccountMetrics) calculateTradingReturn() float64 {
	var baseEquity float64
	if am.sessionStarted && am.sessionStartEquity > 0 {
		baseEquity = am.sessionStartEquity
	} else {
		baseEquity = am.initialEquity
	}
	if baseEquity == 0 {
		return 0
	}
	return (am.sumTradePnL() / baseEquity) * 100
}

// calculatePortfolioPnL returns the total portfolio equity change in ETH.
// This equals positionPnL + tradingPnL.
func (am *AccountMetrics) calculatePortfolioPnL() float64 {
	if len(am.equityCurve) == 0 {
		return 0
	}
	var baseEquity float64
	if am.sessionStarted && am.sessionStartEquity > 0 {
		baseEquity = am.sessionStartEquity
	} else {
		baseEquity = am.initialEquity
	}
	return am.equityCurve[len(am.equityCurve)-1].Equity - baseEquity
}

// sumTradePnL returns the sum of all per-trade PnL values.
func (am *AccountMetrics) sumTradePnL() float64 {
	var total float64
	for _, t := range am.trades {
		total += t.PnL
	}
	return total
}

// calculateTradeReturns computes per-trade percentage returns from trade PnL
// relative to the equity just before that trade. This gives a meaningful
// return series tied to actual trading activity (not mark-to-market noise).
func (am *AccountMetrics) calculateTradeReturns() []float64 {
	if len(am.trades) == 0 {
		return nil
	}

	returns := make([]float64, 0, len(am.trades))
	for i, trade := range am.trades {
		// Equity before this trade = equity after trade minus the PnL from the trade
		var equityBefore float64
		if i == 0 {
			equityBefore = trade.Equity - trade.PnL
		} else {
			equityBefore = am.trades[i-1].Equity
		}
		if equityBefore > 0 {
			returns = append(returns, trade.PnL/equityBefore)
		}
	}
	return returns
}

// calculateDailyVolatility computes rolling daily volatility (%) from trade returns.
// Uses the same approach as TWAPChart.tsx: stdDev(returns) * sqrt(periodsPerDay) * 100.
// periodsPerDay is estimated from the average trade interval observed in the session.
func (am *AccountMetrics) calculateDailyVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	stdDev := standardDeviation(returns)

	// Estimate periods per day from actual trade timestamps
	periodsPerDay := am.estimatePeriodsPerDay()

	return stdDev * math.Sqrt(periodsPerDay) * 100
}

// estimatePeriodsPerDay returns the number of trade-frequency periods per day.
// Computed from the average interval between trades in the session.
// Falls back to 1440 (one trade per minute) when there aren't enough trades.
func (am *AccountMetrics) estimatePeriodsPerDay() float64 {
	if len(am.trades) < 2 {
		return 1440 // default: 1-minute periods → 1440/day
	}
	first := am.trades[0].Timestamp
	last := am.trades[len(am.trades)-1].Timestamp
	elapsed := last.Sub(first).Seconds()
	if elapsed <= 0 {
		return 1440
	}
	avgInterval := elapsed / float64(len(am.trades)-1)
	if avgInterval <= 0 {
		return 1440
	}
	return 86400.0 / avgInterval
}

// calculateSharpe computes the Sharpe ratio as:
//
//	Sharpe = totalReturn% / sessionDailyVol%
//
// Both numerator and denominator are in the same units (percentage).
// Using the session-level market volatility as the denominator ensures
// that the ratio is comparable across all accounts regardless of their
// individual trade frequency or size.
// Assumes risk-free rate = 0.
func (am *AccountMetrics) calculateSharpe(totalReturn, sessionVol float64) float64 {
	if sessionVol == 0 || len(am.trades) < 2 {
		return 0
	}
	return totalReturn / sessionVol
}

// calculateMaxDrawdown returns the maximum drawdown percentage
func (am *AccountMetrics) calculateMaxDrawdown() float64 {
	if len(am.equityCurve) == 0 {
		return 0
	}

	var maxDD float64
	for _, point := range am.equityCurve {
		if point.Drawdown > maxDD {
			maxDD = point.Drawdown
		}
	}
	return maxDD * 100 // Return as percentage
}

// calculateWinRate returns the percentage of winning trades
func (am *AccountMetrics) calculateWinRate() float64 {
	if len(am.trades) == 0 {
		return 0
	}

	var wins int
	for _, trade := range am.trades {
		if trade.PnL > 0 {
			wins++
		}
	}
	return (float64(wins) / float64(len(am.trades))) * 100
}

func (am *AccountMetrics) copyEquityCurve() []EquityPoint {
	result := make([]EquityPoint, len(am.equityCurve))
	copy(result, am.equityCurve)
	return result
}

func (am *AccountMetrics) copyTrades() []TradeRecord {
	result := make([]TradeRecord, len(am.trades))
	copy(result, am.trades)
	return result
}

// Reset clears all account data
func (am *AccountMetrics) Reset(initialEquity float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	am.trades = make([]TradeRecord, 0)
	am.equityCurve = []EquityPoint{{Timestamp: now, Equity: initialEquity, Drawdown: 0}}
	am.initialEquity = initialEquity
	am.peakEquity = initialEquity
	am.ethBalance = initialEquity
	am.appleBalance = 0
	am.sessionStarted = false
	am.sessionStartTime = nil
	am.sessionEndTime = nil
	am.sessionStartEquity = 0
	am.sessionStartETH = 0
	am.sessionClosingPrice = 0
}

// ResetForSession resets account metrics to track session-only performance.
// All accounts use the same normalized starting balance (from config) so that
// returns are comparable across account types.
// The actual on-chain balances are still used for real trade execution; this only
// affects how performance metrics (return %, PnL, equity curve) are calculated.
func (am *AccountMetrics) ResetForSession(currentETHBalance, currentAPPLBalance, currentSpotPrice float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	// Use normalized starting balance from config constants.
	normalizedETH := config.NormalizedStartingETH
	normalizedAPPL := config.NormalizedStartingAPPL
	startingEquity := normalizedETH + (normalizedAPPL * currentSpotPrice)

	// Reset session tracking
	am.trades = make([]TradeRecord, 0)
	am.equityCurve = []EquityPoint{{Timestamp: now, Equity: startingEquity, Drawdown: 0}}
	am.initialEquity = startingEquity
	am.peakEquity = startingEquity
	am.ethBalance = normalizedETH
	am.appleBalance = normalizedAPPL
	am.sessionStarted = true
	am.sessionStartTime = &now
	am.sessionEndTime = nil
	am.sessionStartEquity = startingEquity
	am.sessionStartETH = normalizedETH
	am.sessionClosingPrice = 0 // Will be set when session ends
}

// IsSessionActive returns true if a session is currently active
func (am *AccountMetrics) IsSessionActive() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.sessionStarted && am.sessionEndTime == nil
}

// FinalizeSession finalizes session performance using the last price at session end
func (am *AccountMetrics) FinalizeSession(finalSpotPrice float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.sessionStarted {
		return // Not in a session
	}

	now := time.Now()
	am.sessionEndTime = &now
	am.sessionClosingPrice = finalSpotPrice

	// Calculate final equity using current balances and final spot price
	finalEquity := am.ethBalance + (am.appleBalance * finalSpotPrice)

	// Recalculate PnL for all trades using closing price
	for i := range am.trades {
		trade := &am.trades[i]
		trade.ClosePrice = finalSpotPrice
		// Calculate PnL based on closing price
		if trade.AppleAmount > 0 {
			if trade.IsBuy {
				// Buy: PnL = (closing_price - execution_price) * amount_of_APPL
				trade.PnL = (finalSpotPrice - trade.Price) * trade.AppleAmount
			} else {
				// Sell: PnL = (execution_price - closing_price) * amount_of_APPL
				trade.PnL = (trade.Price - finalSpotPrice) * trade.AppleAmount
			}
		}
	}

	// Update equity curve with final mark-to-market
	if len(am.equityCurve) > 0 {
		// Update last equity point or add new one
		am.equityCurve[len(am.equityCurve)-1].Equity = finalEquity
		am.equityCurve[len(am.equityCurve)-1].Timestamp = now
	} else {
		am.equityCurve = append(am.equityCurve, EquityPoint{
			Timestamp: now,
			Equity:    finalEquity,
			Drawdown:  0,
		})
	}

	// Update peak if needed
	if finalEquity > am.peakEquity {
		am.peakEquity = finalEquity
	}

	// Recalculate drawdown for all points
	for i := range am.equityCurve {
		if am.peakEquity > 0 {
			am.equityCurve[i].Drawdown = (am.peakEquity - am.equityCurve[i].Equity) / am.peakEquity
		}
	}
}

// AccountMetricsManager manages metrics for multiple accounts
type AccountMetricsManager struct {
	mu                sync.RWMutex
	accounts          map[string]*AccountMetrics
	sessionVolatility float64 // observed daily vol (%) for the entire session
}

// NewAccountMetricsManager creates a new manager
func NewAccountMetricsManager() *AccountMetricsManager {
	return &AccountMetricsManager{
		accounts: make(map[string]*AccountMetrics),
	}
}

// SetSessionVolatility stores the market-level daily volatility (%) observed
// during the session. All accounts use this single value as the Sharpe
// denominator so that ratios are comparable across account types.
func (m *AccountMetricsManager) SetSessionVolatility(vol float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessionVolatility = vol
}

// GetSessionVolatility returns the stored session volatility.
func (m *AccountMetricsManager) GetSessionVolatility() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessionVolatility
}

// GetOrCreate gets or creates metrics for an account
func (m *AccountMetricsManager) GetOrCreate(nickname string, address common.Address, initialEquity float64) *AccountMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if am, exists := m.accounts[nickname]; exists {
		return am
	}

	am := NewAccountMetrics(nickname, address, initialEquity)
	m.accounts[nickname] = am
	return am
}

// Get returns metrics for an account
func (m *AccountMetricsManager) Get(nickname string) *AccountMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.accounts[nickname]
}

// GetAllPerformanceWithVol returns performance for all accounts using the
// provided session volatility as the Sharpe denominator.
func (m *AccountMetricsManager) GetAllPerformanceWithVol(vol float64) []PerformanceData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]PerformanceData, 0, len(m.accounts))
	for _, am := range m.accounts {
		result = append(result, am.GetPerformance(vol))
	}
	return result
}

// Reset resets all account metrics back to initial state
// For each account, it resets trades, equity curve, and balances to initial values
func (m *AccountMetricsManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessionVolatility = 0

	// Reset each account back to its stored initial equity
	// This preserves account existence but clears all trading history
	for _, am := range m.accounts {
		// Use the stored initialEquity value that was set when the account was created
		am.mu.RLock()
		initialEquity := am.initialEquity
		am.mu.RUnlock()
		am.Reset(initialEquity)
	}
}

// ResetForSession resets all accounts to track session-only performance
// Requires a function to get current balances for each account and current spot price
func (m *AccountMetricsManager) ResetForSession(getBalance func(nickname string) (ethBalance, applBalance float64, err error), currentSpotPrice float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessionVolatility = 0

	for _, am := range m.accounts {
		ethBalance, applBalance, err := getBalance(am.nickname)
		if err != nil {
			log.Printf("Warning: Could not get balances for %s at session start: %v", am.nickname, err)
			am.mu.RLock()
			ethBalance = am.ethBalance
			applBalance = am.appleBalance
			am.mu.RUnlock()
		}
		am.ResetForSession(ethBalance, applBalance, currentSpotPrice)
	}
}

// FinalizeSession finalizes all accounts using the final spot price at session end
func (m *AccountMetricsManager) FinalizeSession(finalSpotPrice float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, am := range m.accounts {
		am.FinalizeSession(finalSpotPrice)
	}
}

// Helper: round to decimal places
func round(val float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(val*p) / p
}
