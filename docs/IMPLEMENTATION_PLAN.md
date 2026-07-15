# Implementation plan - stable baseline & recovery path

This document is for humans (and agents) who hit repeated **toolchain**, **dependency**, and **“it works on my machine”** failures. It does **not** mean “throw away the simulator logic.” It means: **lock the contract between OS, Node, Go, and the repo** so the *same* demo story runs reliably.

---

## 1. Honest diagnosis (why it feels cursed)

Several unrelated failure modes stacked up:

| Layer | What went wrong (symptom) | Root cause class |
|--------|---------------------------|-------------------|
| **Node / npm** | `tsc` or `vite` “hangs,” `Operation timed out`, `import: command not found` | Wrong interpreter running `.bin` shims (ESM CLIs invoked as shell scripts), Cursor/Homebrew/NVM **PATH order**, corrupted `node_modules` |
| **Frontend stack** | Next canary → Vite migration, lockfile churn | Bleeding-edge pins + interrupted installs |
| **Go** | `packages.Load` / gopls errors | `go` line in `go.mod` vs local `go` patch + **`GOTOOLCHAIN=local`** blocking toolchain fetch |
| **Process hygiene** | Stale listeners on `:3000` / multiple dev servers | tmux + multiple `npm run dev` without `make kill-all` |
| **Cognitive load** | “Which doc is true?” | README/agents drift during rapid fixes |

None of that invalidates **Solidity + Foundry + Go + a small dashboard**. It invalidates an **implicit** toolchain contract.

---

## 2. Non‑negotiables (what we are *not* allowed to break)

- **Demo story:** Anvil → deploy → backend `:8080` → UI `:3000` → WebSocket stream → session/bots.
- **No fake chain data** for core metrics; dashboard reflects backend/chain state.
- **Keep scope:** no new databases, auth, or microservices unless explicitly requested.

---

## 3. Golden path (definition of “project works”)

A maintainer (or CI) should be able to do **exactly this** on a clean machine:

1. Install **Foundry**, **Go** (matches `backend/go.mod` / README), **Node 20 or 22 LTS** (see **`frontend/.nvmrc`**; **Node 25+** currently breaks Vite/Rollup).
2. `make setup` - contracts deps, `frontend` npm install, bindings.
3. `make up` (or manual four terminals from README).
4. Open `http://localhost:3000`, click **Start**, see live updates.
5. From `frontend/`: `npm run verify` - **lint + production bundle** completes without manual hacks.

If any step needs “try five random Stack Overflow fixes,” the **toolchain contract** is still broken.

---

## 4. Dependency constitution (rules going forward)

1. **Pin what demos depend on**
  - **React** / **Vite**: prefer **exact** or tight ranges on anything that broke before; avoid canary unless the whole team opts in.
  - **Go**: `go` line in `go.mod` must match what README claims; prefer **`GOTOOLCHAIN=auto`** in editor + Makefile for local dev (see `.vscode/settings.json`).

2. **One Node on PATH for project commands**
  - Document: “Use Homebrew Node” *or* “Use nvm Node” - pick one primary story in README.
  - Optional: add **`.node-version`** (or `.tool-versions`) with the version you actually test.

3. **Scripts call Node explicitly where shims failed**
  - `package.json` already uses `node ./node_modules/vite/bin/vite.js` and `node ./node_modules/typescript/lib/tsc.js` for fragile environments - **keep that pattern** until the team standardizes on a single Node layout.

4. **Recovery recipe** (paste into README / runbook if needed)
   ```bash
   make kill-all
   cd frontend && rm -rf node_modules package-lock.json dist .vite .next && npm install && npm run verify
   cd ../backend && GOTOOLCHAIN=auto go mod tidy && go build ./...
   ```

5. **Do not vendor half the internet into git**
  - `contracts/lib/*` from `forge install` belongs in `.gitignore` + documented `forge install` step, **or** committed intentionally with a clear policy - pick one and stick to it (massive untracked `lib/` trees confuse agents and humans alike).

