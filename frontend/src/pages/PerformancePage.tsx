// PerformancePage.tsx - Per-account performance analytics view.
//
// Pulls AccountPerformance from REST/WS and renders equity curves plus trade
// blotter rows. Metrics (Sharpe, drawdown, PnL) are computed in Go.
//
// CONNECTIONS:
//  - Backend: metrics/account.go → GET /accounts, /accounts/{nick}/performance
//  - WS:      "performance" / account updates via useWebSocket
//  - Types:   types/index.ts PerformanceData, EquityPoint, TradeRecord

import { useState, useEffect, useCallback, useRef } from 'react';
import { useWebSocket } from '@/hooks/useWebSocket';
import type { PerformanceData, WSMessage, EquityPoint } from '@/types';
import * as api from '@/lib/api';

// Simple equity chart component (inline SVG; no financial calc beyond pathing).
const EquityChart = ({ data, height = 200 }: { data: EquityPoint[]; height?: number }) => {
  if (data.length < 2) {
    return (
      <div 
        className="bg-surface-light rounded flex items-center justify-center text-gray-500"
        style={{ height }}
      >
        Not enough data for chart
      </div>
    );
  }

  const values = data.map(d => d.equity);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;

  // Create SVG path
  const width = 100;
  const points = data.map((d, i) => {
    const x = (i / (data.length - 1)) * width;
    const y = ((max - d.equity) / range) * height;
    return `${x},${y}`;
  }).join(' ');

  const isPositive = values[values.length - 1] >= values[0];

  return (
    <div className="bg-surface-light rounded p-2" style={{ height: height + 20 }}>
      <svg width="100%" height={height} viewBox={`0 0 ${width} ${height}`} preserveAspectRatio="none">
        <polyline
          points={points}
          fill="none"
          stroke={isPositive ? '#22c55e' : '#ef4444'}
          strokeWidth="2"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
    </div>
  );
};

// Drawdown chart component
const DrawdownChart = ({ data, height = 100 }: { data: EquityPoint[]; height?: number }) => {
  if (data.length < 2) {
    return (
      <div 
        className="bg-surface-light rounded flex items-center justify-center text-gray-500"
        style={{ height }}
      >
        Not enough data for chart
      </div>
    );
  }

  const values = data.map(d => d.drawdown * 100);
  const maxDD = Math.max(...values) || 1;

  const width = 100;
  const points = data.map((d, i) => {
    const x = (i / (data.length - 1)) * width;
    const y = (d.drawdown * 100 / maxDD) * height;
    return `${x},${y}`;
  }).join(' ');

  return (
    <div className="bg-surface-light rounded p-2" style={{ height: height + 20 }}>
      <svg width="100%" height={height} viewBox={`0 0 ${width} ${height}`} preserveAspectRatio="none">
        <polyline
          points={points}
          fill="none"
          stroke="#ef4444"
          strokeWidth="2"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
    </div>
  );
};

// KPI Tile component
const KPITile = ({ label, value, subValue, positive, large }: {
  label: string;
  value: string;
  subValue?: string;
  positive?: boolean;
  large?: boolean;
}) => {
  return (
    <div className={`bg-surface rounded-lg border border-border p-4 ${large ? 'col-span-2' : ''}`}>
      <div className="text-sm text-gray-400 mb-1">{label}</div>
      <div className={`${large ? 'text-3xl' : 'text-2xl'} font-bold ${
        positive === undefined ? 'text-white' :
        positive ? 'text-green-500' : 'text-red-500'
      }`}>
        {value}
      </div>
      {subValue && (
        <div className="text-sm text-gray-500 mt-1">{subValue}</div>
      )}
    </div>
  );
};

