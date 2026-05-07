// AccountMetrics.tsx — Per-bot performance dashboard (PnL, Sharpe, drawdown).
//
// Displays each bot's trading performance with equity curves and trade history.
// Data arrives two ways:
//   1. Initial load: GET /accounts via lib/api.ts
//   2. Real-time: WebSocket "account_update" → page.tsx updates accounts state
//
// CONNECTIONS:
//   - Backend data:  metrics/account.go AccountPerformance (computed per trade)
//   - REST endpoint: server/handlers.go handleGetAccounts
//   - WebSocket msg: "account_update" from broadcast.go BroadcastAccountUpdate
//   - Types:         types/index.ts PerformanceData, EquityPoint, TradeRecord

import { useState, useEffect } from 'react';
import type { PerformanceData } from '@/types';
import * as api from '@/lib/api';

interface AccountMetricsProps {
  accounts?: PerformanceData[];
  selectedNickname?: string;
  onSelect?: (nickname: string) => void;
}

export const AccountMetrics = ({ accounts: propAccounts, selectedNickname, onSelect }: AccountMetricsProps) => {
  const [accounts, setAccounts] = useState<PerformanceData[]>(propAccounts || []);
  const [selected, setSelected] = useState<string>(selectedNickname || '');
  const [performance, setPerformance] = useState<PerformanceData | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Fetch accounts list
  useEffect(() => {
    if (!propAccounts) {
      api.getAccounts().then(setAccounts).catch(console.error);
    }
  }, [propAccounts]);

  // Fetch selected account performance
  useEffect(() => {
    if (!selected) return;

    setIsLoading(true);
    api.getAccountPerformance(selected)
      .then(setPerformance)
      .catch(console.error)
      .finally(() => setIsLoading(false));
  }, [selected]);

  const handleSelect = (nickname: string) => {
    setSelected(nickname);
    onSelect?.(nickname);
  };

  const formatPercent = (value: number) => `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;

  return (
    <div className="bg-surface rounded-lg border border-border">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-sm font-medium text-white">Account Metrics</h3>
      </div>
      
      <div className="p-4">
        {/* Account selector */}
        <div className="mb-4">
          <select
            value={selected}
            onChange={(e) => handleSelect(e.target.value)}
            className="w-full bg-surface-light border border-border rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500"
          >
            <option value="">Select account...</option>
            {accounts.map((acc) => (
              <option key={acc.nickname} value={acc.nickname}>
                {acc.nickname}
              </option>
            ))}
          </select>
        </div>

        {/* Performance metrics */}
        {isLoading ? (
          <div className="text-center py-4 text-gray-500">Loading...</div>
        ) : performance ? (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="bg-surface-light rounded p-2">
                <div className="text-xs text-gray-400">Total Return</div>
                <div className={`font-semibold ${performance.totalReturn >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                  {formatPercent(performance.totalReturn)}
                </div>
              </div>
              <div className="bg-surface-light rounded p-2">
                <div className="text-xs text-gray-400">Sharpe Ratio</div>
                <div className="font-semibold text-white">
                  {performance.sharpeRatio.toFixed(2)}
                </div>
              </div>
              <div className="bg-surface-light rounded p-2">
                <div className="text-xs text-gray-400">Volatility</div>
                <div className="font-semibold text-white">
                  {formatPercent(performance.volatility * 100)}
                </div>
              </div>
              <div className="bg-surface-light rounded p-2">
                <div className="text-xs text-gray-400">Max Drawdown</div>
                <div className="font-semibold text-red-500">
                  -{performance.maxDrawdown.toFixed(2)}%
                </div>
              </div>
              <div className="bg-surface-light rounded p-2">
                <div className="text-xs text-gray-400">Win Rate</div>
                <div className="font-semibold text-white">
                  {performance.winRate.toFixed(1)}%
                </div>
              </div>
              <div className="bg-surface-light rounded p-2">
                <div className="text-xs text-gray-400">Trade Count</div>
                <div className="font-semibold text-white">
                  {performance.tradeCount}
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="text-center py-4 text-gray-500 text-sm">
            Select an account to view metrics
          </div>
        )}
      </div>
    </div>
  );
};
