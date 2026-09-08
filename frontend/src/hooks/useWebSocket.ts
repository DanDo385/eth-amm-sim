// useWebSocket.ts - Persistent WebSocket connection to backend /stream.
//
// URL resolution (see lib/backend.ts):
//  - VITE_WS_URL → explicit override
//  - localhost → ws://127.0.0.1:8080/stream (local `make up`)
//  - otherwise → Ubuntu tunnel, then MBP backup if Ubuntu /healthz fails
//
// CONNECTIONS:
//  - Backend: server/broadcast.go handleWebSocket + sendInitialState
//  - Consumers: Dashboard / PerformancePage handleWSMessage

import { useEffect, useRef, useState, useCallback } from 'react';
import type { WSMessage } from '@/types';
import {
  describeBackendTarget,
  ensureHostedOrigin,
  getWebSocketUrl,
  resetHostedOriginProbe,
} from '@/lib/backend';

type MessageHandler = (message: WSMessage) => void;

export function useWebSocket(onMessage?: MessageHandler) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);
  const [backendTarget, setBackendTarget] = useState(() => describeBackendTarget());
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const handlersRef = useRef<Set<MessageHandler>>(new Set());
  const intentionalCloseRef = useRef(false);
  const connectingRef = useRef(false);

  const addHandler = useCallback((handler: MessageHandler) => {
    handlersRef.current.add(handler);
  }, []);

  const removeHandler = useCallback((handler: MessageHandler) => {
    handlersRef.current.delete(handler);
  }, []);

  const connect = useCallback(() => {
    if (
      connectingRef.current ||
      wsRef.current?.readyState === WebSocket.OPEN ||
      wsRef.current?.readyState === WebSocket.CONNECTING
    ) {
      return;
    }

    connectingRef.current = true;
    void (async () => {
      try {
        intentionalCloseRef.current = false;
        await ensureHostedOrigin();
        if (intentionalCloseRef.current) {
          connectingRef.current = false;
          return;
        }
        if (
          wsRef.current?.readyState === WebSocket.OPEN ||
          wsRef.current?.readyState === WebSocket.CONNECTING
        ) {
          connectingRef.current = false;
          return;
        }
        setBackendTarget(describeBackendTarget());
        const ws = new WebSocket(getWebSocketUrl());

        ws.onopen = () => {
          console.log('WebSocket connected', getWebSocketUrl());
          setIsConnected(true);
          setBackendTarget(describeBackendTarget());
        };

        ws.onclose = () => {
          if (!intentionalCloseRef.current) {
            console.log('WebSocket disconnected');
          }
          setIsConnected(false);

          if (!intentionalCloseRef.current) {
            resetHostedOriginProbe();
            reconnectTimeoutRef.current = setTimeout(connect, 2000);
          }
        };

        ws.onerror = (error) => {
          if (!intentionalCloseRef.current) {
            console.error('WebSocket error:', error);
          }
        };

        ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data);
            setLastMessage(message);
            handlersRef.current.forEach((handler) => handler(message));
            onMessage?.(message);
          } catch (e) {
            console.error('Failed to parse WebSocket message:', e);
          }
        };

        wsRef.current = ws;
        connectingRef.current = false;
      } catch (e) {
        connectingRef.current = false;
        if (!intentionalCloseRef.current) {
          console.error('Failed to connect WebSocket:', e);
          resetHostedOriginProbe();
          reconnectTimeoutRef.current = setTimeout(connect, 2000);
        }
      }
    })();
  }, [onMessage]);

  const disconnect = useCallback(() => {
    intentionalCloseRef.current = true;
    connectingRef.current = false;
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      if (
        wsRef.current.readyState === WebSocket.CONNECTING ||
        wsRef.current.readyState === WebSocket.OPEN
      ) {
        wsRef.current.close();
      }
      wsRef.current = null;
    }
  }, []);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  return {
    isConnected,
    backendTarget,
    lastMessage,
    addHandler,
    removeHandler,
    reconnect: connect,
  };
}
