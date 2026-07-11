#!/usr/bin/env bash
# razad install script
#
# Installs the razad-daemon binary, sets up storage directories, creates a
# dedicated system user, and registers the systemd service. Idempotent —
# running it twice on the same host leaves the second run a no-op except
# where the binary itself was updated.
#
# Usage:
#   sudo ./scripts/install.sh [path-to-razad-daemon-binary]
#
# If no binary path is supplied, the script downloads the latest release
# for the current architecture from GitHub. Pass a local path when testing
# or when running from a release tarball.

set -euo pipefail

if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
    echo "install.sh must be run as root (try sudo)" >&2
    exit 1
fi

REPO="halimdirgantara/razad"
DATA_DIR="${RAZAD_DATA_DIR:-/var/lib/razad}"
BIN_DIR="${RAZAD_BIN_DIR:-/usr/local/bin}"
UNIT_DIR="${RAZAD_UNIT_DIR:-/etc/systemd/system}"
SERVICE_USER="${RAZAD_SERVICE_USER:-razad}"
SERVICE_GROUP="${RAZAD_SERVICE_GROUP:-razad}"
BIN_PATH="${BIN_DIR}/razad-daemon"
SERVICE_NAME="razad-daemon.service"

LOCAL_BINARY="${1:-}"

log() { printf '[install] %s\n' "$*"; }
fail() { printf '[install] ERROR: %s\n' "$*" >&2; exit 1; }

# ---- Prereqs ---------------------------------------------------------------

command -v systemctl >/dev/null 2>&1 || fail "systemctl not found (systemd is required)"
command -v curl       >/dev/null 2>&1 || command -v wget >/dev/null 2>&1 \
    || fail "either curl or wget is required"
[[ -d /run/systemd/system ]] || fail "systemd does not appear to be running"

# ---- Service user / group -------------------------------------------------

if ! getent group "${SERVICE_GROUP}" >/dev/null; then
    log "creating group ${SERVICE_GROUP}"
    groupadd --system "${SERVICE_GROUP}"
fi
if ! getent passwd "${SERVICE_USER}" >/dev/null; then
    log "creating user ${SERVICE_USER}"
    useradd --system \
        --gid "${SERVICE_GROUP}" \
        --home-dir "${DATA_DIR}" \
        --shell /usr/sbin/nologin \
        --comment "Razad daemon" \
        "${SERVICE_USER}"
fi

# ---- Binary ---------------------------------------------------------------

install_binary() {
    local src="$1"
    log "installing binary to ${BIN_PATH}"
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

# ---- Storage directories --------------------------------------------------

log "preparing storage directories under ${DATA_DIR}"
mkdir -p \
    "${DATA_DIR}" \
    "${DATA_DIR}/nginx/sites-available" \
    "${DATA_DIR}/nginx/sites-enabled" \
    "${DATA_DIR}/nginx/backups" \
    "${DATA_DIR}/letsencrypt" \
    "${DATA_DIR}/logs" \
    "${DATA_DIR}/apps" \
    "${DATA_DIR}/databases" \
    "${DATA_DIR}/health" \
    "${DATA_DIR}/backups" \
    "${DATA_DIR}/audit"
chown -R "${SERVICE_USER}:${SERVICE_GROUP}" "${DATA_DIR}"
chmod 0750 "${DATA_DIR}"

# ---- Systemd unit ---------------------------------------------------------

log "installing systemd unit to ${UNIT_DIR}/${SERVICE_NAME}"
install -m 0644 "$(dirname "$0")/../deployments/systemd/razad-daemon.service" \
    "${UNIT_DIR}/${SERVICE_NAME}"
systemctl daemon-reload
if ! systemctl is-enabled --quiet "${SERVICE_NAME}"; then
    log "enabling ${SERVICE_NAME}"
    systemctl enable "${SERVICE_NAME}"
fi

# ---- Start ---------------------------------------------------------------

if ! systemctl is-active --quiet "${SERVICE_NAME}"; then
    log "starting ${SERVICE_NAME}"
    systemctl start "${SERVICE_NAME}"
fi

log "install complete"
log "  binary: ${BIN_PATH}"
log "  data:   ${DATA_DIR}"
log "  admin:  admin@razad.local  (password: razadadmin — change on first login)"
log "  UI:     http://localhost:8080  (or https:// if RAZAD_TLS_ENABLED=true)"
