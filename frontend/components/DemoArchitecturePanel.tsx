const layers = [
  {
    name: 'Foundry / Anvil',
    detail: 'Local EVM chain, deterministic accounts, safe test capital.',
  },
  {
    name: 'Solidity AMM',
    detail: 'AppleToken + AppleAMM contracts deployed once. Logic is fixed at the contract address.',
  },
  {
    name: 'Go / geth engine',
    detail: 'Bots and user trades submit transactions through Ethereum-style bindings.',
  },
  {
    name: 'WebSocket stream',
    detail: 'Trades, price candles, LP metrics, and key events stream from backend to UI.',
  },
  {
    name: 'Next.js dashboard',
    detail: 'Recording surface explains causality: trade → reserves → price/TWAP → LP exposure.',
  },
];

export function DemoArchitecturePanel() {
  return (
    <section className="rounded-xl border border-blue-500/30 bg-surface p-4 shadow-xl shadow-blue-950/10">
      <div className="mb-3 flex items-start justify-between gap-3">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-blue-300">Show Architecture</div>
          <h2 className="mt-1 text-sm font-semibold text-white">Foundry/Anvil → Solidity AMM → Go/geth → WebSocket → Next.js</h2>
        </div>
        <span className="rounded-full border border-purple-400/40 bg-purple-500/15 px-2 py-1 text-[10px] font-bold uppercase tracking-wider text-purple-200">
          Sim
        </span>
      </div>

      <div className="space-y-2">
        {layers.map((layer, index) => (
          <div key={layer.name} className="rounded-lg border border-white/10 bg-black/20 p-3">
            <div className="flex items-center gap-2">
              <span className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-500/20 text-xs font-bold text-blue-200">{index + 1}</span>
              <span className="text-sm font-semibold text-white">{layer.name}</span>
            </div>
            <p className="mt-1 text-xs leading-5 text-gray-400">{layer.detail}</p>
          </div>
        ))}
      </div>

      <div className="mt-3 rounded-lg border border-amber-400/30 bg-amber-500/10 p-3 text-xs leading-5 text-amber-100">
        <span className="font-semibold text-amber-200">Contract immutability callout:</span> after deployment, the backend can call the AMM contract, not rewrite its pricing formula. Same software shape, different risk boundary.
      </div>
    </section>
  );
}
