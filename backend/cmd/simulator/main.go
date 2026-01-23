// Package main is the entry point for the AMM simulation backend
package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eth-amm-sim/internal/bots"
	"eth-amm-sim/internal/chain"
	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/server"
	"eth-amm-sim/internal/store"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	log.Println("=== ETH-AMM-SIM Backend Starting ===")
	
	// Load configuration
	cfg := config.DefaultConfig()
	
	// Get contract addresses from environment or use defaults
	tokenAddr := os.Getenv("TOKEN_ADDRESS")
	ammAddr := os.Getenv("AMM_ADDRESS")
	
	if tokenAddr == "" || ammAddr == "" {
		log.Println("Warning: TOKEN_ADDRESS and AMM_ADDRESS not set")
		log.Println("Using placeholder addresses - deploy contracts first!")
		// These will be replaced after deployment
		tokenAddr = "0x5FbDB2315678afecb367f032d93F642f64180aa3"
		ammAddr = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"
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
	executor := engine.NewExecutor(client, nonceManager, cfg.AMMAddress, cfg.TokenAddress)
	orchestrator := engine.NewOrchestrator()
	session := engine.NewSession(orchestrator)
	memStore := store.NewMemoryStore()
	
	// Set up nicknames using new config system
	for _, acc := range config.Accounts {
		executor.SetNickname(acc.Address(), acc.Nickname)
	}
	
	// Create bots using config-driven approach
	createBots(executor, orchestrator)
	
	// Initialize account metrics
	initializeAccountMetrics(memStore)
	
	// Set up trade callback to record trades
	executor.OnTrade(func(trade *engine.Trade) {
		memStore.RecordTrade(*trade)
		log.Printf("[Trade] %s %s %.4f", 
			trade.Nickname, 
			directionStr(trade.IsBuy), 
			toEther(trade.AmountIn))
	})
	
	// Initialize LP metrics with initial pool state
	initLPMetrics(memStore)
	
	// Create and start server
	srv := server.NewServer(session, memStore, executor)
	
	// Start price polling (for demo)
	go pollPrices(client, executor, memStore, srv)
	
	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if session.IsRunning() {
			session.Stop()
		}
		srv.Stop(ctx)
		os.Exit(0)
	}()
	
	// Start HTTP server
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	
	log.Printf("Starting server on %s", addr)
	if err := srv.Start(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// createBots creates all trading bots based on config
func createBots(executor *engine.Executor, orchestrator *engine.Orchestrator) {
	for _, acc := range config.Accounts {
		var bot bots.Bot

		switch acc.Type {
		case config.BotTypeLP:
			// LP doesn't trade, skip
			continue

		case config.BotTypeWhale:
			bot = bots.NewWhaleBot(&acc, executor)

		case config.BotTypeRetail:
			bot = bots.NewRetailBot(&acc, executor)

		case config.BotTypeMeanRev:
			// TODO: Implement MeanRev bot
			log.Printf("MeanRev bot not yet implemented: %s", acc.Nickname)
			continue

		case config.BotTypeMomentum:
			// TODO: Implement Momentum bot
			log.Printf("Momentum bot not yet implemented: %s", acc.Nickname)
			continue

		case config.BotTypeLeverage:
			// Phase 2
			continue

		case config.BotTypeLiquidator:
			// Phase 2
			continue

		default:
			log.Printf("Unknown bot type for %s: %s", acc.Nickname, acc.Type)
			continue
		}

		orchestrator.AddBot(bot)
	}

	log.Printf("Created %d bots", orchestrator.BotCount())
}

// initializeAccountMetrics initializes metrics for all accounts
func initializeAccountMetrics(memStore *store.MemoryStore) {
	for _, acc := range config.Accounts {
		// Initial equity = starting ETH (10000 from Anvil default)
		memStore.GetOrCreateAccountMetrics(acc.Nickname, acc.Address(), 10000)
	}
}

// initLPMetrics initializes LP metrics with pool state
func initLPMetrics(memStore *store.MemoryStore) {
	// Initial pool: 1000 APPL + 1000 ETH (from config.PoolApples and config.PoolETH)
	ether := func(n int64) *big.Int {
		return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18))
	}
	
	memStore.GetLPMetrics().SetInitialState(ether(1000), ether(1000))
	memStore.GetImpactCurve().UpdateReserves(ether(1000), ether(1000))
}

// pollPrices periodically fetches price and updates metrics
func pollPrices(client *chain.Client, executor *engine.Executor, memStore *store.MemoryStore, srv *server.Server) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		ctx := context.Background()
		
		// Get current price
		price, err := executor.GetSpotPrice(ctx)
if err != nil {
			log.Printf("Error getting spot price: %v", err)
			continue
		}
		
		// Convert from wei to float
		priceFloat := toEtherFloat(price)
		memStore.RecordPrice(priceFloat)
		
		// Get reserves
		apples, eth, err := executor.GetReserves(ctx)
		if err != nil {
			continue
		}
		
		// Update LP metrics
		memStore.GetLPMetrics().UpdateState(apples, eth, big.NewInt(0), big.NewInt(0))
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
