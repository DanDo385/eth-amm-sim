// handlers.go — HTTP handler implementations for all REST endpoints.
//
// Each handler reads from MemoryStore or controls Session/Executor and returns
// JSON. Frontend lib/api.ts calls these endpoints. User trade handlers
// (handleUserBuy/Sell) call executor methods directly to submit on-chain
// transactions for the User account (index 29).
//
// CONNECTIONS:
//   - Frontend API client: frontend/lib/api.ts (fetches from these endpoints)
//   - Session control: engine/session.go (Start/Stop/Reset/GetState)
//   - Trade execution: engine/executor.go (user buy/sell → on-chain tx)
//   - Data source: store/memory.go (candles, trades, LP metrics, events, accounts)
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/store"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
)

// Session handlers

func (s *Server) handleSessionStart(w http.ResponseWriter, r *http.Request) {
	// Parse optional duration from body
	var body struct {
		Duration int `json:"duration"` // seconds
	}
	
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
	}
	
	if body.Duration > 0 {
		s.session.SetDuration(time.Duration(body.Duration) * time.Second)
	}
	
	ctx := context.Background()
	
	// Re-initialize LP metrics with current pool state when starting a new session
	// This ensures each session starts with a fresh baseline
	s.reinitializeLPMetrics(ctx)
	
	if err := s.session.Start(ctx); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Record session start event
	state := s.session.GetState()
	event := store.KeyEvent{
		Timestamp:   time.Now(),
		Type:        "strategy_trigger",
		Description: fmt.Sprintf("Session started (duration: %d seconds)", state.Duration),
		Severity:    "info",
	}
	s.store.RecordEvent(event.Type, event.Description, event.Severity)
	s.BroadcastEvent(event)
	
	// Broadcast updated LP metrics
	s.BroadcastLPMetrics()
	
	respondJSON(w, map[string]string{"status": "started"})
}

func (s *Server) handleSessionStop(w http.ResponseWriter, r *http.Request) {
	if err := s.session.Stop(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Record session stop event
	event := store.KeyEvent{
		Timestamp:   time.Now(),
		Type:        "strategy_trigger",
		Description: "Session stopped",
		Severity:    "info",
	}
	s.store.RecordEvent(event.Type, event.Description, event.Severity)
	s.BroadcastEvent(event)
	
	respondJSON(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleSessionReset(w http.ResponseWriter, r *http.Request) {
	// Check for hard reset parameter (clears account metrics too)
	hardReset := r.URL.Query().Get("hard") == "true"
	
	if err := s.session.Reset(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Reset store (clears trades, events, price metrics, LP metrics)
	s.store.Reset()
	
	// If hard reset, also clear account metrics
	if hardReset {
		// Get account metrics manager and reset all accounts
		accountMgr := s.store.GetAccountMetricsManager()
		if accountMgr != nil {
			accountMgr.Reset()
			log.Println("Hard reset: Account metrics cleared")
		}
	}
	
	// Re-initialize LP metrics with current pool state
	ctx := context.Background()
	s.reinitializeLPMetrics(ctx)
	
	// Broadcast updated LP metrics
	s.BroadcastLPMetrics()
	
	// Broadcast empty events list to clear Key Events in frontend
	s.Broadcast(WSMessage{
		Type: "events",
		Data: s.store.GetRecentEvents(0), // Empty events array
	})
	
	// Broadcast empty trades to clear frontend
	s.Broadcast(WSMessage{
		Type: "trades",
		Data: []interface{}{}, // Empty trades array
	})
	
	respondJSON(w, map[string]interface{}{
		"status":    "reset",
		"hardReset": hardReset,
	})
}

func (s *Server) handleSessionState(w http.ResponseWriter, r *http.Request) {
	state := s.session.GetState()
	respondJSON(w, state)
}

// Account handlers

func (s *Server) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := s.store.GetAllAccountPerformance()
	respondJSON(w, accounts)
}

func (s *Server) handleGetAccountPerformance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname := vars["nickname"]
	
	perf := s.store.GetAccountPerformance(nickname)
	if perf == nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}
	
	respondJSON(w, perf)
}

// LP handlers

func (s *Server) handleGetLPMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.store.GetLPData()
	respondJSON(w, metrics)
}

// Market data handlers

func (s *Server) handleGetCandles(w http.ResponseWriter, r *http.Request) {
	candles := s.store.GetCandles()
	respondJSON(w, candles)
}

func (s *Server) handleGetTrades(w http.ResponseWriter, r *http.Request) {
	// Optional limit parameter - default to 1000 to capture full sessions
	limitStr := r.URL.Query().Get("limit")
	limit := 1000 // Increased default to capture all trades in a session
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	trades := s.store.GetRecentTrades(limit)
	respondJSON(w, trades)
}

func (s *Server) handleGetImpactCurve(w http.ResponseWriter, r *http.Request) {
	buyCurve := s.store.GetBuyImpact()
	sellCurve := s.store.GetSellImpact()
	
	respondJSON(w, map[string]interface{}{
		"buy":  buyCurve,
		"sell": sellCurve,
	})
}

func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	// Optional limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	events := s.store.GetRecentEvents(limit)
	respondJSON(w, events)
}

// User trading handlers

