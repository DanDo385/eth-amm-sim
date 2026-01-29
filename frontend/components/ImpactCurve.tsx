// ImpactCurve.tsx — Price impact visualization for hypothetical trade sizes.
//
// Shows price impact (slippage) as a function of trade size.
// Data is computed in the backend (metrics/impact.go) from current reserves and returned
// as chart-ready points:
//   - X-axis: trade size (ETH), negative for sells and positive for buys
//   - Y-axis: impact in basis points (bps), centered at 0 (spot price)
//
// CONNECTIONS:
//   - Backend data:  metrics/impact.go CalculateBuyCurve/CalculateSellCurve
//   - REST endpoint: server/handlers.go handleGetImpactCurve
//   - Reserve updates: main.go pollPrices → impact.UpdateReserves
//   - Types:         types/index.ts ImpactPoint, ImpactCurve
'use client';

import { useEffect, useRef } from 'react';
import { createChart, IChartApi, ISeriesApi, LineData, Time, IPriceLine } from 'lightweight-charts';
import type { ImpactPoint, Candle } from '@/types';

interface ImpactCurveProps {
  buyData: ImpactPoint[];
  sellData: ImpactPoint[];
  candles: Candle[]; // To get last traded price
  session: { status: string; startedAt?: string }; // To detect reset
  height?: number;
}

export function ImpactCurve({ buyData, sellData, candles, session, height = 200 }: ImpactCurveProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const priceSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const spotLineRef = useRef<IPriceLine | null>(null);
  const spotPriceFromBackend =
    buyData.find((p) => p.spotPrice > 0)?.spotPrice ?? sellData.find((p) => p.spotPrice > 0)?.spotPrice;
  const lastTradedPrice = candles.length > 0 ? candles[candles.length - 1].close : undefined;
  const spotPrice = spotPriceFromBackend ?? lastTradedPrice ?? 0;

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
      rightPriceScale: {
        borderColor: '#2d3748',
        autoScale: true,
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      },
      leftPriceScale: {
        visible: false,
      },
      timeScale: {
        visible: true,
        timeVisible: false,
        secondsVisible: false,
        rightOffset: 10,
        lockVisibleTimeRangeOnResize: true,
        tickMarkFormatter: (time: Time) => {
          const size = Number(time);
          return size >= 0 ? `+${size}` : `${size}`;
        },
      },
      localization: {
        priceFormatter: (value: number) => value.toFixed(4),
        timeFormatter: (time: Time) => {
          // Custom formatter for x-axis: show quantity instead of date
          const quantity = Number(time);
          if (isNaN(quantity)) return '';
          return quantity >= 0 ? `+${quantity.toFixed(0)} ETH` : `${quantity.toFixed(0)} ETH`;
        },
      },
      crosshair: {
        mode: 1, // Normal mode
      },
    });

    const priceSeries = chart.addLineSeries({
      color: '#3b82f6',
      lineWidth: 2,
      title: 'Execution Price vs Trade Size',
      priceLineVisible: false, // Don't show price line
      priceFormat: {
        type: 'price',
        precision: 4,
        minMove: 0.0001,
      },
      lineType: 0, // Line type: 0 = line (continuous), 1 = area, 2 = histogram
    });

    // Dashed baseline at spot price (no trade).
    spotLineRef.current = priceSeries.createPriceLine({
      price: 1.0, // updated after mount in data effect
      color: '#9ca3af',
      lineWidth: 1,
      lineStyle: 2, // Dashed
      axisLabelVisible: true,
      title: 'Spot',
    });

    chartRef.current = chart;
    priceSeriesRef.current = priceSeries;

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

  // Update data
  useEffect(() => {
    if (!priceSeriesRef.current || !chartRef.current) return;

    const fallbackSpot = spotPrice > 0 ? spotPrice : 1.0;

    // Keep dashed baseline aligned to spot price.
    spotLineRef.current?.applyOptions({ price: fallbackSpot });

    // If idle, show a flat baseline at spot price.
    if (session.status === 'idle' || !session.startedAt) {
      const baseline: LineData[] = [
        { time: -500 as Time, value: fallbackSpot },
        { time: 0 as Time, value: fallbackSpot },
        { time: 500 as Time, value: fallbackSpot },
      ];
      priceSeriesRef.current.setData(baseline);
      return;
    }

    // Combine buy and sell curves into a single continuous line around the centered spot price:
    // - X: trade size in ETH (negative = sell, positive = buy)
    // - Y: execution price (ETH/APPL), centered visually via the dashed spot baseline
    const buyPoints = buyData
      .filter((p) => p.tradeSize > 0 && p.executePrice > 0)
      .map((p) => ({
        time: p.tradeSize as Time,
        value: p.executePrice,
      }));

    const sellPoints = sellData
      .filter((p) => p.tradeSize > 0 && p.executePrice > 0)
      .map((p) => ({
        time: -p.tradeSize as Time,
        value: p.executePrice,
      }));

    const combinedImpactData: LineData[] = [
      ...sellPoints,
      { time: 0 as Time, value: fallbackSpot },
      ...buyPoints,
    ].sort((a, b) => Number(a.time) - Number(b.time));

    // Set x-axis range to encompass available trade sizes (symmetric if possible)
    const allSizes = [...buyData, ...sellData].map((p) => Math.abs(p.tradeSize)).filter((v) => v > 0);
    const maxSize = allSizes.length > 0 ? Math.max(...allSizes) : 500;
    const range = Math.max(50, Math.min(maxSize, 500));

    const timeScale = chartRef.current.timeScale();
    if (timeScale) {
      timeScale.applyOptions({
        visible: true,
        timeVisible: false,
        secondsVisible: false,
        rightOffset: 10,
        lockVisibleTimeRangeOnResize: true,
      });

      try {
        timeScale.setVisibleRange({
          from: (-range) as Time,
          to: range as Time,
        });
      } catch (e) {
        // Retry after short delay if timeScale is not ready
        setTimeout(() => {
          const ts = chartRef.current?.timeScale();
          if (ts) {
            ts.setVisibleRange({
              from: (-range) as Time,
              to: range as Time,
            });
          }
        }, 100);
      }
    }

    priceSeriesRef.current.setData(combinedImpactData);
  }, [buyData, sellData, session.status, session.startedAt, spotPrice]);

  return (
    <div className="bg-surface rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-sm font-medium text-white">Price Impact Curve</h3>
        <div className="text-xs text-gray-400">
          Spot: {spotPrice > 0 ? `${spotPrice.toFixed(4)} ETH/APPL` : '—'}
        </div>
      </div>
      <div ref={containerRef} />
      <div className="px-4 py-2 text-xs text-gray-400 text-center border-t border-border">
        Execution Price (ETH/APPL) vs Trade Size (ETH) | Dashed line = Spot
      </div>
    </div>
  );
}
