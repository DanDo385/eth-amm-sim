#!/usr/bin/env bash
# Apply Ubuntu-primary / MBP-backup Cloudflare Tunnel hostnames.
#
# This is not automatic DNS failover. Cloudflare maps one hostname to one
# tunnel. The SPA fails over by probing /healthz and switching to the backup
# hostname (see frontend/src/lib/backend.ts).
#
# Usage:
#   ./scripts/apply-hosted-origins.sh print     # show intended configs
#   ./scripts/apply-hosted-origins.sh apply-mbp # write MBP config + route DNS (run ON the MacBook)
#   ./scripts/apply-hosted-origins.sh apply-ubuntu  # write Ubuntu config + route DNS (run ON the VPS as root)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ACTION="${1:-print}"

PRIMARY_HOST="api-staging-eth-amm-sim.magro.dev"
BACKUP_HOST="api-backup-eth-amm-sim.magro.dev"
UBUNTU_TUNNEL="eth-l2-ubuntu"
MBP_TUNNEL="magro-mbp-backends"

print_plan() {
  cat <<EOF
Hosted origins
  primary (Ubuntu): https://${PRIMARY_HOST}  tunnel=${UBUNTU_TUNNEL}  → 127.0.0.1:8103
  backup  (MBP):    https://${BACKUP_HOST}   tunnel=${MBP_TUNNEL}     → 127.0.0.1:8103

The MacBook must keep Anvil :11545 + Go :8103 running (LaunchAgents) but must
NOT advertise ${PRIMARY_HOST}. That hostname is Ubuntu-only.

Templates:
  ${ROOT}/scripts/cloudflared/ubuntu-eth-backends.yml
  ${ROOT}/scripts/cloudflared/mbp-backends.yml
EOF
}

apply_mbp() {
  local dest="${HOME}/.cloudflared/config.yml"
  local src="${ROOT}/scripts/cloudflared/mbp-backends.yml"
  if [[ ! -f "$src" ]]; then
    echo "error: missing $src" >&2
    exit 1
  fi
  mkdir -p "${HOME}/.cloudflared"
  if [[ -f "$dest" ]]; then
    cp "$dest" "${dest}.bak-before-amm-backup-$(date +%Y%m%d%H%M%S)"
  fi
  cp "$src" "$dest"
  chmod 600 "$dest"
  echo "wrote $dest"
  if command -v cloudflared >/dev/null 2>&1; then
    cloudflared tunnel route dns "$MBP_TUNNEL" "$BACKUP_HOST" || \
      cloudflared tunnel route dns --overwrite-dns "$MBP_TUNNEL" "$BACKUP_HOST"
  else
    echo "warn: cloudflared not on PATH; route DNS manually:" >&2
    echo "  cloudflared tunnel route dns ${MBP_TUNNEL} ${BACKUP_HOST}" >&2
  fi
  local label="gui/$(id -u)/com.openclaw.cloudflared-magro-mbp-backends"
  if launchctl print "$label" >/dev/null 2>&1; then
    launchctl kickstart -k "$label"
    echo "reloaded $label"
  else
    echo "warn: LaunchAgent $label not loaded; restart cloudflared yourself" >&2
  fi
}

apply_ubuntu() {
  local dest="/etc/cloudflared/config.yml"
  local src="${ROOT}/scripts/cloudflared/ubuntu-eth-backends.yml"
  if [[ ! -f "$src" ]]; then
    echo "error: missing $src" >&2
    exit 1
  fi
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "error: apply-ubuntu must run as root on the VPS" >&2
    exit 1
  fi
  if [[ -f "$dest" ]]; then
    cp "$dest" "${dest}.bak-before-amm-primary-$(date +%Y%m%d%H%M%S)"
  fi
  cp "$src" "$dest"
  echo "wrote $dest"
  if command -v cloudflared >/dev/null 2>&1; then
    cloudflared tunnel route dns --overwrite-dns "$UBUNTU_TUNNEL" "$PRIMARY_HOST"
  else
    echo "warn: cloudflared not on PATH; route DNS manually" >&2
  fi
  if systemctl is-enabled cloudflared-eth-backends >/dev/null 2>&1 || systemctl is-active cloudflared-eth-backends >/dev/null 2>&1; then
    systemctl restart cloudflared-eth-backends
    echo "restarted cloudflared-eth-backends"
  elif systemctl is-active cloudflared >/dev/null 2>&1; then
    systemctl restart cloudflared
    echo "restarted cloudflared"
  else
    echo "warn: no cloudflared systemd unit found; restart the tunnel process yourself" >&2
  fi
}

case "$ACTION" in
  print) print_plan ;;
  apply-mbp) apply_mbp ;;
  apply-ubuntu) apply_ubuntu ;;
  *)
    echo "usage: $0 {print|apply-mbp|apply-ubuntu}" >&2
    exit 2
    ;;
esac
