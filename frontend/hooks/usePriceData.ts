'use client';

import { useState, useEffect, useCallback } from 'react';
import type { Candle, LPMetrics } from '@/types';
import * as api from '@/lib/api';

export function usePriceData() {
  const [candles, setCandles] = useState<Candle[]>([]);
  const [lpMetrics, setLPMetrics] = useState<LPMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch initial data
  const refresh = useCallback(async () => {
    try {
      const [candleData, lpData] = await Promise.all([
        api.getCandles(),
        api.getLPMetrics(),
      ]);
      setCandles(candleData);
      setLPMetrics(lpData);
    } catch (e) {
      console.error('Failed to fetch price data:', e);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Add a new candle
  const addCandle = useCallback((candle: Candle) => {
    setCandles((prev) => {
      // Check if we should update last candle or add new one
      if (prev.length > 0) {
        const lastCandle = prev[prev.length - 1];
        const lastTime = new Date(lastCandle.timestamp).getTime();
        const newTime = new Date(candle.timestamp).getTime();
        
        // Same candle period - update
        if (Math.abs(newTime - lastTime) < 5000) {
          return [...prev.slice(0, -1), candle];
        }
      }
      return [...prev, candle];
    });
  }, []);

  // Update LP metrics
  const updateLPMetrics = useCallback((metrics: LPMetrics) => {
    setLPMetrics(metrics);
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return {
    candles,
    lpMetrics,
    isLoading,
    refresh,
    addCandle,
    updateLPMetrics,
  };
}
