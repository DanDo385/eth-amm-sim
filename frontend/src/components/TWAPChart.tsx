// TWAPChart.tsx — Time-Weighted Average Price and standard deviation chart.
//
// Computes TWAP and rolling standard deviation client-side from candle data.
// Visualizes price stability and volatility over the simulation window.
// Candle data originates from the same source as PriceChart (metrics/price.go).
//
// CONNECTIONS:
//   - Backend data:  metrics/price.go Candle close prices (used for TWAP calc)
//   - Data hook:     hooks/usePriceData.ts provides candle array
//   - Types:         types/index.ts Candle, Trade, SessionState

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

export const TWAPChart = ({ candles, trades, session, height = 200 }: TWAPChartProps) => {
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

  // Calculate 60s rolling daily volatility
  const stdDevData = useMemo(() => {
    if (sessionCandles.length < 2) return []; // Need at least 2 candles to calculate returns
    
    const data: LineData[] = [];
    const windowSeconds = 60; // 60-second rolling window
    
    // Calculate returns from candle closes
    // Returns = (price_t - price_{t-1}) / price_{t-1}
    const returns: Array<{ time: number; return: number }> = [];
    for (let i = 1; i < sessionCandles.length; i++) {
      const prevPrice = sessionCandles[i - 1].close;
      const currPrice = sessionCandles[i].close;
      const currTime = new Date(sessionCandles[i].timestamp).getTime();
      
      if (prevPrice > 0) {
        const returnValue = (currPrice - prevPrice) / prevPrice;
        returns.push({
          time: currTime,
          return: returnValue,
        });
      }
    }
    
    if (returns.length < 2) return [];
    
    // Calculate rolling volatility for each candle
    for (let i = 0; i < sessionCandles.length; i++) {
      const currentTime = new Date(sessionCandles[i].timestamp).getTime();
      const cutoffTime = currentTime - windowSeconds * 1000;
      
      // Get all returns within the 60s window
      const windowReturns = returns
        .filter((r) => r.time >= cutoffTime && r.time <= currentTime)
        .map((r) => r.return);
      
      if (windowReturns.length < 2) continue; // Need at least 2 returns to calculate std dev
      
      // Calculate mean return
      const meanReturn = windowReturns.reduce((sum, r) => sum + r, 0) / windowReturns.length;
      
      // Calculate variance of returns
      const variance = windowReturns.reduce(
        (sum, r) => sum + Math.pow(r - meanReturn, 2),
        0
      ) / windowReturns.length;
      
      // Standard deviation of returns
      const stdDev = Math.sqrt(variance);
      
      // Scale to daily volatility: sqrt(1440 periods per day) * 100 for percentage
      // 1440 = 86400 seconds per day / 60 seconds per period
      const dailyVol = stdDev * Math.sqrt(1440) * 100;
      
      data.push({
        time: (currentTime / 1000) as Time,
        value: dailyVol,
      });
    }
    
    return data;
  }, [sessionCandles]);

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
      title: metricType === 'twap' ? 'TWAP' : '60s Rolling Daily Vol (%)',
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
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

    // Handle empty data gracefully
    if (data.length === 0) {
      seriesRef.current.setData([]);
      // Don't return early - still update series options
    } else {
      seriesRef.current.setData(data);
    }

    // Update series color, title, and price format
    seriesRef.current.applyOptions({
      color: metricType === 'twap' ? '#3b82f6' : '#f59e0b',
      title: metricType === 'twap' ? 'TWAP' : '60s Rolling Daily Vol (%)',
      priceFormat: metricType === 'stddev' 
        ? {
            type: 'custom',
            formatter: (price: number) => `${price.toFixed(2)}%`,
          }
        : {
            type: 'price',
            precision: 2,
            minMove: 0.01,
          },
    });

    // Set time range to match session
    if (session.startedAt && session.duration > 0) {
      const startTime = new Date(session.startedAt).getTime() / 1000;
      const endTime = startTime + session.duration;
      
      try {
        chartRef.current.timeScale().setVisibleRange({
          from: startTime as Time,
          to: endTime as Time,
        });
      } catch (error) {
        // Ignore errors when setting time range (e.g., invalid range)
        console.warn('Error setting time range:', error);
      }
    } else if (data.length > 0) {
      try {
        chartRef.current.timeScale().fitContent();
      } catch (error) {
        // Ignore errors when fitting content
        console.warn('Error fitting content:', error);
      }
    }
  }, [twapData, stdDevData, metricType, session.startedAt, session.duration]);

  return (
    <div className="bg-surface rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-sm font-medium text-white">
          {metricType === 'twap' ? 'TWAP (60s window)' : '60s Rolling Daily Vol (%)'}
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
            disabled={sessionCandles.length < 2}
            title={sessionCandles.length < 2 ? 'Requires 2+ candles' : ''}
          >
            Daily Vol (%)
          </button>
        </div>
      </div>
      <div ref={containerRef} />
      {metricType === 'stddev' && sessionCandles.length < 2 && (
        <div className="px-4 py-2 text-xs text-amber-400 text-center border-t border-border">
          Waiting for candle data... ({sessionCandles.length}/2)
        </div>
      )}
    </div>
  );
};
