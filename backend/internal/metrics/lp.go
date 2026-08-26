// lp.go - Liquidity Provider economics: impermanent loss, fees, and net PnL.
//
// Clean separation of concepts:
//  - HODL Value: value of initial tokens at current price
//  - Theoretical IL: price-only AMM rebalancing drag (≤ 0, no fees, no path)
//  - LP vs HODL PnL: realized performance vs HODL (can be +/-)
//  - Net PnL: LP vs HODL PnL + fees
package metrics

import (
	"math"
	"math/big"
	"sync"
	"time"
)

// LPMetrics tracks liquidity provider metrics
type LPMetrics struct {
	mu sync.RWMutex

	// Initial state
	initialApples *big.Int
	initialETH    *big.Int
	initialPrice  float64 // ETH per APPL

	// Current state
	currentApples *big.Int
	currentETH    *big.Int

	// Fee tracking
	feesApple        *big.Int
	feesETH          *big.Int
	initialFeesApple *big.Int
	initialFeesETH   *big.Int

	history []LPSnapshot
}

// Snapshot for charts (sampled over time for the LP strip / history series).
type LPSnapshot struct {
	Timestamp     time.Time `json:"timestamp"`
	AppleReserve  float64   `json:"appleReserve"`
	ETHReserve    float64   `json:"ethReserve"`
	SpotPrice     float64   `json:"spotPrice"`
	LPValue       float64   `json:"lpValue"`
	HODLValue     float64   `json:"hodlValue"`
	TheoreticalIL float64   `json:"theoreticalIL"` // ETH (≤ 0)
	LPvsHODLPnL   float64   `json:"lpVsHodlPnL"`   // ETH (+/-)
	FeesEarned    float64   `json:"feesEarned"`
	NetPnL        float64   `json:"netPnL"`
}

// API response payload for REST/WS (frontend types/index.ts LPMetrics).
type LPMetricsData struct {
	InitialApples   float64      `json:"initialApples"`
	InitialETH      float64      `json:"initialETH"`
	CurrentApples   float64      `json:"currentApples"`
	CurrentETH      float64      `json:"currentETH"`
	CurrentPrice    float64      `json:"currentPrice"`
	LPValue         float64      `json:"lpValue"`
	HODLValue       float64      `json:"hodlValue"`
	TheoreticalIL   float64      `json:"theoreticalIL"`
	LPvsHODLPnL     float64      `json:"lpVsHodlPnL"`
	FeesEarnedApple float64      `json:"feesEarnedApple"`
	FeesEarnedETH   float64      `json:"feesEarnedETH"`
	TotalFeesUSD    float64      `json:"totalFeesUSD"`
	NetPnL          float64      `json:"netPnL"`
	NetPnLPercent   float64      `json:"netPnLPercent"`
	History         []LPSnapshot `json:"history"`
}

func NewLPMetrics() *LPMetrics {
	return &LPMetrics{
		initialApples:    big.NewInt(0),
		initialETH:       big.NewInt(0),
		currentApples:    big.NewInt(0),
		currentETH:       big.NewInt(0),
		feesApple:        big.NewInt(0),
		feesETH:          big.NewInt(0),
		initialFeesApple: big.NewInt(0),
		initialFeesETH:   big.NewInt(0),
		history:          []LPSnapshot{},
	}
}

func (lp *LPMetrics) SetInitialState(apples, eth *big.Int) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.initialApples = new(big.Int).Set(apples)
	lp.initialETH = new(big.Int).Set(eth)
	lp.currentApples = new(big.Int).Set(apples)
	lp.currentETH = new(big.Int).Set(eth)

	if apples.Sign() > 0 {
		lp.initialPrice = toEther(eth) / toEther(apples)
	}
}

func (lp *LPMetrics) SetInitialFees(feesApple, feesETH *big.Int) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.initialFeesApple = new(big.Int).Set(feesApple)
	lp.initialFeesETH = new(big.Int).Set(feesETH)
}

