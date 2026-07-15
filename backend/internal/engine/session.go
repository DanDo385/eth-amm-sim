// session.go - Controls the simulation session lifecycle (idle → running → completed).
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
//  - REST API: server/handlers.go calls Start/Stop/Reset/GetState
//  - Frontend: SessionControls reads session state; useSession hook polls it
//  - Orchestrator: session.run() calls orchestrator.Start/Stop
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
	StatusPaused    SessionStatus = "paused"
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
	elapsed   time.Duration

	cancel   context.CancelFunc
	stopMode sessionStopMode

	// Orchestrator manages bots
	orchestrator *Orchestrator

	// Callbacks
	stateCallbacks []func(SessionState)

	endedMu        sync.Mutex
	onSessionEnded func() // optional; invoked once per run after bots stop, before completed
}

type sessionStopMode string

const (
	stopModeComplete sessionStopMode = "complete"
	stopModePause    sessionStopMode = "pause"
)

// NewSession creates a new session
func NewSession(orchestrator *Orchestrator) *Session {
	return &Session{
		status:       StatusIdle,
		duration:     3 * time.Minute,
		orchestrator: orchestrator,
		stopMode:     stopModeComplete,
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

// SetOnSessionEnded registers a hook invoked synchronously from run() after
// orchestrator.Stop() returns and before status moves to completed. Use this
// for metrics finalization so bot teardown finishes before session accounting closes.
func (s *Session) SetOnSessionEnded(fn func()) {
	s.endedMu.Lock()
	s.onSessionEnded = fn
	s.endedMu.Unlock()
}

func (s *Session) invokeSessionEnded() {
	s.endedMu.Lock()
	fn := s.onSessionEnded
	s.endedMu.Unlock()
	if fn != nil {
		fn()
	}
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
	elapsedDur := s.elapsed
	if s.status == StatusRunning && s.startedAt != nil {
		elapsedDur += time.Since(*s.startedAt)
	}
	elapsed = int(elapsedDur.Seconds())

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
	s.elapsed = 0
	s.status = StatusRunning
	s.stopMode = stopModeComplete

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
	if s.startedAt != nil {
		s.elapsed += time.Since(*s.startedAt)
		s.startedAt = nil
	}
	mode := s.stopMode
	s.stopMode = stopModeComplete
	s.cancel = nil
	if mode == stopModePause {
		s.status = StatusPaused
		s.mu.Unlock()
		s.emitState()
		return
	}
	now := time.Now()
	s.endedAt = &now
	s.status = StatusCompleted
	s.mu.Unlock()

	// One-shot accounting hook (e.g. finalize PnL) while bots are fully stopped
	s.invokeSessionEnded()
	s.emitState()
}

// Stop stops the current session, or normalizes a finished session to idle.
//
//  - running: cancel context; run() ends the session as completed (existing path).
//  - completed: move to idle and clear timestamps so the UI matches first load;
//     idempotent for clients that call POST /session/stop after the timer ends.
//  - idle: no-op (idempotent).
func (s *Session) Stop() error {
	s.mu.Lock()

	if s.status == StatusIdle {
		s.mu.Unlock()
		return nil
	}

	if s.status == StatusCompleted || s.status == StatusPaused {
		s.status = StatusIdle
		s.startedAt = nil
		s.endedAt = nil
		s.elapsed = 0
		s.cancel = nil
		s.stopMode = stopModeComplete
		s.mu.Unlock()
		s.emitState()
		return nil
	}

	if s.status != StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("session not running")
	}

	if s.cancel != nil {
		s.stopMode = stopModeComplete
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
	s.elapsed = 0
	s.stopMode = stopModeComplete

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

// Pause temporarily pauses a running session without finalizing account metrics.
func (s *Session) Pause() error {
	s.mu.Lock()

	if s.status != StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("session not running")
	}

	if s.cancel != nil {
		s.stopMode = stopModePause
		s.cancel()
	}
	s.mu.Unlock()

	return nil
}

// Resume resumes a previously paused session for the remaining duration.
func (s *Session) Resume(ctx context.Context) error {
	s.mu.Lock()

	if s.status != StatusPaused {
		s.mu.Unlock()
		return fmt.Errorf("session not paused")
	}

	remaining := s.duration - s.elapsed
	if remaining <= 0 {
		now := time.Now()
		s.endedAt = &now
		s.status = StatusCompleted
		s.mu.Unlock()
		s.emitState()
		return nil
	}

	now := time.Now()
	s.startedAt = &now
	s.endedAt = nil
	s.status = StatusRunning
	s.stopMode = stopModeComplete

	sessionCtx, cancel := context.WithTimeout(ctx, remaining)
	s.cancel = cancel
	s.mu.Unlock()

	s.emitState()

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
