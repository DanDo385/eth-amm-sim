// handlers.go - HTTP handler implementations for all REST endpoints.
//
// Each handler reads from MemoryStore or controls Session/Executor and returns
// JSON. Frontend lib/api.ts calls these endpoints. User trade handlers
// (handleUserBuy/Sell) call executor methods directly to submit on-chain
// transactions for the User account (index 1).
//
// CONNECTIONS:
//  - Frontend API client: frontend/lib/api.ts (fetches from these endpoints)
//  - Session control: engine/session.go (Start/Stop/Reset/GetState)
//  - Trade execution: engine/executor.go (user buy/sell → on-chain tx)
//  - Data source: store/memory.go (candles, trades, LP metrics, events, accounts)
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"eth-amm-sim/internal/config"
	"eth-amm-sim/internal/store"

	"github.com/ethereum/go-ethereum/common"
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
		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&body); err != nil {
			if err != io.EOF {
				respondError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
				return
			}
		}
	}

	if body.Duration > 0 {
		s.session.SetDuration(time.Duration(body.Duration) * time.Second)
	}

	// Reset finalization guard for the new session
	atomic.StoreInt32(&s.sessionFinalized, 0)

	// Use request context with timeout for RPC calls only
	rpcCtx, rpcCancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer rpcCancel()

	// Re-initialize LP metrics with current pool state when starting a new session
	// This ensures each session starts with a fresh baseline
	s.reinitializeLPMetrics(rpcCtx)

	// Get current spot price for session reset
	spotPrice, err := s.executor.GetSpotPrice(rpcCtx)
	if err != nil {
		log.Printf("Warning: Could not get spot price for session start: %v", err)
		spotPrice = big.NewInt(0)
	}

	// Convert spot price to float64
	var spotPriceFloat float64
	if spotPrice != nil && spotPrice.Sign() > 0 {
		spotPriceBigFloat := new(big.Float).SetInt(spotPrice)
		spotPriceBigFloat.Quo(spotPriceBigFloat, big.NewFloat(1e18))
		spotPriceFloat, _ = spotPriceBigFloat.Float64()
	}

	// Reset all accounts for session-only performance tracking
	// Create a function to get balances for each account
	getBalance := func(nickname string) (ethBalance, applBalance float64, err error) {
		// Find account by nickname
		acc := config.GetAccount(nickname)
		if acc == nil {
			return 0, 0, fmt.Errorf("account not found: %s", nickname)
		}

		// Get balances from executor (use RPC context with timeout)
		ethBal, err := s.executor.GetETHBalance(rpcCtx, acc.Address())
		if err != nil {
			return 0, 0, err
		}

		applBal, err := s.executor.GetAPPLBalance(rpcCtx, acc.Address())
		if err != nil {
			return 0, 0, err
		}

		// Convert from wei to ETH/APPL
		ethFloat := new(big.Float).SetInt(ethBal)
		ethFloat.Quo(ethFloat, big.NewFloat(1e18))
		ethBalance, _ = ethFloat.Float64()

		applFloat := new(big.Float).SetInt(applBal)
		applFloat.Quo(applFloat, big.NewFloat(1e18))
		applBalance, _ = applFloat.Float64()

		return ethBalance, applBalance, nil
	}

	s.store.ResetAccountsForSession(getBalance, spotPriceFloat)

	// Start session with background context (not request context)
	// The session manages its own timeout based on duration
	if err := s.session.Start(context.Background()); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Record session start event
	if s.session.IsRunning() {
		state := s.session.GetState()
		event := store.KeyEvent{
			Timestamp:   time.Now(),
			Type:        "strategy_trigger",
			Description: fmt.Sprintf("Session started (duration: %d seconds)", state.Duration),
			Severity:    "info",
		}
		s.store.RecordEvent(event.Type, event.Description, event.Severity)
		s.BroadcastEvent(event)
	}

	// Broadcast updated LP metrics
	s.BroadcastLPMetrics()

	respondJSON(w, map[string]string{"status": "started"})
}

