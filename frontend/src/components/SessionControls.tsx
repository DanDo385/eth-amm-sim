// SessionControls.tsx - Start / Stop / Reset buttons for the simulation.
//
// Drives the simulation lifecycle via hooks/useSession.ts → lib/api.ts
// → backend POST /session/{start,stop,reset}. Real-time state updates
// (elapsed time, status transitions) arrive via WebSocket "session_state".
//
// CONNECTIONS:
//  - Backend:     engine/session.go state machine (idle → running → completed)
//  - REST:        handlers.go handleSessionStart/Stop/Reset
//  - WebSocket:   broadcast.go sends "session_state" on transitions
//  - State hook:  hooks/useSession.ts provides session state + action callbacks

import { useState } from 'react';
import type { ResetMode, SessionState } from '@/types';

interface SessionControlsProps {
  session: SessionState;
  isLoading: boolean;
  error?: string | null;
  onStart: (duration: number) => void;
  onPause: () => void;
  onResume: () => void;
  onStop: () => void;
  onReset: (mode: ResetMode) => void;
}

export const SessionControls = ({ session, isLoading, error, onStart, onPause, onResume, onStop, onReset }: SessionControlsProps) => {
  const [duration, setDuration] = useState(120);
  
  const isRunning = session.status === 'running';
  const isPaused = session.status === 'paused';
  const isCompleted = session.status === 'completed';
  const isIdle = session.status === 'idle';
  
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };
  
  const progress = session.duration > 0 ? (session.elapsed / session.duration) * 100 : 0;
  
  return (
    <div className="bg-surface rounded-lg p-4 border border-border">
      {error && (
        <div
          className="mb-4 rounded border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm text-red-200"
          role="alert"
        >
          {error}
        </div>
      )}

      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-medium text-white">Session Control</h2>
        <div className="flex items-center space-x-2">
          <span className={`w-2 h-2 rounded-full ${
            isRunning ? 'bg-green-500 animate-pulse' : 
            isPaused ? 'bg-amber-500' :
            isCompleted ? 'bg-blue-500' : 'bg-gray-500'
          }`} />
          <span className="text-sm text-gray-400 capitalize">{session.status}</span>
        </div>
      </div>
      
      {/* Duration input */}
      <div className="mb-4">
        <label className="block text-sm text-gray-400 mb-1">Duration (seconds)</label>
        <div className="mb-2 rounded border border-cyan-500/25 bg-cyan-500/10 px-2 py-1 text-[11px] leading-4 text-cyan-200">
          Loom default: 120 seconds. One clean pass, one retake max.
        </div>
        <input
          type="number"
          value={duration}
          onChange={(e) => setDuration(parseInt(e.target.value) || 180)}
          disabled={isRunning || isPaused}
          className="w-full bg-surface-light border border-border rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500 disabled:opacity-50"
          min={30}
          max={600}
        />
      </div>
      
      {/* Progress bar: running (live) or completed (full bar, read-only) */}
      {(isRunning || isPaused || isCompleted) && (
        <div className="mb-4">
          <div className="flex justify-between text-sm text-gray-400 mb-1">
            <span>Elapsed</span>
            <span>{formatTime(session.elapsed)} / {formatTime(session.duration)}</span>
          </div>
          <div className="h-2 bg-surface-light rounded-full overflow-hidden">
            <div 
              className="h-full bg-blue-500 transition-all duration-1000"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}
      
      {/* Control buttons */}
      <div className="flex flex-wrap gap-2">
        {(isIdle || isCompleted) && (
          <button
            type="button"
            onClick={() => onStart(duration)}
            disabled={isLoading}
            className="min-w-0 flex-1 basis-[40%] bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
          >
            {isLoading ? 'Starting...' : 'Start'}
          </button>
        )}
        {isRunning && (
          <button
            type="button"
            onClick={() => onPause()}
            disabled={isLoading}
            className="min-w-0 flex-1 basis-[40%] bg-amber-600 hover:bg-amber-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
            title="Pause simulation and keep remaining time"
          >
            {isLoading ? 'Pausing...' : 'Pause'}
          </button>
        )}
        {isPaused && (
          <button
            type="button"
            onClick={() => onResume()}
            disabled={isLoading}
            className="min-w-0 flex-1 basis-[40%] bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
            title="Resume simulation with positions reset to starting balances"
          >
            {isLoading ? 'Resuming...' : 'Resume'}
          </button>
        )}
        {(isRunning || isPaused || isCompleted) && (
          <button
            type="button"
            onClick={() => onStop()}
            disabled={isLoading}
            className="min-w-0 flex-1 basis-[40%] bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
            title={isCompleted ? 'Dismiss completed session (same as clearing to idle)' : 'Stop simulation'}
          >
            {isLoading ? 'Stopping...' : 'Stop'}
          </button>
        )}
        {!isRunning && !isPaused && (isIdle || isCompleted) && (
          <button
            type="button"
            onClick={() => onReset('soft')}
            disabled={isLoading}
            className="min-w-0 flex-1 basis-[30%] bg-gray-600 hover:bg-gray-700 disabled:opacity-50 text-white font-medium py-2 px-3 rounded transition"
            title="Soft reset: clear charts/events/trades; keep chain state"
          >
            Reset
            <span className="ml-1 text-[10px] text-gray-200" title="Soft reset: clears dashboard session data only.">
              ?
            </span>
          </button>
        )}
        {!isRunning && !isPaused && (isIdle || isCompleted) && (
          <button
            type="button"
            onClick={() => onReset('hard')}
            disabled={isLoading}
            className="min-w-0 flex-1 basis-[30%] bg-slate-600 hover:bg-slate-700 disabled:opacity-50 text-white font-medium py-2 px-3 rounded transition"
            title="Hard reset: clear session/account metrics and reset User wallet balances"
          >
            Hard
            <span className="ml-1 text-[10px] text-gray-200" title="Hard reset: also resets account metrics and user balances.">
              ?
            </span>
          </button>
        )}
        {!isRunning && !isPaused && (isIdle || isCompleted) && (
          <button
            type="button"
            onClick={() => onReset('reseed')}
            disabled={isLoading}
            className="min-w-0 flex-1 basis-[30%] bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white font-medium py-2 px-3 rounded transition"
            title="Reseed reset: anvil_reset + deploy to restore initial pool and ~1.0 starting price"
          >
            Reseed
            <span className="ml-1 text-[10px] text-violet-200" title="Reseed: full chain reset + redeploy to initial balances/pool.">
              ?
            </span>
          </button>
        )}
      </div>
    </div>
  );
};
