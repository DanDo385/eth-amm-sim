// Dashboard.tsx — main client dashboard (lazy-loaded from HomePage).

import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { SessionControls } from '@/components/SessionControls';
import { TradingPanel } from '@/components/TradingPanel';
import { PriceChart } from '@/components/PriceChart';
import { TWAPChart } from '@/components/TWAPChart';
import { Blotter } from '@/components/Blotter';
import { LPStats } from '@/components/LPStats';
import { KeyEvents } from '@/components/KeyEvents';
import { ImpactCurve } from '@/components/ImpactCurve';
import { AmmDetailsPanel } from '@/components/AmmDetailsPanel';
import { useSession } from '@/hooks/useSession';
import { usePriceData } from '@/hooks/usePriceData';
import { useWebSocket } from '@/hooks/useWebSocket';
import type { Trade, KeyEvent, WSMessage, Candle, LPMetrics, ImpactPoint, SessionState, ResetMode } from '@/types';
import * as api from '@/lib/api';

export default function Dashboard() {
  const { session, isLoading, error: sessionError, start, pause, resume, stop, reset, updateFromWS } = useSession();
  const { candles, lpMetrics, addCandle, updateLPMetrics, refresh: refreshPriceData } = usePriceData();
  const [trades, setTrades] = useState<Trade[]>([]);
  const [events, setEvents] = useState<KeyEvent[]>([]);
  const [impactData, setImpactData] = useState<{ buy: ImpactPoint[]; sell: ImpactPoint[] }>({ buy: [], sell: [] });
  const [selectedTrade, setSelectedTrade] = useState<Trade | null>(null);
  const [displaySessionStartedAt, setDisplaySessionStartedAt] = useState<string | undefined>(undefined);
  const [chartResetVersion, setChartResetVersion] = useState(0);
  const [userBalanceRefreshToken, setUserBalanceRefreshToken] = useState(0);
  const [, setPriceRange] = useState<{ min: number; max: number } | undefined>(undefined);
  const sessionStartRef = useRef<string | undefined>(undefined);

  const handleReset = useCallback(
    async (mode: ResetMode = 'soft') => {
      setTrades([]);
      setEvents([]);
      setSelectedTrade(null);
      setDisplaySessionStartedAt(undefined);
      setChartResetVersion((v) => v + 1);
      sessionStartRef.current = undefined;
      await reset(mode);
      await refreshPriceData();
    },
    [reset, refreshPriceData]
  );

  useEffect(() => {
    if (session.startedAt && session.startedAt !== sessionStartRef.current) {
      sessionStartRef.current = session.startedAt;
      setDisplaySessionStartedAt(session.startedAt);
      setSelectedTrade(null);
      refreshPriceData();
    }
  }, [session.startedAt, refreshPriceData]);

  const sessionTrades = useMemo(() => {
    if (!displaySessionStartedAt) {
      return trades;
    }
    const startTime = new Date(displaySessionStartedAt).getTime();
    return trades.filter((t) => {
      const tradeTime = new Date(t.timestamp).getTime();
      return tradeTime >= startTime;
    });
  }, [trades, displaySessionStartedAt]);

  const sessionEvents = useMemo(() => {
    if (!displaySessionStartedAt) {
      return events;
    }
    const startTime = new Date(displaySessionStartedAt).getTime();
    return events.filter((event) => new Date(event.timestamp).getTime() >= startTime);
  }, [events, displaySessionStartedAt]);

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
        setEvents((prev) => [...prev, message.data as KeyEvent]);
        break;
      case 'events':
        setEvents(message.data as KeyEvent[]);
        break;
      case 'user_balance_reset':
        setUserBalanceRefreshToken((v) => v + 1);
        break;
    }
  }, [updateFromWS, addCandle, updateLPMetrics]);

  const handleSelectEvent = useCallback((event: KeyEvent) => {
    if (event.type !== 'trade') return;
    const match = /^(?<nick>\\S+) executed (?<side>BUY|SELL) of/.exec(event.description);
    const evtTime = new Date(event.timestamp).getTime();
    const nickname = match?.groups?.nick;
    const isBuy = match?.groups?.side === 'BUY';

    const candidates = sessionTrades
      .filter((t) => {
        if (nickname && t.nickname !== nickname) return false;
        if (match?.groups?.side && t.isBuy !== isBuy) return false;
        return true;
      })
      .sort((a, b) => Math.abs(new Date(a.timestamp).getTime() - evtTime) - Math.abs(new Date(b.timestamp).getTime() - evtTime));
    if (candidates.length > 0) {
      setSelectedTrade(candidates[0]);
    }
  }, [sessionTrades]);

  const { isConnected } = useWebSocket(handleWSMessage);

  useEffect(() => {
    api.getTrades(1000).then((data) => setTrades(data as Trade[])).catch(console.error);
    api.getEvents(1000).then((data) => setEvents(data as KeyEvent[])).catch(console.error);
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
      <div className="bg-surface rounded-xl border border-border overflow-hidden">
        <div className="px-5 py-4 border-b border-border flex items-center justify-between">
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

        <div className="p-5">
          <div className="grid grid-cols-15 gap-6">
            <div className="col-span-15 lg:col-span-3 space-y-6">
              <SessionControls
                session={session}
                isLoading={isLoading}
                error={sessionError}
                onStart={start}
                onPause={pause}
                onResume={resume}
                onStop={stop}
                onReset={handleReset}
              />
              <TradingPanel session={session} balanceRefreshToken={userBalanceRefreshToken} />
              <LPStats metrics={lpMetrics} />
            </div>

            <div className="col-span-15 lg:col-span-6 space-y-6">
              <PriceChart
                candles={candles}
                session={session}
                height={350}
                onPriceRangeChange={setPriceRange}
                resetVersion={chartResetVersion}
              />
              <TWAPChart candles={candles} trades={sessionTrades} session={session} height={200} />
              <ImpactCurve buyData={impactData.buy} sellData={impactData.sell} lpMetrics={lpMetrics} />
            </div>

            <div className="col-span-15 lg:col-span-6 space-y-6">
              <Blotter trades={sessionTrades} height={350} highlightedTxHash={selectedTrade?.txHash ?? null} />
              <KeyEvents events={sessionEvents} height={200} onSelectEvent={handleSelectEvent} />
              <AmmDetailsPanel trade={selectedTrade} />
            </div>
          </div>
        </div>

        <div className="text-center text-xs text-gray-500 py-4 border-t border-border">
          <p>
            Simulating market microstructure with Whale, Retail, and Strategy bots.
            All metrics computed in real-time.
          </p>
        </div>
      </div>
    </div>
  );
}
