// account.go — Per-account performance analytics (Sharpe ratio, drawdown, win rate).
//
// Tracks equity curves and trade history for every account in the simulation.
// Updated after each trade via the trade callback in main.go. Computes:
//   - Total return, PnL, volatility, Sharpe ratio, max drawdown, win rate
//   - Equity curve snapshots for charting
//
// CONNECTIONS:
//   - Updated by: main.go trade callback → store.RecordTrade → AccountMetrics
//   - Frontend: GET /accounts/{nickname}/performance → AccountMetrics component
//   - Persists across sessions (not cleared on reset unless hard reset)
package metrics

import (
	"log"
	"math"
	"sync"
	"time"

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
	ethBalance  float64 // ETH balance in ETH
	appleBalance float64 // APPL balance in APPL tokens
	
	// Session tracking
	sessionStarted bool
	sessionStartTime *time.Time
	sessionEndTime *time.Time
	sessionStartEquity float64 // Equity at session start (for comparison)
	sessionClosingPrice float64 // Closing price at session end (for PnL calculation)
}

// TradeRecord records a single trade
type TradeRecord struct {
	Timestamp time.Time `json:"timestamp"`
	IsBuy     bool      `json:"isBuy"`
	Size      float64   `json:"size"`
	Price     float64   `json:"price"`
	AppleAmount float64 `json:"appleAmount"` // Amount of APPL traded (for PnL calculation)
	PnL       float64   `json:"pnl"`
	Equity    float64   `json:"equity"`
}

// EquityPoint represents equity at a point in time
type EquityPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Equity    float64   `json:"equity"`
	Drawdown  float64   `json:"drawdown"`
}