func (lp *LPMetrics) UpdateState(apples, eth, feesApple, feesETH *big.Int) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.currentApples = new(big.Int).Set(apples)
	lp.currentETH = new(big.Int).Set(eth)
	lp.feesApple = new(big.Int).Set(feesApple)
	lp.feesETH = new(big.Int).Set(feesETH)

	lp.history = append(lp.history, lp.calculateSnapshotLocked())
}

func (lp *LPMetrics) GetMetrics() LPMetricsData {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	s := lp.calculateSnapshotLocked()

	feesAppleEarned := maxBig(new(big.Int).Sub(lp.feesApple, lp.initialFeesApple))
	feesETHEarned := maxBig(new(big.Int).Sub(lp.feesETH, lp.initialFeesETH))

	return LPMetricsData{
		InitialApples:   toEther(lp.initialApples),
		InitialETH:      toEther(lp.initialETH),
		CurrentApples:   toEther(lp.currentApples),
		CurrentETH:      toEther(lp.currentETH),
		CurrentPrice:    s.SpotPrice,
		LPValue:         s.LPValue,
		HODLValue:       s.HODLValue,
		TheoreticalIL:   s.TheoreticalIL,
		LPvsHODLPnL:     s.LPvsHODLPnL,
		FeesEarnedApple: toEther(feesAppleEarned),
		FeesEarnedETH:   toEther(feesETHEarned),
		TotalFeesUSD:    s.FeesEarned,
		NetPnL:          s.NetPnL,
		NetPnLPercent:   pct(s.NetPnL, s.HODLValue),
		History:         append([]LPSnapshot{}, lp.history...),
	}
}

func (lp *LPMetrics) calculateSnapshotLocked() LPSnapshot {
	apples := toEther(lp.currentApples)
	eth := toEther(lp.currentETH)
	initApples := toEther(lp.initialApples)
	initETH := toEther(lp.initialETH)

	var price float64
	if apples > 0 {
		price = eth / apples
	}

	lpValue := apples*price + eth
	hodlValue := initApples*price + initETH

	lpVsHodl := lpValue - hodlValue
	ilPct := theoreticalIL(price, lp.initialPrice)
	ilETH := hodlValue * ilPct

	feesApple := toEther(maxBig(new(big.Int).Sub(lp.feesApple, lp.initialFeesApple)))
	feesETH := toEther(maxBig(new(big.Int).Sub(lp.feesETH, lp.initialFeesETH)))
	totalFees := feesApple*price + feesETH

	return LPSnapshot{
		Timestamp:     time.Now(),
		AppleReserve:  apples,
		ETHReserve:    eth,
		SpotPrice:     price,
		LPValue:       lpValue,
		HODLValue:     hodlValue,
		TheoreticalIL: ilETH,
		LPvsHODLPnL:   lpVsHodl,
		FeesEarned:    totalFees,
		NetPnL:        lpVsHodl + totalFees,
	}
}

func theoreticalIL(price, initial float64) float64 {
	if initial == 0 {
		return 0
	}
	r := price / initial
	return (2*math.Sqrt(r))/(1+r) - 1
}

func toEther(wei *big.Int) float64 {
	if wei == nil {
		return 0
	}
	f := new(big.Float).SetInt(wei)
	f.Quo(f, big.NewFloat(1e18))
	v, _ := f.Float64()
	return v
}

func maxBig(v *big.Int) *big.Int {
	if v.Sign() < 0 {
		return big.NewInt(0)
	}
	return v
}

func pct(v, base float64) float64 {
	if base == 0 {
		return 0
	}
	return (v / base) * 100
}
// Reset clears all LP metrics state
func (lp *LPMetrics) Reset() {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.initialApples = big.NewInt(0)
	lp.initialETH = big.NewInt(0)
	lp.initialPrice = 0

	lp.currentApples = big.NewInt(0)
	lp.currentETH = big.NewInt(0)

	lp.feesApple = big.NewInt(0)
	lp.feesETH = big.NewInt(0)
	lp.initialFeesApple = big.NewInt(0)
	lp.initialFeesETH = big.NewInt(0)

	lp.history = []LPSnapshot{}
}