const PerformancePage = () => {
  const [accounts, setAccounts] = useState<PerformanceData[]>([]);
  const [selected, setSelected] = useState<string>('');
  const [performance, setPerformance] = useState<PerformanceData | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Re-fetch trigger: bumped when the session stops so the REST call
  // picks up finalized data (close prices, recalculated PnL).
  const [fetchTrigger, setFetchTrigger] = useState(0);

  // Use a ref for selected so the WS callback identity stays stable
  // and doesn't cause WebSocket reconnections when the dropdown changes.
  const selectedRef = useRef(selected);
  useEffect(() => { selectedRef.current = selected; }, [selected]);

  // Handle WebSocket messages - stable identity (no deps on selected)
  const handleWSMessage = useCallback((message: WSMessage) => {
    if (message.type === 'account_update') {
      const data = message.data as PerformanceData;
      if (data.nickname === selectedRef.current) {
        setPerformance(data);
      }
    }
    // When the session stops, trigger a re-fetch so we pick up
    // finalized close prices and recalculated PnL/Sharpe.
    // 1s delay ensures the backend has finished finalization before we fetch.
    if (message.type === 'session_state') {
      const state = message.data as { status?: string };
      if (state.status === 'completed' || state.status === 'idle') {
        setTimeout(() => setFetchTrigger((n) => n + 1), 1000);
      }
    }
  }, []);

  useWebSocket(handleWSMessage);

  // Fetch accounts list
  useEffect(() => {
    api.getAccounts().then(setAccounts).catch(console.error);
  }, []);

  // Fetch selected account performance (also re-runs when session stops)
  useEffect(() => {
    if (!selected) return;

    setIsLoading(true);
    api.getAccountPerformance(selected)
      .then(setPerformance)
      .catch(console.error)
      .finally(() => setIsLoading(false));
  }, [selected, fetchTrigger]);

  const formatPercent = (value: number) => `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
  const formatNumber = (value: number) => value.toFixed(2);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Performance Analytics</h1>
        <p className="text-gray-400 text-sm mt-1">
          TradFi-style risk metrics and performance attribution
        </p>
      </div>

      {/* Account Selector */}
      <div className="bg-surface rounded-lg border border-border p-4">
        <label className="block text-sm text-gray-400 mb-2">Select Account</label>
        <select
          value={selected}
          onChange={(e) => setSelected(e.target.value)}
          className="w-full max-w-md bg-surface-light border border-border rounded px-4 py-2 text-white focus:outline-none focus:border-blue-500"
        >
          <option value="">Choose an account...</option>
          {accounts.map((acc) => (
            <option key={acc.nickname} value={acc.nickname}>
              {acc.nickname} ({acc.tradeCount} trades)
            </option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">Loading performance data...</div>
      ) : performance ? (
        <>
          {/* KPI Grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <KPITile
              label="Trading Return"
              value={formatPercent(performance.totalReturn)}
              subValue="Trading PnL / starting equity"
              positive={performance.totalReturn >= 0}
              large
            />
            <KPITile
              label="Trading PnL"
              value={`${formatNumber(performance.tradingPnL)} ETH`}
              subValue="Sum of per-trade PnL"
              positive={performance.tradingPnL >= 0}
              large
            />
            <KPITile
              label="Position PnL"
              value={`${formatNumber(performance.positionPnL)} ETH`}
              subValue="APPL inventory mark-to-market"
              positive={performance.positionPnL >= 0}
            />
            <KPITile
              label="Total PnL"
              value={`${formatNumber(performance.totalPnL)} ETH`}
              subValue="Position + Trading"
              positive={performance.totalPnL >= 0}
            />
            <KPITile
              label="Sharpe Ratio"
              value={formatNumber(performance.sharpeRatio)}
              subValue="Trading return / vol"
              positive={performance.sharpeRatio > 0}
            />
            <KPITile
              label="Volatility"
              value={`${formatNumber(performance.volatility)}%`}
              subValue="Session realized vol"
            />
            <KPITile
              label="Max Drawdown"
              value={`-${formatNumber(performance.maxDrawdown)}%`}
              subValue="Peak to trough"
              positive={false}
            />
            <KPITile
              label="Win Rate"
              value={`${formatNumber(performance.winRate)}%`}
              subValue={`${performance.tradeCount} trades`}
              positive={performance.winRate > 50}
            />
          </div>

          {/* Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-surface rounded-lg border border-border">
              <div className="px-4 py-3 border-b border-border">
                <h3 className="text-sm font-medium text-white">Equity Curve</h3>
              </div>
              <div className="p-4">
                <EquityChart data={performance.equityCurve} height={200} />
              </div>
            </div>

            <div className="bg-surface rounded-lg border border-border">
              <div className="px-4 py-3 border-b border-border">
                <h3 className="text-sm font-medium text-white">Drawdown Over Time</h3>
              </div>
              <div className="p-4">
                <DrawdownChart data={performance.equityCurve} height={200} />
              </div>
            </div>
          </div>

          {/* Trade History */}
          <div className="bg-surface rounded-lg border border-border">
            <div className="px-4 py-3 border-b border-border">
              <h3 className="text-sm font-medium text-white">Trade History</h3>
            </div>
            <div className="overflow-auto max-h-64">
              <table className="w-full text-sm">
                <thead className="text-gray-400 text-xs uppercase bg-surface-light sticky top-0">
                  <tr>
                    <th className="px-4 py-2 text-left">Time</th>
                    <th className="px-4 py-2 text-left">Side</th>
                    <th className="px-4 py-2 text-right">Size</th>
                    <th className="px-4 py-2 text-right">Price</th>
                    <th className="px-4 py-2 text-right">Close</th>
                    <th className="px-4 py-2 text-right">PnL</th>
                    <th className="px-4 py-2 text-right">Equity</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {performance.trades.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                        No trades recorded
                      </td>
                    </tr>
                  ) : (
                    performance.trades.map((trade, i) => (
                      <tr 
                        key={i} 
                        className={trade.isBuy ? 'bg-green-900/10' : 'bg-red-900/10'}
                      >
                        <td className="px-4 py-2 text-gray-300">
                          {new Date(trade.timestamp).toLocaleTimeString()}
                        </td>
                        <td className={`px-4 py-2 font-medium ${trade.isBuy ? 'text-green-500' : 'text-red-500'}`}>
                          {trade.isBuy ? 'BUY' : 'SELL'}
                        </td>
                        <td className="px-4 py-2 text-right text-gray-300">
                          {trade.size.toFixed(4)}
                        </td>
                        <td className="px-4 py-2 text-right text-gray-300">
                          {trade.price.toFixed(4)}
                        </td>
                        <td className="px-4 py-2 text-right text-gray-300">
                          {trade.closePrice > 0 ? trade.closePrice.toFixed(4) : '-'}
                        </td>
                        <td className={`px-4 py-2 text-right font-medium ${trade.pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                          {trade.pnl >= 0 ? '+' : ''}{trade.pnl.toFixed(4)}
                        </td>
                        <td className="px-4 py-2 text-right text-white">
                          {trade.equity.toFixed(2)}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      ) : (
        <div className="bg-surface rounded-lg border border-border p-12 text-center">
          <div className="text-gray-500 mb-2">Select an account to view performance metrics</div>
          <p className="text-gray-600 text-sm">
            Performance analytics include Sharpe ratio, volatility, max drawdown, and equity curves
          </p>
        </div>
      )}

      {/* Footer */}
      <div className="text-center text-xs text-gray-500 pt-4 border-t border-border">
        <p>
          Metrics calculated using institutional risk management standards. 
          Sharpe ratio assumes risk-free rate = 0.
        </p>
      </div>
    </div>
  );
};

export default PerformancePage;
