// ImpactCurve.tsx — Price impact visualization for hypothetical trade sizes.
//
// Shows price impact curve relative to the last traded price from the APPL/ETH chart.
// Uses the constant product formula x*y=k. Data is computed by backend metrics/impact.go
// using current pool reserves, fetched via GET /impact-curve.
//
// CONNECTIONS:
//   - Backend data:  metrics/impact.go CalculateBuyCurve/CalculateSellCurve
//   - REST endpoint: server/handlers.go handleGetImpactCurve
//   - Reserve updates: main.go pollPrices → impact.UpdateReserves
//   - Types:         types/index.ts ImpactPoint, ImpactCurve
'use client';

import { useEffect, useRef } from 'react';
import { createChart, IChartApi, ISeriesApi, LineData, Time } from 'lightweight-charts';
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
  const resetPriceRef = useRef<number>(1.0); // Last traded price we center the axis around
  const sessionStartRef = useRef<string | undefined>(undefined);

  // Get last traded price from candles (last candle's close price)
  const lastTradedPrice = candles.length > 0 ? candles[candles.length - 1].close : undefined;

  // Detect reset and update reset price
  useEffect(() => {
    if (session.status === 'idle' || !session.startedAt) {
      // Reset detected - set reset price to 1.0
      resetPriceRef.current = 1.0;
      if (sessionStartRef.current !== undefined) {
        sessionStartRef.current = undefined;
      }
    } else if (session.startedAt && session.startedAt !== sessionStartRef.current) {
      // New session started - capture last traded price as reset price
      sessionStartRef.current = session.startedAt;
      if (lastTradedPrice && lastTradedPrice > 0) {
        resetPriceRef.current = lastTradedPrice;
      }
    } else if (lastTradedPrice && lastTradedPrice > 0 && sessionStartRef.current !== undefined) {
      // Update reset price to current last traded price during session
      resetPriceRef.current = lastTradedPrice;
    }
  }, [session.status, session.startedAt, lastTradedPrice]);

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
        autoScale: false, // Manual control: we keep last price centered between 0 and 2.0
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
        priceFormatter: (price: number) => {
          return price.toFixed(4);
        },
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
      title: 'Price vs Trade Size',
      priceLineVisible: false, // Don't show price line
      priceFormat: {
        type: 'price',
        precision: 4,
        minMove: 0.0001,
      },
      lineType: 0, // Line type: 0 = line (continuous), 1 = area, 2 = histogram
    });

    // Add a dashed price line at the reference (last traded) price
    priceSeries.createPriceLine({
      price: resetPriceRef.current,
      color: '#9ca3af',
      lineWidth: 1,
      lineStyle: 2, // Dashed
      axisLabelVisible: true,
      title: 'Last Price',
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
    
    // Use reset price as reference (tracks last traded price)
    const referencePrice = resetPriceRef.current;
    
    if (!referencePrice || referencePrice <= 0) {
      // If no reference price, show a flat baseline at 1.0
      const baseline: LineData[] = [
        { time: -500 as Time, value: 1.0 },
        { time: 0 as Time, value: 1.0 },
        { time: 500 as Time, value: 1.0 },
      ];
      priceSeriesRef.current.setData(baseline);
      return;
    }
    
    // Combine buy and sell data into one series
    // X-axis: Trade Size (in ETH) - negative for sells, positive for buys
    // Y-axis: Price normalized to the reference (last traded) price.
    // 1.0 = current/last price, <1.0 = cheaper, >1.0 = more expensive.
    // Add center point at (0, 1.0) representing current spot price with no trade.
    const combinedPriceData: LineData[] = [
      // Center point: normalized price 1.0 at trade size 0
      {
        time: 0 as Time,
        value: 1.0,
      },
      // Buy data (positive x-axis)
      ...buyData
        .filter(p => p.tradeSize > 0 && p.tradeSize <= 500 && p.executePrice > 0)
        .map((p) => ({
          time: p.tradeSize as Time, // Positive for buys
          value: p.executePrice / referencePrice, // Normalized price
        })),
      // Sell data (negative x-axis)
      ...sellData
        .filter(p => p.tradeSize > 0 && p.tradeSize <= 500 && p.executePrice > 0)
        .map((p) => ({
          time: -p.tradeSize as Time, // Negative for sells
          value: p.executePrice / referencePrice, // Normalized price
        })),
    ].sort((a, b) => Number(a.time) - Number(b.time));

    // Center the y-axis on 1.0 with global [0.0, 2.0] bounds (normalized price space).
    const window = 1.0;
    let yAxisMin = 1.0 - window / 2; // 0.5
    let yAxisMax = 1.0 + window / 2; // 1.5
    
    // Find actual min/max values in the data to ensure they're visible
    const dataValues = combinedPriceData.map(d => d.value);
    const actualMin = Math.min(...dataValues);
    const actualMax = Math.max(...dataValues);
    
    // Adjust range if data extends beyond the default window, but keep within [0, 2.0]
    let finalYMin = yAxisMin;
    let finalYMax = yAxisMax;
    
    if (actualMin < yAxisMin) {
      // Data goes below current min, extend range downward but keep floor at 0
      finalYMin = Math.max(0, actualMin - 0.05);
    }
    if (actualMax > yAxisMax) {
      // Data goes above current max, extend range upward but cap at 2.0
      finalYMax = Math.min(2.0, actualMax + 0.05);
    }

    // Set x-axis range to -500 to +500
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
          from: -500 as Time,
          to: 500 as Time,
        });
      } catch (e) {
        // Retry after short delay if timeScale is not ready
        setTimeout(() => {
          const ts = chartRef.current?.timeScale();
          if (ts) {
            ts.setVisibleRange({
              from: -500 as Time,
              to: 500 as Time,
            });
          }
        }, 100);
      }
    }

    // Add invisible data points at the min/max to force the scale range
    // This is a workaround since lightweight-charts doesn't directly support setting min/max
    // We add points just outside the visible range to ensure they don't conflict with actual data
    // Use -501 and 501 to avoid duplicates with actual trade data at -500 and 500
    const extendedData = [
      ...combinedPriceData,
      // Add points just outside the edges to force the scale
      { time: -501 as Time, value: finalYMin },
      { time: 501 as Time, value: finalYMax },
    ].sort((a, b) => Number(a.time) - Number(b.time)); // Sort by time to ensure ascending order

    // Set data with extended points to force the scale range
    priceSeriesRef.current.setData(extendedData);

    // Set price scale range after data is set
    setTimeout(() => {
      if (!chartRef.current) return;
      
      // Use manual scale with sliding range
      chartRef.current.priceScale('right').applyOptions({
        autoScale: false,
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      });
    }, 50);
  }, [buyData, sellData, lastTradedPrice, session.status, session.startedAt]);

  return (
    <div className="bg-surface rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-sm font-medium text-white">Price Impact Curve</h3>
        <div className="text-xs text-gray-400">
          Reference: {resetPriceRef.current.toFixed(4)} ETH/APPL
        </div>
      </div>
      <div ref={containerRef} />
      <div className="px-4 py-2 text-xs text-gray-400 text-center border-t border-border">
        Price (ETH/APPL) vs Trade Size (ETH) | Dashed line = Last Price (defaults to 1.0)
      </div>
    </div>
  );
}
