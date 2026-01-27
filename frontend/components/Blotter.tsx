'use client';

import { useEffect, useRef, useMemo } from 'react';
import type { Trade } from '@/types';

interface BlotterProps {
  trades: Trade[];
  maxRows?: number; // Optional limit, undefined means show all trades
  height?: number; // Height in pixels (defaults to 350 to match PriceChart)
}

export function Blotter({ trades, maxRows, height = 350 }: BlotterProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const prevLengthRef = useRef(0);
  
  // Auto-scroll to bottom only when new trades are added (not on every render)
  useEffect(() => {
    if (trades.length > prevLengthRef.current && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
    prevLengthRef.current = trades.length;
  }, [trades.length]);
  
  // Memoize display trades - show all if maxRows not specified, otherwise limit
  const displayTrades = useMemo(() => {
    if (maxRows === undefined || maxRows <= 0) {
      return trades; // Show all trades
    }
    return trades.slice(-maxRows); // Show last N trades
  }, [trades, maxRows]);

  const formatAmount = (amount: string | undefined) => {
    if (!amount) return '-';
    const num = parseFloat(amount) / 1e18;
    return num.toFixed(4);
  };

  const formatPrice = (trade: Trade) => {
    // If price is explicitly provided, use it
    if (trade.price) {
      const priceNum = parseFloat(trade.price) / 1e18;
      return priceNum.toFixed(6);
    }
    
    // Otherwise calculate from amountIn and amountOut
    if (trade.amountIn && trade.amountOut) {
      const amountIn = parseFloat(trade.amountIn) / 1e18;
      const amountOut = parseFloat(trade.amountOut) / 1e18;
      
      if (amountIn > 0 && amountOut > 0) {
        if (trade.isBuy) {
          // Buy: price = ETH spent / APPL received
          return (amountIn / amountOut).toFixed(6);
        } else {
          // Sell: price = ETH received / APPL sold
          return (amountOut / amountIn).toFixed(6);
        }
      }
    }
    
    return '-';
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString();
  };

  return (
    <div className="bg-surface rounded-lg border border-border">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-sm font-medium text-white">Trade Blotter</h3>
      </div>
      <div 
        ref={containerRef}
        className="overflow-auto"
        style={{ height: `${height}px` }}
      >
        <table className="w-full text-sm">
          <thead className="text-gray-400 text-xs uppercase bg-surface-light sticky top-0">
            <tr>
              <th className="px-3 py-2 text-left">Time</th>
              <th className="px-3 py-2 text-left">Account</th>
              <th className="px-3 py-2 text-left">Side</th>
              <th className="px-3 py-2 text-right">Size</th>
              <th className="px-3 py-2 text-right">Price</th>
              <th className="px-3 py-2 text-right">Tx</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {displayTrades.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-3 py-8 text-center text-gray-500">
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
                  <td className="px-3 py-2 text-right text-gray-300 font-mono text-xs">
                    {formatPrice(trade)}
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
