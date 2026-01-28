// liquidator.go — Position liquidation bot (account 28).
//
// Strategy: Polls every 3 seconds for leveraged positions that are underwater
// (margin ratio below the leverage-specific threshold). Uses a hybrid
// TWAP/spot price check to prevent false liquidations from temporary price
// spikes. When a position is liquidatable, calls executor.LiquidatePosition
// which invokes AppleAMM.liquidate() on-chain. The liquidator earns a 5%
// fee on the collateral.
//
// CONNECTIONS:
//   - Monitors: leverage bot positions (accounts 25-27)
//   - Reads: executor.IsLiquidatableHybrid (uses TWAP from store/metrics)
//   - Executes: executor.LiquidatePosition → AppleAMM.liquidate()
//   - Frontend: liquidation events appear in KeyEvents component
package bots

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/metrics"
	"eth-amm-sim/internal/store"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// LiquidatorBot monitors leveraged accounts and liquidates them when they're underwater
type LiquidatorBot struct {
	*BaseBot
	privateKey    *ecdsa.PrivateKey
	priceProvider metrics.PriceProvider
	executor      *engine.Executor
}

// NewLiquidatorBot creates a new liquidator bot
func NewLiquidatorBot(cfg *config.AccountConfig, executor *engine.Executor, priceProvider metrics.PriceProvider, store *store.MemoryStore) (*LiquidatorBot, error) {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key for %s: %w", cfg.Nickname, err)
	}

	return &LiquidatorBot{
		BaseBot:       NewBaseBot(cfg, executor, store),
		privateKey:    privateKey,
		priceProvider: priceProvider,
		executor:      executor,
	}, nil
}

// Run starts the liquidator bot monitoring loop
// Now listens for PositionLiquidatable events and also polls as backup
func (l *LiquidatorBot) Run(ctx context.Context) {
	log.Printf("[%s] Liquidator bot started (event-driven + polling)", l.Nickname())

	// Start event monitoring (in background)
	go l.monitorLiquidationEvents(ctx)

	// Also poll every 3 seconds as backup (in case events are missed)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Liquidator bot stopped (context cancelled)", l.Nickname())
			return
		case <-l.stopCh:
			log.Printf("[%s] Liquidator bot stopped", l.Nickname())
			return
		case <-ticker.C:
			// Backup polling - check for liquidatable positions
			l.checkAndLiquidate(ctx)
		}
	}
}

// monitorLiquidationEvents monitors for PositionLiquidatable events
// In a real implementation, this would subscribe to blockchain events
// For now, we'll use polling but structure it to be event-ready
func (l *LiquidatorBot) monitorLiquidationEvents(ctx context.Context) {
	// TODO: Implement actual event subscription using ethclient.SubscribeFilterLogs
	// For now, this is a placeholder that will be called by polling
	// In production, this would listen to: PositionLiquidatable(address indexed trader, ...)
}

// checkAndLiquidate checks all leveraged accounts and liquidates underwater positions
// Uses hybrid TWAP/spot price check to prevent false liquidations from manipulation
func (l *LiquidatorBot) checkAndLiquidate(ctx context.Context) {
	const deviationThreshold = 0.05 // 5% deviation threshold - use TWAP if spot deviates more than this

	// Get all leveraged accounts from config
	leveragedAccounts := config.GetAccountsByType(config.BotTypeLeverage)

	for _, accConfig := range leveragedAccounts {
		// Skip if no collateral configured
		if accConfig.Collateral == nil || accConfig.Collateral.Sign() == 0 {
			continue
		}

		// Get account address
		accountAddr := config.GetAddress(accConfig.Index)

		// Use hybrid check (TWAP if deviated, spot otherwise)
		isLiquidatable, _, priceSource, marginRatio, err :=
			l.executor.IsLiquidatableHybrid(ctx, accountAddr, l.store, deviationThreshold)

		if err != nil {
			// Fallback to contract check if hybrid check fails
			isLiquidatable, err := l.executor.IsLiquidatable(ctx, accountAddr)
			if err != nil || !isLiquidatable {
				continue
			}
			log.Printf("[%s] 🚨 LIQUIDATING %s (contract fallback)", l.Nickname(), accConfig.Nickname)
		} else if !isLiquidatable {
			continue
		} else {
			// Log which price was used (spot or TWAP)
			log.Printf("[%s] 🚨 LIQUIDATING %s (price: %s, margin: %.2f%%)",
				l.Nickname(), accConfig.Nickname, priceSource, marginRatio*100)
		}

		// Execute liquidation through contract
		if err := l.executeLiquidation(ctx, accountAddr, &accConfig); err != nil {
			log.Printf("[%s] Failed to liquidate %s: %v", l.Nickname(), accConfig.Nickname, err)
		} else {
			log.Printf("[%s] ✓ Successfully liquidated %s", l.Nickname(), accConfig.Nickname)

			// Record liquidation event with price source info
			reason := fmt.Sprintf("Liquidated %s using %s (margin: %.2f%%)",
				accConfig.Nickname, priceSource, marginRatio*100)
			l.store.RecordEvent("liquidation", reason, "critical")
		}
	}
}

