// server.go — HTTP REST API and WebSocket server on :8080.
//
// SYSTEM ROLE:
// This is the backend's interface to the frontend. It exposes REST endpoints
// for session control, metrics queries, and user trading, plus a WebSocket
// endpoint for real-time data streaming. The frontend (Vite SPA on :3000)
// connects here for all data.
//
// ROUTES (see handlers.go for implementations):
//
//	POST /session/{start,pause,resume,stop,reset}  — Control simulation lifecycle
//	GET  /session/state               — Current session status
//	GET  /candles, /trades, /events   — Market data
//	GET  /lp/metrics                  — LP performance
//	GET  /accounts                    — All account metrics
//	POST /trade/{buy,sell}            — User manual trading
//	WS   /stream                      — Real-time WebSocket feed
//
// CONNECTIONS:
//   - Frontend: lib/api.ts makes REST calls; hooks/useWebSocket.ts opens /stream
//   - Backend: reads from store/memory.go, controls engine/session.go
//   - Broadcast: server/broadcast.go pushes trade/price/event data to WS clients
package server

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/store"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Server handles HTTP and WebSocket connections
type Server struct {
	router   *mux.Router
	session  *engine.Session
	store    *store.MemoryStore
	executor *engine.Executor

	// WebSocket
	upgrader        websocket.Upgrader
	clients         map[*websocket.Conn]bool
	clientsMu       sync.RWMutex
	broadcast       chan interface{}
	broadcastClosed int32 // Atomic flag to track if broadcast channel is closed
	// Rate-limit noisy "channel full" logs when the writer falls behind.
	broadcastDropLogMu   sync.Mutex
	lastBroadcastDropLog time.Time
	accountUpdateMu      sync.Mutex
	lastAccountUpdateAt  map[string]time.Time

	// Session finalization guard — ensures FinalizeAccountsForSession runs
	// exactly once per session, whether stopped manually or by auto-expire.
	sessionFinalized int32

	httpServer *http.Server

	// allowedOrigins: if non-nil, only these Origin values may use CORS and WS.
	// When nil, permissive dev mode (any origin). Set ETH_AMM_SIM_ALLOWED_ORIGINS.
	allowedOrigins map[string]struct{}
	startedAt      time.Time

	security securityConfig
	limiter  *rateLimiter
}

// NewServer creates a new server
func NewServer(session *engine.Session, store *store.MemoryStore, executor *engine.Executor) *Server {
	sec := loadSecurityConfig()
	s := &Server{
		router:              mux.NewRouter(),
		session:             session,
		store:               store,
		executor:            executor,
		clients:             make(map[*websocket.Conn]bool),
		broadcast:           make(chan interface{}, 1024),
		lastAccountUpdateAt: make(map[string]time.Time),
		allowedOrigins:      parseAllowedOriginsFromEnv(),
		security:            sec,
		limiter:             newRateLimiter(sec.readLimit, sec.mutationLimit),
	}
	// Never CheckOrigin=true; always evaluate against the allowlist (or permissive empty).
	s.upgrader.CheckOrigin = func(r *http.Request) bool {
		return s.webSocketOriginOK(r)
	}

	session.SetOnSessionEnded(func() {
		s.finalizeSession()
	})

	s.logAuthStartupWarning()
	s.setupRoutes()
	return s
}

