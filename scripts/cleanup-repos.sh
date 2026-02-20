#!/usr/bin/env bash
set -euo pipefail

OWNER="DanDo385"
REPO="eth-amm-sim"
FULL_REPO="$OWNER/$REPO"
DRY_RUN="${DRY_RUN:-false}"

log()  { printf "\033[1;34m[INFO]\033[0m  %s\n" "$1"; }
warn() { printf "\033[1;33m[WARN]\033[0m  %s\n" "$1"; }
ok()   { printf "\033[1;32m[ OK ]\033[0m  %s\n" "$1"; }
err()  { printf "\033[1;31m[ERR ]\033[0m  %s\n" "$1"; }
sep()  { printf "\n%s\n" "────────────────────────────────────────────────"; }

if [[ "$DRY_RUN" == "true" ]]; then
  warn "DRY RUN mode — no destructive actions will be taken"
fi

# ── 1. Delete branches whose PRs are already merged or closed ───────────────

sep
log "Scanning remote branches for $FULL_REPO …"

branches=$(gh api "repos/$OWNER/$REPO/branches" --paginate --jq '.[].name')
default_branch=$(gh api "repos/$OWNER/$REPO" --jq '.default_branch')

deleted=0
kept=0

for branch in $branches; do
  if [[ "$branch" == "$default_branch" ]]; then
    ok "Keeping default branch: $branch"
    continue
  fi

  # Skip the current working branch for this task
  if [[ "$branch" == cursor/github-repositories-cleanup-* ]]; then
    ok "Keeping current working branch: $branch"
    continue
  fi

  pr_json=$(gh pr list --repo "$FULL_REPO" --head "$branch" --state all \
    --json number,state,title --limit 1 2>/dev/null || echo "[]")
  pr_state=$(echo "$pr_json" | jq -r '.[0].state // "NONE"')
  pr_number=$(echo "$pr_json" | jq -r '.[0].number // "N/A"')
  pr_title=$(echo "$pr_json" | jq -r '.[0].title // "N/A"')

  case "$pr_state" in
    MERGED)
      log "Branch '$branch' → PR #$pr_number MERGED (\"$pr_title\")"
      if [[ "$DRY_RUN" != "true" ]]; then
        gh api -X DELETE "repos/$OWNER/$REPO/git/refs/heads/$branch" 2>/dev/null \
          && ok "  Deleted branch '$branch'" \
          || err "  Failed to delete branch '$branch'"
      else
        warn "  Would delete branch '$branch'"
      fi
      ((deleted++)) || true
      ;;
    CLOSED)
      log "Branch '$branch' → PR #$pr_number CLOSED without merge (\"$pr_title\")"
      if [[ "$DRY_RUN" != "true" ]]; then
        gh api -X DELETE "repos/$OWNER/$REPO/git/refs/heads/$branch" 2>/dev/null \
          && ok "  Deleted branch '$branch'" \
          || err "  Failed to delete branch '$branch'"
      else
        warn "  Would delete branch '$branch'"
      fi
      ((deleted++)) || true
      ;;
    OPEN)
      ok "Keeping branch '$branch' — PR #$pr_number is still OPEN"
      ((kept++)) || true
      ;;
    NONE)
      warn "Branch '$branch' has no associated PR — skipping (manual review suggested)"
      ((kept++)) || true
      ;;
  esac
done

sep
log "Branch cleanup summary: $deleted deleted, $kept kept"

# ── 2. Report on open Dependabot PRs ───────────────────────────────────────

sep
log "Checking open Dependabot PRs …"

dependabot_prs=$(gh pr list --repo "$FULL_REPO" --state open \
  --json number,title,headRefName,createdAt \
  --jq '.[] | select(.headRefName | startswith("dependabot/"))')

if [[ -z "$dependabot_prs" ]]; then
  ok "No open Dependabot PRs"
else
  count=$(gh pr list --repo "$FULL_REPO" --state open \
    --json number,title,headRefName,createdAt \
    --jq '[.[] | select(.headRefName | startswith("dependabot/"))] | length')
  warn "$count open Dependabot PR(s) found:"
  gh pr list --repo "$FULL_REPO" --state open \
    --json number,title,createdAt \
    --jq '.[] | select(.title | test("(?i)bump|dependabot")) | "  #\(.number)  \(.title)  (opened \(.createdAt | split("T")[0]))"'
fi

# ── 3. Check for stale open PRs (>30 days with no activity) ────────────────

sep
log "Checking for stale open PRs (>30 days old) …"

thirty_days_ago=$(date -u -d '30 days ago' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null \
  || date -u -v-30d +%Y-%m-%dT%H:%M:%SZ 2>/dev/null \
  || echo "2026-01-21T00:00:00Z")

stale_count=0
gh pr list --repo "$FULL_REPO" --state open \
  --json number,title,updatedAt,createdAt \
  --jq ".[] | select(.updatedAt < \"$thirty_days_ago\")" | \
while IFS= read -r line; do
  if [[ -n "$line" ]]; then
    echo "$line" | jq -r '"  #\(.number)  \(.title)  (last updated \(.updatedAt | split("T")[0]))"'
    stale_count=$((stale_count + 1))
  fi
done

if [[ "$stale_count" -eq 0 ]]; then
  ok "No stale open PRs"
fi

# ── 4. Repository settings check ──────────────────────────────────────────

sep
log "Checking repository settings …"

repo_info=$(gh api "repos/$OWNER/$REPO" --jq '{
  has_issues: .has_issues,
  has_wiki: .has_wiki,
  has_projects: .has_projects,
  delete_branch_on_merge: .delete_branch_on_merge,
  allow_squash_merge: .allow_squash_merge,
  allow_merge_commit: .allow_merge_commit,
  allow_rebase_merge: .allow_rebase_merge,
  description: .description,
  homepage: .homepage,
  topics: .topics
}')

delete_on_merge=$(echo "$repo_info" | jq -r '.delete_branch_on_merge')
description=$(echo "$repo_info" | jq -r '.description')
topics=$(echo "$repo_info" | jq -r '.topics | length')

if [[ "$delete_on_merge" != "true" ]]; then
  warn "Auto-delete branches on merge is DISABLED"
  if [[ "$DRY_RUN" != "true" ]]; then
    gh api -X PATCH "repos/$OWNER/$REPO" \
      -f delete_branch_on_merge=true --silent 2>/dev/null \
      && ok "  Enabled auto-delete branches on merge" \
      || err "  Failed to enable auto-delete (may require admin permissions)"
  else
    warn "  Would enable auto-delete branches on merge"
  fi
else
  ok "Auto-delete branches on merge is already enabled"
fi

if [[ -z "$description" || "$description" == "null" ]]; then
  warn "Repository has no description set"
fi

if [[ "$topics" -eq 0 ]]; then
  warn "Repository has no topics set — consider adding some for discoverability"
fi

# ── 5. Final summary ──────────────────────────────────────────────────────

sep
log "Cleanup complete for $FULL_REPO"
remaining_branches=$(gh api "repos/$OWNER/$REPO/branches" --paginate --jq '.[].name' | wc -l)
open_prs=$(gh pr list --repo "$FULL_REPO" --state open --json number --jq 'length')
log "Remaining branches: $remaining_branches"
log "Open PRs: $open_prs"
echo ""
