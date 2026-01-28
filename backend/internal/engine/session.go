// session.go — Controls the simulation session lifecycle (idle → running → completed).
//
// A session wraps an Orchestrator and a timer. When the frontend user clicks
// "Start" (SessionControls → POST /session/start → server/handlers.go), Session
// creates a context with the configured duration (default 3 min), starts all bots
// via Orchestrator, and transitions to StatusRunning. When the timer expires or
// the user clicks "Stop", the context is cancelled, bots exit, and state becomes
// StatusCompleted. State changes are broadcast via WebSocket to update the
// frontend SessionControls component in real time.
//
// CONNECTIONS:
//   - REST API: server/handlers.go calls Start/Stop/Reset/GetState
//   - Frontend: SessionControls reads session state; useSession hook polls it
//   - Orchestrator: session.run() calls orchestrator.Start/Stop
package engine

import (
	"context"
	"fmt"
	"log"
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
	
	var wg sync.WaitGroup
	for _, cb := range callbacks {
		cb := cb // Capture loop variable
		wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Session state callback panicked: %v", r)
				}
				wg.Done()
			}()
			cb(state)
		}()
	}
	
	// Fire-and-forget wait (don't block emitState caller)
	go func() {
		wg.Wait()
	}()
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
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Session run panicked: %v", r)
				s.mu.Lock()
				s.status = StatusError
				s.mu.Unlock()
				s.emitState()
			}
		}()
		s.run(sessionCtx)
	}()
	
	return nil
}

// run executes the session
func (s *Session) run(ctx context.Context) {
	// Start all bots
	s.orchestrator.Start(ctx)
	
	// Wait for context to complete (timeout or cancellation)
	<-ctx.Done()
	
	// Stop all bots
	s.orchestrator.Stop()
	
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
	
	// Ensure orchestrator is fully stopped (in case any bots are lingering)
	s.orchestrator.Stop()
	
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
