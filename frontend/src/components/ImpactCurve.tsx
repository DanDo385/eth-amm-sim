// ImpactCurve.tsx - Constant product curve (ETH vs APPL reserves).
//
// Visualizes the pool reserves along the invariant x*y=k curve, with a highlighted
// trade path from selling 500 APPL to buying 100 APPL. This makes the price
// change visible along the curve as the pool moves through different reserve
// states.
//
// CONNECTIONS:
//  - Pool reserves: lpMetrics (currentApples/currentETH) from /lp/metrics
//  - Fallback:      infer reserves from impact curve points when metrics not ready
//  - Backend data:  metrics/impact.go CalculateBuyCurve/CalculateSellCurve
//  - REST endpoint: server/handlers.go handleGetImpactCurve
//  - Reserve updates: main.go pollPrices → impact.UpdateReserves
//  - Types:         types/index.ts ImpactPoint, LPMetrics

import { useMemo, useRef, useState } from 'react';
import type { ImpactPoint, LPMetrics } from '@/types';

interface ImpactCurveProps {
  buyData: ImpactPoint[];
  sellData: ImpactPoint[];
  lpMetrics?: LPMetrics | null;
  height?: number;
}

const FEE_RATE = 0.003; // 0.30%
const SELL_APPL = 500;
const BUY_APPL = 100;

type Reserves = {
  apples: number;
  eth: number;
  source: 'lp' | 'impact';
};

const formatNumber = (value: number, maxDecimals: number = 2) => {
  if (!Number.isFinite(value)) return '-';
  return value.toLocaleString('en-US', { maximumFractionDigits: maxDecimals });
};

const inferReservesFromImpact = (buyData: ImpactPoint[], sellData: ImpactPoint[]): Reserves | null => {
  const buyPoint = buyData.find((p) => p.tradeSize > 0 && p.executePrice > 0 && p.spotPrice > 0);
  if (buyPoint) {
    const size = buyPoint.tradeSize;
    const spot = buyPoint.spotPrice;
    const amountInAfterFee = size * (1 - FEE_RATE);
    const amountOut = size / buyPoint.executePrice;

    const denom = amountInAfterFee / spot - amountOut;
    if (denom > 0) {
      const ethReserve = (amountOut * amountInAfterFee) / denom;
      const applesReserve = ethReserve / spot;
      if (Number.isFinite(ethReserve) && ethReserve > 0 && Number.isFinite(applesReserve) && applesReserve > 0) {
        return { apples: applesReserve, eth: ethReserve, source: 'impact' };
      }
    }
  }

  const sellPoint = sellData.find((p) => p.tradeSize > 0 && p.executePrice > 0 && p.spotPrice > 0);
  if (!sellPoint) return null;

  const size = sellPoint.tradeSize;
  const spot = sellPoint.spotPrice;
  const amountInAfterFee = size * (1 - FEE_RATE);
  const amountOut = sellPoint.executePrice * size;

  const denom = amountInAfterFee * spot - amountOut;
  if (denom <= 0) return null;

  const applesReserve = (amountOut * amountInAfterFee) / denom;
  const ethReserve = spot * applesReserve;

  if (!Number.isFinite(ethReserve) || ethReserve <= 0 || !Number.isFinite(applesReserve) || applesReserve <= 0) {
    return null;
  }

  return { apples: applesReserve, eth: ethReserve, source: 'impact' };
};

