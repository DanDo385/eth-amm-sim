'use client';

import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { SessionControls } from '@/components/SessionControls';
import { TradingPanel } from '@/components/TradingPanel';
import { PriceChart } from '@/components/PriceChart';
import { TWAPChart } from '@/components/TWAPChart';
import { Blotter } from '@/components/Blotter';
import { LPStats } from '@/components/LPStats';
import { KeyEvents } from '@/components/KeyEvents';
import { AccountMetrics } from '@/components/AccountMetrics';
import { ImpactCurve } from '@/components/ImpactCurve';
import { useSession } from '@/hooks/useSession';
import { usePriceData } from '@/hooks/usePriceData';
import { useWebSocket } from '@/hooks/useWebSocket';
import type { Trade, KeyEvent, WSMessage, Candle, LPMetrics, ImpactPoint } from '@/types';
import * as api from '@/lib/api';

export default function Dashboard() {
  const { session, isLoading, start, stop, reset, updateFromWS } = useSession();
  const { candles, lpMetrics, addCandle, updateLPMetrics, refresh: refreshPriceData } = usePriceData();
  const [trades, setTrades] = useState<Trade[]>([]);
  const [events, setEvents] = useState<KeyEvent[]>([]);
  const [impactData, setImpactData] = useState<{ buy: ImpactPoint[]; sell: ImpactPoint[] }>({ buy: [], sell: [] });
  const [priceRange, setPriceRange] = useState<{ min: number; max: number } | undefined>(undefined);
  const sessionStartRef = useRef<string | undefined>(undefined);

  // Track session start and handle reset
  useEffect(() => {
    if (session.startedAt && session.startedAt !== sessionStartRef.current) {
      // New session started - track start time but DON'T clear trades yet
      // Trades will accumulate throughout the session
      sessionStartRef.current = session.startedAt;
      refreshPriceData(); // Refresh LP metrics and candles
    } else if (session.status === 'idle' || !session.startedAt) {
      // Session is idle/reset - clear trades, events, and refresh LP metrics
      if (sessionStartRef.current !== undefined) {
        sessionStartRef.current = undefined;
        setTrades([]); // Clear trades only on reset
        setEvents([]); // Clear Key Events on reset
        refreshPriceData(); // Refresh LP metrics to get current state
      }
    }
  }, [session.startedAt, session.status, refreshPriceData]);

  // Filter trades to only show those from the current session
  const sessionTrades = useMemo(() => {
    if (!session.startedAt) {
      return []; // No session started, show no trades
    }

    const startTime = new Date(session.startedAt).getTime();
    return trades.filter((t) => {
      const tradeTime = new Date(t.timestamp).getTime();
      return tradeTime >= startTime;
    });
  }, [trades, session.startedAt]);

  // Handle WebSocket messages
  const handleWSMessage = useCallback((message: WSMessage) => {
    switch (message.type) {
      case 'session_state':
        updateFromWS(message.data as typeof session);
        break;
      case 'trade':
        setTrades((prev) => {
          // Append new trade - no limit, trades accumulate throughout session
          return [...prev, message.data as Trade];
        });
        break;
      case 'trades':
        setTrades(message.data as Trade[]);
        break;
      case 'price':
        addCandle(message.data as Candle);
        break;
      case 'lp_metrics':
        updateLPMetrics(message.data as LPMetrics);
        break;
      case 'key_event':
        setEvents((prev) => [...prev.slice(-49), message.data as KeyEvent]);
        break;
      case 'events':
        setEvents(message.data as KeyEvent[]);
        break;
    }
  }, [updateFromWS, addCandle, updateLPMetrics]);

  const { isConnected } = useWebSocket(handleWSMessage);

  // Fetch initial data
  useEffect(() => {
    // Fetch more trades to ensure we get all session trades
    api.getTrades(1000).then(setTrades).catch(console.error);
    api.getEvents(20).then(setEvents).catch(console.error);
    api.getImpactCurve().then(setImpactData).catch(console.error);
  }, []);

  // Refresh impact curve when LP metrics change
  useEffect(() => {
    if (lpMetrics) {
      api.getImpactCurve().then(setImpactData).catch(console.error);
    }
  }, [lpMetrics?.currentApples, lpMetrics?.currentETH]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Market Dashboard</h1>
          <p className="text-gray-400 text-sm mt-1">
            Real-time AMM simulation with heterogeneous agents
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <span className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="text-sm text-gray-400">
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
      </div>

      {/* Main Grid */}
      <div className="grid grid-cols-12 gap-6">
        {/* Left Column - Controls & LP Stats */}
        <div className="col-span-12 lg:col-span-3 space-y-6">
          <SessionControls
            session={session}
            isLoading={isLoading}
            onStart={start}
            onStop={stop}
            onReset={reset}
          />
          <TradingPanel />
          <LPStats metrics={lpMetrics} />
          <AccountMetrics />
        </div>

        {/* Center Column - Charts */}
        <div className="col-span-12 lg:col-span-6 space-y-6">
          <PriceChart candles={candles} session={session} height={350} onPriceRangeChange={setPriceRange} />
          <TWAPChart candles={candles} trades={sessionTrades} session={session} height={200} />
          <ImpactCurve 
            buyData={impactData.buy} 
            sellData={impactData.sell} 
            currentPrice={lpMetrics?.currentPrice}
          />
        </div>

        {/* Right Column - Blotter & Events */}
        <div className="col-span-12 lg:col-span-3 space-y-6">
          <Blotter trades={sessionTrades} height={350} />
          <KeyEvents events={events} />
        </div>
      </div>

      {/* Footer Info */}
      <div className="text-center text-xs text-gray-500 pt-4 border-t border-border">
        <p>
          Simulating market microstructure with Whale, Retail, and Strategy bots.
          All metrics computed in real-time.
        </p>
      </div>
    </div>
  );
}
