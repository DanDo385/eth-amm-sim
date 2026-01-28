// PriceChart.tsx — OHLC candlestick chart for APPL/ETH price.
//
// Uses TradingView's lightweight-charts library to render candles.
// Initial candle history comes from GET /candles; live updates arrive
// via WebSocket "price" messages → usePriceData.addCandle → this component.
//
// CONNECTIONS:
//   - Backend data:  metrics/price.go Candle (OHLC aggregated per interval)
//   - WebSocket msg: "price" from broadcast.go BroadcastPrice
//   - Data hook:     hooks/usePriceData.ts manages candle state
//   - Types:         types/index.ts Candle
'use client';

import { useEffect, useRef, useMemo } from 'react';
import { createChart, IChartApi, ISeriesApi, CandlestickData, Time } from 'lightweight-charts';
import type { Candle, SessionState } from '@/types';

interface PriceChartProps {
  candles: Candle[];
  session: SessionState;
  height?: number;
  onPriceRangeChange?: (range: { min: number; max: number }) => void;
}

export function PriceChart({ candles, session, height = 300, onPriceRangeChange }: PriceChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null);
  const sessionStartRef = useRef<string | undefined>(undefined);

  // Filter candles to only show those from the current session
  const sessionCandles = useMemo(() => {
    if (!session.startedAt) {
      return []; // No session started, show blank chart
    }

    const startTime = new Date(session.startedAt).getTime();
    return candles.filter((c) => {
      const candleTime = new Date(c.timestamp).getTime();
      return candleTime >= startTime;
    });
  }, [candles, session.startedAt]);

  // Reset chart when a new session starts or when session is reset
  useEffect(() => {
    if (session.status === 'idle' || !session.startedAt) {
      // Session is idle/reset - clear the chart
      if (sessionStartRef.current !== undefined) {
        sessionStartRef.current = undefined;
        if (seriesRef.current) {
          seriesRef.current.setData([]);
        }
      }
    } else if (session.startedAt && session.startedAt !== sessionStartRef.current) {
      // New session started - clear the chart
      sessionStartRef.current = session.startedAt;
      if (seriesRef.current) {
        seriesRef.current.setData([]);
      }
    }
  }, [session.startedAt, session.status]);

  // Initialize chart
  useEffect(() => {
    if (!containerRef.current) return;

    const chart = createChart(containerRef.current, {
      layout: {
        background: { color: '#1a1f26' },
        textColor: '#9ca3af',
      },
      grid: {
        vertLines: { color: '#2d3748' },
        horzLines: { color: '#2d3748' },
      },
      width: containerRef.current.clientWidth,
      height,
      timeScale: {
        timeVisible: true,
        secondsVisible: true,
        rightOffset: 10,
        // Prevent auto-scrolling - we'll control the visible range manually
        rightBarStaysOnScroll: true,
        lockVisibleTimeRangeOnResize: true,
      },
      rightPriceScale: {
        borderColor: '#2d3748',
        autoScale: true,
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      },
    });

    const series = chart.addCandlestickSeries({
      upColor: '#22c55e',
      downColor: '#ef4444',
      borderUpColor: '#22c55e',
      borderDownColor: '#ef4444',
      wickUpColor: '#22c55e',
      wickDownColor: '#ef4444',
    });

    chartRef.current = chart;
    seriesRef.current = series;

    // Handle resize
    const handleResize = () => {
      if (containerRef.current) {
        chart.applyOptions({ width: containerRef.current.clientWidth });
      }
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chart.remove();
    };
  }, [height]);

  // Update data and set fixed time window
  useEffect(() => {
    if (!seriesRef.current || !chartRef.current) return;

    if (sessionCandles.length === 0) {
      // No candles yet - show blank chart
      seriesRef.current.setData([]);
      return;
    }

    const data: CandlestickData[] = sessionCandles.map((c) => ({
      time: (new Date(c.timestamp).getTime() / 1000) as Time,
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
    }));

    seriesRef.current.setData(data);

    // Set fixed time window based on session duration
    // This ensures the chart view never changes - it always shows the full session duration
    if (session.startedAt && session.duration > 0) {
      const startTime = new Date(session.startedAt).getTime() / 1000;
      const endTime = startTime + session.duration;
      
      // Always set the visible range to the full session duration
      // This keeps the view constant - as candles come in, they fill from left to right
      chartRef.current.timeScale().setVisibleRange({
        from: startTime as Time,
        to: endTime as Time,
      });
      
      // Lock the visible range so it doesn't auto-adjust
      chartRef.current.timeScale().applyOptions({
        rightBarStaysOnScroll: true,
        lockVisibleTimeRangeOnResize: true,
      });
    } else {
      // Fallback: fit content if no session info
      chartRef.current.timeScale().fitContent();
    }
    
    // Calculate and notify exact price range for Impact Curve
    // Always calculate from data to ensure consistency
    if (onPriceRangeChange && data.length > 0) {
      const prices = data.flatMap(c => [c.low, c.high]);
      const minPrice = Math.min(...prices);
      const maxPrice = Math.max(...prices);
      const range = maxPrice - minPrice;
      // Apply same margins as chart (10% top/bottom) to match Price Chart scale
      const margin = range * 0.1;
      onPriceRangeChange({ 
        min: Math.max(0, minPrice - margin), 
        max: maxPrice + margin 
      });
    }
  }, [sessionCandles, session.startedAt, session.duration, onPriceRangeChange]);

  return (
    <div className="bg-surface rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-sm font-medium text-white">APPL/ETH Price</h3>
      </div>
      <div ref={containerRef} />
    </div>
  );
}
