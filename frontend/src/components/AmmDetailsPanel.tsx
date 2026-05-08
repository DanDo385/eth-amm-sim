import type { Trade } from '@/types';

interface AmmDetailsPanelProps {
  trade: Trade | null;
}

const toToken = (raw?: string | null): number => {
  if (!raw) return 0;
  return Number(raw) / 1e18;
};

const fmt = (n: number, d = 4) => n.toLocaleString('en-US', { maximumFractionDigits: d });

export const AmmDetailsPanel = ({ trade }: AmmDetailsPanelProps) => {
  if (!trade) {
    return (
      <div className="bg-surface rounded-lg border border-border p-4">
        <h3 className="text-sm font-medium text-white">AMM Details</h3>
        <p className="mt-2 text-xs text-gray-400">
          Click a trade entry in <span className="text-gray-300">Key Events</span> to inspect before/after spot and reserve impact.
        </p>
      </div>
    );
  }

  const beforeSpot = toToken(trade.priceBefore);
  const afterSpot = toToken(trade.priceAfter);
  const beforeAppl = toToken(trade.reservesBeforeAPPL);
  const beforeEth = toToken(trade.reservesBeforeETH);
  const inAmount = toToken(trade.amountIn);
  const outAmount = toToken(trade.amountOut);
  const applDelta = trade.isBuy ? -outAmount : inAmount;
  const ethDelta = trade.isBuy ? inAmount : -outAmount;
  const afterAppl = beforeAppl + applDelta;
  const afterEth = beforeEth + ethDelta;
  const pct = beforeSpot > 0 ? ((afterSpot - beforeSpot) / beforeSpot) * 100 : 0;
  const bps = pct * 100;
  const applSize = trade.isBuy ? outAmount : inAmount;
  const summary = `${trade.nickname.toLowerCase()}-${trade.isBuy ? 'buy' : 'sell'}-${fmt(applSize, 2)}-appl-${fmt(afterSpot, 4)}`;

  return (
    <div className="bg-surface rounded-lg border border-border p-4">
      <h3 className="text-sm font-medium text-white">AMM Details:</h3>
      <div className="mt-2 text-xs text-cyan-300 font-mono break-all">{summary}</div>

      <div className="mt-3 grid grid-cols-2 gap-3 text-xs">
        <div className="rounded border border-border bg-surface-light p-3">
          <div className="text-gray-400">Market Before</div>
          <div className="mt-1 text-gray-200">Spot: {fmt(beforeSpot, 4)} ETH/APPL</div>
          <div className="text-gray-200">Reserves: {fmt(beforeAppl, 2)} APPL / {fmt(beforeEth, 2)} ETH</div>
        </div>
        <div className="rounded border border-border bg-surface-light p-3">
          <div className="text-gray-400">Market After</div>
          <div className="mt-1 text-gray-200">Spot: {fmt(afterSpot, 4)} ETH/APPL</div>
          <div className="text-gray-200">Reserves: {fmt(afterAppl, 2)} APPL / {fmt(afterEth, 2)} ETH</div>
        </div>
      </div>

      <div className="mt-3 text-xs text-gray-300">
        Spot impact:{' '}
        <span className={pct >= 0 ? 'text-green-400' : 'text-red-400'}>
          {pct >= 0 ? '+' : ''}{fmt(pct, 3)}% ({pct >= 0 ? '+' : ''}{fmt(bps, 1)} bps)
        </span>
      </div>
    </div>
  );
};
