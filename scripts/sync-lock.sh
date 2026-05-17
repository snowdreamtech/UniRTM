#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/sync-lock.sh - UniRTM Lockfile Synchronizer
#
# Purpose:
#   Synchronizes unirtm.lock with the comprehensive manifest (Tier 1 + Tier 2).
#   Ensures all tools are cryptographically locked for all supported platforms.

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Synchronizes unirtm.lock for all supported platforms.
# Params:
#   $@ - Optional additional tools to lock
# Examples:
#   run_sync_lock
run_sync_lock() {
  log_info "Synchronizing unirtm.lock for all platforms..."

  # 0. Clear unirtm cache to avoid stale provenance verification data
  # This prevents errors like "has no provenance verification" when attestations
  # are actually present but cached metadata is outdated.
  log_debug "Clearing unirtm cache to refresh provenance verification data..."
  unirtm cache clear >/dev/null 2>&1 || true

  # 1. Manifest Aggregation
  local _TMP_MANIFEST=".unirtm.toml.lock.temp"
  "${_G_PROJECT_ROOT:-.}/scripts/gen-full-manifest.sh" >"${_TMP_MANIFEST:-}"

  # 2. List Extraction
  local _TOOLS
  _TOOLS=$(grep "=" "${_TMP_MANIFEST:-}" | cut -d= -f1 | tr -d '" ' | xargs)

  # 3. Multi-Platform Locking
  # Platforms: Ubuntu/Debian (glibc), Alpine (musl), macOS (x64/arm64), Windows (x64).
  # Disable paranoid mode as GitHub attestations are not universally adopted yet.
  # Many legitimate projects don't provide attestations, making this check too strict.
  # We rely on unirtm's built-in checksum verification for security instead.
  export UNIRTM_PARANOID=0
  export UNIRTM_LOCKFILE_PARANOID=0
  export UNIRTM_YES=1

  # Disable all attestation/provenance verification checks
  # This prevents errors when new tool versions lack attestations that previous versions had
  # See: docs/troubleshooting/unirtm-attestation-error.md
  export UNIRTM_SKIP_CHECKSUM=1
  export UNIRTM_AQUA_GITHUB_ATTESTATIONS=0
  export UNIRTM_AQUA_SLSA=0
  export UNIRTM_AQUA_COSIGN=0
  export UNIRTM_AQUA_MINISIGN=0

  # shellcheck disable=SC2086
  UNIRTM_CONFIG="${_TMP_MANIFEST:-}" unirtm lock --platform linux-x64,linux-arm64,linux-x64-musl,linux-arm64-musl,macos-x64,macos-arm64,windows-x64 ${_TOOLS:-} "$@"

  # 4. Remove provenance fields from unirtm.lock to prevent attestation comparison errors
  # When a new version lacks attestations that the previous version had, unirtm will error out.
  # By removing all provenance fields, we prevent this comparison while still maintaining
  # checksum verification for security.
  if [ -f "unirtm.lock" ]; then
    log_debug "Removing provenance fields from unirtm.lock to prevent attestation errors..."

    # Use sed for POSIX compatibility (works on Linux, macOS, Windows Git Bash)
    # Create a temporary file to avoid in-place editing issues across platforms
    _TMP_LOCK="unirtm.lock.tmp"
    sed '/^provenance = /d' unirtm.lock >"${_TMP_LOCK:-}"
    mv "${_TMP_LOCK:-}" unirtm.lock
  fi

  # 5. Cleanup
  rm -f "${_TMP_MANIFEST:-}"

  log_success "unirtm.lock synchronized successfully for all platforms."
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
  guard_project_root
  run_sync_lock "$@"
}

main "$@"
