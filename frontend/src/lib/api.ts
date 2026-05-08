// api.ts — REST client for the Go backend on :8080.
//
// URL resolution:
//   - VITE_API_URL set → use that base URL (explicit / remote deploy).
//   - Browser, unset → same-origin `/api/...` (Vite dev/preview proxy → Go;
//     configure your host’s reverse proxy the same way for production static hosting).
//
// CONNECTIONS:
//   - Backend routes: server/server.go setupRoutes → server/handlers.go
//   - Dev proxy:       vite.config.ts `/api` → VITE_DEV_BACKEND_URL or :8080

function restUrl(path: string): string {
  const direct = (import.meta.env.VITE_API_URL || '').trim().replace(/\/$/, '');
  if (direct) {
    return `${direct}${path}`;
  }
  return `/api${path}`;
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

export async function pauseSession() {
  return fetchAPI('/session/pause', { method: 'POST' });
}

export async function resumeSession() {
  return fetchAPI('/session/resume', { method: 'POST' });
}

export async function resetSession(mode: import('@/types').ResetMode = 'soft') {
  const url = `/session/reset?mode=${mode}`;
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
