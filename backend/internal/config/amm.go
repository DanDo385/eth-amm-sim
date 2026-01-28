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

	// Initial account equity (in ETH)
	InitialAccountEquityETH = 10000

	// Memory bounds for store (prevent unbounded growth)
	MaxTrades = 50000
	MaxEvents = 5000
)
