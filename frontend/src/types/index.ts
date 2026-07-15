// types/index.ts - Shared frontend types

// --------------------
// Session types
// --------------------
export interface SessionState {
  status: 'idle' | 'running' | 'paused' | 'completed' | 'error';
  startedAt?: string;
  endedAt?: string;
  duration: number;
  elapsed: number;
  error?: string;
}

export type ResetMode = 'soft' | 'hard' | 'reseed';

// --------------------
// Trading types
// --------------------
export interface TradeResponse {
  txHash: string;
  ethAmount?: string;
  appleAmount?: string;
  ethBalance: string;
  appleBalance: string;
  status: string;
}

export interface UserBalance {
  ethBalance: string;
  appleBalance: string;
}

// --------------------
// LP Metrics types
// --------------------
export interface LPMetrics {
  initialApples: number;
  initialETH: number;
  currentApples: number;
  currentETH: number;
  currentPrice: number;

  lpValue: number;
  hodlValue: number;

  theoreticalIL: number;   // ETH (≤ 0)
  lpVsHodlPnL: number;     // ETH (+ / −)

  feesEarnedApple: number;
  feesEarnedETH: number;
  totalFeesUSD: number;

  netPnL: number;
  netPnLPercent: number;

  history: LPSnapshot[];
}

export interface LPSnapshot {
  timestamp: string;
  appleReserve: number;
  ethReserve: number;
  spotPrice: number;
  lpValue: number;
  hodlValue: number;
  theoreticalIL: number;
  lpVsHodlPnL: number;
  feesEarned: number;
  netPnL: number;
}

// --------------------
// Trade types
// --------------------
export interface Trade {
  txHash: string;
  trader: string;
  nickname: string;
  isBuy: boolean;
  amountIn: string;
  amountOut: string | null;
  price: string | null;
  fee: string | null;
  timestamp: string;
  blockNum: number;
  priceBefore: string | null;
  priceAfter: string | null;
  reservesBeforeETH: string | null;
  reservesBeforeAPPL: string | null;
}

// --------------------
// Price/Candle types
// --------------------
export interface Candle {
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  timestamp: string;
}

// --------------------
// Event types
// --------------------
export interface KeyEvent {
  timestamp: string;
  type: string;
  description: string;
  severity: string;
}

// --------------------
// WebSocket types
// --------------------
export interface WSMessage {
  type: string;
  data: any;
}

// --------------------
// Impact Curve types
// --------------------
export interface ImpactPoint {
  tradeSize: number;
  impactBps: number;
  executePrice: number;
  spotPrice: number;
}

// --------------------
// Account Performance types
// --------------------
export interface EquityPoint {
  timestamp: string;
  equity: number;
  drawdown: number;
}

export interface TradeRecord {
  timestamp: string;
  isBuy: boolean;
  size: number;
  price: number;
  closePrice: number;
  appleAmount: number;
  pnl: number;
  equity: number;
}

export interface PerformanceData {
  nickname: string;
  address: string;
  totalReturn: number;
  totalPnL: number;
  positionPnL: number;
  tradingPnL: number;
  volatility: number;
  sharpeRatio: number;
  maxDrawdown: number;
  winRate: number;
  tradeCount: number;
  equityCurve: EquityPoint[];
  trades: TradeRecord[];
}
