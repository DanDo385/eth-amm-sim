import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'ETH-AMM-SIM | DeFi Market Simulator',
  description: 'Portfolio-grade DeFi AMM simulation with TradFi analytics',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className="min-h-screen">
        <nav className="border-b border-border bg-surface">
          <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <span className="text-xl font-bold text-white">ETH-AMM-SIM</span>
              <span className="text-xs px-2 py-0.5 bg-blue-600 rounded text-white">DEMO</span>
            </div>
            <div className="flex items-center space-x-6">
              <a href="/" className="text-gray-300 hover:text-white transition">Dashboard</a>
              <a href="/performance" className="text-gray-300 hover:text-white transition">Performance</a>
            </div>
          </div>
        </nav>
        <main className="max-w-7xl mx-auto px-4 py-6">
          {children}
        </main>
      </body>
    </html>
  )
}
