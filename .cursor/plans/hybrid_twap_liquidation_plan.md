# Hybrid TWAP/Spot Price Liquidation Plan

## Overview
Implement a hybrid liquidation mechanism that uses TWAP when spot price deviates significantly from TWAP (>X%), otherwise uses spot price. This prevents false liquidations from temporary price manipulation while maintaining fast liquidation in normal conditions.

## Current Architecture

**Liquidation Flow**:
1. Liquidator bot polls/checks positions every 3 seconds
2. Calls `executor.IsLiquidatable()` which queries contract's `isLiquidatable()`
3. Contract uses `getSpotPrice()` directly for position valuation
4. If liquidatable, calls `executor.LiquidatePosition()` → contract's `liquidate()`

**Price Sources**:
- Spot price: Available in contract via `getSpotPrice()`
- TWAP: Available in backend via `store.GetTWAP()` (60-second window)

## Design Decision: Backend Hybrid Check

**Why backend, not contract**:
- TWAP calculation requires historical data (not available on-chain)
- Contract remains simple and gas-efficient
- Backend can make intelligent decisions before calling contract
- Easier to tune deviation threshold without contract redeployment

## Implementation Plan

### 1. Add Hybrid Price Function

**File**: `backend/internal/engine/executor.go`

**New Function**:
```go
// GetHybridPrice returns TWAP if spot deviates >threshold% from TWAP, otherwise spot
// This prevents false liquidations from temporary price manipulation
func (e *Executor) GetHybridPrice(ctx context.Context, store *store.MemoryStore, deviationThreshold float64) (float64, string, error) {
    // Get spot price
    spotPriceBig, err := e.GetSpotPrice(ctx)
    if err != nil {
        return 0, "", err
    }
    spotPrice := toEtherFloat(spotPriceBig)
    
    // Get TWAP
    twap := store.GetTWAP()
    if twap == 0 {
        // No TWAP data yet, use spot
        return spotPrice, "spot (no TWAP)", nil
    }
    
    // Calculate deviation: |spot - twap| / twap
    deviation := math.Abs(spotPrice - twap) / twap
    
    if deviation > deviationThreshold {
        // Significant deviation - use TWAP (more stable)
        return twap, fmt.Sprintf("TWAP (deviation: %.2f%%)", deviation*100), nil
    }
    
    // Normal conditions - use spot (faster)
    return spotPrice, "spot", nil
}
```

### 2. Add Hybrid Liquidation Check

**File**: `backend/internal/engine/executor.go`

**New Function**:
```go
// IsLiquidatableHybrid checks liquidation using hybrid price (TWAP if deviated, spot otherwise)
// Returns (liquidatable, priceUsed, priceSource, marginRatio, error)
func (e *Executor) IsLiquidatableHybrid(
    ctx context.Context, 
    trader common.Address, 
    store *store.MemoryStore,
    deviationThreshold float64,
) (bool, float64, string, float64, error) {
    // Get hybrid price
    hybridPrice, priceSource, err := e.GetHybridPrice(ctx, store, deviationThreshold)
    if err != nil {
        return false, 0, "", 0, err
    }
    
    // Get position from contract
    pos, err := e.GetTraderPosition(ctx, trader)
    if err != nil {
        return false, 0, "", 0, err
    }
    
    if pos.Collateral == nil || pos.Collateral.Sign() == 0 {
        return false, hybridPrice, priceSource, 0, nil
    }
    
    // Calculate position value using hybrid price
    // Convert hybridPrice (float64) to *big.Int scaled by 1e18
    hybridPriceBig := new(big.Int).SetUint64(uint64(hybridPrice * 1e18))
    
    // Calculate position value: APPL_amount * hybrid_price
    applValue := new(big.Int).Mul(pos.Size, hybridPriceBig)
    applValue.Div(applValue, big.NewInt(1e18))
    
    // Calculate leveraged cost
    leveragedCost := new(big.Int).Mul(pos.Collateral, big.NewInt(int64(pos.Leverage)))
    
    // Calculate margin ratio (in basis points)
    if leveragedCost.Sign() == 0 {
        return false, hybridPrice, priceSource, 0, nil
    }
    
    marginRatioBig := new(big.Int).Mul(applValue, big.NewInt(10000))
    marginRatioBig.Div(marginRatioBig, leveragedCost)
    marginRatio := float64(marginRatioBig.Uint64()) / 10000.0
    
    // Check if liquidatable (margin ratio < 10%)
    liquidatable := marginRatio < 0.10
    
    return liquidatable, hybridPrice, priceSource, marginRatio, nil
}
```

### 3. Update Liquidator Bot

**File**: `backend/internal/bots/liquidator.go`

**Changes**:
- Add deviation threshold constant (e.g., 5% = 0.05)
- Use `IsLiquidatableHybrid()` instead of `IsLiquidatable()`
- Log which price source was used (spot vs TWAP)
- Only call contract liquidation if hybrid check passes

