#!/bin/sh
# install.sh — UniRTM installer for Linux and macOS
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | sh
#   sh install.sh --version v0.0.10
#
# Environment variables:
#   GITHUB_PROXY  — Optional proxy prefix for GitHub downloads
#                   Default: https://gh-proxy.sn0wdr1am.com/
#   INSTALL_DIR   — Directory to install the binary (default: ~/.unirtm/bin, or /usr/local/bin if root)
#   UNIRTM_VERSION — Target version (overridden by --version flag)

set -eu

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
REPO="snowdreamtech/UniRTM"
BINARY="unirtm"
GITHUB_PROXY="${GITHUB_PROXY:-https://gh-proxy.sn0wdr1am.com/}"
UNIRTM_VERSION="${UNIRTM_VERSION:-}"

# Retry config
CURL_RETRY_COUNT=5
CURL_RETRY_DELAY=2
CURL_CONNECT_TIMEOUT=15
CURL_MAX_TIME=120

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
info()    { printf '\033[0;32m[INFO]\033[0m  %s\n' "$*"; }
warn()    { printf '\033[0;33m[WARN]\033[0m  %s\n' "$*"; }
error()   { printf '\033[0;31m[ERROR]\033[0m %s\n' "$*" >&2; }
die()     { error "$*"; exit 1; }

need_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        die "Required command not found: $1. Please install it and try again."
    fi
}

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --version|-v)
                shift
                UNIRTM_VERSION="$1"
                ;;
            --install-dir)
                shift
                INSTALL_DIR="$1"
                ;;
            --no-proxy)
                GITHUB_PROXY=""
                ;;
            --help|-h)
                printf 'Usage: install.sh [--version <tag>] [--install-dir <dir>] [--no-proxy]\n'
                exit 0
                ;;
            *)
                warn "Unknown argument: $1"
                ;;
        esac
        shift
    done
}

# ---------------------------------------------------------------------------
# Detect OS / Arch
# ---------------------------------------------------------------------------
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux)   OS_NAME="Linux" ;;
        Darwin)  OS_NAME="Darwin" ;;
        *)       die "Unsupported operating system: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64)   ARCH_NAME="x86_64" ;;
        aarch64|arm64)  ARCH_NAME="arm64" ;;
        armv7*|armv6*)  ARCH_NAME="armv7" ;;
        i386|i686)      ARCH_NAME="i386" ;;
        *)              die "Unsupported architecture: $ARCH" ;;
    esac

    info "Detected platform: ${OS_NAME}/${ARCH_NAME}"
}

# ---------------------------------------------------------------------------
# Resolve target version
# ---------------------------------------------------------------------------
resolve_version() {
    if [ -n "$UNIRTM_VERSION" ]; then
        # Normalize: strip leading 'v' for comparison but keep tag form for URL
        VERSION="$UNIRTM_VERSION"
        # Ensure v-prefix for tag
        case "$VERSION" in
            v*) ;;
            *)  VERSION="v${VERSION}" ;;
        esac
        info "Target version: ${VERSION}"
        return
    fi

    info "Fetching latest release from GitHub API..."
    API_URL="https://api.github.com/repos/${REPO}/releases/latest"

    # Try with proxy first, fall back to direct
    LATEST=""
    LATEST="$(curl_with_retry "$API_URL" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')"

    if [ -z "$LATEST" ]; then
        die "Failed to determine latest version. Please specify one with --version."
    fi

    VERSION="$LATEST"
    info "Latest version: ${VERSION}"
}

# ---------------------------------------------------------------------------
# Download with retry and proxy
# ---------------------------------------------------------------------------
curl_with_retry() {
    URL="$1"
    OUTPUT="${2:-}"

    # Apply proxy prefix to github.com URLs
    PROXIED_URL="$URL"
    case "$URL" in
        https://github.com/*|https://objects.githubusercontent.com/*|https://raw.githubusercontent.com/*)
            if [ -n "$GITHUB_PROXY" ]; then
                PROXIED_URL="${GITHUB_PROXY}${URL}"
            fi
            ;;
    esac

    if [ -n "$OUTPUT" ]; then
        curl \
            --retry "$CURL_RETRY_COUNT" \
            --retry-delay "$CURL_RETRY_DELAY" \
            --retry-connrefused \
            --connect-timeout "$CURL_CONNECT_TIMEOUT" \
            --max-time "$CURL_MAX_TIME" \
            --fail \
            --location \
            --silent \
            --show-error \
            -o "$OUTPUT" \
            "$PROXIED_URL" || {
            # On failure, retry without proxy
            if [ "$PROXIED_URL" != "$URL" ]; then
                warn "Proxy download failed, retrying without proxy..."
                curl \
                    --retry "$CURL_RETRY_COUNT" \
                    --retry-delay "$CURL_RETRY_DELAY" \
                    --retry-connrefused \
                    --connect-timeout "$CURL_CONNECT_TIMEOUT" \
                    --max-time "$CURL_MAX_TIME" \
                    --fail \
                    --location \
                    --silent \
                    --show-error \
                    -o "$OUTPUT" \
                    "$URL"
            else
                return 1
            fi
        }
    else
        curl \
            --retry "$CURL_RETRY_COUNT" \
            --retry-delay "$CURL_RETRY_DELAY" \
            --retry-connrefused \
            --connect-timeout "$CURL_CONNECT_TIMEOUT" \
            --max-time "$CURL_MAX_TIME" \
            --fail \
            --location \
            --silent \
            --show-error \
            "$PROXIED_URL" || {
            if [ "$PROXIED_URL" != "$URL" ]; then
                warn "Proxy request failed, retrying without proxy..."
                curl \
                    --retry "$CURL_RETRY_COUNT" \
                    --retry-delay "$CURL_RETRY_DELAY" \
                    --retry-connrefused \
                    --connect-timeout "$CURL_CONNECT_TIMEOUT" \
                    --max-time "$CURL_MAX_TIME" \
                    --fail \
                    --location \
                    --silent \
                    --show-error \
                    "$URL"
            else
                return 1
            fi
        }
    fi
}

