// Package engine provides session management
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SessionStatus represents the current session state
type SessionStatus string

const (
	StatusIdle      SessionStatus = "idle"
	StatusRunning   SessionStatus = "running"
	StatusCompleted SessionStatus = "completed"
	StatusError     SessionStatus = "error"
)

// SessionState holds the current session state
type SessionState struct {
	Status    SessionStatus `json:"status"`
	StartedAt *time.Time    `json:"startedAt,omitempty"`
	EndedAt   *time.Time    `json:"endedAt,omitempty"`
	Duration  int           `json:"duration"` // in seconds
	Elapsed   int           `json:"elapsed"`  // in seconds
	Error     string        `json:"error,omitempty"`
}

// Session manages a simulation session
type Session struct {
	mu sync.RWMutex
	
	status    SessionStatus
	startedAt *time.Time
	endedAt   *time.Time
	duration  time.Duration
	
	cancel context.CancelFunc
	
	// Orchestrator manages bots
	orchestrator *Orchestrator
	
	// Callbacks
	stateCallbacks []func(SessionState)
}

// NewSession creates a new session
func NewSession(orchestrator *Orchestrator) *Session {
	return &Session{
		status:       StatusIdle,
		duration:     3 * time.Minute,
		orchestrator: orchestrator,
	}
}

// SetDuration sets the session duration
func (s *Session) SetDuration(d time.Duration) {
	s.mu.Lock()
	s.duration = d
	s.mu.Unlock()
}

// OnStateChange registers a callback for state changes
func (s *Session) OnStateChange(callback func(SessionState)) {
	s.mu.Lock()
	s.stateCallbacks = append(s.stateCallbacks, callback)
	s.mu.Unlock()
}

// emitState notifies all callbacks of state change
func (s *Session) emitState() {
	state := s.GetState()
	
	s.mu.RLock()
	callbacks := make([]func(SessionState), len(s.stateCallbacks))
	copy(callbacks, s.stateCallbacks)
	s.mu.RUnlock()
	
	for _, cb := range callbacks {
		go cb(state)
	}
}

// GetState returns the current session state
func (s *Session) GetState() SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var elapsed int
	if s.startedAt != nil {
		if s.endedAt != nil {
			elapsed = int(s.endedAt.Sub(*s.startedAt).Seconds())
		} else {
			elapsed = int(time.Since(*s.startedAt).Seconds())
		}
	}
	
	return SessionState{
		Status:    s.status,
		StartedAt: s.startedAt,
		EndedAt:   s.endedAt,
		Duration:  int(s.duration.Seconds()),
		Elapsed:   elapsed,
	}
}

// Start begins a new session
func (s *Session) Start(ctx context.Context) error {
	s.mu.Lock()
	
	if s.status == StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("session already running")
	}
	
	now := time.Now()
	s.startedAt = &now
	s.endedAt = nil
	s.status = StatusRunning
	
	// Create cancellable context for this session
	sessionCtx, cancel := context.WithTimeout(ctx, s.duration)
	s.cancel = cancel
	
	s.mu.Unlock()
	
	s.emitState()
	
	// Start the orchestrator
	go s.run(sessionCtx)
	
	return nil
}

// run executes the session
func (s *Session) run(ctx context.Context) {
	// Start all bots
	s.orchestrator.StartAll(ctx)
	
	// Wait for context to complete (timeout or cancellation)
	<-ctx.Done()
	
	// Stop all bots
	s.orchestrator.StopAll()
	
	s.mu.Lock()
	now := time.Now()
	s.endedAt = &now
	s.status = StatusCompleted
	s.mu.Unlock()
	
	s.emitState()
}

// Stop stops the current session
func (s *Session) Stop() error {
	s.mu.Lock()
	
	if s.status != StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("session not running")
	}
	
	if s.cancel != nil {
		s.cancel()
	}
	
	s.mu.Unlock()
	
	return nil
}

// Reset resets the session for a new run
func (s *Session) Reset() error {
	s.mu.Lock()
	
	if s.status == StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("cannot reset running session")
	}
	
	s.status = StatusIdle
	s.startedAt = nil
	s.endedAt = nil
	
	s.mu.Unlock()
	
	s.emitState()
	
	return nil
}

// IsRunning returns whether the session is running
func (s *Session) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status == StatusRunning
}
