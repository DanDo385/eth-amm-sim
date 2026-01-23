// Package metrics provides price impact calculations
package metrics

import (
	"math/big"
	"sync"
)

// ImpactPoint represents price impact at a trade size
type ImpactPoint struct {
	TradeSize   float64 `json:"tradeSize"`   // in ETH
	ImpactBps   float64 `json:"impactBps"`   // impact in basis points
	ExecutePrice float64 `json:"executePrice"` // execution price
	SpotPrice   float64 `json:"spotPrice"`   // spot price before trade
}

// ImpactCurve calculates price impact at various trade sizes
type ImpactCurve struct {
	mu sync.RWMutex
	
	appleReserve *big.Int
	ethReserve   *big.Int
}

// NewImpactCurve creates a new impact curve calculator
func NewImpactCurve() *ImpactCurve {
	return &ImpactCurve{
		appleReserve: big.NewInt(0),
		ethReserve:   big.NewInt(0),
	}
}

// UpdateReserves updates the current reserves
func (ic *ImpactCurve) UpdateReserves(apples, eth *big.Int) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	
	ic.appleReserve = new(big.Int).Set(apples)
	ic.ethReserve = new(big.Int).Set(eth)
}

// CalculateBuyCurve calculates impact for buying APPL with various ETH amounts
func (ic *ImpactCurve) CalculateBuyCurve(sizes []float64) []ImpactPoint {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	
	if ic.appleReserve.Sign() == 0 || ic.ethReserve.Sign() == 0 {
		return nil
	}
	
	spotPrice := toEther(ic.ethReserve) / toEther(ic.appleReserve)
	
	points := make([]ImpactPoint, len(sizes))
	for i, size := range sizes {
		// Calculate amount out using constant product formula
		// Fee: 0.30%
		sizeWei := new(big.Int).Mul(
			big.NewInt(int64(size*1e18)),
			big.NewInt(1),
		)
		
		fee := new(big.Int).Div(
			new(big.Int).Mul(sizeWei, big.NewInt(30)),
			big.NewInt(10000),
		)
		amountInAfterFee := new(big.Int).Sub(sizeWei, fee)
		
		// amountOut = (amountIn * reserveOut) / (reserveIn + amountIn)
		numerator := new(big.Int).Mul(amountInAfterFee, ic.appleReserve)
		denominator := new(big.Int).Add(ic.ethReserve, amountInAfterFee)
		
		if denominator.Sign() == 0 {
			continue
		}
		
		amountOut := new(big.Int).Div(numerator, denominator)
		amountOutFloat := toEther(amountOut)
		
		// Execution price = ETH spent / APPL received
		var execPrice float64
		if amountOutFloat > 0 {
			execPrice = size / amountOutFloat
		}
		
		// Impact in bps = (execPrice - spotPrice) / spotPrice * 10000
		impactBps := 0.0
		if spotPrice > 0 {
			impactBps = ((execPrice - spotPrice) / spotPrice) * 10000
		}
		
		points[i] = ImpactPoint{
			TradeSize:    size,
			ImpactBps:    impactBps,
			ExecutePrice: execPrice,
			SpotPrice:    spotPrice,
		}
	}
	
	return points
}

// CalculateSellCurve calculates impact for selling APPL for various amounts
func (ic *ImpactCurve) CalculateSellCurve(sizes []float64) []ImpactPoint {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	
	if ic.appleReserve.Sign() == 0 || ic.ethReserve.Sign() == 0 {
		return nil
	}
	
	spotPrice := toEther(ic.ethReserve) / toEther(ic.appleReserve)
	
	points := make([]ImpactPoint, len(sizes))
	for i, size := range sizes {
		sizeWei := new(big.Int).Mul(
			big.NewInt(int64(size*1e18)),
			big.NewInt(1),
		)
		
		fee := new(big.Int).Div(
			new(big.Int).Mul(sizeWei, big.NewInt(30)),
			big.NewInt(10000),
		)
		amountInAfterFee := new(big.Int).Sub(sizeWei, fee)
		
		// amountOut = (amountIn * reserveOut) / (reserveIn + amountIn)
		numerator := new(big.Int).Mul(amountInAfterFee, ic.ethReserve)
		denominator := new(big.Int).Add(ic.appleReserve, amountInAfterFee)
		
		if denominator.Sign() == 0 {
			continue
		}
		
		amountOut := new(big.Int).Div(numerator, denominator)
		amountOutFloat := toEther(amountOut)
		
		// Execution price = ETH received / APPL sold
		var execPrice float64
		if size > 0 {
			execPrice = amountOutFloat / size
		}
		
		// Impact in bps (negative for sells)
		impactBps := 0.0
		if spotPrice > 0 {
			impactBps = ((execPrice - spotPrice) / spotPrice) * 10000
		}
		
		points[i] = ImpactPoint{
			TradeSize:    size,
			ImpactBps:    impactBps,
			ExecutePrice: execPrice,
			SpotPrice:    spotPrice,
		}
	}
	
	return points
}

// GetDefaultSizes returns default trade sizes for impact curve
func GetDefaultSizes() []float64 {
	return []float64{
		1, 5, 10, 25, 50, 100, 200, 300, 500, 750, 1000,
	}
}
