// backend.ts - Canonical API / WebSocket endpoints for eth-amm-sim.
//
// Hosted topology (production):
//   Browser (https://eth-amm-sim.vercel.app)
//     REST → same-origin /api/*  (Vercel rewrite)
//          → https://api-staging-eth-amm-sim.magro.dev  (Cloudflare Tunnel)
//          → 127.0.0.1:8103 Go + 127.0.0.1:11545 Anvil on Ubuntu
//     WS   → wss://api-staging-eth-amm-sim.magro.dev/stream
//          Vercel static rewrites cannot proxy WebSockets, so the browser
//          opens the tunnel hostname for /stream only.
//
// Local topology (`npm run dev` / `make up`):
//   Browser → Vite /api proxy → http://127.0.0.1:8080
//   WebSocket → ws://127.0.0.1:8080/stream

/** Public Cloudflare Tunnel origin (Ubuntu Go). Used for WebSocket only in production. */
export const PUBLIC_TUNNEL_ORIGIN = 'https://api-staging-eth-amm-sim.magro.dev';

/** @deprecated Use PUBLIC_TUNNEL_ORIGIN. Kept so older comments/imports still compile. */
export const PUBLIC_API_ORIGIN = PUBLIC_TUNNEL_ORIGIN;

/** Public WebSocket URL (Ubuntu via Cloudflare Tunnel). */
export const PUBLIC_WS_URL = `${PUBLIC_TUNNEL_ORIGIN.replace(/^http/i, 'ws')}/stream`;

/** Hosted Vite SPA — also the public REST origin via /api/*. */
export const PUBLIC_UI_ORIGIN = 'https://eth-amm-sim.vercel.app';

/** Default local Go listen address used by `make backend` / Vite proxy. */
export const LOCAL_API_ORIGIN = 'http://127.0.0.1:8080';

export const LOCAL_WS_URL = 'ws://127.0.0.1:8080/stream';

export function isLocalHostname(hostname: string): boolean {
  return (
    hostname === 'localhost' ||
    hostname === '127.0.0.1' ||
    hostname === '[::1]' ||
    hostname.endsWith('.local')
  );
}

/**
 * Resolve the WebSocket URL.
 * Priority: VITE_WS_URL → local default on localhost → public Ubuntu tunnel.
 */
export function getWebSocketUrl(): string {
  const fromEnv = import.meta.env.VITE_WS_URL?.trim();
  if (fromEnv) {
    return fromEnv;
  }
  if (typeof window === 'undefined') {
    return LOCAL_WS_URL;
  }
  if (isLocalHostname(window.location.hostname)) {
    return LOCAL_WS_URL;
  }
  return PUBLIC_WS_URL;
}

/**
 * Public REST origin shown in the UI.
 * Hosted SPA uses the Vercel hostname; local uses the Vite origin.
 */
export function getPublicRestOrigin(): string {
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin;
  }
  return PUBLIC_UI_ORIGIN;
}

/** Short label for the connection chip in the dashboard. */
export function describeBackendTarget(): string {
  if (typeof window !== 'undefined' && !isLocalHostname(window.location.hostname)) {
    return window.location.host;
  }
  const ws = getWebSocketUrl();
  if (ws.includes('api-staging-eth-amm-sim.magro.dev')) {
    return 'eth-amm-sim.vercel.app';
  }
  if (ws.includes('127.0.0.1') || ws.includes('localhost')) {
    return 'local backend';
  }
  try {
    return new URL(ws).host;
  } catch {
    return 'remote backend';
  }
}

/**
 * REST base for direct calls. Empty means same-origin `/api/*`
 * (Vite proxy locally, Vercel rewrite in production).
 */
export function getDirectApiBase(): string {
  return (import.meta.env.VITE_API_URL || '').trim().replace(/\/$/, '');
}
