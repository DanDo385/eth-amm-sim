// Dashboard.tsx — main client dashboard (loaded from app/page.tsx with ssr:false).
'use client';

import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { SessionControls } from '@/components/SessionControls';
import { LoomDemoDirector } from '@/components/LoomDemoDirector';
import { DemoArchitecturePanel } from '@/components/DemoArchitecturePanel';
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
import type { Trade, KeyEvent, WSMessage, Candle, LPMetrics, ImpactPoint, SessionState } from '@/types';
import * as api from '@/lib/api';

export default function Dashboard() {
  const { session, isLoading, error: sessionError, start, stop, reset, updateFromWS } = useSession();
  const { candles, lpMetrics, addCandle, updateLPMetrics, refresh: refreshPriceData } = usePriceData();
  const [trades, setTrades] = useState<Trade[]>([]);
  const [events, setEvents] = useState<KeyEvent[]>([]);
  const [impactData, setImpactData] = useState<{ buy: ImpactPoint[]; sell: ImpactPoint[] }>({ buy: [], sell: [] });
  const [, setPriceRange] = useState<{ min: number; max: number } | undefined>(undefined);
  const sessionStartRef = useRef<string | undefined>(undefined);

  const handleReset = useCallback(
    async (hardReset: boolean = false) => {
      setTrades([]);
      setEvents([]);
      await reset(hardReset);
      await refreshPriceData();
    },
    [reset, refreshPriceData]
  );

  useEffect(() => {
    if (session.startedAt && session.startedAt !== sessionStartRef.current) {
      sessionStartRef.current = session.startedAt;
      refreshPriceData();
    } else if (session.status === 'idle' || !session.startedAt) {
      if (sessionStartRef.current !== undefined) {
        sessionStartRef.current = undefined;
        setTrades([]);
        setEvents([]);
        refreshPriceData();
      }
    }
  }, [session.startedAt, session.status, refreshPriceData]);

  const sessionTrades = useMemo(() => {
    if (!session.startedAt) {
      return [];
    }
    const startTime = new Date(session.startedAt).getTime();
    return trades.filter((t) => {
      const tradeTime = new Date(t.timestamp).getTime();
      return tradeTime >= startTime;
    });
  }, [trades, session.startedAt]);

  const handleWSMessage = useCallback((message: WSMessage) => {
    switch (message.type) {
      case 'session_state':
        updateFromWS(message.data as SessionState);
        break;
      case 'trade':
        setTrades((prev) => [...prev, message.data as Trade]);
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
      case 'user_balance_reset':
        break;
    }
  }, [updateFromWS, addCandle, updateLPMetrics]);

  const { isConnected } = useWebSocket(handleWSMessage);

  useEffect(() => {
    api.getTrades(1000).then((data) => setTrades(data as Trade[])).catch(console.error);
    api.getEvents(20).then((data) => setEvents(data as KeyEvent[])).catch(console.error);
    api.getImpactCurve()
      .then((data) => setImpactData(data as { buy: ImpactPoint[]; sell: ImpactPoint[] }))
      .catch(console.error);
  }, []);

  // Backfill blotter after refresh or Start: the mount-only GET /trades can race before
  // session state (and startedAt) exists, so sessionTrades stays empty. Refetch whenever
  // we have a running session anchor so filters match the server store.
  useEffect(() => {
    if (session.status !== 'running' || !session.startedAt) return;
    api.getTrades(1000).then((data) => setTrades(data as Trade[])).catch(console.error);
  }, [session.status, session.startedAt]);

  useEffect(() => {
    if (lpMetrics) {
      api.getImpactCurve()
        .then((data) => setImpactData(data as { buy: ImpactPoint[]; sell: ImpactPoint[] }))
        .catch(console.error);
    }
  }, [lpMetrics?.currentApples, lpMetrics?.currentETH]);

  return (
    <div className="space-y-6">
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

      <LoomDemoDirector />

      <div className="grid grid-cols-15 gap-6">
        <div className="col-span-15 lg:col-span-3 space-y-6">
          <SessionControls
            session={session}
            isLoading={isLoading}
            error={sessionError}
            onStart={start}
            onStop={stop}
            onReset={handleReset}
          />
          <DemoArchitecturePanel />
          <TradingPanel session={session} />
          <LPStats metrics={lpMetrics} />
          <AccountMetrics />
        </div>

        <div className="col-span-15 lg:col-span-6 space-y-6">
          <PriceChart candles={candles} session={session} height={350} onPriceRangeChange={setPriceRange} />
          <TWAPChart candles={candles} trades={sessionTrades} session={session} height={200} />
          <ImpactCurve buyData={impactData.buy} sellData={impactData.sell} lpMetrics={lpMetrics} />
        </div>

        <div className="col-span-15 lg:col-span-6 space-y-6">
          <Blotter trades={sessionTrades} height={350} />
          <KeyEvents events={events} height={200} />
        </div>
      </div>

      <div className="text-center text-xs text-gray-500 pt-4 border-t border-border">
        <p>
          Simulating market microstructure with Whale, Retail, and Strategy bots.
          All metrics computed in real-time.
        </p>
      </div>
    </div>
  );
}
