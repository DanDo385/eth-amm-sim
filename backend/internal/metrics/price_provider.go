// Package metrics provides market and performance analytics
package metrics

// PriceDataStore defines the interface for accessing price data
// This interface breaks the import cycle between metrics and store packages
type PriceDataStore interface {
	GetCandles() []Candle
	GetTWAP() float64
	GetVolatility() float64
}

// PriceProvider gives bots access to price history
type PriceProvider interface {
	GetRecentPrices(n int) []float64
	GetCurrentPrice() float64
	GetTWAP() float64
	GetVolatility() float64
}

// StorePriceProvider wraps a PriceDataStore to provide price data
type StorePriceProvider struct {
	store PriceDataStore
}

// NewPriceProvider creates a price provider from a price data store
func NewPriceProvider(s PriceDataStore) PriceProvider {
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
