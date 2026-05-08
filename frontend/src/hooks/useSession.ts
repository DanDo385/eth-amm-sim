// useSession.ts — Manages simulation lifecycle (start / stop / reset).
//
// Provides the session state to page.tsx and SessionControls. REST calls
// go through lib/api.ts → backend POST /session/{start,stop,reset}.
// Real-time state updates arrive via WebSocket "session_state" messages
// (dispatched by page.tsx handleWSMessage → updateFromWS).
//
// CONNECTIONS:
//   - REST:      lib/api.ts → server/handlers.go handleSession{Start,Stop,Reset}
//   - WebSocket: page.tsx receives "session_state" → calls updateFromWS
//   - Backend:   engine/session.go manages the actual state machine

import { useState, useEffect, useCallback } from 'react';
import type { ResetMode, SessionState } from '@/types';
import * as api from '@/lib/api';

export function useSession() {
  const [session, setSession] = useState<SessionState>({
    status: 'idle',
    duration: 180,
    elapsed: 0,
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch session state
  const refresh = useCallback(async () => {
    try {
      const state = await api.getSessionState();
      setSession(state);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch session state');
    }
  }, []);

  // Start session
  const start = useCallback(async (duration?: number) => {
    setIsLoading(true);
    setError(null);
    try {
      await api.startSession(duration);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to start session');
    } finally {
      setIsLoading(false);
    }
  }, [refresh]);

  // Stop session
  const stop = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      await api.stopSession();
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to stop session');
    } finally {
      setIsLoading(false);
    }
  }, [refresh]);

  const pause = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      await api.pauseSession();
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to pause session');
    } finally {
      setIsLoading(false);
    }
  }, [refresh]);

  const resume = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      await api.resumeSession();
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to resume session');
    } finally {
      setIsLoading(false);
    }
  }, [refresh]);

  // Reset session
  const reset = useCallback(async (mode: ResetMode = 'soft') => {
    setIsLoading(true);
    setError(null);
    try {
      await api.resetSession(mode);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to reset session');
    } finally {
      setIsLoading(false);
    }
  }, [refresh]);

  // Update session state from WebSocket
  const updateFromWS = useCallback((state: SessionState) => {
    setSession(state);
    setError(null);
  }, []);

  // Poll for elapsed time when running
  useEffect(() => {
    if (session.status !== 'running') return;

    const interval = setInterval(() => {
      setSession((prev) => ({
        ...prev,
        elapsed: Math.min(prev.elapsed + 1, prev.duration),
      }));
    }, 1000);

    return () => clearInterval(interval);
  }, [session.status]);

  // Initial fetch
  useEffect(() => {
    refresh();
  }, [refresh]);

  return {
    session,
    isLoading,
    error,
    start,
    stop,
    pause,
    resume,
    reset,
    refresh,
    updateFromWS,
  };
}
