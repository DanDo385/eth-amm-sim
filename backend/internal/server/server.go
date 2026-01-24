// Package server provides HTTP and WebSocket server
package server

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"sync"
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
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]bool
	clientsMu   sync.RWMutex
	broadcast   chan interface{}
	
	httpServer *http.Server
}

// NewServer creates a new server
func NewServer(session *engine.Session, store *store.MemoryStore, executor *engine.Executor) *Server {
	s := &Server{
		router:   mux.NewRouter(),
		session:  session,
		store:    store,
		executor: executor,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan interface{}, 100),
	}
	
	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Enable CORS
	s.router.Use(corsMiddleware)
	
	// Session endpoints
	s.router.HandleFunc("/session/start", s.handleSessionStart).Methods("POST", "OPTIONS")
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
	
	// WebSocket
	s.router.HandleFunc("/stream", s.handleWebSocket)
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	
	// Start broadcast goroutine
	go s.runBroadcast()
	
	// Register session state callback
	s.session.OnStateChange(func(state engine.SessionState) {
		s.Broadcast(WSMessage{Type: "session_state", Data: state})
	})
	
	log.Printf("Server starting on %s", addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Broadcast sends a message to all WebSocket clients
func (s *Server) Broadcast(msg interface{}) {
	select {
	case s.broadcast <- msg:
	default:
		log.Println("Broadcast channel full, dropping message")
	}
}

// runBroadcast handles broadcasting messages to all clients
func (s *Server) runBroadcast() {
	for msg := range s.broadcast {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling broadcast: %v", err)
			continue
		}
		
		s.clientsMu.RLock()
		for client := range s.clients {
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("Error sending to client: %v", err)
				client.Close()
				delete(s.clients, client)
			}
		}
		s.clientsMu.RUnlock()
	}
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
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
	s.store.GetImpactCurve().UpdateReserves(apples, eth)
	
	log.Printf("LP metrics re-initialized: APPL=%s, ETH=%s", apples.String(), eth.String())
}
