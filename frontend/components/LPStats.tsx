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
    return <div className="text-gray-500 text-center py-4">Loading…</div>;
  }

  const f = (v: number | undefined | null) => {
    if (v === undefined || v === null || isNaN(v)) return '0.00';
    return v.toFixed(2);
  };
  const pct = (v: number | undefined | null) => {
    if (v === undefined || v === null || isNaN(v)) return '+0.00%';
    return `${v >= 0 ? '+' : ''}${v.toFixed(2)}%`;
  };
  const isPositive = (v: number | undefined | null): boolean => {
    if (v === undefined || v === null || isNaN(v)) return false;
    return v >= 0;
  };

  return (
    <div className="bg-surface rounded-lg border border-border p-4 grid grid-cols-2 gap-3">
      <MetricCard
        label="LP Value"
        value={`${f(metrics.lpValue)} ETH`}
        subValue={`HODL: ${f(metrics.hodlValue)} ETH`}
      />

      <MetricCard
        label="Impermanent Loss (Price Only)"
        value={`${f(metrics.theoreticalIL)} ETH`}
        subValue="AMM rebalancing drag"
        positive={false}
      />

      <MetricCard
        label="LP vs HODL PnL"
        value={`${f(metrics.lpVsHodlPnL)} ETH`}
        subValue="Path dependent"
        positive={isPositive(metrics.lpVsHodlPnL)}
      />

      <MetricCard
        label="Fees Earned"
        value={`${f(metrics.totalFeesUSD)} ETH`}
        positive={isPositive(metrics.totalFeesUSD)}
      />

      <div className="col-span-2">
        <MetricCard
          label="Net PnL"
          value={`${f(metrics.netPnL)} ETH`}
          subValue={pct(metrics.netPnLPercent)}
          positive={isPositive(metrics.netPnL)}
        />
      </div>
    </div>
  );
}
