// session_test.go - Session timer, pause/resume, and status transitions.
package engine

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func waitForSessionStatus(t *testing.T, s *Session, want SessionStatus, timeout time.Duration) SessionState {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		st := s.GetState()
		if st.Status == want {
			return st
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for status %s; last=%+v", want, s.GetState())
	return SessionState{}
}

func TestSessionPauseResumeCompletesWithRemainingTime(t *testing.T) {
	o := NewOrchestrator()
	o.AddBot(&testBot{n: "pause-resume-bot"})

	s := NewSession(o)
	s.SetDuration(3 * time.Second)

	var endedCalls atomic.Int32
	s.SetOnSessionEnded(func() {
		endedCalls.Add(1)
	})

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	waitForSessionStatus(t, s, StatusRunning, 250*time.Millisecond)

	time.Sleep(1200 * time.Millisecond)
	if err := s.Pause(); err != nil {
		t.Fatalf("pause failed: %v", err)
	}
	paused := waitForSessionStatus(t, s, StatusPaused, 500*time.Millisecond)
	if paused.Elapsed <= 0 {
		t.Fatalf("expected elapsed > 0 after pause, got %d", paused.Elapsed)
	}
	if got := endedCalls.Load(); got != 0 {
		t.Fatalf("session ended hook should not fire on pause, got %d", got)
	}

	if err := s.Resume(context.Background()); err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	completed := waitForSessionStatus(t, s, StatusCompleted, 3*time.Second)
	if completed.Elapsed < paused.Elapsed {
		t.Fatalf("elapsed regressed after resume: paused=%d completed=%d", paused.Elapsed, completed.Elapsed)
	}
	if got := endedCalls.Load(); got != 1 {
		t.Fatalf("session ended hook should fire once on completion, got %d", got)
	}
}

func TestSessionStopFromPausedReturnsToIdle(t *testing.T) {
	o := NewOrchestrator()
	o.AddBot(&testBot{n: "stop-paused-bot"})

	s := NewSession(o)
	s.SetDuration(2 * time.Second)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	waitForSessionStatus(t, s, StatusRunning, 250*time.Millisecond)

	time.Sleep(150 * time.Millisecond)
	if err := s.Pause(); err != nil {
		t.Fatalf("pause failed: %v", err)
	}
	waitForSessionStatus(t, s, StatusPaused, 500*time.Millisecond)

	if err := s.Stop(); err != nil {
		t.Fatalf("stop from paused failed: %v", err)
	}
	idle := waitForSessionStatus(t, s, StatusIdle, 250*time.Millisecond)
	if idle.Elapsed != 0 {
		t.Fatalf("expected idle elapsed reset to 0, got %d", idle.Elapsed)
	}
}
