'use client';

import dynamic from 'next/dynamic';

const Dashboard = dynamic(() => import('@/components/Dashboard'), {
  ssr: false,
  loading: () => (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-3 text-gray-400">
      <p className="text-lg">Loading dashboard…</p>
      <p className="text-sm">
        When Next prints <span className="font-mono text-gray-300">Ready</span>, open{' '}
        <a href="http://localhost:3000" className="text-blue-400 underline">
          http://localhost:3000
        </a>
      </p>
    </div>
  ),
});

export default function HomeClient() {
  return <Dashboard />;
}
