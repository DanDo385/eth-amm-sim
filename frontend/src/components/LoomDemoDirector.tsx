const beats = [
  {
    time: "0:00-0:20",
    title: "Local EVM + immutable contracts",
    detail: "Ubuntu runs Anvil + Go behind Cloudflare Tunnel; Vercel serves the SPA that proves state changes in real time.",
  },
  {
    time: "0:20-0:55",
    title: "Go/geth bot infrastructure",
    detail: "Show the Ubuntu backend as production-shaped trading infrastructure: signed transactions, RPC calls, logs, websocket events, and deterministic sessions.",
  },
  {
    time: "0:55-1:35",
    title: "Market microstructure under stress",
    detail: "Trigger normal flow first, then a whale trade. Narrate price impact, LP inventory shift, TWAP, and trade blotter causality.",
  },
  {
    time: "1:35-2:25",
    title: "Risk boundary: simulation vs production",
    detail: "Same software shape, different risk boundary: dev chain and test accounts here; audited contracts, keys, monitoring, and circuit breakers in production.",
  },
  {
    time: "2:25-3:00",
    title: "Why employers should care",
    detail: "This is not just a chart. It is an end-to-end protocol lab across Solidity, Go, geth, websockets, state indexing, and frontend observability.",
  },
];

export function LoomDemoDirector() {
  return (
    <section className="rounded-xl border border-cyan-500/30 bg-gradient-to-r from-cyan-950/50 via-slate-950/80 to-purple-950/50 p-4 shadow-2xl shadow-cyan-950/20">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div className="max-w-3xl">
          <div className="text-xs uppercase tracking-[0.25em] text-cyan-300">Loom recording mode · 180 seconds</div>
          <h2 className="mt-1 text-xl font-bold text-white">AMM Execution Lab: local chain → contracts → bots → market state</h2>
          <p className="mt-2 text-sm leading-6 text-slate-300">
            Use this overlay as the demo spine. The visual hook is the whale shock, but the hiring signal is the full engineering loop:
            local EVM-compatible blockchain, immutable Solidity contracts, Go/geth execution, websocket observability, and a UI that explains cause and effect.
          </p>
        </div>
        <div className="rounded-lg border border-white/10 bg-black/30 px-4 py-3 text-sm text-slate-300">
          <div className="font-semibold text-white">Demo focus</div>
          <div className="text-slate-400">AMM execution lab</div>
          <div className="mt-2 font-semibold text-white">Best GIF loop</div>
          <div className="text-slate-400">Start bots → whale row pulses → price/TWAP/LP panels move</div>
        </div>
      </div>
      <div className="mt-4 grid gap-3 lg:grid-cols-5">
        {beats.map((beat) => (
          <div key={beat.time} className="rounded-lg border border-white/10 bg-black/25 p-3">
            <div className="text-xs font-mono text-cyan-300">{beat.time}</div>
            <div className="mt-1 text-sm font-semibold text-white">{beat.title}</div>
            <p className="mt-1 text-xs leading-5 text-slate-400">{beat.detail}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
