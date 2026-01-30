// Package config contains AMM-specific configuration constants.
//
// This file centralizes magic numbers used throughout the codebase for AMM fees,
// trade thresholds, initial equity, and memory bounds. This makes configuration
// values discoverable and easier to adjust.
package config

const (
	// AMM fee configuration (0.30% = 30/10000)
	AMMFeeNumerator   = 30
	AMMFeeDenominator = 10000

	// Trade size thresholds for key event generation (in ETH)
	LargeTradeThresholdETH    = 100.0
	CriticalTradeThresholdETH = 300.0

	// Initial account equity (in ETH) — normalized starting balance for performance
	// tracking: 1,000 ETH + 1,000 APPL (valued at spot price, ~2,000 ETH at 1:1).
	InitialAccountEquityETH = 2000

	// Normalized starting balances for session performance tracking
	NormalizedStartingETH  = 1000.0
	NormalizedStartingAPPL = 1000.0

	// Memory bounds for store (prevent unbounded growth)
	MaxTrades = 50000
	MaxEvents = 5000
)
