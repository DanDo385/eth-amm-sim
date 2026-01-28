// Package engine provides shared types
package engine

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
