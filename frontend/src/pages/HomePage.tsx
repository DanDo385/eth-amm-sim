// HomePage.tsx - Lazy-loads the live AMM dashboard (code-split for Vite).
//
// Suspense fallback explains local vs hosted entry points so cold loads
// on Vercel still orient viewers during Loom demos.
//
// CONNECTIONS:
//  - Dashboard: components/Dashboard.tsx (heavy charts + WS)
//  - Route:     App.tsx index route

import { Suspense, lazy } from 'react';

const Dashboard = lazy(() => import('@/components/Dashboard'));

export default function HomePage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-[50vh] flex-col items-center justify-center gap-3 text-gray-400">
          <p className="text-lg">Loading dashboard…</p>
          <p className="text-sm">
            Loading the dashboard from this origin. For local demos run{' '}
            <span className="font-mono text-blue-400">make up</span>; hosted UI is{' '}
            <a
              href="https://eth-amm-sim.vercel.app"
              className="font-mono text-blue-400 underline"
              target="_blank"
              rel="noreferrer"
            >
              eth-amm-sim.vercel.app
            </a>
            .
          </p>
        </div>
      }
    >
      <Dashboard />
    </Suspense>
  );
}