**Modified Function**:
```go
func (l *LiquidatorBot) checkAndLiquidate(ctx context.Context) {
    const deviationThreshold = 0.05 // 5% deviation threshold
    
    leveragedAccounts := config.GetAccountsByType(config.BotTypeLeverage)
    
    for _, accConfig := range leveragedAccounts {
        if accConfig.Collateral == nil || accConfig.Collateral.Sign() == 0 {
            continue
        }
        
        accountAddr := config.GetAddress(accConfig.Index)
        
        // Use hybrid check (TWAP if deviated, spot otherwise)
        isLiquidatable, priceUsed, priceSource, marginRatio, err := 
            l.executor.IsLiquidatableHybrid(ctx, accountAddr, l.store, deviationThreshold)
        
        if err != nil {
            // Fallback to contract check
            isLiquidatable, _ := l.executor.IsLiquidatable(ctx, accountAddr)
            if !isLiquidatable {
                continue
            }
        } else if !isLiquidatable {
            continue
        }
        
        // Log which price was used
        log.Printf("[%s] 🚨 LIQUIDATING %s (price: %s, margin: %.2f%%)", 
            l.Nickname(), accConfig.Nickname, priceSource, marginRatio*100)
        
        // Execute liquidation
        if err := l.executeLiquidation(ctx, accountAddr, &accConfig); err != nil {
            log.Printf("[%s] Failed to liquidate %s: %v", l.Nickname(), accConfig.Nickname, err)
        } else {
            log.Printf("[%s] ✓ Successfully liquidated %s", l.Nickname(), accConfig.Nickname)
            l.store.RecordEvent("liquidation", 
                fmt.Sprintf("Liquidated %s using %s", accConfig.Nickname, priceSource), 
                "critical")
        }
    }
}
```

### 4. Add Configuration

**File**: `backend/internal/config/accounts.go` (optional)

**Option**: Add `TWAPDeviationThreshold float64` to `AccountConfig` for per-account thresholds.

**Recommendation**: Use global constant for simplicity (5% deviation).

### 5. Update Price Provider Interface (if needed)

**File**: `backend/internal/metrics/price_provider.go`

**Check**: Ensure `GetTWAP()` is available (already exists).

## Deviation Threshold Selection

**Recommended: 5% (0.05)**

**Reasoning**:
- Small enough to catch manipulation attempts
- Large enough to avoid false positives from normal volatility
- Common in production DeFi protocols (Uniswap V3 uses similar thresholds)

**Alternative thresholds**:
- 3%: More sensitive (catches smaller manipulations, more false positives)
- 10%: Less sensitive (fewer false positives, might miss some manipulations)

## Example Scenarios

### Scenario 1: Normal Volatility (Use Spot)
- Spot price: 100 ETH/APPL
- TWAP: 101 ETH/APPL
- Deviation: |100-101|/101 = 0.99% < 5%
- **Result**: Use spot price (fast liquidation)

### Scenario 2: Price Manipulation (Use TWAP)
- Spot price: 95 ETH/APPL (manipulated down)
- TWAP: 100 ETH/APPL
- Deviation: |95-100|/100 = 5% ≥ 5%
- **Result**: Use TWAP (prevents false liquidation)

### Scenario 3: Legitimate Price Move (Use TWAP)
- Spot price: 110 ETH/APPL (real move)
- TWAP: 100 ETH/APPL
- Deviation: |110-100|/100 = 10% ≥ 5%
- **Result**: Use TWAP (more stable, prevents premature liquidation)

## Benefits

1. **Manipulation Resistance**: Prevents false liquidations from flash loan attacks
2. **Fast Normal Liquidations**: Uses spot price when conditions are normal
3. **Production-Like**: Mimics real DeFi protocol behavior
4. **Demo Value**: Shows sophisticated risk management
5. **No Contract Changes**: Keeps contract simple and gas-efficient

## Testing Considerations

- Test with normal price movements (<5% deviation)
- Test with manipulated prices (>5% deviation)
- Test with no TWAP data (fallback to spot)
- Test with TWAP = 0 (edge case)
- Verify liquidations use correct price source

## Files to Modify

1. `backend/internal/engine/executor.go` - Add `GetHybridPrice()` and `IsLiquidatableHybrid()`
2. `backend/internal/bots/liquidator.go` - Use hybrid check instead of direct contract check
3. `backend/internal/store/memory.go` - Ensure `GetTWAP()` is accessible (already exists)

## Implementation Order

1. Add `GetHybridPrice()` helper function
2. Add `IsLiquidatableHybrid()` function
3. Update liquidator bot to use hybrid check
4. Test with various price scenarios
5. Add logging to show which price source was used

## Notes

- Contract's `isLiquidatable()` remains unchanged (backward compatible)
- Hybrid check is done in backend before calling contract
- Contract still uses spot price for actual liquidation (execution price)
- This is a pre-check to prevent unnecessary liquidation calls
