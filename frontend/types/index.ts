// types/index.ts — TypeScript interfaces mirroring Go backend JSON structs.
//
// Each interface here corresponds to a struct serialized by the backend:
//   SessionState   ← engine/session.go    SessionState
//   Trade          ← engine/types.go      Trade
//   Candle         ← metrics/price.go     Candle
//   LPMetrics      ← metrics/lp.go        LPData / LPSnapshot
//   PerformanceData← metrics/account.go   AccountPerformance
//   ImpactPoint    ← metrics/impact.go    ImpactPoint
//   KeyEvent       ← store/memory.go      KeyEvent
//   WSMessage      ← server/server.go     WSMessage
//
// CONNECTIONS:
//   - Consumed by: lib/api.ts (return types), hooks/, components/
//   - Produced by: Go backend JSON serialization via encoding/json

// Session types
export interface SessionState {
  status: 'idle' | 'running' | 'completed' | 'error';
  startedAt?: string;
  endedAt?: string;
  duration: number;
  elapsed: number;
  error?: string;
}

// Trade types
export interface Trade {
  txHash: string;
  trader: string;
  nickname: string;
  isBuy: boolean;
  amountIn: string;
  amountOut?: string;
  price?: string;
  fee?: string;
  timestamp: string;
  blockNum?: number;
}

// Candle (OHLC) types
export interface Candle {
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  timestamp: string;
}

// LP Metrics types
export interface LPMetrics {
  initialApples: number;
  initialETH: number;
  currentApples: number;
  currentETH: number;
  currentPrice: number;
  lpValue: number;
  hodlValue: number;
  impermanentLoss: number;
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
  il: number;
  feesEarned: number;
  netPnL: number;
}

// Account Performance types
export interface PerformanceData {
  nickname: string;
  address: string;
  totalReturn: number;
  totalPnL: number;
  volatility: number;
  sharpeRatio: number;
  maxDrawdown: number;
  winRate: number;
  tradeCount: number;
  equityCurve: EquityPoint[];
  trades: TradeRecord[];
}

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
  pnl: number;
  equity: number;
}

// Impact curve types
export interface ImpactPoint {
  tradeSize: number;
  impactBps: number;
  executePrice: number;
  spotPrice: number;
}

export interface ImpactCurve {
  buy: ImpactPoint[];
  sell: ImpactPoint[];
}

// Key event types
export interface KeyEvent {
  timestamp: string;
  type: 'trade' | 'strategy_trigger';
  description: string;
  severity: 'info' | 'warning' | 'critical';
}

// WebSocket message types
export interface WSMessage {
  type: 'trade' | 'price' | 'lp_metrics' | 'account_update' | 'key_event' | 'session_state' | 'candles' | 'trades' | 'events';
  data: unknown;
}

// User trading types
export interface UserBalance {
  ethBalance: string;
  appleBalance: string;
}

export interface TradeResponse {
  txHash: string;
  ethAmount?: string;
  appleAmount?: string;
  ethBalance: string;
  appleBalance: string;
  status: string;
}
