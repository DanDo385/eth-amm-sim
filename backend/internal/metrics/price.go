// Package metrics provides market and performance analytics
package metrics

import (
	"math"
	"sync"
	"time"
)

// Candle represents OHLC price data
type Candle struct {
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// PriceMetrics tracks price-related metrics
type PriceMetrics struct {
	mu sync.RWMutex
	
	// Raw price observations
	prices     []pricePoint
	
	// Candle data
	candles    []Candle
	candleSize time.Duration
	
	// TWAP state
	twapWindow time.Duration
}

type pricePoint struct {
	price     float64
	timestamp time.Time
}

// NewPriceMetrics creates a new price metrics tracker
func NewPriceMetrics(candleSize, twapWindow time.Duration) *PriceMetrics {
	return &PriceMetrics{
		prices:     make([]pricePoint, 0),
		candles:    make([]Candle, 0),
		candleSize: candleSize,
		twapWindow: twapWindow,
	}
}

// RecordPrice records a new price observation
func (pm *PriceMetrics) RecordPrice(price float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	now := time.Now()
	pm.prices = append(pm.prices, pricePoint{price: price, timestamp: now})
	
	// Update candles
	pm.updateCandles(price, now)
}

// updateCandles updates the candle data
func (pm *PriceMetrics) updateCandles(price float64, now time.Time) {
	if len(pm.candles) == 0 {
		// Create first candle
		pm.candles = append(pm.candles, Candle{
			Open:      price,
			High:      price,
			Low:       price,
			Close:     price,
			Volume:    0,
			Timestamp: now.Truncate(pm.candleSize),
		})
		return
	}
	
	last := &pm.candles[len(pm.candles)-1]
	candleStart := now.Truncate(pm.candleSize)
	
	if candleStart.After(last.Timestamp) {
		// New candle
		pm.candles = append(pm.candles, Candle{
			Open:      price,
			High:      price,
			Low:       price,
			Close:     price,
			Volume:    0,
			Timestamp: candleStart,
		})
	} else {
		// Update current candle
		if price > last.High {
			last.High = price
		}
		if price < last.Low {
			last.Low = price
		}
		last.Close = price
	}
}

// AddVolume adds volume to the current candle
func (pm *PriceMetrics) AddVolume(volume float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if len(pm.candles) > 0 {
		pm.candles[len(pm.candles)-1].Volume += volume
	}
}

// GetTWAP returns the Time-Weighted Average Price
func (pm *PriceMetrics) GetTWAP() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	if len(pm.prices) == 0 {
		return 0
	}
	
	cutoff := time.Now().Add(-pm.twapWindow)
	
	var weightedSum float64
	var totalWeight float64
	
	for i := len(pm.prices) - 1; i >= 0; i-- {
		if pm.prices[i].timestamp.Before(cutoff) {
			break
		}
		
		// Weight by time (simple equal weight for now)
		weight := 1.0
		weightedSum += pm.prices[i].price * weight
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		return pm.prices[len(pm.prices)-1].price
	}
	
	return weightedSum / totalWeight
}

// GetVolatility returns the price volatility (standard deviation of returns)
func (pm *PriceMetrics) GetVolatility() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	if len(pm.prices) < 2 {
		return 0
	}
	
	// Calculate returns
	returns := make([]float64, len(pm.prices)-1)
	for i := 1; i < len(pm.prices); i++ {
		if pm.prices[i-1].price > 0 {
			returns[i-1] = (pm.prices[i].price - pm.prices[i-1].price) / pm.prices[i-1].price
		}
	}
	
	return standardDeviation(returns)
}

// GetCandles returns all candles
func (pm *PriceMetrics) GetCandles() []Candle {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	result := make([]Candle, len(pm.candles))
	copy(result, pm.candles)
	return result
}

// GetLatestPrice returns the most recent price
func (pm *PriceMetrics) GetLatestPrice() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	if len(pm.prices) == 0 {
		return 0
	}
	return pm.prices[len(pm.prices)-1].price
}

// Reset clears all price data
func (pm *PriceMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.prices = make([]pricePoint, 0)
	pm.candles = make([]Candle, 0)
}

// standardDeviation calculates the standard deviation of a slice
func standardDeviation(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	// Calculate mean
	var sum float64
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(len(data))
	
	// Calculate variance
	var variance float64
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(data))
	
	return math.Sqrt(variance)
}