func (s *Server) handleSessionStop(w http.ResponseWriter, r *http.Request) {
	if err := s.session.Stop(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Finalization runs in Session.run after bots stop (see SetOnSessionEnded).

	respondJSON(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleSessionPause(w http.ResponseWriter, r *http.Request) {
	if err := s.session.Pause(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, map[string]string{"status": "paused"})
}

func (s *Server) handleSessionResume(w http.ResponseWriter, r *http.Request) {
	// Before resuming, normalize all trading account positions back to their
	// configured starting balances so resumed flow is deterministic.
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := s.executor.ResetTradingAccounts(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to reset trading positions: "+err.Error())
		return
	}

	spotPrice, err := s.executor.GetSpotPrice(ctx)
	if err != nil {
		log.Printf("Warning: Could not get spot price during resume: %v", err)
		spotPrice = big.NewInt(0)
	}
	var spotPriceFloat float64
	if spotPrice != nil && spotPrice.Sign() > 0 {
		spotPriceBigFloat := new(big.Float).SetInt(spotPrice)
		spotPriceBigFloat.Quo(spotPriceBigFloat, big.NewFloat(1e18))
		spotPriceFloat, _ = spotPriceBigFloat.Float64()
	}
	s.store.ResetAccountsForSession(s.buildBalanceLookup(ctx), spotPriceFloat)

	s.Broadcast(WSMessage{
		Type: "user_balance_reset",
		Data: map[string]interface{}{"reset": true},
	})

	if err := s.session.Resume(context.Background()); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if s.session.IsRunning() {
		event := store.KeyEvent{
			Timestamp:   time.Now(),
			Type:        "strategy_trigger",
			Description: "Session resumed with trading positions reset",
			Severity:    "info",
		}
		s.store.RecordEvent(event.Type, event.Description, event.Severity)
		s.BroadcastEvent(event)
	}
	s.BroadcastAllAccountUpdates()

	respondJSON(w, map[string]string{"status": "resumed"})
}

func (s *Server) handleSessionReset(w http.ResponseWriter, r *http.Request) {
	// Reset mode:
	// - soft: clear session/store data only
	// - hard: also reset account metrics + user balance
	// - reseed: hard reset + anvil_reset + redeploy to recover initial 1.0 pool
	mode, hardReset, reseed, err := parseResetMode(r.URL.Query())
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.session.Reset(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Reset store (clears trades, events, price metrics, LP metrics)
	s.store.Reset()

	// If hard reset, also clear account metrics and reset LP metrics to zero
	if hardReset {
		// Get account metrics manager and reset all accounts back to initial state
		accountMgr := s.store.GetAccountMetricsManager()
		if accountMgr != nil {
			accountMgr.Reset()
			log.Println("Hard reset: Account metrics reset to initial state")
		}

		// Reset LP metrics to zero (don't re-initialize with current pool state)
		s.store.GetLPMetrics().Reset()
		log.Println("Hard reset: LP metrics reset to zero")

		userAccount := config.GetAccountByIndex(config.UserAccountIndex)
		if userAccount != nil {
			normalizeCtx, normalizeCancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer normalizeCancel()
			if err := s.normalizeUserAccount(normalizeCtx, userAccount.Address()); err != nil {
				log.Printf("Warning: Failed to normalize User account balances: %v", err)
			}
		}

		if reseed {
			reseedCtx, reseedCancel := context.WithTimeout(r.Context(), 3*time.Minute)
			defer reseedCancel()
			if err := s.reseedChainAndRedeploy(reseedCtx); err != nil {
				respondError(w, http.StatusInternalServerError, "failed to reseed chain: "+err.Error())
				return
			}
			if userAccount != nil {
				postReseedCtx, postReseedCancel := context.WithTimeout(context.Background(), 45*time.Second)
				defer postReseedCancel()
				if err := s.normalizeUserAccount(postReseedCtx, userAccount.Address()); err != nil {
					log.Printf("Warning: Failed to normalize User account balances after reseed: %v", err)
				}
			}
			// After reseed/redeploy, establish a fresh LP baseline and reset account metrics
			// to on-chain starting balances at the current spot price.
			s.reinitializeLPMetrics(reseedCtx)
			spotPrice, err := s.executor.GetSpotPrice(reseedCtx)
			if err != nil {
				log.Printf("Warning: Could not get spot price after reseed: %v", err)
				spotPrice = big.NewInt(0)
			}
			var spotPriceFloat float64
			if spotPrice != nil && spotPrice.Sign() > 0 {
				spf := new(big.Float).SetInt(spotPrice)
				spf.Quo(spf, big.NewFloat(1e18))
				spotPriceFloat, _ = spf.Float64()
			}
			s.store.ResetAccountsForSession(s.buildBalanceLookup(reseedCtx), spotPriceFloat)
			log.Println("Hard reset reseed: chain reset + redeploy complete")
		}
	} else {
		// Soft reset: Re-initialize LP metrics with current pool state
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		s.reinitializeLPMetrics(ctx)
	}

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
		"mode":      mode,
		"hardReset": hardReset,
		"reseeded":  reseed,
	})
}

func parseResetMode(q map[string][]string) (mode string, hardReset bool, reseed bool, err error) {
	mode = strings.ToLower(strings.TrimSpace(firstQuery(q, "mode")))
	switch mode {
	case "", "soft":
		mode = "soft"
	case "hard":
		mode = "hard"
		hardReset = true
	case "reseed":
		mode = "reseed"
		hardReset = true
		reseed = true
	default:
		return "", false, false, fmt.Errorf("invalid reset mode: use soft, hard, or reseed")
	}

	// Backward-compatible query params:
	// /session/reset?hard=true   -> hard
	// /session/reset?reseed=true -> reseed
	if mode == "soft" {
		if firstQuery(q, "hard") == "true" {
			mode = "hard"
			hardReset = true
		}
		if firstQuery(q, "reseed") == "true" {
			mode = "reseed"
			hardReset = true
			reseed = true
		}
	}
	return mode, hardReset, reseed, nil
}

func firstQuery(q map[string][]string, key string) string {
	values := q[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (s *Server) handleSessionState(w http.ResponseWriter, r *http.Request) {
	state := s.session.GetState()
	respondJSON(w, state)
}

func (s *Server) buildBalanceLookup(ctx context.Context) func(nickname string) (ethBalance, applBalance float64, err error) {
	return func(nickname string) (ethBalance, applBalance float64, err error) {
		acc := config.GetAccount(nickname)
		if acc == nil {
			return 0, 0, fmt.Errorf("account not found: %s", nickname)
		}

		ethBal, err := s.executor.GetETHBalance(ctx, acc.Address())
		if err != nil {
			return 0, 0, err
		}
		applBal, err := s.executor.GetAPPLBalance(ctx, acc.Address())
		if err != nil {
			return 0, 0, err
		}

		ethFloat := new(big.Float).SetInt(ethBal)
		ethFloat.Quo(ethFloat, big.NewFloat(1e18))
		ethBalance, _ = ethFloat.Float64()

		applFloat := new(big.Float).SetInt(applBal)
		applFloat.Quo(applFloat, big.NewFloat(1e18))
		applBalance, _ = applFloat.Float64()

		return ethBalance, applBalance, nil
	}
}

func (s *Server) normalizeUserAccount(ctx context.Context, userAddress common.Address) error {
	expectedETH := new(big.Int).Mul(big.NewInt(config.UserStartingETH), big.NewInt(1e18))
	expectedAPPL := new(big.Int).Mul(big.NewInt(config.UserStartingAPPL), big.NewInt(1e18))
	resetAndVerify := func() error {
		if err := s.executor.ResetUserAccount(ctx, userAddress); err != nil {
			return fmt.Errorf("reset user account failed: %w", err)
		}
		ethAfter, err := s.executor.GetETHBalance(ctx, userAddress)
		if err != nil {
			return fmt.Errorf("read user ETH balance failed: %w", err)
		}
		applAfter, err := s.executor.GetAPPLBalance(ctx, userAddress)
		if err != nil {
			return fmt.Errorf("read user APPL balance failed: %w", err)
		}
		if ethAfter.Cmp(expectedETH) != 0 || applAfter.Cmp(expectedAPPL) != 0 {
			return fmt.Errorf(
				"user balance mismatch after normalization (ETH=%s, APPL=%s, expectedETH=%s, expectedAPPL=%s)",
				ethAfter.String(),
				applAfter.String(),
				expectedETH.String(),
				expectedAPPL.String(),
			)
		}
		return nil
	}

	if err := resetAndVerify(); err != nil {
		log.Printf("Warning: First user normalization attempt failed: %v", err)
		if retryErr := resetAndVerify(); retryErr != nil {
			return retryErr
		}
	}

	log.Printf("User account balances normalized to %d ETH and %d APPL", config.UserStartingETH, config.UserStartingAPPL)
	s.Broadcast(WSMessage{
		Type: "user_balance_reset",
		Data: map[string]interface{}{"reset": true},
	})
	return nil
}

func (s *Server) reseedChainAndRedeploy(ctx context.Context) error {
	if err := s.executor.ResetChain(ctx); err != nil {
		return fmt.Errorf("anvil reset failed: %w", err)
	}
	projectRoot, err := findProjectRootForScripts()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "bash", "scripts/deploy.sh")
	cmd.Dir = projectRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deploy script failed: %w; output: %s", err, strings.TrimSpace(string(out)))
	}

	// Clear cached nonces after chain reset/redeploy to avoid stale nonce usage.
	if err := s.executor.ResetNonceCache(ctx); err != nil {
		return fmt.Errorf("nonce cache reset failed: %w", err)
	}
	return nil
}

func findProjectRootForScripts() (string, error) {
	startDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not determine working directory: %w", err)
	}

	dir := startDir
	for {
		if hasDir(filepath.Join(dir, "contracts")) && hasDir(filepath.Join(dir, "backend")) && hasDir(filepath.Join(dir, "scripts")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find project root for deploy script")
}

func hasDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Account handlers

func (s *Server) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	// Ensure the User account exists in account metrics so it shows up in the UI
	// even before the user has executed any trades.
	if acc := config.GetAccount("User"); acc != nil {
		s.store.GetOrCreateAccountMetrics(acc.Nickname, acc.Address(), config.InitialAccountEquityETH)
	}

	accounts := s.store.GetAllAccountPerformance()
	respondJSON(w, accounts)
}

func (s *Server) handleGetAccountPerformance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	perf := s.store.GetAccountPerformance(nickname)
	if perf == nil && nickname == "User" {
		// Lazily create the User account metrics if missing (e.g., after a hard reset).
		if acc := config.GetAccount("User"); acc != nil {
			s.store.GetOrCreateAccountMetrics(acc.Nickname, acc.Address(), config.InitialAccountEquityETH)
			perf = s.store.GetAccountPerformance(nickname)
		}
	}
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
	// Use request context with timeout for external RPC calls
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

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

	// Get reserves before trade to calculate amount out
	appleReserveBefore, ethReserveBefore, err := s.executor.GetReserves(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get reserves: %v", err))
		return
	}

	// Calculate fee and amount out (same formula as executor)
	fee := new(big.Int).Div(new(big.Int).Mul(ethAmount, big.NewInt(config.AMMFeeNumerator)), big.NewInt(config.AMMFeeDenominator))
	amountInAfterFee := new(big.Int).Sub(ethAmount, fee)

	// Calculate amount out: (amountInAfterFee * appleReserve) / (ethReserve + amountInAfterFee)
	numerator := new(big.Int).Mul(amountInAfterFee, appleReserveBefore)
	denominator := new(big.Int).Add(ethReserveBefore, amountInAfterFee)
	appleAmountOut := new(big.Int).Div(numerator, denominator)

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
		"txHash":       txHash,
		"ethAmount":    ethAmount.String(),
		"appleAmount":  appleAmountOut.String(),
		"ethBalance":   ethBalanceAfter.String(),
		"appleBalance": appleBalance.String(),
		"status":       "success",
	})
}