# ---------------------------------------------------------------------------
# Download and verify checksum
# ---------------------------------------------------------------------------
download_and_verify() {
    ARCHIVE_NAME="${BINARY}_${OS_NAME}_${ARCH_NAME}.tar.gz"
    ARCHIVE_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"
    CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

    TMP_DIR="$(mktemp -d)"
    ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE_NAME}"
    CHECKSUM_PATH="${TMP_DIR}/checksums.txt"

    # Ensure cleanup on exit
    trap 'rm -rf "$TMP_DIR"' EXIT

    info "Downloading ${ARCHIVE_NAME}..."
    curl_with_retry "$ARCHIVE_URL" "$ARCHIVE_PATH"

    # Verify file is not empty
    if [ ! -s "$ARCHIVE_PATH" ]; then
        die "Downloaded archive is empty: ${ARCHIVE_PATH}"
    fi

    info "Downloading checksums..."
    if curl_with_retry "$CHECKSUM_URL" "$CHECKSUM_PATH" 2>/dev/null; then
        info "Verifying checksum..."
        # Extract the expected checksum for our archive
        EXPECTED="$(grep "${ARCHIVE_NAME}" "$CHECKSUM_PATH" | awk '{print $1}')"
        if [ -z "$EXPECTED" ]; then
            warn "Checksum entry not found for ${ARCHIVE_NAME}, skipping verification."
        else
            # Compute actual checksum
            if command -v sha256sum >/dev/null 2>&1; then
                ACTUAL="$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}')"
            elif command -v shasum >/dev/null 2>&1; then
                ACTUAL="$(shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}')"
            else
                warn "No sha256sum or shasum available. Skipping checksum verification."
                ACTUAL="$EXPECTED"
            fi

            if [ "$EXPECTED" != "$ACTUAL" ]; then
                die "Checksum mismatch! Expected: ${EXPECTED}, Got: ${ACTUAL}"
            fi
            info "Checksum verified OK."
        fi
    else
        warn "Could not download checksums.txt. Skipping checksum verification."
    fi

    echo "$ARCHIVE_PATH"
    echo "$TMP_DIR"
}

# ---------------------------------------------------------------------------
# Install binary
# ---------------------------------------------------------------------------
install_binary() {
    ARCHIVE_PATH="$1"
    TMP_DIR="$2"

    # Determine install directory
    if [ -z "${INSTALL_DIR:-}" ]; then
        if [ "$(id -u)" = "0" ]; then
            INSTALL_DIR="/usr/local/bin"
        else
            INSTALL_DIR="$HOME/.unirtm/bin"
        fi
    fi

    info "Installing to ${INSTALL_DIR}..."
    mkdir -p "$INSTALL_DIR"

    # Extract archive
    tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

    # Find the binary (may be inside a subdirectory)
    BINARY_PATH="$(find "$TMP_DIR" -name "$BINARY" -type f | head -1)"
    if [ -z "$BINARY_PATH" ]; then
        die "Binary '${BINARY}' not found in archive."
    fi

    chmod +x "$BINARY_PATH"
    cp "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"

    info "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
}

# ---------------------------------------------------------------------------
# Update PATH hint
# ---------------------------------------------------------------------------
suggest_path() {
    # Check if INSTALL_DIR is already in PATH
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) return ;;
    esac

    warn "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    printf '  export PATH="%s:$PATH"\n' "$INSTALL_DIR"
}

# ---------------------------------------------------------------------------
# Post-install verification
# ---------------------------------------------------------------------------
verify_install() {
    INSTALLED="${INSTALL_DIR}/${BINARY}"
    if [ ! -x "$INSTALLED" ]; then
        die "Verification failed: binary not found at ${INSTALLED}"
    fi

    INSTALLED_VER="$("$INSTALLED" version 2>/dev/null | head -1 || echo 'unknown')"
    info "Installed version: ${INSTALLED_VER}"
    info "Installation complete!"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    parse_args "$@"

    need_cmd curl
    need_cmd tar

    detect_platform
    resolve_version
    read -r ARCHIVE_PATH TMP_DIR <<EOF
$(download_and_verify)
EOF
    install_binary "$ARCHIVE_PATH" "$TMP_DIR"
    suggest_path
    verify_install
}

main "$@"
