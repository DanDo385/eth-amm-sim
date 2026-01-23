// Package metrics provides TradFi-style account performance analytics
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
	}
}

// RecordTrade records a trade and updates metrics
func (am *AccountMetrics) RecordTrade(isBuy bool, size, price, newEquity float64) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	now := time.Now()
	
	// Calculate PnL from this trade
	var pnl float64
	if len(am.equityCurve) > 0 {
		lastEquity := am.equityCurve[len(am.equityCurve)-1].Equity
		pnl = newEquity - lastEquity
	}
	
	// Record trade
	trade := TradeRecord{
		Timestamp: now,
		IsBuy:     isBuy,
		Size:      size,
		Price:     price,
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
	volatility := standardDeviation(returns)
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

// Reset resets all account metrics
func (m *AccountMetricsManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts = make(map[string]*AccountMetrics)
}

// Helper: round to decimal places
func round(val float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(val*p) / p
}
