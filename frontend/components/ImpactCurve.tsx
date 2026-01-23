'use client';

import { useEffect, useRef } from 'react';
import { createChart, IChartApi, ISeriesApi, LineData, Time } from 'lightweight-charts';
import type { ImpactPoint } from '@/types';

interface ImpactCurveProps {
  buyData: ImpactPoint[];
  sellData: ImpactPoint[];
  height?: number;
}

export function ImpactCurve({ buyData, sellData, height = 200 }: ImpactCurveProps) {
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
      },
      timeScale: {
        visible: false,
      },
    });

    const buySeries = chart.addLineSeries({
      color: '#22c55e',
      lineWidth: 2,
      title: 'Buy Impact',
    });

    const sellSeries = chart.addLineSeries({
      color: '#ef4444',
      lineWidth: 2,
      title: 'Sell Impact',
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
    if (!buySeriesRef.current || !sellSeriesRef.current) return;

    // Convert to chart format (using trade size as x-axis)
    const buyChartData: LineData[] = buyData.map((p, i) => ({
      time: i as Time,
      value: p.impactBps,
    }));

    const sellChartData: LineData[] = sellData.map((p, i) => ({
      time: i as Time,
      value: Math.abs(p.impactBps), // Show absolute value for sells
    }));

    buySeriesRef.current.setData(buyChartData);
    sellSeriesRef.current.setData(sellChartData);
  }, [buyData, sellData]);

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
        Impact (bps) vs Trade Size
      </div>
    </div>
  );
}
