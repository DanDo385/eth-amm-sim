// useWebSocket.ts — Persistent WebSocket connection to backend /stream.
//
// Opens a WebSocket to the Go server (server/broadcast.go handleWebSocket).
// On connect, the backend sends initial state (session, LP metrics, candles,
// trades, events). During the simulation, broadcast.go fans out real-time
// messages which this hook dispatches to registered handlers.
//
// MESSAGE FLOW:
//   backend executor callback → server.BroadcastTrade/Price/etc
//   → broadcast channel → runBroadcast → WebSocket frame
//   → this hook → page.tsx handleWSMessage → component state updates
//
// CONNECTIONS:
//   - Backend endpoint: server/broadcast.go handleWebSocket + sendInitialState
//   - Consumer:         page.tsx registers handleWSMessage as the onMessage callback
//   - Message types:    "trade", "price", "lp_metrics", "account_update",
//                       "key_event", "session_state", "candles", "trades", "events"
'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import type { WSMessage } from '@/types';

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/stream';

type MessageHandler = (message: WSMessage) => void;

export function useWebSocket(onMessage?: MessageHandler) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
  const handlersRef = useRef<Set<MessageHandler>>(new Set());

  // Add a message handler
  const addHandler = useCallback((handler: MessageHandler) => {
    handlersRef.current.add(handler);
  }, []);

  // Remove a message handler
  const removeHandler = useCallback((handler: MessageHandler) => {
    handlersRef.current.delete(handler);
  }, []);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    try {
      const ws = new WebSocket(WS_URL);

      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
        setIsConnected(false);
        
        // Reconnect after 2 seconds
        reconnectTimeoutRef.current = setTimeout(connect, 2000);
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          setLastMessage(message);
          
          // Call all registered handlers
          handlersRef.current.forEach((handler) => handler(message));
          
          // Call the direct handler if provided
          onMessage?.(message);
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e);
        }
      };

      wsRef.current = ws;
    } catch (e) {
      console.error('Failed to connect WebSocket:', e);
      reconnectTimeoutRef.current = setTimeout(connect, 2000);
    }
  }, [onMessage]);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  return {
    isConnected,
    lastMessage,
    addHandler,
    removeHandler,
    reconnect: connect,
  };
}
