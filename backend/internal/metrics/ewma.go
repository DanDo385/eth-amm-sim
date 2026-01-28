// ewma.go — Exponentially Weighted Moving Average calculator for mean reversion.
//
// Maintains a running mean and variance of liquidity-normalized log returns
// using exponential decay. The half-life parameter controls how quickly old
// observations lose weight (e.g., 50 trades = weight decays to 50% after 50
// observations). MeanRev bots use the z-score (deviation / std dev) to
// decide when to trade.
//
// CONNECTIONS:
//   - Used by: bots/meanrev.go (each MeanRev bot has its own EWMACalculator)
//   - Fed by: metrics/trade_flow.go emits trade flow events with normalized returns
package metrics

import (
	"math"
	"sync"
)

// EWMAState maintains exponentially weighted moving average statistics
// Used for mean reversion strategies that adapt to changing market conditions
//
// Why EWMA instead of rolling windows:
// - Adaptive: automatically adjusts to current market regime
// - No fixed lookback: doesn't require arbitrary window size
// - Efficient: O(1) updates vs O(n) for rolling windows
// - Better for non-stationary markets: recent data weighted more heavily
type EWMAState struct {
	mu       float64 // Mean of rTilde (liquidity-normalized returns)
	variance float64 // Variance of rTilde
	sigma    float64 // Standard deviation = sqrt(variance)

	tradeCount int     // Number of trades observed (for this bot)
	halfLife   int     // Half-life in trades (decay parameter)
	lambda     float64 // Decay factor = exp(ln(0.5) / halfLife)

	mutex sync.RWMutex // Mutex for thread-safe access
}

// NewEWMAState creates a new EWMA state with specified half-life
// halfLifeTrades: Number of trades for the weight to decay to 50%
// Example: halfLifeTrades=50 means after 50 trades, weight of old data is 50% of new data
func NewEWMAState(halfLifeTrades int) *EWMAState {
	if halfLifeTrades <= 0 {
		halfLifeTrades = 50 // Default half-life
	}

	// Calculate decay factor: lambda = exp(ln(0.5) / halfLife)
	// This ensures that after halfLife trades, the weight is 0.5
	lambda := math.Exp(math.Log(0.5) / float64(halfLifeTrades))

	return &EWMAState{
		halfLife: halfLifeTrades,
		lambda:   lambda,
	}
}

// UpdateEWMA updates the EWMA statistics with a new rTilde observation
// rTilde: Liquidity-normalized return = logReturn / normalizedFlow
//
// Update formulas:
//
//	mu_new = lambda * mu_old + (1 - lambda) * rTilde
//	var_new = lambda * var_old + (1 - lambda) * (rTilde - mu_old)^2
//	sigma = sqrt(var_new)
func (e *EWMAState) UpdateEWMA(rTilde float64) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Update mean: exponentially weighted moving average
	e.mu = e.lambda*e.mu + (1-e.lambda)*rTilde

	// Update variance: exponentially weighted moving variance
	// Use the old mean for the squared deviation (standard EWMA variance formula)
	deviation := rTilde - e.mu
	e.variance = e.lambda*e.variance + (1-e.lambda)*deviation*deviation

	// Update standard deviation
	e.sigma = math.Sqrt(e.variance)

	// Increment trade count
	e.tradeCount++
}

// GetZScore calculates the z-score for a given rTilde
// z = (rTilde - mu) / sigma
//
// Z-score measures how many standard deviations rTilde is from the mean
// Used for mean reversion signal generation:
// - |z| > threshold → price has deviated significantly from mean
// - z > 0 → price moved up (fade by selling)
// - z < 0 → price moved down (fade by buying)
func (e *EWMAState) GetZScore(rTilde float64) float64 {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if e.sigma == 0 {
		return 0 // No volatility yet
	}

	return (rTilde - e.mu) / e.sigma
}

// GetMu returns the current mean
func (e *EWMAState) GetMu() float64 {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.mu
}

// GetSigma returns the current standard deviation
func (e *EWMAState) GetSigma() float64 {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.sigma
}

// GetTradeCount returns the number of trades observed
func (e *EWMAState) GetTradeCount() int {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.tradeCount
}

// ResetEWMA resets the EWMA statistics to initial state
// Used by MeanRev1 after each executed trade (fast reset)
func (e *EWMAState) ResetEWMA() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.mu = 0
	e.variance = 0
	e.sigma = 0
	e.tradeCount = 0
}

// ResetVarianceOnly resets only variance (keeps mean)
// Can be used for partial resets if needed
func (e *EWMAState) ResetVarianceOnly() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.variance = 0
	e.sigma = 0
}
