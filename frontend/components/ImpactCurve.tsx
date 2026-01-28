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
      lineType: 0, // Line type: 0 = line (continuous), 1 = area, 2 = histogram
    });

    // Add a price line at 1.0 to show the reference price (center of y-axis)
    series.createPriceLine({
      price: 1.0,
      color: '#9ca3af',
      lineWidth: 1,
      lineStyle: 2, // Dashed
      axisLabelVisible: true,
      title: 'Current Price (1.00)',
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
    // Add center point at (0, 1.0) representing current spot price with no trade
    const combinedData: LineData[] = [
      // Center point: current spot price at trade size 0
      {
        time: 0 as Time,
        value: 1.0, // Normalized reference price (current spot price)
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

    // Center the axes on the last traded price (normalized to 1.0)
    // The center point is at (0, 1.0) representing current spot price with no trade
    const centerPrice = 1.0; // Always use 1.0 as center (normalized reference price)
    
    // Fixed range centered on 1.0: [0.5, 1.5] or wider if needed
    // This ensures the center (1.0) is always in the middle of the y-axis
    const range = 2.0;
    const yAxisMin = Math.max(0, centerPrice - range / 2); // 0.0
    const yAxisMax = centerPrice + range / 2; // 2.0
    
    // Find actual min/max values in the data to ensure they're visible
    const dataValues = combinedData.map(d => d.value);
    const actualMin = Math.min(...dataValues);
    const actualMax = Math.max(...dataValues);
    
    // Adjust range if data extends beyond the default range, but keep center at 1.0
    let finalYMin = yAxisMin;
    let finalYMax = yAxisMax;
    
    if (actualMin < yAxisMin) {
      // Data goes below 0, extend range downward but keep center visible
      finalYMin = Math.max(0, actualMin - 0.1);
    }
    if (actualMax > yAxisMax) {
      // Data goes above 2.0, extend range upward but keep center visible
      finalYMax = actualMax + 0.1;
    }
    
    // Ensure center (1.0) is always in the middle of the visible range
    const currentRange = finalYMax - finalYMin;
    const centerOffset = centerPrice - (finalYMin + finalYMax) / 2;
    
    // Adjust range to center on 1.0
    if (Math.abs(centerOffset) > 0.01) {
      finalYMin = Math.max(0, centerPrice - currentRange / 2);
      finalYMax = centerPrice + currentRange / 2;
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
      ...combinedData,
      // Add points just outside the edges to force the scale
      { time: -501 as Time, value: finalYMin },
      { time: 501 as Time, value: finalYMax },
    ].sort((a, b) => Number(a.time) - Number(b.time)); // Sort by time to ensure ascending order

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
        Price (normalized) vs Trade Size (ETH) | Center (0, 1.0) = Current Spot Price
      </div>
    </div>
  );
}