// shouldLiquidate checks if an account should be liquidated
// Returns (shouldLiquidate, reason)
func (l *LiquidatorBot) shouldLiquidate(ctx context.Context, accountAddr common.Address, accConfig *config.AccountConfig, currentPrice float64) (bool, string) {
	// Get account balances
	ethBalance, err := l.executor.GetETHBalance(ctx, accountAddr)
	if err != nil {
		return false, ""
	}

	appleBalance, err := l.executor.GetAPPLBalance(ctx, accountAddr)
	if err != nil {
		return false, ""
	}

	// Convert balances to float64
	ethBalanceFloat := toEtherFloat(ethBalance)
	appleBalanceFloat := toEtherFloat(appleBalance)

	// Calculate current equity: ETH + APPL * current_price
	currentEquity := ethBalanceFloat + (appleBalanceFloat * currentPrice)

	// Get liquidation threshold (default 0.8 if not set)
	liquidationThreshold := accConfig.LiquidationThreshold
	if liquidationThreshold == 0 {
		liquidationThreshold = 0.8 // Default: liquidate at 80% of collateral
	}

	// Calculate liquidation threshold value
	collateralFloat := toEtherFloat(accConfig.Collateral)
	liquidationValue := collateralFloat * liquidationThreshold

	// Check if equity is below liquidation threshold
	if currentEquity < liquidationValue {
		return true, formatLiquidationReason(currentEquity, collateralFloat, liquidationValue, ethBalanceFloat, appleBalanceFloat, currentPrice)
	}

	return false, ""
}

// executeLiquidation liquidates a position through the contract
// The contract handles closing the position and paying the liquidator fee
func (l *LiquidatorBot) executeLiquidation(ctx context.Context, accountAddr common.Address, accConfig *config.AccountConfig) error {
	// Call contract's liquidate function
	// The liquidator bot uses its own private key (gets liquidation fee)
	txHash, err := l.executor.LiquidatePosition(ctx, l.privateKey, accountAddr)
	if err != nil {
		return fmt.Errorf("failed to call liquidate: %w", err)
	}

	log.Printf("[%s] Liquidated %s via contract (tx: %s)",
		l.Nickname(), accConfig.Nickname, txHash[:10]+"...")

	return nil
}

// Helper functions

func toEtherFloat(wei *big.Int) float64 {
	if wei == nil {
		return 0
	}
	weiFloat := new(big.Float).SetInt(wei)
	ethFloat := new(big.Float).Quo(weiFloat, big.NewFloat(1e18))
	result, _ := ethFloat.Float64()
	return result
}

func formatLiquidationReason(equity, collateral, liquidationValue, ethBalance, appleBalance, currentPrice float64) string {
	return formatEtherFloat(equity) + " ETH equity < " + formatEtherFloat(liquidationValue) +
		" ETH threshold (collateral: " + formatEtherFloat(collateral) +
		" ETH, ETH: " + formatEtherFloat(ethBalance) +
		", APPL: " + formatEtherFloat(appleBalance) +
		", price: " + formatEtherFloat(currentPrice) + ")"
}

func formatEtherFloat(val float64) string {
	wei := big.NewInt(int64(val * 1e18))
	return formatEther(wei)
}
