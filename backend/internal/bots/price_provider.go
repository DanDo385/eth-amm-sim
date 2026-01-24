// Package bots provides price data access for trading bots
package bots

import (
	"eth-amm-sim/internal/store"
)

// PriceProvider gives bots access to price history
type PriceProvider interface {
	GetRecentPrices(n int) []float64
	GetCurrentPrice() float64
	GetTWAP() float64
	GetVolatility() float64
}

// StorePriceProvider wraps MemoryStore to provide price data
type StorePriceProvider struct {
	store *store.MemoryStore
}

// NewPriceProvider creates a price provider from a store
func NewPriceProvider(s *store.MemoryStore) PriceProvider {
	return &StorePriceProvider{store: s}
}

// GetRecentPrices returns the last N price observations
func (p *StorePriceProvider) GetRecentPrices(n int) []float64 {
	// Get candles and extract close prices
	candles := p.store.GetCandles()
	if len(candles) == 0 {
		return []float64{}
	}
	
	prices := make([]float64, 0, n)
	start := len(candles) - n
	if start < 0 {
		start = 0
	}
	
	for i := start; i < len(candles); i++ {
		prices = append(prices, candles[i].Close)
	}
	
	return prices
}

// GetCurrentPrice returns the most recent price
func (p *StorePriceProvider) GetCurrentPrice() float64 {
	candles := p.store.GetCandles()
	if len(candles) == 0 {
		return 0
	}
	return candles[len(candles)-1].Close
}

// GetTWAP returns the Time-Weighted Average Price
func (p *StorePriceProvider) GetTWAP() float64 {
	return p.store.GetTWAP()
}

// GetVolatility returns price volatility
func (p *StorePriceProvider) GetVolatility() float64 {
	return p.store.GetVolatility()
}
