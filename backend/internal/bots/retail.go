// retail.go — Small, frequent noise trader bot (accounts 10-24).
//
// Strategy: Pure random buy/sell (50/50 coin flip) with small trade sizes
// (up to 15 ETH) at high frequency (1-4s intervals). 15 retail bots generate
// the bulk of trading volume and provide the baseline "noise" that mean
// reversion bots trade against.
//
// Parameters come from config/accounts.go. Trades execute via the Executor
// which calls the on-chain AppleAMM contract.
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
	"eth-amm-sim/internal/store"
	"github.com/ethereum/go-ethereum/crypto"
)

// RetailBot executes small, frequent random trades
type RetailBot struct {
	*BaseBot
	privateKey *ecdsa.PrivateKey
}

// NewRetailBot creates a new retail bot
func NewRetailBot(cfg *config.AccountConfig, executor *engine.Executor, store *store.MemoryStore) (*RetailBot, error) {
	privateKeyHex := cfg.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key for %s: %w", cfg.Nickname, err)
	}

	return &RetailBot{
		BaseBot:    NewBaseBot(cfg, executor, store),
		privateKey: privateKey,
	}, nil
}

// Run starts the retail bot trading loop
func (r *RetailBot) Run(ctx context.Context) {
	log.Printf("[%s] Retail bot started", r.Nickname())

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Retail bot stopped (context cancelled)", r.Nickname())
			return
		default:
			// Wait random interval from config (2-5 seconds)
			delay := r.RandomDelay()
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}

			// Random direction (pure noise)
			var side engine.TradeSide
			if r.rng.Float64() < 0.5 {
				side = engine.BUY
			} else {
				side = engine.SELL
			}

			// Random size from config
			size := r.RandomSize()

			if size.Sign() > 0 {
				// Check balance before attempting trade
				if r.hasSufficientBalance(ctx, side, size) {
					r.executeTrade(ctx, side, size)
				}
			}
		}
	}
}

// hasSufficientBalance checks if the bot has enough balance for the trade
func (r *RetailBot) hasSufficientBalance(ctx context.Context, side engine.TradeSide, size *big.Int) bool {
	addr := crypto.PubkeyToAddress(r.privateKey.PublicKey)
	gasReserve := new(big.Int).Mul(big.NewInt(1e16), big.NewInt(1)) // 0.01 ETH
	
	if side == engine.BUY {
		ethBalance, err := r.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		required := new(big.Int).Add(size, gasReserve)
		return ethBalance.Cmp(required) >= 0
	} else {
		// For SELL: Check APPL balance (size is in ETH, need to convert to APPL)
		// Get current spot price to convert ETH size to APPL
		spotPrice, err := r.executor.GetSpotPrice(ctx)
		if err != nil || spotPrice == nil || spotPrice.Sign() == 0 {
			return false
		}
		
		// Convert ETH size to APPL: appleAmount = ethAmount / price
		// spotPrice is ETH per APPL, so APPL = ETH / price
		sizeFloat := new(big.Float).SetInt(size)
		priceFloat := new(big.Float).SetInt(spotPrice)
		priceFloat.Quo(priceFloat, big.NewFloat(1e18)) // Convert from wei to ETH
		appleNeeded := new(big.Float).Quo(sizeFloat, priceFloat)
		appleNeededInt, _ := appleNeeded.Int(nil)
		
		// Check APPL balance
		appleBalance, err := r.executor.GetAPPLBalance(ctx, addr)
		if err != nil {
			return false
		}
		
		// Check ETH for gas
		ethBalance, err := r.executor.GetETHBalance(ctx, addr)
		if err != nil {
			return false
		}
		
		return appleBalance.Cmp(appleNeededInt) >= 0 && ethBalance.Cmp(gasReserve) >= 0
	}
}

// executeTrade performs a single random trade
func (r *RetailBot) executeTrade(ctx context.Context, side engine.TradeSide, size *big.Int) {
	var err error
	var txHash string

	if side == engine.BUY {
		// Buy: swap ETH for APPLES
		txHash, err = r.executor.SwapETHForApples(ctx, r.privateKey, size)
	} else {
		// Sell: swap APPLES for ETH
		// size is in ETH, need to convert to APPL
		spotPrice, err := r.executor.GetSpotPrice(ctx)
		if err != nil || spotPrice == nil || spotPrice.Sign() == 0 {
			log.Printf("[%s] Cannot execute SELL: failed to get spot price: %v", r.Nickname(), err)
			return
		}
		
		// Convert ETH size to APPL: appleAmount = ethAmount / price
		sizeFloat := new(big.Float).SetInt(size)
		priceFloat := new(big.Float).SetInt(spotPrice)
		priceFloat.Quo(priceFloat, big.NewFloat(1e18)) // Convert from wei to ETH
		appleAmount := new(big.Float).Quo(sizeFloat, priceFloat)
		appleAmountInt, _ := appleAmount.Int(nil)
		
		txHash, err = r.executor.SwapApplesForETH(ctx, r.privateKey, appleAmountInt)
	}

	if err != nil {
		log.Printf("[%s] Trade failed: %v", r.Nickname(), err)
		return
	}

	log.Printf("[%s] ✓ Trade executed: %s (%s, size: %s)", r.Nickname(), txHash[:10]+"...", side, formatEther(size))
}
