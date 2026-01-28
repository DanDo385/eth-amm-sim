// ImpactCurve.tsx — Price impact visualization for hypothetical trade sizes.
//
// Shows buy and sell slippage curves (in basis points) using the constant
// product formula x*y=k. Data is computed by backend metrics/impact.go
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
import type { ImpactPoint } from '@/types';

interface ImpactCurveProps {
  buyData: ImpactPoint[];
  sellData: ImpactPoint[];
  currentPrice?: number; // Current spot price from metrics
  height?: number;
}

export function ImpactCurve({ buyData, sellData, currentPrice, height = 200 }: ImpactCurveProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const buySeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const sellSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);

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
    });

    const buySeries = chart.addLineSeries({
      color: '#22c55e',
      lineWidth: 2,
      title: 'Buy Price',
    });

    const sellSeries = chart.addLineSeries({
      color: '#ef4444',
      lineWidth: 2,
      title: 'Sell Price',
    });

    chartRef.current = chart;
    buySeriesRef.current = buySeries;
    sellSeriesRef.current = sellSeries;

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
    if (!buySeriesRef.current || !sellSeriesRef.current || !chartRef.current) return;
    
    // Use currentPrice from metrics as the reference spot price
    if (!currentPrice || currentPrice <= 0) {
      // If no current price, clear chart
      buySeriesRef.current.setData([]);
      sellSeriesRef.current.setData([]);
      return;
    }
    
    // If no data, clear the chart
    if (buyData.length === 0 && sellData.length === 0) {
      buySeriesRef.current.setData([]);
      sellSeriesRef.current.setData([]);
      return;
    }
    
    // Convert to chart format
    // X-axis: Trade Size (in ETH) - negative for sells, positive for buys
    // Y-axis: Execution Price (ETH per APPL) - absolute price
    const buyChartData: LineData[] = buyData
      .filter(p => p.tradeSize > 0 && p.tradeSize <= 500 && p.executePrice > 0)
      .map((p) => ({
        time: p.tradeSize as Time, // Positive for buys
        value: p.executePrice, // Execution price on y-axis
      }))
      .sort((a, b) => Number(a.time) - Number(b.time));

    const sellChartData: LineData[] = sellData
      .filter(p => p.tradeSize > 0 && p.tradeSize <= 500 && p.executePrice > 0)
      .map((p) => ({
        time: -p.tradeSize as Time, // Negative for sells
        value: p.executePrice, // Execution price on y-axis
      }))
      .sort((a, b) => Number(a.time) - Number(b.time));

    // Calculate price range from all execution prices
    const allPrices = buyChartData.length > 0 || sellChartData.length > 0
      ? [...buyChartData.map(d => d.value), ...sellChartData.map(d => d.value)]
      : [currentPrice];
    
    const minPrice = allPrices.length > 0 ? Math.min(...allPrices) : currentPrice;
    const maxPrice = allPrices.length > 0 ? Math.max(...allPrices) : currentPrice;
    
    // Calculate y-axis range with padding
    const priceRange = maxPrice - minPrice;
    const padding = Math.max(priceRange * 0.1, currentPrice * 0.05); // 10% of range or 5% of current price
    const yAxisMin = Math.max(0.01, minPrice - padding);
    const yAxisMax = maxPrice + padding;

    // Ensure both series use the same price scale
    buySeriesRef.current.applyOptions({
      priceScaleId: 'right',
    });
    
    sellSeriesRef.current.applyOptions({
      priceScaleId: 'right',
    });

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

    // Set data for both series
    buySeriesRef.current.setData(buyChartData);
    sellSeriesRef.current.setData(sellChartData);
  }, [buyData, sellData, currentPrice]);

  return (
    <div className="bg-surface rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-sm font-medium text-white">Price Impact Curve</h3>
        <div className="flex items-center space-x-4 text-xs">
          <span className="flex items-center">
            <span className="w-3 h-0.5 bg-green-500 mr-1"></span>
            Buy
          </span>
          <span className="flex items-center">
            <span className="w-3 h-0.5 bg-red-500 mr-1"></span>
            Sell
          </span>
        </div>
      </div>
      <div ref={containerRef} />
      <div className="px-4 py-2 text-xs text-gray-400 text-center border-t border-border">
        Execution Price (ETH/APPL) vs Trade Size (ETH) | Reference: {currentPrice ? currentPrice.toFixed(4) : 'N/A'} ETH/APPL
      </div>
    </div>
  );
}
