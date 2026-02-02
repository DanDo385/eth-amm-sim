# Language Comparison: Go vs TypeScript vs Python

## Current stack: Go

The backend is a concurrent trading simulator with 21+ bot goroutines, real-time WebSocket streaming, constant RPC polling, big-integer AMM math, and an entirely in-memory data store. Go is a strong fit for this workload.

## Would TypeScript (Node.js) or Python be noticeably worse?

### Memory

| | Go | Node.js (TS) | Python |
|---|---|---|---|
| Base process | ~10-20 MB | ~50-80 MB | ~30-50 MB |
| Per-object overhead | Low (structs, value types) | High (everything is a heap object) | Higher (everything is an object + dict) |
| Big integers | `math/big` (efficient) | Native `BigInt` (decent) | Native `int` (decent) |
| GC pressure | Low (value types, goroutines are cheap) | Moderate (V8 GC handles it but more allocations) | Higher (reference counting + cycle collector) |

For the unbounded price history and trade blotter arrays, Node.js would use roughly **2-4x** more memory per entry and Python roughly **3-5x** more, due to object overhead.

### Concurrency (the biggest difference)

- **Go**: 21 goroutines are trivial (~2-8 KB stack each). Goroutines are designed exactly for this workload — many lightweight concurrent tasks doing I/O and periodic computation.
- **Node.js**: Single-threaded event loop. The bot logic would need to be restructured as async/await. It would *work* since the bottleneck is I/O (RPC calls), but CPU-bound work (EWMA, volatility, impact curves) would block the event loop. You'd need `worker_threads` for the heavier math, adding complexity.
- **Python**: The GIL means true parallelism requires `multiprocessing` or `asyncio` for I/O. Running 21 bots concurrently with real-time WebSocket broadcasting would be significantly more complex and slower. `asyncio` works for I/O but CPU-bound stats calculations would stall the loop.

### Computation (EWMA, volatility, TWAP, impact curves)

Go is roughly **5-20x** faster than Python and **2-5x** faster than Node.js for tight numerical loops. The impact curve calculation (50+ data points with big-integer math) and per-tick volatility updates would be meaningfully slower in both alternatives.

## Bottom line

- **TypeScript/Node.js**: Would work but use ~2-3x more memory overall and require architectural changes for CPU-bound work. The concurrency model is less natural for this kind of multi-agent simulation.
- **Python**: Would be noticeably worse — higher memory, slower numerics, and the GIL makes the concurrent bot orchestration genuinely painful. You'd likely need to reach for C extensions (NumPy) or multiprocessing to keep it responsive.

Go is a natural fit for this project. The combination of cheap goroutines, value-type structs, efficient big-integer math, and low GC overhead aligns well with a real-time multi-bot trading simulator. The alternatives would work for a simpler version but would require more engineering to achieve the same responsiveness.
