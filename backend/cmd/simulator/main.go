// Package main is the entry point for the AMM simulation backend.
//
// SYSTEM ROLE:
// This is the central orchestrator that wires together every backend component.
// It connects the Go backend to the on-chain contracts (via Anvil RPC) and serves
// real-time data to the Next.js frontend (via HTTP REST + WebSocket on :8080).
//
// STARTUP SEQUENCE:
//  1. Load config and contract addresses (from broadcast JSON written by deploy.sh)
//  2. Connect to Anvil at localhost:8545 (chain/client.go)
//  3. Verify contracts are deployed by reading pool reserves
//  4. Create Executor (engine/executor.go) — the bridge to on-chain contract calls
//  5. Create bots from config/accounts.go definitions, each wired to the Executor
//  6. Start price polling goroutine (every 2s reads reserves → updates metrics → broadcasts)
//  7. Start HTTP+WebSocket server (server/server.go) on :8080 for the frontend
//
// DATA FLOW:
//
//	Contract (Anvil) → Executor → Trade callback → MemoryStore + WebSocket → Frontend
//	Price poll loop  → Executor.GetSpotPrice → MemoryStore → WebSocket → Frontend
//
// CONNECTIONS TO OTHER PARTS:
//
//	contracts/: Addresses loaded from contracts/broadcast/Deploy.s.sol/31337/run-latest.json
//	frontend/:  Connects to :8080 for REST API and WebSocket (see frontend/lib/api.ts, hooks/useWebSocket.ts)
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eth-amm-sim/internal/bots"
	"eth-amm-sim/internal/chain"
	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/metrics"
	"eth-amm-sim/internal/server"
	"eth-amm-sim/internal/store"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

