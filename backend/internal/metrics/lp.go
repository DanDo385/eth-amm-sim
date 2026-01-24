// Package metrics provides LP (Liquidity Provider) analytics
package metrics

import (
	"math/big"
	"sync"
	"time"
)

// LPMetrics tracks liquidity provider metrics
type LPMetrics struct {
	mu sync.RWMutex
	
	// Initial state (when LP deposited)
	initialApples *big.Int
	initialETH    *big.Int
	initialPrice  float64 // ETH per APPL
	depositTime   time.Time
	
	// Current state
	currentApples *big.Int
	currentETH    *big.Int
	
	// Fee tracking
	feesApple     *big.Int
	feesETH       *big.Int
	initialFeesApple *big.Int
	initialFeesETH   *big.Int
	
	// History for charts
	history []LPSnapshot
}

// LPSnapshot represents LP state at a point in time
type LPSnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	AppleReserve float64   `json:"appleReserve"`
	ETHReserve   float64   `json:"ethReserve"`
	SpotPrice    float64   `json:"spotPrice"`
	LPValue      float64   `json:"lpValue"`
	HODLValue    float64   `json:"hodlValue"`
	IL           float64   `json:"il"`
	FeesEarned   float64   `json:"feesEarned"`
	NetPnL       float64   `json:"netPnL"`
}

// LPMetricsData is the API response for LP metrics
type LPMetricsData struct {
	InitialApples    float64      `json:"initialApples"`
	InitialETH       float64      `json:"initialETH"`
	CurrentApples    float64      `json:"currentApples"`
	CurrentETH       float64      `json:"currentETH"`
	CurrentPrice     float64      `json:"currentPrice"`
	LPValue          float64      `json:"lpValue"`
	HODLValue        float64      `json:"hodlValue"`
	ImpermanentLoss  float64      `json:"impermanentLoss"`
	FeesEarnedApple  float64      `json:"feesEarnedApple"`
	FeesEarnedETH    float64      `json:"feesEarnedETH"`
	TotalFeesUSD     float64      `json:"totalFeesUSD"`
	NetPnL           float64      `json:"netPnL"`
	NetPnLPercent    float64      `json:"netPnLPercent"`
	History          []LPSnapshot `json:"history"`
}

// NewLPMetrics creates a new LP metrics tracker
func NewLPMetrics() *LPMetrics {
	return &LPMetrics{
		initialApples:    big.NewInt(0),
		initialETH:       big.NewInt(0),
		currentApples:    big.NewInt(0),
		currentETH:       big.NewInt(0),
		feesApple:         big.NewInt(0),
		feesETH:          big.NewInt(0),
		initialFeesApple: big.NewInt(0),
		initialFeesETH:   big.NewInt(0),
		history:          make([]LPSnapshot, 0),
	}
}

// SetInitialState sets the LP's initial deposit
func (lp *LPMetrics) SetInitialState(apples, eth *big.Int) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	lp.initialApples = new(big.Int).Set(apples)
	lp.initialETH = new(big.Int).Set(eth)
	lp.currentApples = new(big.Int).Set(apples)
	lp.currentETH = new(big.Int).Set(eth)
	lp.depositTime = time.Now()
	
	// Calculate initial price (ETH per APPL)
	if apples.Sign() > 0 {
		ethFloat := new(big.Float).SetInt(eth)
		appleFloat := new(big.Float).SetInt(apples)
		price := new(big.Float).Quo(ethFloat, appleFloat)
		lp.initialPrice, _ = price.Float64()
	}
}

// SetInitialFees sets the initial fees from the contract (for tracking fees earned since deposit)
func (lp *LPMetrics) SetInitialFees(feesApple, feesETH *big.Int) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	lp.initialFeesApple = new(big.Int).Set(feesApple)
	lp.initialFeesETH = new(big.Int).Set(feesETH)
}

// UpdateState updates current reserves and fees
func (lp *LPMetrics) UpdateState(apples, eth, feesApple, feesETH *big.Int) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	lp.currentApples = new(big.Int).Set(apples)
	lp.currentETH = new(big.Int).Set(eth)
	lp.feesApple = new(big.Int).Set(feesApple)
	lp.feesETH = new(big.Int).Set(feesETH)
	
	// Record snapshot
	snapshot := lp.calculateSnapshotLocked()
	lp.history = append(lp.history, snapshot)
}

