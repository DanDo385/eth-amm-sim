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
  const seriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const resetPriceRef = useRef<number>(1.0); // Price to reset to (default 1.0)
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
        autoScale: false, // Manual control for reset to 1.0
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

    const series = chart.addLineSeries({
      color: '#3b82f6',
      lineWidth: 2,
      title: 'Price Impact',
      priceLineVisible: false, // Don't show price line
      priceFormat: {
        type: 'price',
        precision: 4,
        minMove: 0.0001,
      },
    });

    // Add a price line at 1.0 to show the reference price
    series.createPriceLine({
      price: 1.0,
      color: '#9ca3af',
      lineWidth: 1,
      lineStyle: 2, // Dashed
      axisLabelVisible: true,
      title: 'Reference (1.00)',
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

  // Update data
  useEffect(() => {
    if (!seriesRef.current || !chartRef.current) return;
    
    // Use reset price as reference (defaults to 1.0, updates to last traded price)
    const referencePrice = resetPriceRef.current;
    
    if (!referencePrice || referencePrice <= 0) {
      // If no reference price, clear chart
      seriesRef.current.setData([]);
      return;
    }
    
    // If no data, clear the chart
    if (buyData.length === 0 && sellData.length === 0) {
      seriesRef.current.setData([]);
      return;
    }
    
    // Combine buy and sell data into one series
    // X-axis: Trade Size (in ETH) - negative for sells, positive for buys
    // Y-axis: Price relative to reference price (normalized)
    const combinedData: LineData[] = [
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

    // Calculate the center price for sliding scale
    // Use the price at trade size 0 (or closest to 0) as the center
    // This represents the current normalized spot price (should be ~1.0)
    let centerPrice = 1.0; // Default center
    
    if (combinedData.length > 0) {
      // Find the price closest to trade size 0 (the current spot price)
      const closestToZero = combinedData.reduce((closest, point) => {
        const currentDist = Math.abs(Number(point.time));
        const closestDist = Math.abs(Number(closest.time));
        return currentDist < closestDist ? point : closest;
      });
      centerPrice = closestToZero.value;
    }
    
    // Sliding scale: center around current price with range of 2.0
    // Scale slides from center - 1.0 to center + 1.0, but never below 0
    const range = 2.0;
    let yAxisMin = Math.max(0, centerPrice - range / 2);
    let yAxisMax = centerPrice + range / 2;
    
    // If min is clamped to 0, adjust max to maintain range
    if (yAxisMin === 0 && centerPrice < 1.0) {
      // If center is below 1.0, keep range at [0, 2.0]
      yAxisMax = range;
    } else if (yAxisMin > 0) {
      // If center is above 1.0, slide the range up
      // e.g., center=1.5 → [0.5, 2.5]
      yAxisMax = yAxisMin + range;
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
    // We add points at the edges of the x-axis to ensure the y-axis shows the full range [yAxisMin, yAxisMax]
    const extendedData = [
      ...combinedData,
      // Add points at the edges to force the scale
      { time: -500 as Time, value: yAxisMin },
      { time: 500 as Time, value: yAxisMax },
    ];

    // Set data with extended points to force the scale range
    seriesRef.current.setData(extendedData);

    // Set price scale range after data is set
    // Sliding scale: always centered around current price with range of 2.0
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
        Price (normalized) vs Trade Size (ETH) | Y-axis resets to 1.00 on Reset
      </div>
    </div>
  );
}
