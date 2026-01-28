// trade_flow.go — Observes all trades and emits flow events for EWMA mean reversion.
//
// When a trade occurs, TradeFlowTracker computes:
//   - Log return: r = ln(priceAfter / priceBefore)
//   - Normalized flow: q = deltaReserve / reserve (liquidity-normalized trade size)
//   - Liquidity-normalized return: rTilde = r / |q|
//
// MeanRev bots subscribe to these events and feed them into their EWMA calculators
// to compute z-scores for trading signals.
//
// CONNECTIONS:
//   - Fed by: main.go trade callback → store.RecordTradeFlow()
//   - Subscribers: bots/meanrev.go (each MeanRev bot subscribes independently)
//   - Data source: executor.go Trade struct carries PriceBefore/After and reserves
package metrics

import (
	"math"
	"math/big"
	"sync"
	"time"

	"eth-amm-sim/internal/engine"
)

// TradeFlowEvent represents a trade with price and flow information
// Used for calculating log returns and liquidity-normalized returns
type TradeFlowEvent struct {
	PriceBefore    float64   `json:"priceBefore"`
	PriceAfter     float64   `json:"priceAfter"`
	TradeSize      float64   `json:"tradeSize"`      // dX in ETH (for BUY) or APPL (for SELL)
	ReservesBefore struct {
		ETH  float64 `json:"eth"`
		APPL float64 `json:"appl"`
	} `json:"reservesBefore"`
	TraderNickname string    `json:"traderNickname"`
	IsBuy          bool      `json:"isBuy"`
	Timestamp      time.Time `json:"timestamp"`
	
	// Calculated fields
	LogReturn      float64 `json:"logReturn"`      // r = ln(priceAfter / priceBefore)
	NormalizedFlow float64 `json:"normalizedFlow"` // q = dX / reserveX
	RTilde         float64 `json:"rTilde"`        // rTilde = r / q (liquidity-normalized return)
}

// TradeFlowTracker tracks trade flow events and provides subscription interface
// for MeanRev bots to observe external market flow
type TradeFlowTracker struct {
	mu sync.RWMutex
	
	// Trade history (for debugging/analysis)
	events []TradeFlowEvent
	
	// Subscribers (MeanRev bots that want to observe trade flow)
	subscribers []TradeFlowSubscriber
	
	// Global trade counter (excludes MeanRev bot's own trades)
	tradeCount int
}

// TradeFlowSubscriber interface for bots that want to observe trade flow
type TradeFlowSubscriber interface {
	OnTradeFlow(event TradeFlowEvent)
	GetNickname() string
}

// NewTradeFlowTracker creates a new trade flow tracker
func NewTradeFlowTracker() *TradeFlowTracker {
	return &TradeFlowTracker{
		events:     make([]TradeFlowEvent, 0),
		subscribers: make([]TradeFlowSubscriber, 0),
	}
}

// Subscribe adds a subscriber (MeanRev bot) to receive trade flow events
func (tft *TradeFlowTracker) Subscribe(subscriber TradeFlowSubscriber) {
	tft.mu.Lock()
	defer tft.mu.Unlock()
	tft.subscribers = append(tft.subscribers, subscriber)
}

