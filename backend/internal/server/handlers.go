// Package server provides HTTP handlers
package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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
	if err := s.session.Start(ctx); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	respondJSON(w, map[string]string{"status": "started"})
}

func (s *Server) handleSessionStop(w http.ResponseWriter, r *http.Request) {
	if err := s.session.Stop(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	respondJSON(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleSessionReset(w http.ResponseWriter, r *http.Request) {
	if err := s.session.Reset(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Reset store
	s.store.Reset()
	
	respondJSON(w, map[string]string{"status": "reset"})
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
	// Optional limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 100
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
