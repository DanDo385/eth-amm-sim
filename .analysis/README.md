# .analysis/ — oss-project-analyzer output

Structural analysis of **eth-amm-sim** — a Go + Solidity + React AMM execution
simulation running against a local Anvil chain.

- **Generated:** 2026-08-31T16:43:28Z
- **Commit:** `ec8e3d9e0c3bf1b3e836539716b7671408c76154` (detached HEAD)
- **Analyzer:** oss-project-analyzer skill (static/structural pass)

## Outputs

| File | Contents |
|------|----------|
| [`metrics.md`](./metrics.md) | Lines of code by language (`cloc` over git-tracked files; `pygount` unavailable), per-component breakdown, test LOC and test-to-source ratios. |
| [`architecture-map.txt`](./architecture-map.txt) | ASCII architecture diagram: header (project/scope/type/root/commit/timestamp), top-level layout, runtime component topology across the browser / Go backend / Anvil boundaries, critical execution paths, and coupling notes. |
| [`deep-analysis.md`](./deep-analysis.md) | Full structural analysis: commit + working-tree state, language/LOC summary, top-level module map (backend packages, contracts, frontend), critical execution path, 8 dependency/coupling observations, test evidence (Go + Foundry both pass), 7 optimization hypotheses with confidence levels, and a structural risk register. |
| [`refinement-logs/`](./refinement-logs/) | Reserved for iterative refinement notes (empty on first pass). |

## How this was produced

- LOC: `cloc --vcs=git` (excludes `.git/`, `frontend/node_modules/`,
  `contracts/lib/` submodules, build output).
- Structure: read of `AGENTS.md`, `Makefile`, `backend/cmd/simulator/main.go`,
  engine/server/store/bots/config sources, `contracts/src/*.sol`,
  `contracts/script/Deploy.s.sol`, and `frontend/src/` entrypoints + hooks.
- Test evidence: `cd backend && go test ./...` (PASS) and
  `cd contracts && forge test` (22/22 PASS) executed during analysis.

## Key takeaways

1. Clean layered Go backend; two `errgroup`s cleanly own all goroutines
   (process-level + per-session bot pool).
2. `internal/config` is the cross-language hub — the 30-account roster is
   re-derived independently in Go and in `Deploy.s.sol` with no shared schema.
3. Largest test gap: `internal/bots/` (~850 LOC of strategy logic, 0 tests) and
   the entire frontend.
4. Two representations of the AMM ABI (hand-packed calldata in `executor.go` +
   generated bindings) invite silent drift.
5. In-memory stores are unbounded — fine for a 3-minute demo, a slow leak for the
   always-on staging instance.
