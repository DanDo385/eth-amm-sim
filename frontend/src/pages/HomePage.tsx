import { Suspense, lazy } from 'react';

const Dashboard = lazy(() => import('@/components/Dashboard'));

export default function HomePage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-[50vh] flex-col items-center justify-center gap-3 text-gray-400">
          <p className="text-lg">Loading dashboard…</p>
          <p className="text-sm">
            Open{' '}
            <a href="http://localhost:3000" className="font-mono text-blue-400 underline">
              http://localhost:3000
            </a>{' '}
            once the Vite dev server is ready.
          </p>
        </div>
      }
    >
      <Dashboard />
    </Suspense>
  );
}
