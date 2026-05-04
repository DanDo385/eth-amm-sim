# Implementation Plan (Simple-Model Friendly)

This plan is designed so a simpler coding model (Composer/Codex Spark) can execute reliably with low context.

## Objectives

1. Raise project quality from solid prototype toward professional demo quality.
2. Add guardrails (tests + CI) so improvements are durable.
3. Preserve demo velocity and existing architecture narrative.

---

## Global Execution Rules

- Follow order: `P0 -> P1 -> P2 -> P3(optional)`.
- Keep each phase in a separate PR.
- Avoid unrelated refactors while implementing a phase.
- After each phase:
  - run required validation commands,
  - update docs if behavior changed,
  - include evidence of successful validation in PR notes.
- If blocked, stop and report:
  - what failed,
  - exact command/error,
  - smallest proposed fix.

---

## Preflight (before Phase P0)

### Steps

1. Create branch:
   - `git checkout -b chore/portfolio-hardening-pass-1`
2. Capture baseline:
   - `cd backend && GOTOOLCHAIN=auto go build ./...`
   - `cd frontend && npm run build`
   - `cd contracts && forge test -vvv`
3. Optional: save command outputs in `docs/IMPLEMENTATION_NOTES.md`.

### Exit criteria

- Baseline build/test status is documented.

---

## Phase P0: Safety + Demo UX Quick Wins

### Scope

#### P0.1 Surface session API errors in UI

- Files:
  - `frontend/components/Dashboard.tsx`
  - `frontend/components/SessionControls.tsx`
- Work:
  - Pass `error` from `useSession()` through `Dashboard` into `SessionControls`.
  - Render a clear inline error panel in `SessionControls`.
  - Ensure error clears on next successful action/state refresh.
- Done when:
  - failed session action visibly reports error in UI.

#### P0.2 Strict request validation for session start

- Files:
  - `backend/internal/server/handlers.go`
- Work:
  - Check JSON decode errors in `handleSessionStart`.
  - Return `400` with clear message on malformed payload.
- Done when:
  - malformed `POST /session/start` returns `400`.

#### P0.3 Fail fast on contract binding construction

- Files:
  - `backend/internal/engine/executor.go`
  - `backend/cmd/simulator/main.go`
  - any other `NewExecutor` callsites
- Work:
  - Change constructor signature to `NewExecutor(...) (*Executor, error)`.
  - Handle errors from `contracts.NewAppleAMM` and `contracts.NewAppleToken`.
  - Exit startup with wrapped error if bindings fail.
- Done when:
  - startup aborts cleanly on binding failure.

#### P0.4 Align documented toolchain requirements

- Files:
  - `README.md`
- Work:
  - Ensure prerequisites match current `backend/go.mod` + frontend reality.
  - Clarify that frontend currently uses a Next canary release.
- Done when:
  - docs and build requirements are consistent.

### Validation

- `cd backend && GOTOOLCHAIN=auto go build ./...`
- `cd frontend && npm run lint && npm run build`

### Suggested commit message

- `fix: improve startup/input safety and session error visibility`

---

## Phase P1: Testing Foundation + CI

### Scope

#### P1.1 Add backend unit tests

- Files (new):
  - `backend/internal/config/config_test.go`
  - `backend/internal/metrics/ewma_test.go`
  - `backend/internal/engine/orchestrator_test.go`
  - optional: `backend/internal/chain/client_test.go`
- Coverage targets:
  - address-loading behavior,
  - EWMA update/reset behavior,
  - orchestrator cancellation behavior,
  - optional JSON-RPC payload semantics for `SetBalance`.

#### P1.2 Add frontend hook tests (deferred)

- **Status:** not added in-repo yet. CI runs `npm run lint` + `npm run build` for the frontend.
- When revisited, add Vitest (or Jest) + `useSession` tests and wire `npm test` into `.github/workflows/ci.yml`.

#### P1.3 Add project-level GitHub Actions

- Files (new):
  - `.github/workflows/ci.yml`
- Jobs:
  - backend: `go test ./...` + `go build ./...`
  - frontend: `npm ci`, `npm run lint`, `npm run build`, tests
  - contracts: `forge test -vvv`

### Validation

- Run all CI commands locally once before opening PR.

### Suggested commit message

- `test: add baseline backend/frontend tests and CI pipeline`

---

## Phase P2: Security + Operability Hardening

### Scope

#### P2.1 Origin allowlist controls

- Files:
  - `backend/internal/server/server.go`
  - `backend/internal/config/*` (if needed)
  - `README.md`
- Work:
  - Add env-driven allowlist for CORS + WebSocket origin checks.
  - Keep local dev workflow functional by default.
- Done when:
  - unknown origins are rejected in non-dev mode.

#### P2.2 Health endpoints

- Files:
  - backend server/handler files
- Work:
  - Add `/healthz` and `/readyz`.
  - Include minimal server health fields (uptime, ws clients, queue depth).
- Done when:
  - endpoints return machine-readable JSON quickly.

#### P2.3 Backpressure behavior docs

- Files:
  - `README.md`
  - optional: `docs/SESSION_BOT_LIFECYCLE.md`
- Work:
  - document WS queue/drop behavior and operator playbook.

### Validation

- `cd backend && GOTOOLCHAIN=auto go build ./...`
- Manual smoke:
  - start stack,
  - confirm ws + reset still work,
  - hit health endpoints.

### Suggested commit message

- `feat: add origin controls and backend health observability`

---

## Phase P3 (Optional): Demo Polish

### Scope

1. Unify account update behavior between dashboard and performance page.
2. Add top KPI strip for immediate business signal.
3. Improve loading/empty/error states in chart-heavy panels.
4. Add explicit "Demo reset" UX copy and feedback.

### Suggested commit message

- `feat: improve demo UX clarity and account update consistency`

---

## Recommended PR Sequence

1. `PR-1`: P0 safety + UX quick wins
2. `PR-2`: P1 tests + CI
3. `PR-3`: P2 security + operability
4. `PR-4` (optional): P3 demo polish

---

## Project Done Definition

- [x] Backend builds with `GOTOOLCHAIN=auto`.
- [x] Frontend lint + build pass.
- [x] Contracts tests pass (local / CI `forge test`).
- [x] Root CI workflow exists (backend test+build, frontend lint+build, contracts tests).
- [x] Session/API failures are visible in UI.
- [x] Malformed session input returns `400` with clear error.
- [x] Executor constructor fails fast on binding errors.
- [x] Origin policy is configurable and documented.
- [x] README prerequisites reflect actual required versions.
- [ ] Frontend unit tests for hooks (P1.2 deferred).

---

## Copy/Paste Prompt for a Simple Model

Use this prompt one phase at a time.

```text
Implement Phase P0 from IMPLEMENTATION_PLAN.md.

Constraints:
- Only change files required by P0 scope.
- Keep diffs minimal and focused.
- Do not refactor unrelated code.
- Run:
  1) cd backend && GOTOOLCHAIN=auto go build ./...
  2) cd frontend && npm run lint && npm run build
- If any command fails, fix before finishing.

Deliverables:
1) files changed,
2) concise rationale per file,
3) validation command outputs summary,
4) suggested commit message.
```

