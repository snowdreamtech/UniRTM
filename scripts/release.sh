#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/release.sh - Standardized Release Manager
#
# Purpose:
#   Automates semantic versioning, git tagging, and pre-release verification.
#   Ensures a reliable and documented release flow for all project artifacts.
#
# Usage:
#   sh scripts/release.sh [OPTIONS] [VERSION]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General, Network), Rule 07 (Git).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Supports multiple package ecosystems (Node.js, Python, Go, etc.).
#   - Works with release-please for automated versioning.
#   - Guarded git operations with dry-run support.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Displays usage information for the release manager.
# Examples:
#   show_help
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS] [VERSION]

Standardized release manager for versioning and tagging.

Options:
  --dry-run        Preview release actions without executing them.
  --git-tag        Explicitly create a git tag and push to origin (default: skip).
  --sync-lockfiles Synchronize lockfiles for all package ecosystems.
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

VERSION:
  A semantic version (e.g., 1.2.3 or v1.2.3). If omitted,
  the version is extracted from project manifests (SSoT).

Note:
  This script supports multiple package ecosystems:
  - Node.js (package.json, lockfile)
  - Python (pyproject.toml, requirements.txt)
  - Go (go.mod)
  - Docs (docs/package.json)

EOF
}

# Purpose: Executes pre-flight health checks before allowing a release.
# Examples:
#   run_release_verify
run_release_verify() {
  log_info "── Verification: Running pre-flight checks ──"
  log_info "Skipping environment check (check-env removed)."
}

# Purpose: Handles the actual git tagging and remote synchronization logic.
# Params:
#   $1 - The target version string
#   $2 - Skip tag flag (1 to skip, 0 to perform)
# Examples:
#   perform_git_release "v1.0.0" 0
perform_git_release() {
  local _LV_RAW_VERSION="${1:-}"
  local _LV_DO_TAG="${2:-0}"

  if [ -z "${_LV_RAW_VERSION:-}" ]; then
    log_info "Synchronizing with manifest version..."
    _LV_RAW_VERSION=$(get_project_version)
  fi

  # Normalize version: Manifests use numeric, Git Tags use 'v' prefix
  local _LV_TAG_VERSION
  if echo "${_LV_RAW_VERSION:-}" | grep -q "^v"; then
    _LV_TAG_VERSION="${_LV_RAW_VERSION:-}"
  else
    _LV_TAG_VERSION="v${_LV_RAW_VERSION:-}"
  fi

  log_info "── Target Release: ${_LV_TAG_VERSION:-} ──"

  if [ "${_LV_DO_TAG:-}" -eq 0 ]; then
    log_info "Default mode: Local verification only. Skipping git tag operation."
    log_info "Tip: Run with --git-tag to perform actual tagging and pushing."
    return 0
  fi

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_info "DRY-RUN: Would tag version ${_LV_TAG_VERSION:-} and push to origin."
  else
    # Safety Gate: Check if tag already exists
    if git rev-parse "${_LV_TAG_VERSION:-}" >/dev/null 2>&1; then
      log_warn "Warning: Git tag ${_LV_TAG_VERSION:-} already exists. Skipping tagging."
    else
      log_info "Tagging local repository as ${_LV_TAG_VERSION:-}..."
      git tag -a "${_LV_TAG_VERSION:-}" -m "chore(release): ${_LV_TAG_VERSION:-}"
      log_info "Pushing tag to origin..."
      git push origin "${_LV_TAG_VERSION:-}"
    fi
  fi
}

# Purpose: Synchronizes lockfiles for all supported package ecosystems.
# Examples:
#   sync_all_lockfiles
sync_all_lockfiles() {
  log_info "── Lockfile Synchronization ──"

  # Node.js (root)
  sync_node_lockfile "."

  # Docs (docs/)
  if [ -f "docs/package.json" ]; then
    log_info "Syncing docs dependencies..."
    sync_node_lockfile "docs"
  fi

  log_success "Lockfile synchronization complete."
}

# Purpose: Main entry point for the release management engine.
# Params:
#   $@ - Command line arguments and optional version
# Examples:
#   main v1.2.3
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  local _REL_TARGET_VERSION=""
  local _DO_TAG_MAIN=0
  local _SYNC_LOCKFILES=0
  local _arg_rel
  for _arg_rel in "$@"; do
    case "${_arg_rel:-}" in
    --git-tag)
      _DO_TAG_MAIN=1
      ;;
    --sync-lockfiles)
      _SYNC_LOCKFILES=1
      ;;
    [0-9]* | v[0-9]*)
      _REL_TARGET_VERSION="${_arg_rel:-}"
      ;;
    esac
  done
  parse_common_args "$@"

  log_info "📦 Starting Standardized Release Process...\n"

  # 3. Sync lockfiles if requested
  if [ "${_SYNC_LOCKFILES:-}" -eq 1 ]; then
    sync_all_lockfiles
    printf "\n"
  fi

  # 4. Pre-release verification
  run_release_verify

  printf "\n"

  # 5. Versioning & Tagging logic
  perform_git_release "${_REL_TARGET_VERSION:-}" "${_DO_TAG_MAIN:-}"

  printf "\n"

  # 6. Sync lockfiles after version bump
  sync_all_lockfiles

  log_success "\n✨ Release process completed successfully!"

  # 7. Standardized Next Actions
  if [ "${DRY_RUN:-0}" -eq 0 ] && [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
    printf "  - Run %bgit push --tags%b to synchronize version tags with the remote.\n" "${GREEN:-}" "${NC:-}"
    printf "  - Run %bmake cleanup%b to remove temporary release artifacts.\n" "${GREEN:-}" "${NC:-}"
  fi
}

main "$@"
