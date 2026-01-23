// Package engine provides shared types
package engine

import (
	"math/big"
	"time"
)

// TradeSide represents buy or sell
type TradeSide int

const (
	BUY TradeSide = iota
	SELL
)

func (s TradeSide) String() string {
	if s == BUY {
		return "BUY"
	}
	return "SELL"
}

// TradeAction represents a trade to be executed
type TradeAction struct {
	AccountIndex int
	Nickname     string
	Side         TradeSide
	Size         *big.Int   // Amount of APPLES (for sell) or ETH (for buy)
	MinOutput    *big.Int   // Slippage protection (optional)
	Timestamp    time.Time
}
