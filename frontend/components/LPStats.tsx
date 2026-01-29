// LPStats.tsx — Liquidity provider performance metrics display.
//
// Shows pool reserves, impermanent loss, fees earned, and net PnL.
// Data flows from the backend's metrics/lp.go calculations, which track
// how the LP position performs vs simply holding the initial assets.
//
// CONNECTIONS:
//   - Backend data:  metrics/lp.go LPData (IL, fees, net PnL calculations)
//   - WebSocket msg: "lp_metrics" from broadcast.go BroadcastLPMetrics
//   - REST fallback: GET /lp/metrics via hooks/usePriceData.ts
//   - Types:         types/index.ts LPMetrics, LPSnapshot
'use client';

import type { LPMetrics } from '@/types';

interface LPStatsProps {
  metrics: LPMetrics | null;
}

function MetricCard({ label, value, subValue, positive }: {
  label: string;
  value: string;
  subValue?: string;
  positive?: boolean;
}) {
  return (
    <div className="bg-surface-light rounded-lg p-3">
      <div className="text-xs text-gray-400 mb-1">{label}</div>
      <div className={`text-lg font-semibold ${
        positive === undefined ? 'text-white' :
        positive ? 'text-green-500' : 'text-red-500'
      }`}>
        {value}
      </div>
      {subValue && (
        <div className="text-xs text-gray-500 mt-0.5">{subValue}</div>
      )}
    </div>
  );
}

export function LPStats({ metrics }: LPStatsProps) {
  if (!metrics) {
    return (
      <div className="bg-surface rounded-lg border border-border p-4">
        <h3 className="text-sm font-medium text-white mb-4">LP Metrics</h3>
        <div className="text-gray-500 text-center py-4">Loading...</div>
      </div>
    );
  }

  const formatETH = (value: number) => value.toFixed(2);
  const formatPercent = (value: number) => `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
  const formatPrice = (value: number) => value.toFixed(4);

  return (
    <div className="bg-surface rounded-lg border border-border">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-sm font-medium text-white">LP Metrics</h3>
        <span className="text-xs text-gray-400">
          Spot: {formatPrice(metrics.currentPrice)} ETH
        </span>
      </div>
      
      <div className="p-4 grid grid-cols-2 gap-3">
        <MetricCard
          label="Pool Reserves"
          value={`${formatETH(metrics.currentApples)} APPL`}
          subValue={`${formatETH(metrics.currentETH)} ETH`}
        />
        
        <MetricCard
          label="LP Value"
          value={`${formatETH(metrics.lpValue)} ETH`}
          subValue={`HODL: ${formatETH(metrics.hodlValue)} ETH`}
        />
        
        <MetricCard
          label="Impermanent Loss"
          value={`${formatETH(metrics.impermanentLoss)} ETH`}
          positive={false}
        />
        
        <MetricCard
          label="Fees Earned"
          value={`${formatETH(metrics.totalFeesUSD)} ETH`}
          positive={metrics.totalFeesUSD > 0}
        />
        
        <div className="col-span-2">
          <MetricCard
            label="Net PnL"
            value={`${formatETH(metrics.netPnL)} ETH`}
            subValue={formatPercent(metrics.netPnLPercent)}
            positive={metrics.netPnL >= 0}
          />
        </div>
      </div>
    </div>
  );
}
