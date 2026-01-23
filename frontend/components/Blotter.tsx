'use client';

import { useEffect, useRef } from 'react';
import type { Trade } from '@/types';

interface BlotterProps {
  trades: Trade[];
  maxRows?: number;
}

export function Blotter({ trades, maxRows = 50 }: BlotterProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  
  // Auto-scroll to bottom on new trades
  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [trades.length]);

  const formatAmount = (amount: string | undefined) => {
    if (!amount) return '-';
    const num = parseFloat(amount) / 1e18;
    return num.toFixed(4);
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString();
  };

  const displayTrades = trades.slice(-maxRows);

  return (
    <div className="bg-surface rounded-lg border border-border">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-sm font-medium text-white">Trade Blotter</h3>
      </div>
      <div 
        ref={containerRef}
        className="overflow-auto max-h-64"
      >
        <table className="w-full text-sm">
          <thead className="text-gray-400 text-xs uppercase bg-surface-light sticky top-0">
            <tr>
              <th className="px-3 py-2 text-left">Time</th>
              <th className="px-3 py-2 text-left">Account</th>
              <th className="px-3 py-2 text-left">Side</th>
              <th className="px-3 py-2 text-right">Size</th>
              <th className="px-3 py-2 text-right">Tx</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {displayTrades.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-3 py-8 text-center text-gray-500">
                  No trades yet
                </td>
              </tr>
            ) : (
              displayTrades.map((trade, i) => (
                <tr 
                  key={trade.txHash || i} 
                  className={trade.isBuy ? 'trade-row-buy' : 'trade-row-sell'}
                >
                  <td className="px-3 py-2 text-gray-300">
                    {formatTime(trade.timestamp)}
                  </td>
                  <td className="px-3 py-2 text-white font-medium">
                    {trade.nickname}
                  </td>
                  <td className={`px-3 py-2 font-medium ${trade.isBuy ? 'text-green-500' : 'text-red-500'}`}>
                    {trade.isBuy ? 'BUY' : 'SELL'}
                  </td>
                  <td className="px-3 py-2 text-right text-gray-300">
                    {formatAmount(trade.amountIn)}
                  </td>
                  <td className="px-3 py-2 text-right">
                    <a 
                      href={`#${trade.txHash}`}
                      className="text-blue-400 hover:text-blue-300 font-mono text-xs"
                    >
                      {trade.txHash?.slice(0, 8)}...
                    </a>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