export const ImpactCurve = ({ buyData, sellData, lpMetrics, height = 280 }: ImpactCurveProps) => {
  const svgRef = useRef<SVGSVGElement | null>(null);
  const [hoverPoint, setHoverPoint] = useState<{
    svgX: number;
    svgY: number;
    reserveX: number;
    price: number;
    deltaPct: number;
  } | null>(null);

  const reserves = useMemo<Reserves | null>(() => {
    if (lpMetrics && lpMetrics.currentApples > 0 && lpMetrics.currentETH > 0) {
      return { apples: lpMetrics.currentApples, eth: lpMetrics.currentETH, source: 'lp' };
    }
    return inferReservesFromImpact(buyData, sellData);
  }, [buyData, sellData, lpMetrics]);

  const {
    curvePath,
    segmentPath,
    xSell,
    xBuy,
    xSpot,
    ySell,
    yBuy,
    ySpot,
    priceSell,
    priceBuy,
    priceSpot,
    xTicks,
    yTicks,
    adjustedSell,
    adjustedBuy,
    viewBox,
  } = useMemo(() => {
    if (!reserves) {
      return {
        curvePath: '',
        segmentPath: '',
        xSell: 0,
        xBuy: 0,
        xSpot: 0,
        ySell: 0,
        yBuy: 0,
        ySpot: 0,
        priceSell: 0,
        priceBuy: 0,
        priceSpot: 0,
        xTicks: [],
        yTicks: [],
        adjustedSell: SELL_APPL,
        adjustedBuy: BUY_APPL,
        viewBox: '0 0 1000 520',
      };
    }

    const width = 1000;
    const heightBox = 520;
    const margin = { left: 90, right: 30, top: 40, bottom: 80 };
    const plotWidth = width - margin.left - margin.right;
    const plotHeight = heightBox - margin.top - margin.bottom;

    const apples = reserves.apples;
    const eth = reserves.eth;
    const k = apples * eth;

    let effectiveSell = SELL_APPL;
    let effectiveBuy = BUY_APPL;
    if (apples - effectiveBuy <= 0) {
      effectiveBuy = Math.max(1, apples * 0.1);
    }

    const xSpotValue = apples;
    const xSellValue = apples + effectiveSell;
    const xBuyValue = Math.max(1, apples - effectiveBuy);

    const xMinBase = Math.min(xBuyValue, xSpotValue, xSellValue);
    const xMaxBase = Math.max(xBuyValue, xSpotValue, xSellValue);
    const xPadding = (xMaxBase - xMinBase) * 0.15;
    const xMin = Math.max(1, xMinBase - xPadding);
    const xMax = xMaxBase + xPadding;

    const yMinBase = k / xMax;
    const yMaxBase = k / xMin;
    const yPadding = (yMaxBase - yMinBase) * 0.12;
    const yMin = Math.max(0.0001, yMinBase - yPadding);
    const yMax = yMaxBase + yPadding;

    const xToPx = (x: number) => margin.left + ((x - xMin) / (xMax - xMin)) * plotWidth;
    const yToPx = (y: number) => margin.top + (1 - (y - yMin) / (yMax - yMin)) * plotHeight;

    const samplePoints = (from: number, to: number, steps: number) => {
      const points: Array<{ x: number; y: number }> = [];
      for (let i = 0; i <= steps; i += 1) {
        const t = i / steps;
        const x = from + (to - from) * t;
        const y = k / x;
        points.push({ x, y });
      }
      return points;
    };

    const buildPath = (points: Array<{ x: number; y: number }>) => {
      return points
        .map((p, index) => {
          const cmd = index === 0 ? 'M' : 'L';
          return `${cmd}${xToPx(p.x).toFixed(2)},${yToPx(p.y).toFixed(2)}`;
        })
        .join(' ');
    };

    const curvePoints = samplePoints(xMin, xMax, 160);
    const segmentPoints = samplePoints(xSellValue, xBuyValue, 80);

    const priceAt = (x: number) => k / (x * x);

    const priceSellValue = priceAt(xSellValue);
    const priceBuyValue = priceAt(xBuyValue);
    const priceSpotValue = priceAt(xSpotValue);

    const buildTicks = (min: number, max: number, count: number) => {
      const ticks: number[] = [];
      for (let i = 0; i <= count; i += 1) {
        const t = i / count;
        ticks.push(min + (max - min) * t);
      }
      return ticks;
    };

    return {
      curvePath: buildPath(curvePoints),
      segmentPath: buildPath(segmentPoints),
      xSell: xSellValue,
      xBuy: xBuyValue,
      xSpot: xSpotValue,
      ySell: k / xSellValue,
      yBuy: k / xBuyValue,
      ySpot: k / xSpotValue,
      priceSell: priceSellValue,
      priceBuy: priceBuyValue,
      priceSpot: priceSpotValue,
      xTicks: buildTicks(xMin, xMax, 4),
      yTicks: buildTicks(yMin, yMax, 4),
      adjustedSell: effectiveSell,
      adjustedBuy: effectiveBuy,
      viewBox: `0 0 ${width} ${heightBox}`,
    };
  }, [reserves]);

  const xMin = xTicks.length > 0 ? xTicks[0] : 0;
  const xMax = xTicks.length > 0 ? xTicks[xTicks.length - 1] : 0;
  const yMin = yTicks.length > 0 ? yTicks[0] : 0;
  const yMax = yTicks.length > 0 ? yTicks[yTicks.length - 1] : 0;
  const xSpan = xMax - xMin;
  const ySpan = yMax - yMin;
  const k = xSpot * ySpot;
  const xToPx = (x: number) => 90 + ((x - xMin) / xSpan) * 880;
  const yToPx = (y: number) => 40 + (1 - (y - yMin) / ySpan) * 400;
  const pxToX = (px: number) => xMin + ((px - 90) / 880) * xSpan;

  const handleMouseMove = (evt: React.MouseEvent<SVGSVGElement>) => {
    if (!reserves || !svgRef.current || xSpan <= 0 || ySpan <= 0 || k <= 0) {
      return;
    }
    const rect = svgRef.current.getBoundingClientRect();
    const localX = ((evt.clientX - rect.left) / rect.width) * 1000;
    const clampedX = Math.min(970, Math.max(90, localX));
    const reserveX = pxToX(clampedX);
    if (reserveX <= 0) {
      setHoverPoint(null);
      return;
    }
    const price = k / (reserveX * reserveX);
    const reserveY = k / reserveX;
    const svgY = yToPx(reserveY);
    const deltaPct = priceSpot > 0 ? ((price - priceSpot) / priceSpot) * 100 : 0;
    setHoverPoint({ svgX: xToPx(reserveX), svgY, reserveX, price, deltaPct });
  };

  return (
    <div className="bg-surface rounded-lg border border-border overflow-hidden">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-sm font-medium text-white">AMM Reserve Curve</h3>
        <div className="text-xs text-gray-400">
          {reserves ? (
            <span>
              Spot: {formatNumber(priceSpot, 4)} ETH/APPL
              {hoverPoint ? (
                <span className="ml-3 text-cyan-300">
                  Hover: {formatNumber(hoverPoint.price, 4)} ({hoverPoint.deltaPct >= 0 ? '+' : ''}{formatNumber(hoverPoint.deltaPct, 3)}%)
                </span>
              ) : null}
            </span>
          ) : 'Waiting for pool reserves'}
        </div>
      </div>
      <div className="px-4 py-4">
        {!reserves ? (
          <div className="h-[220px] flex items-center justify-center text-sm text-gray-500">
            Loading pool reserves...
          </div>
        ) : (
          <div className="w-full" style={{ height }}>
            <svg
              ref={svgRef}
              viewBox={viewBox}
              width="100%"
              height="100%"
              onMouseMove={handleMouseMove}
              onMouseLeave={() => setHoverPoint(null)}
            >
              <defs>
                <linearGradient id="impactSegment" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stopColor="#ef4444" />
                  <stop offset="50%" stopColor="#f59e0b" />
                  <stop offset="100%" stopColor="#22c55e" />
                </linearGradient>
              </defs>

              {/* Axes */}
              <line x1="90" y1="40" x2="90" y2="440" stroke="#374151" strokeWidth="2" />
              <line x1="90" y1="440" x2="970" y2="440" stroke="#374151" strokeWidth="2" />

              {/* Ticks + grid */}
              {xTicks.map((tick, index) => {
                const x = 90 + ((tick - xTicks[0]) / (xTicks[xTicks.length - 1] - xTicks[0])) * 880;
                return (
                  <g key={`x-${index}`}>
                    <line x1={x} y1="440" x2={x} y2="448" stroke="#4b5563" />
                    <line x1={x} y1="440" x2={x} y2="40" stroke="#1f2937" strokeDasharray="4 6" />
                    <text x={x} y="468" fill="#9ca3af" fontSize="12" textAnchor="middle">
                      {formatNumber(tick, 0)}
                    </text>
                  </g>
                );
              })}

              {yTicks.map((tick, index) => {
                const y = 40 + (1 - (tick - yTicks[0]) / (yTicks[yTicks.length - 1] - yTicks[0])) * 400;
                return (
                  <g key={`y-${index}`}>
                    <line x1="84" y1={y} x2="90" y2={y} stroke="#4b5563" />
                    <line x1="90" y1={y} x2="970" y2={y} stroke="#1f2937" strokeDasharray="4 6" />
                    <text x="78" y={y + 4} fill="#9ca3af" fontSize="12" textAnchor="end">
                      {formatNumber(tick, 0)}
                    </text>
                  </g>
                );
              })}

              {/* Constant product curve */}
              <path d={curvePath} fill="none" stroke="#60a5fa" strokeWidth="2.5" />

              {/* Highlighted trade path */}
              <path d={segmentPath} fill="none" stroke="url(#impactSegment)" strokeWidth="4" />

              {/* Spot / Sell / Buy markers */}
              <circle cx={90 + ((xSpot - xTicks[0]) / (xTicks[xTicks.length - 1] - xTicks[0])) * 880} cy={40 + (1 - (ySpot - yTicks[0]) / (yTicks[yTicks.length - 1] - yTicks[0])) * 400} r="6" fill="#38bdf8" />
              <circle cx={90 + ((xSell - xTicks[0]) / (xTicks[xTicks.length - 1] - xTicks[0])) * 880} cy={40 + (1 - (ySell - yTicks[0]) / (yTicks[yTicks.length - 1] - yTicks[0])) * 400} r="7" fill="#ef4444" />
              <circle cx={90 + ((xBuy - xTicks[0]) / (xTicks[xTicks.length - 1] - xTicks[0])) * 880} cy={40 + (1 - (yBuy - yTicks[0]) / (yTicks[yTicks.length - 1] - yTicks[0])) * 400} r="7" fill="#22c55e" />

              {hoverPoint && (
                <>
                  <line x1={hoverPoint.svgX} y1="40" x2={hoverPoint.svgX} y2="440" stroke="#22d3ee" strokeWidth="1.5" strokeDasharray="5 4" />
                  <circle cx={hoverPoint.svgX} cy={hoverPoint.svgY} r="5" fill="#22d3ee" />
                </>
              )}

              {/* Labels */}
              <text x="90" y="24" fill="#9ca3af" fontSize="12">ETH in Pool</text>
              <text x="970" y="480" fill="#9ca3af" fontSize="12" textAnchor="end">APPL in Pool</text>

              <text
                x={90 + ((xSell - xTicks[0]) / (xTicks[xTicks.length - 1] - xTicks[0])) * 880}
                y={40 + (1 - (ySell - yTicks[0]) / (yTicks[yTicks.length - 1] - yTicks[0])) * 400 - 16}
                fill="#f87171"
                fontSize="12"
                textAnchor="middle"
              >
                Sell {formatNumber(adjustedSell, 0)} APPL ({formatNumber(priceSell, 4)} ETH/APPL)
              </text>
              <text
                x={90 + ((xBuy - xTicks[0]) / (xTicks[xTicks.length - 1] - xTicks[0])) * 880}
                y={40 + (1 - (yBuy - yTicks[0]) / (yTicks[yTicks.length - 1] - yTicks[0])) * 400 + 20}
                fill="#4ade80"
                fontSize="12"
                textAnchor="middle"
              >
                Buy {formatNumber(adjustedBuy, 0)} APPL ({formatNumber(priceBuy, 4)} ETH/APPL)
              </text>
            </svg>
          </div>
        )}
      </div>
      <div className="px-4 py-2 text-xs text-gray-400 text-center border-t border-border">
        {reserves
          ? `Reserves: ${formatNumber(reserves.apples, 2)} APPL / ${formatNumber(reserves.eth, 2)} ETH${reserves.source === 'impact' ? ' (estimated)' : ''}`
          : 'Calculating reserves from impact curve'}
      </div>
    </div>
  );
};