func (s *Server) handleTradeSell(w http.ResponseWriter, r *http.Request) {
	// Use request context with timeout for external RPC calls
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

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

	// Get reserves before trade to calculate amount out
	appleReserveBefore, ethReserveBefore, err := s.executor.GetReserves(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get reserves: %v", err))
		return
	}

	// Calculate fee and amount out (same formula as executor)
	fee := new(big.Int).Div(new(big.Int).Mul(appleAmount, big.NewInt(config.AMMFeeNumerator)), big.NewInt(config.AMMFeeDenominator))
	amountInAfterFee := new(big.Int).Sub(appleAmount, fee)

	// Calculate amount out: (amountInAfterFee * ethReserve) / (appleReserve + amountInAfterFee)
	numerator := new(big.Int).Mul(amountInAfterFee, ethReserveBefore)
	denominator := new(big.Int).Add(appleReserveBefore, amountInAfterFee)
	ethAmountOut := new(big.Int).Div(numerator, denominator)

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
		"txHash":       txHash,
		"ethAmount":    ethAmountOut.String(),
		"appleAmount":  appleAmount.String(),
		"ethBalance":   ethBalanceAfter.String(),
		"appleBalance": appleBalanceAfter.String(),
		"status":       "success",
	})
}

func (s *Server) handleGetUserBalance(w http.ResponseWriter, r *http.Request) {
	// Use request context with timeout for external RPC calls
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

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

	// Self-heal known reseed leak signature:
	// Anvil defaults every account to 30,000 ETH, while User target is configured in config.
	// If a reseed flow returns early, UI can show 30,000 ETH + User baseline APPL.
	// Only normalize this exact mismatch while session is not running.
	if !s.session.IsRunning() && isAnvilDefaultUserLeak(ethBalance, appleBalance) {
		normalizeCtx, normalizeCancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer normalizeCancel()
		if err := s.normalizeUserAccount(normalizeCtx, userAddr); err != nil {
			log.Printf("Warning: Auto-normalization from leaked default balance failed: %v", err)
		} else {
			ethBalance, _ = s.executor.GetETHBalance(normalizeCtx, userAddr)
			appleBalance, _ = s.executor.GetAPPLBalance(normalizeCtx, userAddr)
		}
	}

	respondJSON(w, map[string]interface{}{
		"ethBalance":   ethBalance.String(),
		"appleBalance": appleBalance.String(),
	})
}

func isAnvilDefaultUserLeak(ethBalance, appleBalance *big.Int) bool {
	if ethBalance == nil || appleBalance == nil {
		return false
	}
	userAPPLTarget := new(big.Int).Mul(big.NewInt(config.UserStartingAPPL), big.NewInt(1e18))
	thirtyThousand := new(big.Int).Mul(big.NewInt(30000), big.NewInt(1e18))
	return ethBalance.Cmp(thirtyThousand) == 0 && appleBalance.Cmp(userAPPLTarget) == 0
}
