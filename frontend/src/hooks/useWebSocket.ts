// useWebSocket.ts — Persistent WebSocket connection to backend /stream.
//
// URL resolution:
//   - VITE_WS_URL set → use it (required for HTTPS UI → wss backend).
//   - Otherwise in the browser → ws(s)://<current-host>:8080/stream so LAN /
//     phone / tunnel hits the same machine as the UI host (Go listens on :8080).
//
// CONNECTIONS:
//   - Backend: server/broadcast.go handleWebSocket + sendInitialState
//   - Consumer: page.tsx handleWSMessage

import { useEffect, useRef, useState, useCallback } from 'react';
import type { WSMessage } from '@/types';

function getWebSocketUrl(): string {
  const env = import.meta.env.VITE_WS_URL?.trim();
  if (env) {
    return env;
  }
  if (typeof window === 'undefined') {
    return 'ws://127.0.0.1:8080/stream';
  }
  const { protocol, hostname } = window.location;
  const wsProto = protocol === 'https:' ? 'wss:' : 'ws:';
  return `${wsProto}//${hostname}:8080/stream`;
}

type MessageHandler = (message: WSMessage) => void;

export function useWebSocket(onMessage?: MessageHandler) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
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
      const ws = new WebSocket(getWebSocketUrl());

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
