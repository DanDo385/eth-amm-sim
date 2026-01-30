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
}

// TradeRecord records a single trade
type TradeRecord struct {
	Timestamp time.Time `json:"timestamp"`
	IsBuy     bool      `json:"isBuy"`
	Size      float64   `json:"size"`
	Price     float64   `json:"price"`
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
		Timestamp: now,
		IsBuy:     isBuy,
		Size:      tradeSize,
		Price:     executionPrice,
		PnL:       pnl,
		Equity:    newEquity,
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
	if am.initialEquity == 0 || len(am.equityCurve) == 0 {
		return 0
	}
	currentEquity := am.equityCurve[len(am.equityCurve)-1].Equity
	return ((currentEquity - am.initialEquity) / am.initialEquity) * 100
}

// calculateTotalPnL returns the total PnL
func (am *AccountMetrics) calculateTotalPnL() float64 {
	if len(am.equityCurve) == 0 {
		return 0
	}
	return am.equityCurve[len(am.equityCurve)-1].Equity - am.initialEquity
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

// Helper: round to decimal places
func round(val float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(val*p) / p
}
