// KeyEvents.tsx - Real-time feed of notable simulation events.
//
// Shows trades and strategy triggers with severity badges.
// Events arrive via WebSocket "key_event" messages, dispatched from
// broadcast.go BroadcastEvent whenever a bot executes a notable action.
//
// CONNECTIONS:
//  - Backend data:  store/memory.go AddEvent (called by main.go trade callback)
//  - WebSocket msg: "key_event" from broadcast.go BroadcastEvent
//  - Initial load:  sendInitialState sends "events" batch on WS connect
//  - Types:         types/index.ts KeyEvent

import type { KeyEvent } from '@/types';

interface KeyEventsProps {
  events: KeyEvent[];
  height?: number; // Height in pixels (defaults to 200 to match TWAPChart)
  onSelectEvent?: (event: KeyEvent) => void;
}

const eventIcons: Record<string, string> = {
  trade: '📊',
  strategy_trigger: '🎯',
};

const severityColors: Record<string, string> = {
  info: 'text-blue-400',
  warning: 'text-yellow-400',
  critical: 'text-red-400',
};

export const KeyEvents = ({ events, height = 200, onSelectEvent }: KeyEventsProps) => {
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString();
  };

  return (
    <div className="bg-surface rounded-lg border border-border">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-sm font-medium text-white">Key Events</h3>
      </div>
      <div 
        className="overflow-auto"
        style={{ height: `${height}px` }}
      >
        {events.length === 0 ? (
          <div className="px-4 py-8 text-center text-gray-500 text-sm">
            No events yet
          </div>
        ) : (
          <ul className="divide-y divide-border">
            {events.map((event, i) => (
              <li
                key={i}
                className={`px-4 py-2 flex items-start space-x-3 ${event.type === 'trade' ? 'cursor-pointer hover:bg-surface-light/60' : ''}`}
                onClick={() => {
                  if (event.type === 'trade') {
                    onSelectEvent?.(event);
                  }
                }}
                title={event.type === 'trade' ? 'Click to inspect AMM details' : undefined}
              >
                <span className="text-lg">{eventIcons[event.type] || '📌'}</span>
                <div className="flex-1 min-w-0">
                  <p className={`text-sm ${severityColors[event.severity] || 'text-gray-300'}`}>
                    {event.description}
                  </p>
                  <p className="text-xs text-gray-500 mt-0.5">
                    {formatTime(event.timestamp)}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};