// RecordTrade processes a trade and emits flow events to subscribers
// priceBefore: Price before the trade (ETH per APPL)
// priceAfter: Price after the trade (ETH per APPL)
// trade: The trade details
// reservesBefore: Pool reserves before the trade
func (tft *TradeFlowTracker) RecordTrade(
	priceBefore, priceAfter float64,
	trade *engine.Trade,
	reservesBeforeETH, reservesBeforeAPPL *float64,
) {
	tft.mu.Lock()
	defer tft.mu.Unlock()
	
	// Calculate trade size in ETH
	// For BUY: amountIn is ETH, so tradeSize = amountIn
	// For SELL: amountIn is APPL, convert to ETH equivalent using priceBefore
	var tradeSizeETH float64
	if trade.IsBuy {
		// BUY: tradeSize is the ETH spent
		tradeSizeETH = toEtherFloat(trade.AmountIn)
	} else {
		// SELL: tradeSize is APPL sold, convert to ETH equivalent
		appleAmount := toEtherFloat(trade.AmountIn)
		tradeSizeETH = appleAmount * priceBefore
	}
	
	// Calculate log return: r = ln(priceAfter / priceBefore)
	// Use log returns for stationarity and scale-invariance
	var logReturn float64
	if priceBefore > 0 && priceAfter > 0 {
		logReturn = math.Log(priceAfter / priceBefore)
	}
	
	// Calculate normalized flow: q = dX / reserveX
	// For BUY: dX is ETH, reserveX is ETH reserve
	// For SELL: dX is APPL, reserveX is APPL reserve (but we normalize by ETH reserve for consistency)
	var normalizedFlow float64
	if trade.IsBuy {
		// BUY: normalize ETH trade size by ETH reserve
		if *reservesBeforeETH > 0 {
			normalizedFlow = tradeSizeETH / *reservesBeforeETH
		}
	} else {
		// SELL: normalize APPL trade size by APPL reserve
		// Then convert to ETH-equivalent flow for consistency
		if *reservesBeforeAPPL > 0 {
			appleFlow := toEtherFloat(trade.AmountIn) / *reservesBeforeAPPL
			// Convert to ETH-equivalent normalized flow
			normalizedFlow = appleFlow * priceBefore
		}
	}
	
	// Calculate liquidity-normalized return: rTilde = r / q
	// This normalizes returns by trade flow, making statistics comparable across different trade sizes
	// In AMMs, price impact is flow-dependent, so this normalization is essential
	var rTilde float64
	if normalizedFlow > 0 {
		rTilde = logReturn / normalizedFlow
	}
	
	event := TradeFlowEvent{
		PriceBefore:    priceBefore,
		PriceAfter:     priceAfter,
		TradeSize:      tradeSizeETH,
		TraderNickname: trade.Nickname,
		IsBuy:          trade.IsBuy,
		Timestamp:      trade.Timestamp,
		LogReturn:      logReturn,
		NormalizedFlow: normalizedFlow,
		RTilde:         rTilde,
	}
	event.ReservesBefore.ETH = *reservesBeforeETH
	event.ReservesBefore.APPL = *reservesBeforeAPPL
	
	// Store event
	tft.events = append(tft.events, event)
	
	// Notify subscribers (MeanRev bots) - exclude the trader's own bots
	// This ensures MeanRev bots only observe external market flow, not their own trades
	for _, sub := range tft.subscribers {
		if sub.GetNickname() != trade.Nickname {
			// Increment trade count for this subscriber (external trade)
			tft.tradeCount++
			// Notify subscriber
			go sub.OnTradeFlow(event)
		}
	}
}

// GetTradeCount returns the global trade count (excluding MeanRev bot's own trades)
func (tft *TradeFlowTracker) GetTradeCount() int {
	tft.mu.RLock()
	defer tft.mu.RUnlock()
	return tft.tradeCount
}

// GetRecentEvents returns recent trade flow events
func (tft *TradeFlowTracker) GetRecentEvents(n int) []TradeFlowEvent {
	tft.mu.RLock()
	defer tft.mu.RUnlock()
	
	if n >= len(tft.events) {
		result := make([]TradeFlowEvent, len(tft.events))
		copy(result, tft.events)
		return result
	}
	
	result := make([]TradeFlowEvent, n)
	copy(result, tft.events[len(tft.events)-n:])
	return result
}

// Helper function to convert big.Int to float64 (ETH)
func toEtherFloat(wei *big.Int) float64 {
	if wei == nil {
		return 0
	}
	weiFloat := new(big.Float).SetInt(wei)
	ethFloat := new(big.Float).Quo(weiFloat, big.NewFloat(1e18))
	result, _ := ethFloat.Float64()
	return result
}
