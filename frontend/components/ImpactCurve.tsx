'use client';

import { useEffect, useRef } from 'react';
import { createChart, IChartApi, ISeriesApi, LineData, Time } from 'lightweight-charts';
import type { ImpactPoint } from '@/types';

interface ImpactCurveProps {
  buyData: ImpactPoint[];
  sellData: ImpactPoint[];
  currentPrice?: number; // Current spot price to center y-axis
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
        autoScale: true, // Auto-scale for y-axis (execution price)
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      },
      leftPriceScale: {
        visible: false,
      },
      timeScale: {
        visible: true, // Make time scale visible for custom formatting
        timeVisible: false, // Hide default time labels
        secondsVisible: false,
        rightOffset: 10,
        lockVisibleTimeRangeOnResize: true, // Keep range constant
        // Custom formatter for x-axis to show trade sizes
        tickMarkFormatter: (time: Time, tickMarkType: number, locale: string) => {
          // 'time' here is our tradeSize (negative for sells, positive for buys)
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

    // Note: setVisibleRange will be called after data is loaded in the update effect

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
    
    // Use provided currentPrice, or fallback to data spot price, or default
    const spotPrice = currentPrice || buyData[0]?.spotPrice || sellData[0]?.spotPrice || 1.0;
    
    // If no data, clear the chart but keep it visible
    if (buyData.length === 0 && sellData.length === 0) {
      buySeriesRef.current.setData([]);
      sellSeriesRef.current.setData([]);
      return;
    }
    
    // Convert to chart format
    // X-axis: Trade Size (in ETH) - negative for sells, positive for buys
    // Y-axis: Execution Price (ETH per APPL) - centered at spot price
    const buyChartData: LineData[] = buyData
      .filter(p => p.tradeSize > 0 && p.tradeSize <= 200 && p.executePrice > 0) // Only show valid data up to +200
      .map((p) => ({
        time: p.tradeSize as Time, // Positive for buys
        value: p.executePrice, // Execution price on y-axis
      }))
      .sort((a, b) => Number(a.time) - Number(b.time)); // Ensure ascending order

    const sellChartData: LineData[] = sellData
      .filter(p => p.tradeSize > 0 && p.tradeSize <= 250 && p.executePrice > 0) // Only show valid data up to 250
      .map((p) => ({
        time: -p.tradeSize as Time, // Negative for sells (centered at 0)
        value: p.executePrice, // Execution price on y-axis
      }))
      .sort((a, b) => Number(a.time) - Number(b.time)); // Ensure ascending order (e.g., -250, -200, -100, -50, -1)

    // Calculate price range from all execution prices
    const allPrices = buyChartData.length > 0 || sellChartData.length > 0
      ? [...buyChartData.map(d => d.value), ...sellChartData.map(d => d.value)]
      : [spotPrice];
    
    const minPrice = allPrices.length > 0 ? Math.min(...allPrices) : spotPrice;
    const maxPrice = allPrices.length > 0 ? Math.max(...allPrices) : spotPrice;
    
    // Center the range around spot price (current market price)
    // Calculate deviation from spot price
    const maxDeviation = Math.max(
      Math.abs(maxPrice - spotPrice),
      Math.abs(minPrice - spotPrice)
    );
    
    // Set y-axis range centered at spot price
    // Use at least 10% margin on each side, or actual deviation + 20% padding, whichever is larger
    const margin = Math.max(maxDeviation * 1.2, spotPrice * 0.1);
    const yAxisMin = Math.max(0.01, spotPrice - margin); // Minimum 0.01 to avoid zero
    const yAxisMax = spotPrice + margin;

    // Ensure both series use the same price scale
    buySeriesRef.current.applyOptions({
      priceScaleId: 'right',
    });
    
    sellSeriesRef.current.applyOptions({
      priceScaleId: 'right',
    });

    // Set y-axis scale centered at spot price
    if (chartRef.current) {
      const timeScale = chartRef.current.timeScale();
      if (!timeScale) return; // Safety check
      
      // Always center y-axis at spot price with calculated range
      chartRef.current.priceScale('right').applyOptions({
        autoScale: false,
        minValue: yAxisMin,
        maxValue: yAxisMax,
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      });
      
      // Configure time scale to show trade sizes
      timeScale.applyOptions({
        visible: true,
        timeVisible: false, // Hide default time labels
        secondsVisible: false,
        rightOffset: 10,
        lockVisibleTimeRangeOnResize: true, // Keep range constant
        // Custom formatter for x-axis to show trade sizes
        tickMarkFormatter: (time: Time, tickMarkType: number, locale: string) => {
          // 'time' here is our tradeSize (negative for sells, positive for buys)
          const size = Number(time);
          return size >= 0 ? `+${size}` : `${size}`;
        },
      });
      
      // Set x-axis to show -250 to +200 centered at 0 (current price)
      // This must be set AFTER applyOptions and after data is set
      try {
        timeScale.setVisibleRange({
          from: -250 as Time,
          to: 200 as Time,
        });
      } catch (e) {
        // If setVisibleRange fails, try again after a short delay
        setTimeout(() => {
          const ts = chartRef.current?.timeScale();
          if (ts) {
            ts.setVisibleRange({
              from: -250 as Time,
              to: 200 as Time,
            });
          }
        }, 100);
      }
    }

    // Set data for both series
    buySeriesRef.current.setData(buyChartData);
    sellSeriesRef.current.setData(sellChartData);
    
    // Update price scale centered at spot price (current market price)
    if (chartRef.current) {
      const priceScale = chartRef.current.priceScale('right');
      if (priceScale) {
        priceScale.applyOptions({
          autoScale: false,
          minValue: yAxisMin,
          maxValue: yAxisMax,
          scaleMargins: {
            top: 0.1,
            bottom: 0.1,
          },
        });
      }
      
      // Always ensure x-axis range stays constant at -250 to +200
      const timeScale = chartRef.current.timeScale();
      if (timeScale) {
        try {
          timeScale.setVisibleRange({
            from: -250 as Time,
            to: 200 as Time,
          });
        } catch (e) {
          // Ignore errors if timeScale is not ready - will retry on next update
        }
      }
    }
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
        Price Impact: Execution Price (centered at {currentPrice ? currentPrice.toFixed(4) : 'spot'} ETH/APPL) vs Trade Size (ETH)
      </div>
    </div>
  );
}
