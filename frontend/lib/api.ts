// api.ts — REST client for the Go backend on :8080.
//
// URL resolution:
//   - NEXT_PUBLIC_API_URL set → use that host (explicit / remote deploy).
//   - Browser, unset → same-origin `/api/...` (Next rewrites to Go; works for
//     LAN IPs, tunnel hostnames, and Vercel when BACKEND_PROXY_URL is set).
//   - Server (SSR/build), unset → http://127.0.0.1:8080 (direct to local Go).
//
// CONNECTIONS:
//   - Backend routes:  server/server.go setupRoutes → server/handlers.go
//   - Next rewrites:    next.config.js /api/* → BACKEND_PROXY_URL or :8080

function restUrl(path: string): string {
  const direct = (process.env.NEXT_PUBLIC_API_URL || '').trim().replace(/\/$/, '');
  if (direct) {
    return `${direct}${path}`;
  }
  if (typeof window !== 'undefined') {
    return `/api${path}`;
  }
  return `http://127.0.0.1:8080${path}`;
}

// Generic fetch wrapper
async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(restUrl(path), {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

// Session API
export async function startSession(duration?: number) {
  return fetchAPI('/session/start', {
    method: 'POST',
    body: JSON.stringify({ duration }),
  });
}

export async function stopSession() {
  return fetchAPI('/session/stop', { method: 'POST' });
}

export async function resetSession(hardReset: boolean = false) {
  const url = hardReset ? '/session/reset?hard=true' : '/session/reset';
  return fetchAPI(url, { method: 'POST' });
}

export async function getSessionState() {
  return fetchAPI<import('@/types').SessionState>('/session/state');
}

// Account API
export async function getAccounts() {
  return fetchAPI<import('@/types').PerformanceData[]>('/accounts');
}

export async function getAccountPerformance(nickname: string) {
  return fetchAPI<import('@/types').PerformanceData>(`/accounts/${nickname}/performance`);
}

// LP API
export async function getLPMetrics() {
  return fetchAPI<import('@/types').LPMetrics>('/lp/metrics');
}

// Market data API
export async function getCandles() {
  return fetchAPI<import('@/types').Candle[]>('/candles');
}

export async function getTrades(limit?: number) {
  const params = new URLSearchParams();
  if (limit) {
    params.set('limit', limit.toString());
  }
  const query = params.toString();
  return fetchAPI<import('@/types').Trade[]>(`/trades${query ? `?${query}` : ''}`);
}

export async function getImpactCurve() {
  return fetchAPI('/impact-curve');
}

export async function getEvents(limit?: number) {
  const params = new URLSearchParams();
  if (limit) {
    params.set('limit', limit.toString());
  }
  const query = params.toString();
  return fetchAPI<import('@/types').KeyEvent[]>(`/events${query ? `?${query}` : ''}`);
}

// User trading API
export async function tradeBuy(ethAmount: string) {
  return fetchAPI<import('@/types').TradeResponse>('/trade/buy', {
    method: 'POST',
    body: JSON.stringify({ ethAmount }),
  });
}

export async function tradeSell(appleAmount: string) {
  return fetchAPI<import('@/types').TradeResponse>('/trade/sell', {
    method: 'POST',
    body: JSON.stringify({ appleAmount }),
  });
}

export async function getUserBalance() {
  return fetchAPI<import('@/types').UserBalance>('/user/balance');
}
