// SessionControls.tsx — Start / Stop / Reset buttons for the simulation.
//
// Drives the simulation lifecycle via hooks/useSession.ts → lib/api.ts
// → backend POST /session/{start,stop,reset}. Real-time state updates
// (elapsed time, status transitions) arrive via WebSocket "session_state".
//
// CONNECTIONS:
//   - Backend:     engine/session.go state machine (idle → running → completed)
//   - REST:        handlers.go handleSessionStart/Stop/Reset
//   - WebSocket:   broadcast.go sends "session_state" on transitions
//   - State hook:  hooks/useSession.ts provides session state + action callbacks
'use client';

import { useState } from 'react';
import type { SessionState } from '@/types';

interface SessionControlsProps {
  session: SessionState;
  isLoading: boolean;
  onStart: (duration: number) => void;
  onStop: () => void;
  onReset: (hardReset?: boolean) => void;
}

export const SessionControls = ({ session, isLoading, onStart, onStop, onReset }: SessionControlsProps) => {
  const [duration, setDuration] = useState(120);
  
  const isRunning = session.status === 'running';
  const isCompleted = session.status === 'completed';
  
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };
  
  const progress = session.duration > 0 ? (session.elapsed / session.duration) * 100 : 0;
  
  return (
    <div className="bg-surface rounded-lg p-4 border border-border">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-medium text-white">Session Control</h2>
        <div className="flex items-center space-x-2">
          <span className={`w-2 h-2 rounded-full ${
            isRunning ? 'bg-green-500 animate-pulse' : 
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
          disabled={isRunning}
          className="w-full bg-surface-light border border-border rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500 disabled:opacity-50"
          min={30}
          max={600}
        />
      </div>
      
      {/* Progress bar */}
      {isRunning && (
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
      <div className="flex space-x-2">
        {!isRunning ? (
          <>
            <button
              onClick={() => onStart(duration)}
              disabled={isLoading}
              className="flex-1 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
            >
              {isLoading ? 'Starting...' : 'Start'}
            </button>
            {(isCompleted || session.status === 'idle') && (
              <button
                onClick={() => onReset(true)}
                disabled={isLoading}
                className="flex-1 bg-gray-600 hover:bg-gray-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
                title="Hard reset: Clears all data including account metrics"
              >
                Reset
              </button>
            )}
          </>
        ) : (
          <button
            onClick={onStop}
            disabled={isLoading}
            className="flex-1 bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
          >
            {isLoading ? 'Stopping...' : 'Stop'}
          </button>
        )}
      </div>
    </div>
  );
};
