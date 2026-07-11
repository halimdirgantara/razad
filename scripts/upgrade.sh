#!/usr/bin/env bash
# razad upgrade script
#
# Replaces the razad-daemon binary, lets the daemon run its auto-migrations
# on the next start, and restarts the systemd service. Idempotent for the
# file-system steps (binary replacement is always overwrite), and safe to
# re-run after a partially-failed upgrade.
#
# Usage:
#   sudo ./scripts/upgrade.sh [path-to-new-razad-daemon-binary]
#
# If no binary path is supplied, the script downloads the latest release
# from GitHub (same semantics as install.sh).

set -euo pipefail

if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
    echo "upgrade.sh must be run as root (try sudo)" >&2
    exit 1
fi

REPO="halimdirgantara/razad"
BIN_DIR="${RAZAD_BIN_DIR:-/usr/local/bin}"
BIN_PATH="${BIN_DIR}/razad-daemon"
SERVICE_NAME="razad-daemon.service"

LOCAL_BINARY="${1:-}"

log() { printf '[upgrade] %s\n' "$*"; }
fail() { printf '[upgrade] ERROR: %s\n' "$*" >&2; exit 1; }

command -v systemctl >/dev/null 2>&1 || fail "systemctl not found (systemd is required)"
[[ -f "${BIN_PATH}" ]] || fail "existing binary not found at ${BIN_PATH}; run install.sh first"

# ---- Acquire new binary ---------------------------------------------------

install_binary() {
    local src="$1"
    log "installing new binary to ${BIN_PATH}"
    install -m 0755 -o root -g root "${src}" "${BIN_PATH}"
}

if [[ -n "${LOCAL_BINARY}" ]]; then
    [[ -f "${LOCAL_BINARY}" ]] || fail "binary not found at ${LOCAL_BINARY}"
    install_binary "${LOCAL_BINARY}"
else
    log "no binary supplied; downloading latest release from GitHub"
    TMP="$(mktemp -d)"
    trap 'rm -rf "${TMP}"' EXIT
    ARCH="$(uname -m)"
    case "${ARCH}" in
        x86_64)  ARCH=amd64 ;;
        aarch64) ARCH=arm64 ;;
        armv7l)  ARCH=armv7 ;;
        *) fail "unsupported architecture: ${ARCH}" ;;
    esac
    ASSET="razad-daemon-linux-${ARCH}.tar.gz"
    URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
    log "fetching ${URL}"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "${URL}" -o "${TMP}/${ASSET}"
    else
        wget -q "${URL}" -O "${TMP}/${ASSET}"
    fi
    tar -xzf "${TMP}/${ASSET}" -C "${TMP}"
    install_binary "${TMP}/razad-daemon"
fi

# ---- Restart (which triggers auto-migrations on next start) --------------

if systemctl list-unit-files "${SERVICE_NAME}" >/dev/null 2>&1 \
   && systemctl is-enabled --quiet "${SERVICE_NAME}"; then
    log "restarting ${SERVICE_NAME}"
    systemctl restart "${SERVICE_NAME}"

    log "waiting for daemon to come back up"
    for i in $(seq 1 30); do
        if systemctl is-active --quiet "${SERVICE_NAME}"; then
            log "upgrade complete (active after ${i}s)"
            exit 0
        fi
        sleep 1
    done
    fail "${SERVICE_NAME} did not become active within 30s; check 'journalctl -u ${SERVICE_NAME}'"
else
    log "service not registered; skipping restart (run install.sh to enable)"
    log "binary updated at ${BIN_PATH}"
fi