---

## 5. Phased roadmap

### Phase 0 - Inventory (½ day)

- [ ] List **one** supported matrix: macOS version, Node x.y, Go x.y, Foundry version.
- [x] Confirm `docs/SESSION_BOT_LIFECYCLE.md` exists and is linked from `AGENTS.md` and README.
- [ ] Grep repo for duplicate/contradictory instructions (Next vs Vite, old ports).

**Exit:** README + `AGENTS.md` agree on stack and commands.

### Phase 1 - Toolchain lock (½-1 day)

- [ ] Add **`.node-version`** (optional but high leverage).
- [ ] Ensure **`.vscode/settings.json`** `GOTOOLCHAIN=auto` is committed (already done) or document alternative for Goland/Cursor users.
- [ ] Add **`scripts/check-toolchain.sh`** (optional): prints `node -v`, `npm -v`, `go version`, `forge --version`, fails fast if missing.

**Exit:** New clone + documented installs → `make setup` succeeds.

### Phase 2 - Frontend “boring mode” (1-2 days)

Goal: **boring** `package.json`, predictable builds.

- [ ] Eliminate remaining sharp edges: audit `npm audit` *consciously* (no blind `--force`).
- [ ] Ensure `npm run verify` is the **only** blessed CI frontend check.
- [ ] Optional **greenfield UI folder**: `frontend-v2/` from official `npm create vite@latest` template, then **copy** `src/components`, `hooks`, `lib`, `types` in small PRs - only if incremental cleanup fails.

**Exit:** `npm run verify` < 2 minutes on a normal laptop; no intermittent `.bin` timeouts.

### Phase 3 - Backend & contracts hygiene (1 day)

- [ ] `go test ./...` where tests exist; document “no tests” vs add smoke tests for config parsing.
- [ ] Standardize **broadcast JSON** path and document failure modes (deploy window).
- [ ] Keep **bindings generation** as a single scripted step (`make bindings`).

**Exit:** Backend builds with documented Go policy; deploy/bindings story is one linear narrative.

### Phase 4 - Single demo script (½ day)

- [ ] `make demo-120` (or `scripts/demo.sh`) is the **canonical** Loom path; everything else is “advanced.”

**Exit:** Someone who missed six months of commits can still run **one** command story.

---

## 6. What agents should do with this repo

1. Read **`AGENTS.md`** + this plan + **`README.md`** before editing.
2. Prefer **small diffs** that improve reproducibility over novel architecture.
3. After frontend changes: `cd frontend && npm run verify`.
4. After backend changes: `cd backend && go build ./...`.
5. Do **not** introduce new package managers or frameworks without updating this plan.

---

## 7. Definition of done (project “untangled”)

- [ ] Golden path (section 3) works on a **fresh clone** with only README prerequisites.
- [ ] No duplicate/conflicting “how to run frontend” stories.
- [ ] CI (if re-added) runs `frontend npm run verify` + `backend go test`/`go build` + `forge test` as separate explicit jobs.
- [ ] This file is outdated **on purpose** only when phases complete - then update README and shrink this doc to a short “maintenance” section.

---

## 8. Optional: nuclear reset (only if Phase 2 keeps failing)

1. `git checkout main && git pull`
2. Tag current UI: `git tag ui-before-reset`
3. Create **`frontend/` from official Vite React TS template** in a branch.
4. Copy **`src/`** slices incrementally: `types` → `lib/api` → `hooks` → components bottom-up.
5. Restore **`vite.config.ts` proxy** for `/api` and **`VITE_*`** env names.
6. Merge when `npm run verify` is green and `make up` demo matches prior behavior.

This is **deliberate surgery**, not vibe churn.

---

*Last intent: turn “agents can’t figure this out” into “the README + Makefile + three commands are the source of truth.”*
