// broadcast.go — Pushes real-time data to all connected WebSocket clients.
//
// The frontend opens a WebSocket to /stream (via hooks/useWebSocket.ts).
// When trades execute, prices update, or sessions change, main.go calls
// BroadcastTrade, BroadcastPrice, BroadcastLPMetrics, etc. These methods
// serialize the data as JSON and fan it out to every connected client.
//
// MESSAGE TYPES (consumed by frontend page.tsx handleWSMessage):
//
//	"trade"          → Blotter component
//	"price"          → PriceChart component
//	"lp_metrics"     → LPStats component
//	"key_event"      → KeyEvents component
//	"session_state"  → SessionControls component
//	"account_update" → AccountMetrics component
package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Send initial state before registration so runBroadcast doesn't write
	// concurrently to the same conn during bootstrap.
	s.sendInitialState(conn)

	// Register client
	s.clientsMu.Lock()
	s.clients[conn] = true
	clientCount := len(s.clients)
	s.clientsMu.Unlock()

	log.Printf("WebSocket client connected. Total clients: %d", clientCount)

	// Keep connection alive and handle disconnection
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("WebSocket handler panicked: %v", r)
			}
		}()
		s.handleConnection(conn)
	}()
}

// sendInitialState sends the current state to a new client
func (s *Server) sendInitialState(conn *websocket.Conn) {
	// Send session state
	_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	conn.WriteJSON(WSMessage{
		Type: "session_state",
		Data: s.session.GetState(),
	})

	// Send LP metrics
	_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	conn.WriteJSON(WSMessage{
		Type: "lp_metrics",
		Data: s.store.GetLPData(),
	})

	// Send recent candles
	_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	conn.WriteJSON(WSMessage{
		Type: "candles",
		Data: s.store.GetCandles(),
	})

	// Send recent trades
	_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	conn.WriteJSON(WSMessage{
		Type: "trades",
		Data: s.store.GetRecentTrades(50),
	})

	// Send recent events
	_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	conn.WriteJSON(WSMessage{
		Type: "events",
		Data: s.store.GetRecentEvents(20),
	})
}

// handleConnection manages a single WebSocket connection
func (s *Server) handleConnection(conn *websocket.Conn) {
	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		clientCount := len(s.clients)
		s.clientsMu.Unlock()
		conn.Close()
		log.Printf("WebSocket client disconnected. Total clients: %d", clientCount)
	}()

	// Read messages until disconnect. Browser clients don't send periodic messages,
	// so avoid a hard read deadline that would force disconnect churn.
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// BroadcastTrade sends a trade to all clients
func (s *Server) BroadcastTrade(trade interface{}) {
	s.Broadcast(WSMessage{Type: "trade", Data: trade})
}

// BroadcastPrice sends a price update to all clients
func (s *Server) BroadcastPrice(candle interface{}) {
	s.Broadcast(WSMessage{Type: "price", Data: candle})
}

// BroadcastLPMetrics sends LP metrics to all clients
func (s *Server) BroadcastLPMetrics() {
	s.Broadcast(WSMessage{Type: "lp_metrics", Data: s.store.GetLPData()})
}

// BroadcastAccountUpdate sends an account update to all clients
func (s *Server) BroadcastAccountUpdate(nickname string) {
	// Account performance payloads grow with session length (equity/trade history).
	// Throttle per-account pushes to avoid saturating the WS queue under heavy flow.
	const minAccountUpdateInterval = 1 * time.Second
	now := time.Now()
	s.accountUpdateMu.Lock()
	last := s.lastAccountUpdateAt[nickname]
	if now.Sub(last) < minAccountUpdateInterval {
		s.accountUpdateMu.Unlock()
		return
	}
	s.lastAccountUpdateAt[nickname] = now
	s.accountUpdateMu.Unlock()

	perf := s.store.GetAccountPerformance(nickname)
	if perf != nil {
		s.Broadcast(WSMessage{Type: "account_update", Data: perf})
	}
}

// BroadcastAllAccountUpdates sends updated performance data for every account.
// Called after session finalization so the frontend picks up close prices and
// recalculated PnL/Sharpe.
func (s *Server) BroadcastAllAccountUpdates() {
	allPerf := s.store.GetAllAccountPerformance()
	for i := range allPerf {
		s.Broadcast(WSMessage{Type: "account_update", Data: &allPerf[i]})
	}
}

// BroadcastEvent sends a key event to all clients
func (s *Server) BroadcastEvent(event interface{}) {
	s.Broadcast(WSMessage{Type: "key_event", Data: event})
}
