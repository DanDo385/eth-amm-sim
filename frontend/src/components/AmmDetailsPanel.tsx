import type { Trade } from '@/types';

interface AmmDetailsPanelProps {
  trade: Trade | null;
}

const toToken = (raw?: string | null): number => {
  if (!raw) return 0;
  return Number(raw) / 1e18;
};

const fmt = (n: number, d = 4) => n.toLocaleString('en-US', { maximumFractionDigits: d });
const feeRateFallback = 0.003;

const quoteBuyOut = (ethIn: number, ethReserve: number, applReserve: number, feeRate: number) => {
  if (ethIn <= 0 || ethReserve <= 0 || applReserve <= 0) return 0;
  const inAfterFee = ethIn * (1 - feeRate);
  return (inAfterFee * applReserve) / (ethReserve + inAfterFee);
};

const quoteSellOut = (applIn: number, ethReserve: number, applReserve: number, feeRate: number) => {
  if (applIn <= 0 || ethReserve <= 0 || applReserve <= 0) return 0;
  const inAfterFee = applIn * (1 - feeRate);
  return (inAfterFee * ethReserve) / (applReserve + inAfterFee);
};

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
  const fee = toToken(trade.fee);
  const feeRate = inAmount > 0 && fee > 0 ? Math.max(0, Math.min(0.2, fee / inAmount)) : feeRateFallback;

  const execPrice = trade.isBuy
    ? (outAmount > 0 ? inAmount / outAmount : 0) // ETH/APPL paid
    : (inAmount > 0 ? outAmount / inAmount : 0); // ETH/APPL received
  const execSlipBps = beforeSpot > 0 ? ((execPrice - beforeSpot) / beforeSpot) * 10_000 : 0;

  // Build a 2-way quote from pre-trade reserves:
  // if this trade side was BUY, show what SELL would have looked like with equivalent notional (and vice versa).
  const oppositeEthIn = trade.isBuy ? 0 : inAmount * beforeSpot;
  const oppositeApplIn = trade.isBuy ? (beforeSpot > 0 ? inAmount / beforeSpot : 0) : 0;
  const oppositeOut = trade.isBuy
    ? quoteSellOut(oppositeApplIn, beforeEth, beforeAppl, feeRate)
    : quoteBuyOut(oppositeEthIn, beforeEth, beforeAppl, feeRate);
  const oppositeExecPrice = trade.isBuy
    ? (oppositeApplIn > 0 ? oppositeOut / oppositeApplIn : 0) // ETH/APPL received on sell
    : (oppositeOut > 0 ? oppositeEthIn / oppositeOut : 0); // ETH/APPL paid on buy
  const oppositeSlipBps = beforeSpot > 0 ? ((oppositeExecPrice - beforeSpot) / beforeSpot) * 10_000 : 0;

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

      <div className="mt-3 grid grid-cols-1 gap-3 text-xs">
        <div className="rounded border border-border bg-surface-light p-3">
          <div className="text-gray-400">Executed Trade Quality</div>
          <div className="mt-1 text-gray-200">
            Effective execution price: {fmt(execPrice, 5)} ETH/APPL
          </div>
          <div className="text-gray-200">
            Slippage vs pre-trade spot:{' '}
            <span className={execSlipBps > 0 ? 'text-red-400' : 'text-green-400'}>
              {execSlipBps > 0 ? '+' : ''}{fmt(execSlipBps, 1)} bps
            </span>
          </div>
          <div className="text-gray-500 mt-1">
            Fee estimate used in quote math: {fmt(feeRate * 100, 2)}%
          </div>
        </div>

        <div className="rounded border border-border bg-surface-light p-3">
          <div className="text-gray-400">Pre-Trade 2-Way Quote (What-if opposite side)</div>
          {trade.isBuy ? (
            <>
              <div className="mt-1 text-gray-200">
                If sold {fmt(oppositeApplIn, 2)} APPL instead:
                receive {` ${fmt(oppositeOut, 4)} ETH`}
              </div>
              <div className="text-gray-200">
                Opposite-side execution price: {fmt(oppositeExecPrice, 5)} ETH/APPL
              </div>
            </>
          ) : (
            <>
              <div className="mt-1 text-gray-200">
                If bought with {fmt(oppositeEthIn, 2)} ETH instead:
                receive {` ${fmt(oppositeOut, 4)} APPL`}
              </div>
              <div className="text-gray-200">
                Opposite-side execution price: {fmt(oppositeExecPrice, 5)} ETH/APPL
              </div>
            </>
          )}
          <div className="text-gray-200">
            Opposite-side slippage vs pre-trade spot:{' '}
            <span className={oppositeSlipBps > 0 ? 'text-red-400' : 'text-green-400'}>
              {oppositeSlipBps > 0 ? '+' : ''}{fmt(oppositeSlipBps, 1)} bps
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};
