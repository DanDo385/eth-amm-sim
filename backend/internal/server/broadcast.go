// Package server provides WebSocket broadcasting
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
	
	// Register client
	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()
	
	log.Printf("WebSocket client connected. Total clients: %d", len(s.clients))
	
	// Send initial state
	s.sendInitialState(conn)
	
	// Keep connection alive and handle disconnection
	go s.handleConnection(conn)
}

// sendInitialState sends the current state to a new client
func (s *Server) sendInitialState(conn *websocket.Conn) {
	// Send session state
	conn.WriteJSON(WSMessage{
		Type: "session_state",
		Data: s.session.GetState(),
	})
	
	// Send LP metrics
	conn.WriteJSON(WSMessage{
		Type: "lp_metrics",
		Data: s.store.GetLPData(),
	})
	
	// Send recent candles
	conn.WriteJSON(WSMessage{
		Type: "candles",
		Data: s.store.GetCandles(),
	})
	
	// Send recent trades
	conn.WriteJSON(WSMessage{
		Type: "trades",
		Data: s.store.GetRecentTrades(50),
	})
	
	// Send recent events
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
		s.clientsMu.Unlock()
		conn.Close()
		log.Printf("WebSocket client disconnected. Total clients: %d", len(s.clients))
	}()
	
	// Configure ping/pong for connection health
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	// Read messages (mainly for ping/pong)
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
	perf := s.store.GetAccountPerformance(nickname)
	if perf != nil {
		s.Broadcast(WSMessage{Type: "account_update", Data: perf})
	}
}

// BroadcastEvent sends a key event to all clients
func (s *Server) BroadcastEvent(event interface{}) {
	s.Broadcast(WSMessage{Type: "key_event", Data: event})
}