// GetMetrics returns the current LP metrics
func (lp *LPMetrics) GetMetrics() LPMetricsData {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	snapshot := lp.calculateSnapshotLocked()
	
	// Fees earned = current fees - initial fees
	feesAppleEarned := new(big.Int).Sub(lp.feesApple, lp.initialFeesApple)
	feesETHEarned := new(big.Int).Sub(lp.feesETH, lp.initialFeesETH)
	if feesAppleEarned.Sign() < 0 {
		feesAppleEarned = big.NewInt(0)
	}
	if feesETHEarned.Sign() < 0 {
		feesETHEarned = big.NewInt(0)
	}
	
	return LPMetricsData{
		InitialApples:   toEther(lp.initialApples),
		InitialETH:      toEther(lp.initialETH),
		CurrentApples:   toEther(lp.currentApples),
		CurrentETH:      toEther(lp.currentETH),
		CurrentPrice:    snapshot.SpotPrice,
		LPValue:         snapshot.LPValue,
		HODLValue:       snapshot.HODLValue,
		ImpermanentLoss: snapshot.IL,
		FeesEarnedApple: toEther(feesAppleEarned),
		FeesEarnedETH:   toEther(feesETHEarned),
		TotalFeesUSD:    snapshot.FeesEarned,
		NetPnL:          snapshot.NetPnL,
		NetPnLPercent:   calculatePnLPercent(snapshot),
		History:         lp.getHistoryCopy(),
	}
}

// calculateSnapshotLocked calculates a snapshot (caller must hold lock)
func (lp *LPMetrics) calculateSnapshotLocked() LPSnapshot {
	now := time.Now()
	
	currentApples := toEther(lp.currentApples)
	currentETH := toEther(lp.currentETH)
	initialApples := toEther(lp.initialApples)
	initialETH := toEther(lp.initialETH)
	
	// Spot price (ETH per APPL)
	var spotPrice float64
	if currentApples > 0 {
		spotPrice = currentETH / currentApples
	}
	
	// LP value = current_apples * price + current_eth
	lpValue := currentApples*spotPrice + currentETH
	
	// HODL value = initial_apples * current_price + initial_eth
	hodlValue := initialApples*spotPrice + initialETH
	
	// Impermanent Loss = LP_value - HODL_value
	il := lpValue - hodlValue
	
	// Fees earned (convert to ETH equivalent)
	// Calculate fees earned since deposit
	feesAppleEarned := new(big.Int).Sub(lp.feesApple, lp.initialFeesApple)
	feesETHEarned := new(big.Int).Sub(lp.feesETH, lp.initialFeesETH)
	if feesAppleEarned.Sign() < 0 {
		feesAppleEarned = big.NewInt(0)
	}
	if feesETHEarned.Sign() < 0 {
		feesETHEarned = big.NewInt(0)
	}
	
	feesApple := toEther(feesAppleEarned)
	feesETH := toEther(feesETHEarned)
	totalFees := feesApple*spotPrice + feesETH
	
	// Net PnL = IL + Fees
	netPnL := il + totalFees
	
	return LPSnapshot{
		Timestamp:    now,
		AppleReserve: currentApples,
		ETHReserve:   currentETH,
		SpotPrice:    spotPrice,
		LPValue:      lpValue,
		HODLValue:    hodlValue,
		IL:           il,
		FeesEarned:   totalFees,
		NetPnL:       netPnL,
	}
}

func (lp *LPMetrics) getHistoryCopy() []LPSnapshot {
	result := make([]LPSnapshot, len(lp.history))
	copy(result, lp.history)
	return result
}

func calculatePnLPercent(snapshot LPSnapshot) float64 {
	if snapshot.HODLValue == 0 {
		return 0
	}
	return (snapshot.NetPnL / snapshot.HODLValue) * 100
}

// toEther converts wei to ETH (float64)
func toEther(wei *big.Int) float64 {
	if wei == nil {
		return 0
	}
	f := new(big.Float).SetInt(wei)
	f.Quo(f, big.NewFloat(1e18))
	result, _ := f.Float64()
	return result
}

// Reset clears all LP data
func (lp *LPMetrics) Reset() {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	lp.initialApples = big.NewInt(0)
	lp.initialETH = big.NewInt(0)
	lp.currentApples = big.NewInt(0)
	lp.currentETH = big.NewInt(0)
	lp.feesApple = big.NewInt(0)
	lp.feesETH = big.NewInt(0)
	lp.initialFeesApple = big.NewInt(0)
	lp.initialFeesETH = big.NewInt(0)
	lp.history = make([]LPSnapshot, 0)
}
