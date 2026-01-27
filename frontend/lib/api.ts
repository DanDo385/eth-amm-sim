// API client for the Go backend

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Generic fetch wrapper
async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
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
  const query = limit ? `?limit=${limit}` : '';
  return fetchAPI<import('@/types').Trade[]>(`/trades${query}`);
}

export async function getImpactCurve() {
  return fetchAPI<import('@/types').ImpactCurve>('/impact-curve');
}

export async function getEvents(limit?: number) {
  const query = limit ? `?limit=${limit}` : '';
  return fetchAPI<import('@/types').KeyEvent[]>(`/events${query}`);
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
