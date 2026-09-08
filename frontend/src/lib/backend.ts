// backend.ts - Canonical API / WebSocket endpoints for eth-amm-sim.
//
// Hosted topology (production):
//   Browser (https://eth-amm-sim.vercel.app)
//     REST → same-origin /api/*  (Vercel rewrite → Ubuntu primary tunnel)
//          → fallback https://api-backup-eth-amm-sim.magro.dev (MBP) if primary /healthz fails
//     WS   → wss://api-staging-eth-amm-sim.magro.dev/stream (Ubuntu)
//          → wss://api-backup-eth-amm-sim.magro.dev/stream (MBP) if primary is down
//          Vercel static rewrites cannot proxy WebSockets.
//
// Local topology (`npm run dev` / `make up`):
//   Browser → Vite /api proxy → http://127.0.0.1:8080
//   WebSocket → ws://127.0.0.1:8080/stream
//   No MBP failover (local Go is the origin).

const HEALTH_PROBE_MS = 2500;

/** Ubuntu Cloudflare Tunnel — public primary origin. */
export const PRIMARY_TUNNEL_ORIGIN = 'https://api-staging-eth-amm-sim.magro.dev';

/** MacBook Cloudflare Tunnel — public backup origin (must not claim the primary hostname). */
export const BACKUP_TUNNEL_ORIGIN = 'https://api-backup-eth-amm-sim.magro.dev';

/** @deprecated Use PRIMARY_TUNNEL_ORIGIN. */
export const PUBLIC_TUNNEL_ORIGIN = PRIMARY_TUNNEL_ORIGIN;

/** @deprecated Use PRIMARY_TUNNEL_ORIGIN. */
export const PUBLIC_API_ORIGIN = PRIMARY_TUNNEL_ORIGIN;

export const PRIMARY_WS_URL = `${PRIMARY_TUNNEL_ORIGIN.replace(/^http/i, 'ws')}/stream`;
export const BACKUP_WS_URL = `${BACKUP_TUNNEL_ORIGIN.replace(/^http/i, 'ws')}/stream`;

/** Public WebSocket URL (Ubuntu primary). Prefer getWebSocketUrl() after ensureHostedOrigin(). */
export const PUBLIC_WS_URL = PRIMARY_WS_URL;

/** Hosted Vite SPA — also the public REST origin via /api/* while Ubuntu is healthy. */
export const PUBLIC_UI_ORIGIN = 'https://eth-amm-sim.vercel.app';

/** Default local Go listen address used by `make backend` / Vite proxy. */
export const LOCAL_API_ORIGIN = 'http://127.0.0.1:8080';

export const LOCAL_WS_URL = 'ws://127.0.0.1:8080/stream';

type HostedLane = 'primary' | 'backup';

let hostedLane: HostedLane = 'primary';
let originProbe: Promise<void> | null = null;

export function isLocalHostname(hostname: string): boolean {
  return (
    hostname === 'localhost' ||
    hostname === '127.0.0.1' ||
    hostname === '[::1]' ||
    hostname.endsWith('.local')
  );
}

async function probeHealth(url: string): Promise<boolean> {
  const ctrl = new AbortController();
  const timer = window.setTimeout(() => ctrl.abort(), HEALTH_PROBE_MS);
  try {
    const res = await fetch(url, { signal: ctrl.signal, cache: 'no-store' });
    return res.ok;
  } catch {
    return false;
  } finally {
    window.clearTimeout(timer);
  }
}

function envOverridesFailover(): boolean {
  return Boolean(import.meta.env.VITE_API_URL?.trim() || import.meta.env.VITE_WS_URL?.trim());
}

function shouldProbeHostedFailover(): boolean {
  if (typeof window === 'undefined') {
    return false;
  }
  if (envOverridesFailover()) {
    return false;
  }
  return !isLocalHostname(window.location.hostname);
}

async function selectHostedLane(): Promise<void> {
  if (await probeHealth('/api/healthz')) {
    hostedLane = 'primary';
    return;
  }
  if (await probeHealth(`${BACKUP_TUNNEL_ORIGIN}/healthz`)) {
    hostedLane = 'backup';
    return;
  }
  hostedLane = 'primary';
}

/**
 * Probe Ubuntu (via Vercel /api/healthz) then the MBP backup tunnel.
 * Safe to call repeatedly; concurrent callers share one probe.
 * Call resetHostedOriginProbe() after a dropped WebSocket so the next
 * connect can fail back to Ubuntu.
 */
export function ensureHostedOrigin(): Promise<void> {
  if (!shouldProbeHostedFailover()) {
    return Promise.resolve();
  }
  if (!originProbe) {
    originProbe = selectHostedLane();
  }
  return originProbe;
}

/** Drop the cached probe so the next ensureHostedOrigin() re-checks Ubuntu first. */
export function resetHostedOriginProbe(): void {
  originProbe = null;
}

/**
 * Resolve the WebSocket URL.
 * Priority: VITE_WS_URL → local default on localhost → probed hosted lane.
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
  if (hostedLane === 'backup') {
    return BACKUP_WS_URL;
  }
  return PRIMARY_WS_URL;
}

/**
 * Public REST origin shown in the UI.
 * Hosted SPA uses the Vercel hostname while Ubuntu is healthy; backup REST
 * is the MBP tunnel (browser must call it directly — Vercel rewrite stays on Ubuntu).
 */
export function getPublicRestOrigin(): string {
  if (typeof window !== 'undefined' && window.location?.origin) {
    if (hostedLane === 'backup' && shouldProbeHostedFailover()) {
      return BACKUP_TUNNEL_ORIGIN;
    }
    return window.location.origin;
  }
  return PUBLIC_UI_ORIGIN;
}

/** Short label for the connection chip in the dashboard. */
export function describeBackendTarget(): string {
  if (typeof window !== 'undefined' && !isLocalHostname(window.location.hostname)) {
    if (hostedLane === 'backup' && !envOverridesFailover()) {
      return 'eth-amm-sim.vercel.app · MBP backup';
    }
    return window.location.host;
  }
  const ws = getWebSocketUrl();
  if (ws.includes('api-backup-eth-amm-sim.magro.dev')) {
    return 'eth-amm-sim.vercel.app · MBP backup';
  }
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
 * (Vite proxy locally, Vercel rewrite to Ubuntu in production).
 * After failover, this is the MBP backup origin.
 */
export function getDirectApiBase(): string {
  const fromEnv = (import.meta.env.VITE_API_URL || '').trim().replace(/\/$/, '');
  if (fromEnv) {
    return fromEnv;
  }
  if (hostedLane === 'backup' && shouldProbeHostedFailover()) {
    return BACKUP_TUNNEL_ORIGIN;
  }
  return '';
}