func main() {
	log.Println("=== ETH-AMM-SIM Backend Starting ===")

	// Load configuration
	cfg := config.DefaultConfig()

	// Get contract addresses (priority: env vars > broadcast JSON > defaults)
	tokenAddr := os.Getenv("TOKEN_ADDRESS")
	ammAddr := os.Getenv("AMM_ADDRESS")

	if tokenAddr == "" || ammAddr == "" {
		// Try to load from broadcast JSON
		broadcastToken, broadcastAMM, err := config.LoadAddressesFromBroadcast()
		if err == nil {
			tokenAddr = broadcastToken
			ammAddr = broadcastAMM
			log.Println("✓ Loaded contract addresses from broadcast output")
		} else {
			log.Printf("Note: Could not load from broadcast (%v), using defaults", err)
			// Use default addresses (Anvil's deterministic addresses)
			tokenAddr = "0x5FbDB2315678afecb367f032d93F642f64180aa3"
			ammAddr = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"
		}
	} else {
		log.Println("✓ Using contract addresses from environment variables")
	}

	cfg.TokenAddress = common.HexToAddress(tokenAddr)
	cfg.AMMAddress = common.HexToAddress(ammAddr)

	log.Printf("Token Address: %s", cfg.TokenAddress.Hex())
	log.Printf("AMM Address: %s", cfg.AMMAddress.Hex())

	// Connect to Anvil
	client, err := chain.NewClient(cfg.RPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to chain: %v", err)
	}
	log.Printf("Connected to chain (ID: %s)", client.ChainID().String())

	// Initialize components
	nonceManager := chain.NewNonceManager(client)
	executor, err := engine.NewExecutor(client, nonceManager, cfg.AMMAddress, cfg.TokenAddress)
	if err != nil {
		log.Fatalf("Failed to create executor: %v", err)
	}

	// Verify contracts are deployed and initialized
	ctx := context.Background()
	log.Println("Verifying contracts...")

	// Check if we can read from the AMM contract
	apples, eth, err := executor.GetReserves(ctx)
	if err != nil {
		log.Fatalf("Failed to read AMM reserves. Make sure contracts are deployed and pool is initialized. Error: %v", err)
	}

	if apples.Sign() == 0 || eth.Sign() == 0 {
		log.Fatalf("AMM pool is empty! Reserves: APPL=%s, ETH=%s. Deploy contracts and seed liquidity first.", apples.String(), eth.String())
	}

	log.Printf("✓ Contracts verified. Pool reserves: APPL=%.0f, ETH=%.0f", toEther(apples), toEther(eth))

	// Get initial spot price
	spotPrice, err := executor.GetSpotPrice(ctx)
	if err != nil {
		log.Printf("Warning: Could not get spot price: %v", err)
	} else {
		log.Printf("✓ Initial spot price: %.6f ETH per APPL", toEther(spotPrice))
	}

	// Set up nicknames using new config system
	for _, acc := range config.Accounts {
		executor.SetNickname(acc.Address(), acc.Nickname)
	}

	orchestrator := engine.NewOrchestrator()
	session := engine.NewSession(orchestrator)
	memStore := store.NewMemoryStore()

	// Create price provider for strategy bots
	priceProvider := metrics.NewPriceProvider(memStore)

	// Create bots using config-driven approach
	createBots(executor, orchestrator, priceProvider, memStore)

	// Initialize account metrics with actual on-chain balances
	initializeAccountMetrics(ctx, executor, memStore)

	// Reset User account balances to initial state on startup
	// This ensures a fresh start every time services restart
	userAccount := config.GetAccountByIndex(config.UserAccountIndex)
	if userAccount != nil {
		log.Println("Resetting User account balances to initial state (1,000 ETH and 1,000 APPL)...")
		if err := executor.ResetUserAccount(ctx, userAccount.Address()); err != nil {
			log.Printf("Warning: Failed to reset User account balances on startup: %v", err)
			log.Println("  User account may have different balances from previous session")
		} else {
			log.Println("✓ User account balances reset to 1,000 ETH and 1,000 APPL")
		}
	}

	// Initialize LP metrics with initial pool state and fees
	initLPMetrics(ctx, executor, memStore)

	// Create server
	srv := server.NewServer(session, memStore, executor)

	// Set up trade callback to record trades and broadcast
	executor.OnTrade(func(trade *engine.Trade) {
		memStore.RecordTrade(*trade)
		log.Printf("[Trade] %s %s %.4f",
			trade.Nickname,
			directionStr(trade.IsBuy),
			toEther(trade.AmountIn))

		// Record trade flow event for EWMA mean reversion (if price/reserves data available)
		if trade.PriceBefore != nil && trade.PriceAfter != nil &&
			trade.ReservesBeforeETH != nil && trade.ReservesBeforeAPPL != nil {
			priceBefore := toEtherFloat(trade.PriceBefore)
			priceAfter := toEtherFloat(trade.PriceAfter)
			reservesBeforeETH := toEtherFloat(trade.ReservesBeforeETH)
			reservesBeforeAPPL := toEtherFloat(trade.ReservesBeforeAPPL)

			// Emit trade flow event (will notify MeanRev bots)
			memStore.RecordTradeFlow(priceBefore, priceAfter, trade, reservesBeforeETH, reservesBeforeAPPL)
		}

		// Broadcast trade to WebSocket clients
		srv.BroadcastTrade(trade)

		// Record key events for large trades
		amountInETH := toEther(trade.AmountIn)
		if amountInETH >= config.LargeTradeThresholdETH {
			severity := "warning"
			if amountInETH >= config.CriticalTradeThresholdETH {
				severity = "critical"
			}
			event := store.KeyEvent{
				Timestamp:   time.Now(),
				Type:        "trade",
				Description: fmt.Sprintf("%s executed %s of %.2f ETH", trade.Nickname, directionStr(trade.IsBuy), amountInETH),
				Severity:    severity,
			}
			memStore.RecordEvent(event.Type, event.Description, event.Severity)
			srv.BroadcastEvent(event)
		}

		// Broadcast account update
		srv.BroadcastAccountUpdate(trade.Nickname)
	})

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create main context and errgroup for coordinating services
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	eg, egCtx := errgroup.WithContext(mainCtx)

	// Start price polling
	eg.Go(func() error {
		return pollPricesWithError(egCtx, executor, memStore, srv)
	})

	// Start HTTP server
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	eg.Go(func() error {
		log.Printf("Starting server on %s", addr)
		err := srv.Start(addr)

		// Ignore ErrServerClosed as it's expected during graceful shutdown
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	// Add shutdown handler to errgroup - triggers when context is cancelled
	eg.Go(func() error {
		<-egCtx.Done()
		log.Println("\n=== Shutting down gracefully ===")
		
		// Stop active sessions
		log.Println("Stopping active sessions...")
		if session.IsRunning() {
			if stopErr := session.Stop(); stopErr != nil {
				log.Printf("Error stopping session: %v", stopErr)
			}
			// Give bots a moment to finish
			time.Sleep(500 * time.Millisecond)
		}

		// Stop server with timeout
		log.Println("Stopping HTTP server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if stopErr := srv.Stop(shutdownCtx); stopErr != nil {
			log.Printf("Error shutting down server: %v", stopErr)
			return fmt.Errorf("server shutdown error: %w", stopErr)
		}
		log.Println("✓ Server stopped gracefully")
		return nil
	})

	log.Println("Server running. Press Ctrl+C to stop gracefully...")

	// Wait for shutdown signal or service error
	go func() {
		<-sigCh
		mainCancel() // Cancel context to stop all services
	}()

	// Wait for errgroup (will return first error or nil if context cancelled)
	serviceErr := eg.Wait()

	log.Println("=== Shutdown complete ===")

	// If there was an error from services, exit with error
	if serviceErr != nil && !errors.Is(serviceErr, context.Canceled) {
		log.Fatalf("Service error: %v", serviceErr)
	}

	// Exit cleanly
	os.Exit(0)
}

// createBots iterates over config.Accounts and creates a bot goroutine for each
// trading account. Each bot is wired to the Executor (for on-chain contract calls)
// and the MemoryStore (for metrics). MeanRev bots also receive a PriceProvider
// to access TWAP/volatility from the store without an import cycle.
// See config/accounts.go for the full account roster and trading parameters.
func createBots(executor *engine.Executor, orchestrator *engine.Orchestrator, priceProvider metrics.PriceProvider, store *store.MemoryStore) {
	for _, acc := range config.Accounts {
		// Skip user account (not a bot)
		if acc.Nickname == "User" {
			continue
		}

		var bot bots.Bot
		var err error

		switch acc.Type {
		case config.BotTypeLP:
			// LP doesn't trade, skip
			continue

		case config.BotTypeWhale:
			bot, err = bots.NewWhaleBot(&acc, executor, store)

		case config.BotTypeRetail:
			bot, err = bots.NewRetailBot(&acc, executor, store)

		case config.BotTypeMeanRev:
			bot, err = bots.NewMeanRevBot(&acc, executor, priceProvider, store)

		default:
			log.Printf("Unknown bot type for %s: %s", acc.Nickname, acc.Type)
			continue
		}

		if err != nil {
			log.Printf("Failed to create bot %s: %v", acc.Nickname, err)
			continue
		}

		orchestrator.AddBot(bot)
	}

	log.Printf("Created %d bots", orchestrator.BotCount())
}

// initializeAccountMetrics initializes metrics for all accounts using actual on-chain balances
func initializeAccountMetrics(ctx context.Context, executor *engine.Executor, memStore *store.MemoryStore) {
	// Get current spot price for mark-to-market valuation
	spotPrice, err := executor.GetSpotPrice(ctx)
	if err != nil {
		log.Printf("Warning: Could not get spot price for account initialization: %v", err)
		log.Printf("Falling back to config InitialAccountEquityETH for all accounts")
		spotPrice = nil
	}

	for _, acc := range config.Accounts {
		var initialEquity float64

		if spotPrice != nil && spotPrice.Sign() > 0 {
			// Read actual on-chain balances
			ethBalance, err := executor.GetETHBalance(ctx, acc.Address())
			if err != nil {
				log.Printf("Warning: Could not get ETH balance for %s: %v, using config default", acc.Nickname, err)
				initialEquity = config.InitialAccountEquityETH
			} else {
				appleBalance, err := executor.GetAPPLBalance(ctx, acc.Address())
				if err != nil {
					log.Printf("Warning: Could not get APPL balance for %s: %v, using config default", acc.Nickname, err)
					initialEquity = config.InitialAccountEquityETH
				} else {
					// Calculate initial equity: ethBalance + (appleBalance * spotPrice)
					// Convert from wei to ETH/APPL
					ethFloat := new(big.Float).SetInt(ethBalance)
					ethFloat.Quo(ethFloat, big.NewFloat(1e18))
					eth, _ := ethFloat.Float64()

					appleFloat := new(big.Float).SetInt(appleBalance)
					appleFloat.Quo(appleFloat, big.NewFloat(1e18))
					apple, _ := appleFloat.Float64()

					// Spot price is ETH per APPL (scaled by 1e18)
					spotPriceFloat := new(big.Float).SetInt(spotPrice)
					spotPriceFloat.Quo(spotPriceFloat, big.NewFloat(1e18))
					price, _ := spotPriceFloat.Float64()

					initialEquity = eth + (apple * price)
				}
			}
		} else {
			// Fallback to config if we can't get spot price
			initialEquity = config.InitialAccountEquityETH
		}

		memStore.GetOrCreateAccountMetrics(acc.Nickname, acc.Address(), initialEquity)
	}
}

// initLPMetrics initializes LP metrics with pool state
func initLPMetrics(ctx context.Context, executor *engine.Executor, memStore *store.MemoryStore) {
	// Initialize LP metrics from the actual on-chain pool reserves so the UI
	// always reflects the deployed state (Deploy.s.sol seeds 10,000 APPL + 10,000 ETH).

	// Get initial fees from contract (should be 0 at start, but track for accuracy)
	initialFeesApple, initialFeesETH, err := executor.GetTotalFees(ctx)
	if err != nil {
		log.Printf("Warning: Could not get initial fees: %v", err)
		initialFeesApple = big.NewInt(0)
		initialFeesETH = big.NewInt(0)
	}

	apples, eth, err := executor.GetReserves(ctx)
	if err != nil {
		log.Printf("Warning: Could not read initial reserves, falling back to config: %v", err)
		apples = config.PoolApples
		eth = config.PoolETH
	}

	memStore.GetLPMetrics().SetInitialState(apples, eth)
	memStore.GetLPMetrics().SetInitialFees(initialFeesApple, initialFeesETH)
	memStore.GetLPMetrics().UpdateState(apples, eth, initialFeesApple, initialFeesETH)
	memStore.GetImpactCurve().UpdateReserves(apples, eth)
}

// pollPricesWithError is the main market data loop. Every 2 seconds it:
//  1. Reads spot price from the on-chain AMM contract (executor.GetSpotPrice)
//  2. Records it in MemoryStore (updates candles, TWAP, volatility)
//  3. Reads reserves and fees from the contract
//  4. Updates LP metrics (impermanent loss, fees earned, net PnL)
//  5. Broadcasts price and LP data via WebSocket to all connected frontend clients
//
// This is the primary path that feeds real-time data to the frontend charts
// (PriceChart, LPStats, ImpactCurve components).
func pollPricesWithError(ctx context.Context, executor *engine.Executor, memStore *store.MemoryStore, srv *server.Server) error {
	pollPrices(ctx, executor, memStore, srv)
	// Return nil on context cancellation (expected shutdown)
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}

func pollPrices(ctx context.Context, executor *engine.Executor, memStore *store.MemoryStore, srv *server.Server) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastPrice float64
	var priceInitialized bool

	for {
		select {
		case <-ctx.Done():
			log.Println("Price polling stopped")
			return
		case <-ticker.C:
		}
		// Use the parent context for RPC calls so cancellation propagates
		pollCtx := ctx

		// Get current price
		price, err := executor.GetSpotPrice(pollCtx)
		if err != nil {
			log.Printf("Error getting spot price: %v", err)
			continue
		}

		// Convert from wei to float
		priceFloat := toEtherFloat(price)
		memStore.RecordPrice(priceFloat)

		// Track price movements for key events
		if !priceInitialized {
			lastPrice = priceFloat
			priceInitialized = true
		} else if lastPrice > 0 {
			priceChange := (priceFloat - lastPrice) / lastPrice
			if math.Abs(priceChange) >= 0.05 { // 5% price move
				severity := "info"
				if math.Abs(priceChange) >= 0.10 {
					severity = "warning"
				}
				if math.Abs(priceChange) >= 0.20 {
					severity = "critical"
				}
				direction := "up"
				if priceChange < 0 {
					direction = "down"
				}
				event := store.KeyEvent{
					Timestamp:   time.Now(),
					Type:        "strategy_trigger",
					Description: fmt.Sprintf("Price moved %s %.2f%% (%.6f → %.6f ETH)", direction, math.Abs(priceChange)*100, lastPrice, priceFloat),
					Severity:    severity,
				}
				memStore.RecordEvent(event.Type, event.Description, event.Severity)
				srv.BroadcastEvent(event)
				lastPrice = priceFloat
			}
		}

		// Get reserves
		apples, eth, err := executor.GetReserves(pollCtx)
		if err != nil {
			continue
		}

		// Get total fees collected
		feesApple, feesETH, err := executor.GetTotalFees(pollCtx)
		if err != nil {
			log.Printf("Error getting total fees: %v", err)
			// Continue with zero fees if we can't read them
			feesApple = big.NewInt(0)
			feesETH = big.NewInt(0)
		}

		// Update LP metrics with actual fees
		memStore.GetLPMetrics().UpdateState(apples, eth, feesApple, feesETH)
		memStore.GetImpactCurve().UpdateReserves(apples, eth)

		// Broadcast updates
		srv.BroadcastLPMetrics()

		candles := memStore.GetCandles()
		if len(candles) > 0 {
			srv.BroadcastPrice(candles[len(candles)-1])
		}
	}
}

func toEther(wei *big.Int) float64 {
	if wei == nil {
		return 0
	}
	f := new(big.Float).SetInt(wei)
	f.Quo(f, big.NewFloat(1e18))
	result, _ := f.Float64()
	return result
}

func toEtherFloat(wei *big.Int) float64 {
	return toEther(wei)
}

func directionStr(isBuy bool) string {
	if isBuy {
		return "BUY"
	}
	return "SELL"
}