func parseAllowedOriginsFromEnv() map[string]struct{} {
	raw := strings.TrimSpace(os.Getenv("ETH_AMM_SIM_ALLOWED_ORIGINS"))
	if raw == "" {
		return nil
	}
	out := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		o := strings.TrimSpace(part)
		if o != "" {
			out[o] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (s *Server) originAllowed(origin string) bool {
	if len(s.allowedOrigins) == 0 {
		return true
	}
	if origin == "" {
		return true
	}
	_, ok := s.allowedOrigins[origin]
	return ok
}

func (s *Server) webSocketOriginOK(r *http.Request) bool {
	return s.originAllowed(r.Header.Get("Origin"))
}

// Handler returns the HTTP handler (routes + middleware) for tests and embedding.
func (s *Server) Handler() http.Handler {
	return s.router
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	s.router.Use(func(next http.Handler) http.Handler {
		return s.corsMiddleware(s.securityMiddleware(next))
	})

	s.router.HandleFunc("/healthz", s.handleHealthz).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/readyz", s.handleReadyz).Methods("GET", "OPTIONS")

	// Session endpoints
	s.router.HandleFunc("/session/start", s.handleSessionStart).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/session/pause", s.handleSessionPause).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/session/resume", s.handleSessionResume).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/session/stop", s.handleSessionStop).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/session/reset", s.handleSessionReset).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/session/state", s.handleSessionState).Methods("GET", "OPTIONS")

	// Account endpoints
	s.router.HandleFunc("/accounts", s.handleGetAccounts).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/accounts/{nickname}/performance", s.handleGetAccountPerformance).Methods("GET", "OPTIONS")

	// LP endpoints
	s.router.HandleFunc("/lp/metrics", s.handleGetLPMetrics).Methods("GET", "OPTIONS")

	// Market data endpoints
	s.router.HandleFunc("/candles", s.handleGetCandles).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/trades", s.handleGetTrades).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/impact-curve", s.handleGetImpactCurve).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/events", s.handleGetEvents).Methods("GET", "OPTIONS")

	// User trading endpoints
	s.router.HandleFunc("/trade/buy", s.handleTradeBuy).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/trade/sell", s.handleTradeSell).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/user/balance", s.handleGetUserBalance).Methods("GET", "OPTIONS")

	// WebSocket
	s.router.HandleFunc("/stream", s.handleWebSocket)
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	if s.startedAt.IsZero() {
		s.startedAt = time.Now()
	}
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Start broadcast goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Broadcast loop panicked: %v", r)
			}
		}()
		s.runBroadcast()
	}()

	// Register session state callback (broadcast only; finalization runs in
	// Session.run via SetOnSessionEnded after orchestrator.Stop completes).
	s.session.OnStateChange(func(state engine.SessionState) {
		s.Broadcast(WSMessage{Type: "session_state", Data: state})
	})

	log.Printf("Server starting on %s", addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	// Close all WebSocket connections
	s.clientsMu.Lock()
	clientCount := len(s.clients)
	for client := range s.clients {
		// Unblock ReadMessage in handleConnection so HTTP Shutdown does not hang.
		_ = client.SetReadDeadline(time.Now())
		client.Close()
		delete(s.clients, client)
	}
	s.clientsMu.Unlock()
	if clientCount > 0 {
		log.Printf("Closed %d WebSocket connection(s)", clientCount)
	}

	// Close broadcast channel (this will stop the broadcast goroutine)
	// Use atomic flag to prevent double-close and check in Broadcast method
	if atomic.CompareAndSwapInt32(&s.broadcastClosed, 0, 1) {
		close(s.broadcast)
	}

	// Shutdown HTTP server
	return s.httpServer.Shutdown(ctx)
}

// Broadcast sends a message to all WebSocket clients
func (s *Server) Broadcast(msg interface{}) {
	// Check if channel is closed before sending
	if atomic.LoadInt32(&s.broadcastClosed) == 1 {
		return // Channel is closed, silently drop message
	}

	// Nothing to fan out to; avoid queuing work when no clients are connected.
	s.clientsMu.RLock()
	hasClients := len(s.clients) > 0
	s.clientsMu.RUnlock()
	if !hasClients {
		return
	}

	select {
	case s.broadcast <- msg:
	default:
		s.broadcastDropLogMu.Lock()
		if time.Since(s.lastBroadcastDropLog) > 5*time.Second {
			log.Printf("broadcast: outbound queue full, dropping (writer blocked or clients too slow; see write deadline in runBroadcast)")
			s.lastBroadcastDropLog = time.Now()
		}
		s.broadcastDropLogMu.Unlock()
	}
}

