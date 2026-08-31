# Code Metrics — eth-amm-sim

- **Generated:** 2026-08-31T16:43:28Z
- **Commit:** `ec8e3d9e0c3bf1b3e836539716b7671408c76154` (detached HEAD, tip of `main` history)
- **Tool:** `cloc` v2.02 (`pygount` not installed). Scope limited to git-tracked files
  (`cloc --vcs=git`), which structurally excludes `.git/`, `frontend/node_modules/`,
  `contracts/lib/` (Foundry submodules: forge-std, openzeppelin-contracts),
  `frontend/dist/`, and other build output.

## Lines of code by language (tracked source)

| Language      | Files | Blank | Comment | Code   |
|---------------|------:|------:|--------:|-------:|
| Go            |    37 | 1,463 |   1,834 |  7,143 |
| TypeScript    |    27 |   362 |     400 |  2,907 |
| JSON          |     6 |     0 |       0 |  2,889 |
| Markdown      |     4 |   211 |       0 |    678 |
| Solidity      |     5 |   199 |     301 |    599 |
| Bourne Shell  |     6 |   122 |     555 |    466 |
| CSS           |     1 |    16 |       8 |     69 |
| TOML          |     3 |     7 |       6 |     61 |
| Make          |     1 |    18 |      24 |     44 |
| JavaScript    |     3 |     2 |       1 |     37 |
| HTML          |     1 |     0 |       0 |     20 |
| SVG           |     1 |     0 |       1 |     17 |
| **SUM**       |  **95** | **2,400** | **3,130** | **14,930** |

Notes:
- JSON is dominated by `frontend/package-lock.json` and generated ABI/broadcast
  fixtures; not hand-written source.
- The three JavaScript files are frontend config (`postcss.config.cjs`,
  `tailwind.config.cjs`, `scripts/check-node.cjs`).

## Application code by component

| Component            | Language | Code (approx) | Path |
|----------------------|----------|--------------:|------|
| Backend simulator    | Go       | ~7,143 (all Go) | `backend/` |
| — entrypoint / wiring | Go      | ~570 | `backend/cmd/simulator/main.go` |
| — engine (executor/orchestrator/session) | Go | ~1,700 | `backend/internal/engine/` |
| — metrics (price/LP/EWMA/impact/account) | Go | ~1,340 | `backend/internal/metrics/` |
| — server (REST + WebSocket) | Go | ~1,500 | `backend/internal/server/` |
| — bots (whale/retail/meanrev/base) | Go | ~850 | `backend/internal/bots/` |
| — chain / config / store / contracts bindings | Go | ~1,200 | `backend/internal/{chain,config,store,contracts}/` |
| Contracts            | Solidity | ~599 | `contracts/src/`, `contracts/script/` |
| — `AppleAMM.sol`     | Solidity | 394 (file) | `contracts/src/AppleAMM.sol` |
| — `Deploy.s.sol`     | Solidity | 279 (file) | `contracts/script/Deploy.s.sol` |
| — `AppleToken.sol`   | Solidity |  30 (file) | `contracts/src/AppleToken.sol` |
| Frontend SPA         | TS/TSX   | ~2,907 | `frontend/src/` |
| Dev / deploy scripts | Shell    | ~466 | `scripts/`, `contracts/`, `frontend/scripts/` |

## Test code

| Suite              | Files | Test funcs | Test LOC | Command |
|--------------------|------:|-----------:|---------:|---------|
| Go (`*_test.go`)   |     7 |         24 |      696 | `cd backend && go test ./...` |
| Solidity (`*.t.sol`) |   2 |         22 |      396 | `cd contracts && forge test` |

Go test files: `config/config_test.go`, `server/handlers_test.go`,
`server/security_test.go`, `chain/client_test.go`, `engine/orchestrator_test.go`,
`engine/session_test.go`, `metrics/ewma_test.go`.

Test-to-source ratio (application code): Go ≈ 696 / 7,143 ≈ **9.7%**;
Solidity ≈ 396 / 599 ≈ **66%**.