func (s *Server) handleTradeBuy(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Parse request body
	var req struct {
		EthAmount string `json:"ethAmount"` // Amount in ETH (will be converted to wei)
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	// Get user account
	userAccount := config.GetAccountByIndex(config.UserAccountIndex)
	if userAccount == nil {
		respondError(w, http.StatusInternalServerError, "user account not found")
		return
	}
	
	// Parse ETH amount and convert to wei
	ethAmountFloat, err := strconv.ParseFloat(strings.TrimSpace(req.EthAmount), 64)
	if err != nil || ethAmountFloat <= 0 {
		respondError(w, http.StatusBadRequest, "invalid ETH amount")
		return
	}
	
	// Convert ETH to wei (multiply by 1e18)
	ethAmountWei := new(big.Float).SetFloat64(ethAmountFloat)
	ethAmountWei.Mul(ethAmountWei, big.NewFloat(1e18))
	ethAmount, _ := ethAmountWei.Int(nil)
	
	// Check balance
	userAddr := userAccount.Address()
	ethBalance, err := s.executor.GetETHBalance(ctx, userAddr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get balance: %v", err))
		return
	}
	
	// Reserve ETH for gas (0.1 ETH to ensure sufficient buffer)
	// The transaction sends ethAmount as value, and gas is deducted separately
	// We need: ethAmount (for trade) + gas (for transaction fee)
	gasReserve := new(big.Int).Mul(big.NewInt(1e17), big.NewInt(1)) // 0.1 ETH
	required := new(big.Int).Add(ethAmount, gasReserve)
	
	if ethBalance.Cmp(required) < 0 {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("insufficient ETH balance: have %s wei, need %s wei (trade: %s + gas reserve: %s)", 
			ethBalance.String(), required.String(), ethAmount.String(), gasReserve.String()))
		return
	}
	
	// Load private key
	privateKeyHex := userAccount.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load private key")
		return
	}
	
	// Execute trade
	txHash, err := s.executor.SwapETHForApples(ctx, privateKey, ethAmount)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("trade failed: %v", err))
		return
	}
	
	// Get updated balances
	ethBalanceAfter, _ := s.executor.GetETHBalance(ctx, userAddr)
	appleBalance, _ := s.executor.GetAPPLBalance(ctx, userAddr)
	
	respondJSON(w, map[string]interface{}{
		"txHash":        txHash,
		"ethAmount":     ethAmount.String(),
		"ethBalance":    ethBalanceAfter.String(),
		"appleBalance":  appleBalance.String(),
		"status":        "success",
	})
}

func (s *Server) handleTradeSell(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Parse request body
	var req struct {
		AppleAmount string `json:"appleAmount"` // Amount in APPL (will be converted to wei)
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	// Get user account
	userAccount := config.GetAccountByIndex(config.UserAccountIndex)
	if userAccount == nil {
		respondError(w, http.StatusInternalServerError, "user account not found")
		return
	}
	
	// Parse APPL amount and convert to wei
	appleAmountFloat, err := strconv.ParseFloat(strings.TrimSpace(req.AppleAmount), 64)
	if err != nil || appleAmountFloat <= 0 {
		respondError(w, http.StatusBadRequest, "invalid APPL amount")
		return
	}
	
	// Convert APPL to wei (multiply by 1e18)
	appleAmountWei := new(big.Float).SetFloat64(appleAmountFloat)
	appleAmountWei.Mul(appleAmountWei, big.NewFloat(1e18))
	appleAmount, _ := appleAmountWei.Int(nil)
	
	// Check balance
	userAddr := userAccount.Address()
	appleBalance, err := s.executor.GetAPPLBalance(ctx, userAddr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get balance: %v", err))
		return
	}
	
	if appleBalance.Cmp(appleAmount) < 0 {
		respondError(w, http.StatusBadRequest, "insufficient APPL balance")
		return
	}
	
	// Check ETH for gas
	ethBalance, err := s.executor.GetETHBalance(ctx, userAddr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get ETH balance: %v", err))
		return
	}
	
	gasReserve := new(big.Int).Mul(big.NewInt(1e16), big.NewInt(1)) // 0.01 ETH
	if ethBalance.Cmp(gasReserve) < 0 {
		respondError(w, http.StatusBadRequest, "insufficient ETH for gas")
		return
	}
	
	// Load private key
	privateKeyHex := userAccount.PrivateKey()
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load private key")
		return
	}
	
	// Execute trade
	txHash, err := s.executor.SwapApplesForETH(ctx, privateKey, appleAmount)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("trade failed: %v", err))
		return
	}
	
	// Get updated balances
	ethBalanceAfter, _ := s.executor.GetETHBalance(ctx, userAddr)
	appleBalanceAfter, _ := s.executor.GetAPPLBalance(ctx, userAddr)
	
	respondJSON(w, map[string]interface{}{
		"txHash":        txHash,
		"appleAmount":   appleAmount.String(),
		"ethBalance":    ethBalanceAfter.String(),
		"appleBalance":  appleBalanceAfter.String(),
		"status":        "success",
	})
}

func (s *Server) handleGetUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Get user account
	userAccount := config.GetAccountByIndex(config.UserAccountIndex)
	if userAccount == nil {
		respondError(w, http.StatusInternalServerError, "user account not found")
		return
	}
	
	userAddr := userAccount.Address()
	
	// Get balances
	ethBalance, err := s.executor.GetETHBalance(ctx, userAddr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get ETH balance: %v", err))
		return
	}
	
	appleBalance, err := s.executor.GetAPPLBalance(ctx, userAddr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get APPL balance: %v", err))
		return
	}
	
	respondJSON(w, map[string]interface{}{
		"ethBalance":   ethBalance.String(),
		"appleBalance": appleBalance.String(),
	})
}