// runBroadcast handles broadcasting messages to all clients
func (s *Server) runBroadcast() {
	const writeWait = 3 * time.Second
	for msg := range s.broadcast {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling broadcast: %v", err)
			continue
		}

		s.clientsMu.RLock()
		clients := make([]*websocket.Conn, 0, len(s.clients))
		for c := range s.clients {
			clients = append(clients, c)
		}
		s.clientsMu.RUnlock()

		var failed []*websocket.Conn
		for _, client := range clients {
			_ = client.SetWriteDeadline(time.Now().Add(writeWait))
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				failed = append(failed, client)
			}
		}

		if len(failed) > 0 {
			s.clientsMu.Lock()
			for _, client := range failed {
				client.Close()
				delete(s.clients, client)
			}
			s.clientsMu.Unlock()
		}
	}
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		permissive := len(s.allowedOrigins) == 0

		if permissive {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			if origin != "" && !s.originAllowed(origin) {
				if r.Method == http.MethodOptions {
					http.Error(w, "origin not allowed", http.StatusForbidden)
					return
				}
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	s.clientsMu.RLock()
	n := len(s.clients)
	s.clientsMu.RUnlock()

	qLen := len(s.broadcast)

	uptime := int64(0)
	if !s.startedAt.IsZero() {
		uptime = int64(time.Since(s.startedAt).Seconds())
	}

	respondJSON(w, map[string]interface{}{
		"status":              "ok",
		"uptime_seconds":      uptime,
		"ws_clients":          n,
		"broadcast_queue_len": qLen,
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if s.startedAt.IsZero() {
		respondError(w, http.StatusServiceUnavailable, "server not started")
		return
	}
	s.handleHealthz(w, r)
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// finalizeSession fetches the final spot price and finalizes all account
// metrics (close prices, PnL recalculation, session volatility). It is invoked
// from Session.run via SetOnSessionEnded after orchestrator.Stop(). Guarded by
// sessionFinalized so it runs once per completed session; handleSessionStart
// resets the guard when a new session begins.
func (s *Server) finalizeSession() {
	if !atomic.CompareAndSwapInt32(&s.sessionFinalized, 0, 1) {
		return // already finalized
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	spotPrice, err := s.executor.GetSpotPrice(ctx)
	if err != nil {
		log.Printf("Warning: Could not get spot price for session finalization: %v", err)
		spotPrice = big.NewInt(0)
	}

	var spotPriceFloat float64
	if spotPrice != nil && spotPrice.Sign() > 0 {
		spf := new(big.Float).SetInt(spotPrice)
		spf.Quo(spf, big.NewFloat(1e18))
		spotPriceFloat, _ = spf.Float64()
	}

	log.Printf("Finalizing session: spotPrice=%.6f", spotPriceFloat)

	s.store.FinalizeAccountsForSession(spotPriceFloat)
	s.BroadcastAllAccountUpdates()
}

// SetDuration allows setting session duration (for API)
func (s *Server) SetDuration(seconds int) {
	s.session.SetDuration(time.Duration(seconds) * time.Second)
}

// reinitializeLPMetrics re-initializes LP metrics with current pool state
func (s *Server) reinitializeLPMetrics(ctx context.Context) {
	// Get current reserves from contract
	apples, eth, err := s.executor.GetReserves(ctx)
	if err != nil {
		log.Printf("Warning: Could not get reserves for LP metrics reinit: %v", err)
		return
	}

	// Get current fees from contract
	feesApple, feesETH, err := s.executor.GetTotalFees(ctx)
	if err != nil {
		log.Printf("Warning: Could not get fees for LP metrics reinit: %v", err)
		feesApple = big.NewInt(0)
		feesETH = big.NewInt(0)
	}

	// Set initial state to current reserves (this is the new baseline)
	s.store.GetLPMetrics().SetInitialState(apples, eth)
	s.store.GetLPMetrics().SetInitialFees(feesApple, feesETH)

	// Also update current state to match (ensures metrics show current pool state)
	s.store.GetLPMetrics().UpdateState(apples, eth, feesApple, feesETH)
	s.store.GetImpactCurve().UpdateReserves(apples, eth)

	log.Printf("LP metrics re-initialized: APPL=%s, ETH=%s", apples.String(), eth.String())
}
