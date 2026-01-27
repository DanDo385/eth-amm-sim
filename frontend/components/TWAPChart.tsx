'use client';

import { useState, useEffect, useRef, useMemo } from 'react';
import { createChart, IChartApi, ISeriesApi, LineData, Time } from 'lightweight-charts';
import type { Candle, Trade, SessionState } from '@/types';

interface TWAPChartProps {
  candles: Candle[];
  trades: Trade[];
  session: SessionState;
  height?: number;
}

type MetricType = 'twap' | 'stddev';

export function TWAPChart({ candles, trades, session, height = 200 }: TWAPChartProps) {
  const [metricType, setMetricType] = useState<MetricType>('twap');
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const sessionStartRef = useRef<string | undefined>(undefined);

  // Filter candles and trades to current session
  const sessionCandles = useMemo(() => {
    if (!session.startedAt) return [];
    const startTime = new Date(session.startedAt).getTime();
    return candles.filter((c) => new Date(c.timestamp).getTime() >= startTime);
  }, [candles, session.startedAt]);

  const sessionTrades = useMemo(() => {
    if (!session.startedAt) return [];
    const startTime = new Date(session.startedAt).getTime();
    return trades.filter((t) => new Date(t.timestamp).getTime() >= startTime);
  }, [trades, session.startedAt]);

  // Calculate TWAP from candles (simple average of close prices in window)
  const twapData = useMemo(() => {
    if (sessionCandles.length === 0) return [];
    
    const data: LineData[] = [];
    const twapWindow = 60; // 60 seconds window
    
    for (let i = 0; i < sessionCandles.length; i++) {
      const currentTime = new Date(sessionCandles[i].timestamp).getTime();
      const cutoffTime = currentTime - twapWindow * 1000;
      
      // Get all candles within the window
      const windowCandles = sessionCandles.filter((c) => {
        const candleTime = new Date(c.timestamp).getTime();
        return candleTime >= cutoffTime && candleTime <= currentTime;
      });
      
      if (windowCandles.length > 0) {
        const avgPrice = windowCandles.reduce((sum, c) => sum + c.close, 0) / windowCandles.length;
        data.push({
          time: (currentTime / 1000) as Time,
          value: avgPrice,
        });
      }
    }
    
    return data;
  }, [sessionCandles]);

  // Calculate rolling standard deviation from last 100 trades
  const stdDevData = useMemo(() => {
    if (sessionTrades.length < 100) return []; // Only start after 100 trades
    
    const data: LineData[] = [];
    const windowSize = 100;
    
    // Extract prices from trades (convert from wei to ETH)
    const prices = sessionTrades
      .map((t) => {
        if (!t.price) return null;
        return parseFloat(t.price) / 1e18;
      })
      .filter((p): p is number => p !== null);
    
    if (prices.length < windowSize) return [];
    
    // Calculate rolling std dev
    for (let i = windowSize - 1; i < prices.length; i++) {
      const window = prices.slice(i - windowSize + 1, i + 1);
      
      // Calculate mean
      const mean = window.reduce((sum, p) => sum + p, 0) / window.length;
      
      // Calculate variance
      const variance = window.reduce((sum, p) => sum + Math.pow(p - mean, 2), 0) / window.length;
      
      // Standard deviation
      const stdDev = Math.sqrt(variance);
      
      const trade = sessionTrades[i];
      if (trade) {
        data.push({
          time: (new Date(trade.timestamp).getTime() / 1000) as Time,
          value: stdDev,
        });
      }
    }
    
    return data;
  }, [sessionTrades]);

  // Reset chart when session changes
  useEffect(() => {
    if (session.status === 'idle' || !session.startedAt) {
      if (sessionStartRef.current !== undefined) {
        sessionStartRef.current = undefined;
        if (seriesRef.current) {
          seriesRef.current.setData([]);
        }
      }
    } else if (session.startedAt && session.startedAt !== sessionStartRef.current) {
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

    const series = chart.addLineSeries({
      color: '#3b82f6',
      lineWidth: 2,
      title: metricType === 'twap' ? 'TWAP' : 'Rolling Std Dev (100 trades)',
    });

    chartRef.current = chart;
    seriesRef.current = series;

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
  }, [height, metricType]);

  // Update data based on selected metric
  useEffect(() => {
    if (!seriesRef.current || !chartRef.current) return;

    const data = metricType === 'twap' ? twapData : stdDevData;

    if (data.length === 0) {
      seriesRef.current.setData([]);
      return;
    }

    // Update series color and title
    seriesRef.current.applyOptions({
      color: metricType === 'twap' ? '#3b82f6' : '#f59e0b',
      title: metricType === 'twap' ? 'TWAP' : 'Rolling Std Dev (100 trades)',
    });

    seriesRef.current.setData(data);

    // Set time range to match session
    if (session.startedAt && session.duration > 0) {
      const startTime = new Date(session.startedAt).getTime() / 1000;
      const endTime = startTime + session.duration;
      
      chartRef.current.timeScale().setVisibleRange({
        from: startTime as Time,
        to: endTime as Time,
      });
    } else if (data.length > 0) {
      chartRef.current.timeScale().fitContent();
    }
  }, [twapData, stdDevData, metricType, session.startedAt, session.duration]);

  const hasEnoughTrades = sessionTrades.length >= 100;

  return (
    <div className="bg-surface rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-sm font-medium text-white">
          {metricType === 'twap' ? 'TWAP (60s window)' : 'Rolling Std Dev (100 trades)'}
        </h3>
        <div className="flex items-center space-x-2">
          <button
            onClick={() => setMetricType('twap')}
            className={`px-3 py-1 text-xs rounded ${
              metricType === 'twap'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
          >
            TWAP
          </button>
          <button
            onClick={() => setMetricType('stddev')}
            className={`px-3 py-1 text-xs rounded ${
              metricType === 'stddev'
                ? 'bg-amber-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
            disabled={!hasEnoughTrades}
            title={!hasEnoughTrades ? 'Requires 100+ trades' : ''}
          >
            Std Dev
          </button>
        </div>
      </div>
      <div ref={containerRef} />
      {metricType === 'stddev' && !hasEnoughTrades && (
        <div className="px-4 py-2 text-xs text-amber-400 text-center border-t border-border">
          Waiting for 100 trades... ({sessionTrades.length}/100)
        </div>
      )}
    </div>
  );
}