// PerformanceData is the API response for account performance
type PerformanceData struct {
	Nickname     string        `json:"nickname"`
	Address      string        `json:"address"`
	TotalReturn  float64       `json:"totalReturn"`
	TotalPnL     float64       `json:"totalPnL"`
	Volatility   float64       `json:"volatility"`
	SharpeRatio  float64       `json:"sharpeRatio"`
	MaxDrawdown  float64       `json:"maxDrawdown"`
	WinRate      float64       `json:"winRate"`
	TradeCount   int           `json:"tradeCount"`
	EquityCurve  []EquityPoint `json:"equityCurve"`
	Trades       []TradeRecord `json:"trades"`
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

// GetPerformance returns the performance metrics
func (am *AccountMetrics) GetPerformance() PerformanceData {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	// Calculate all metrics
	totalReturn := am.calculateTotalReturn()
	totalPnL := am.calculateTotalPnL()
	returns := am.calculateReturns()
	volatility := am.calculateVolatility(returns)
	sharpe := am.calculateSharpe(returns)
	maxDD := am.calculateMaxDrawdown()
	winRate := am.calculateWinRate()
	
	return PerformanceData{
		Nickname:     am.nickname,
		Address:      am.address.Hex(),
		TotalReturn:  totalReturn,
		TotalPnL:     totalPnL,
		Volatility:   volatility,
		SharpeRatio:  sharpe,
		MaxDrawdown:  maxDD,
		WinRate:      winRate,
		TradeCount:   len(am.trades),
		EquityCurve:  am.copyEquityCurve(),
		Trades:       am.copyTrades(),
	}
}

// calculateTotalReturn returns the total return percentage
func (am *AccountMetrics) calculateTotalReturn() float64 {
	// For session-based tracking, use sessionStartEquity instead of initialEquity
	// (which is 0 for session-only tracking)
	var baseEquity float64
	if am.sessionStarted && am.sessionStartEquity > 0 {
		baseEquity = am.sessionStartEquity
	} else {
		baseEquity = am.initialEquity
	}
	
	if baseEquity == 0 || len(am.equityCurve) == 0 {
		return 0
	}
	currentEquity := am.equityCurve[len(am.equityCurve)-1].Equity
	return ((currentEquity - baseEquity) / baseEquity) * 100
}

// calculateTotalPnL returns the total PnL
func (am *AccountMetrics) calculateTotalPnL() float64 {
	if len(am.equityCurve) == 0 {
		return 0
	}
	// For session-based tracking, use sessionStartEquity instead of initialEquity
	var baseEquity float64
	if am.sessionStarted && am.sessionStartEquity > 0 {
		baseEquity = am.sessionStartEquity
	} else {
		baseEquity = am.initialEquity
	}
	return am.equityCurve[len(am.equityCurve)-1].Equity - baseEquity
}

// calculateReturns calculates the period-over-period returns
func (am *AccountMetrics) calculateReturns() []float64 {
	if len(am.equityCurve) < 2 {
		return nil
	}
	
	returns := make([]float64, len(am.equityCurve)-1)
	for i := 1; i < len(am.equityCurve); i++ {
		prev := am.equityCurve[i-1].Equity
		curr := am.equityCurve[i].Equity
		if prev > 0 {
			returns[i-1] = (curr - prev) / prev
		}
	}
	return returns
}

// calculateVolatility calculates annualized volatility from returns
// Assumes returns are per-trade, annualizes assuming ~252 trading days
// For simplicity, we'll annualize based on number of trades
func (am *AccountMetrics) calculateVolatility(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	
	// Calculate standard deviation of returns
	stdDev := standardDeviation(returns)
	
	// Annualize: multiply by sqrt(252) for daily returns
	// Since we don't have exact time periods, we'll use a simple annualization
	// Assuming trades happen roughly every few seconds, we'll scale by sqrt(number_of_periods_per_year)
	// For demo purposes, we'll use sqrt(252) to annualize daily-like returns
	annualizedVol := stdDev * math.Sqrt(252) * 100 // Convert to percentage
	
	return annualizedVol
}

// calculateSharpe calculates the Sharpe ratio
// Sharpe = mean_return / std_dev(returns)
// Assumes risk-free rate = 0 per project rules
func (am *AccountMetrics) calculateSharpe(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	
	// Calculate mean return
	var sum float64
	for _, r := range returns {
		sum += r
	}
	meanReturn := sum / float64(len(returns))
	
	// Calculate standard deviation
	stdDev := standardDeviation(returns)
	
	if stdDev == 0 {
		return 0
	}
	
	// Sharpe ratio (risk-free rate = 0)
	return meanReturn / stdDev
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
	am.sessionClosingPrice = 0
}

// ResetForSession resets account metrics to track session-only performance
// Starts with current balances but sets initial equity to 0 for performance calculation
func (am *AccountMetrics) ResetForSession(currentETHBalance, currentAPPLBalance, currentSpotPrice float64) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	now := time.Now()
	// Calculate starting equity using current balances and spot price
	startingEquity := currentETHBalance + (currentAPPLBalance * currentSpotPrice)
	
	// Reset session tracking
	am.trades = make([]TradeRecord, 0)
	am.equityCurve = []EquityPoint{{Timestamp: now, Equity: 0, Drawdown: 0}}
	am.initialEquity = 0 // Start from 0 for performance calculation
	am.peakEquity = 0
	am.ethBalance = currentETHBalance
	am.appleBalance = currentAPPLBalance
	am.sessionStarted = true
	am.sessionStartTime = &now
	am.sessionEndTime = nil
	am.sessionStartEquity = startingEquity // Store for reference, but don't use in calculations
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
	mu       sync.RWMutex
	accounts map[string]*AccountMetrics
}

// NewAccountMetricsManager creates a new manager
func NewAccountMetricsManager() *AccountMetricsManager {
	return &AccountMetricsManager{
		accounts: make(map[string]*AccountMetrics),
	}
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

// GetAllPerformance returns performance for all accounts
func (m *AccountMetricsManager) GetAllPerformance() []PerformanceData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make([]PerformanceData, 0, len(m.accounts))
	for _, am := range m.accounts {
		result = append(result, am.GetPerformance())
	}
	return result
}

// Reset resets all account metrics back to initial state
// For each account, it resets trades, equity curve, and balances to initial values
func (m *AccountMetricsManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
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
	
	for _, am := range m.accounts {
		ethBalance, applBalance, err := getBalance(am.nickname)
		if err != nil {
			log.Printf("Warning: Could not get balances for %s at session start: %v", am.nickname, err)
			// Use stored balances as fallback
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
