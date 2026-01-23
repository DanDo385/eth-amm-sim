// Package store provides in-memory storage for session data
package store

import (
	"sync"
	"time"

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
	
	// Trade blotter
	trades []engine.Trade
	
	// Key events
	events []KeyEvent
}

// KeyEvent represents a significant event
type KeyEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // trade, liquidation, strategy_trigger
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // info, warning, critical
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		priceMetrics:   metrics.NewPriceMetrics(5*time.Second, 60*time.Second),
		lpMetrics:      metrics.NewLPMetrics(),
		accountMetrics: metrics.NewAccountMetricsManager(),
		impactCurve:    metrics.NewImpactCurve(),
		trades:         make([]engine.Trade, 0),
		events:         make([]KeyEvent, 0),
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
	return s.accountMetrics.GetAllPerformance()
}

func (s *MemoryStore) GetAccountPerformance(nickname string) *metrics.PerformanceData {
	am := s.accountMetrics.Get(nickname)
	if am == nil {
		return nil
	}
	perf := am.GetPerformance()
	return &perf
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

// Reset clears all stored data
func (s *MemoryStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.priceMetrics.Reset()
	s.lpMetrics.Reset()
	s.accountMetrics.Reset()
	s.trades = make([]engine.Trade, 0)
	s.events = make([]KeyEvent, 0)
}
