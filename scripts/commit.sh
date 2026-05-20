#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/commit.sh - Guided Commit Manager
#
# Purpose:
#   Facilitates high-quality, conventional commits with Commitizen and health checks.
#   Ensures all changes meet the project's quality standards before version control entry.
#
# Usage:
#   sh scripts/commit.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General), Rule 02 (Coding Style), Rule 07 (Git).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Pre-commit verification before guided entry.
#   - Node.js dependency detection and routing.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# Purpose: Displays usage information for the guided commit manager.
# Examples:
#   show_help
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Starts the interactive Commitizen CLI to create a structured commit message.
Performs a quick environment check before starting.

Options:
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

Environment Variables:

EOF
}

# Purpose: Main entry point for the guided commit experience.
#          Ensures staged changes exist and launches the Commitizen CLI.
# Params:
#   $@ - Command line arguments
# Examples:
#   main --verbose
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  parse_common_args "$@"

  log_info "📝 Starting Structured Commit Guide...\n"

  # 3. Check for staged files
  if [ "${DRY_RUN:-0}" -eq 0 ]; then
    if ! git diff --cached --quiet; then
      log_debug "Staged changes detected."
    else
      # Check if there are ANY changes at all
      if [ -z "$(git status --porcelain)" ]; then
        log_success "Nothing to commit, working tree clean. ✨"
        exit 0
      else
        log_warn "⚠️  No files added to staging! Your changes are currently unstaged."
        log_info "Modified files:"
        git status --porcelain | grep -E '^ [MADRC]' || true
        printf "\n"
        log_info "💡 Run 'git add <file>' before committing."
        exit 0
      fi
    fi
  fi

  # 4. Check for dependencies
  local _NPM_LOCAL_CMT
  _NPM_LOCAL_CMT="${NPM:-}"
  if ! resolve_bin "${_NPM_LOCAL_CMT:-}" >/dev/null 2>&1; then
    log_error "Error: ${_NPM_LOCAL_CMT:-} client not found."
    exit 1
  fi

  # 5. Launch Commitizen
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_success "DRY-RUN: Would launch interactive Commitizen CLI."
    if [ -f "package.json" ] && grep -q '"commit":' package.json; then
      log_info "Command: ${_NPM_LOCAL_CMT:-} run commit"
    else
      log_info "Command: ${_NPM_LOCAL_CMT:-} exec cz"
    fi
    exit 0
  fi

  log_info "Launching interactive CLI..."
  # We use direct exec to avoid recursion if the npm script points back here
  "${_NPM_LOCAL_CMT:-}" exec cz

  # 6. Standardized Next Actions
  if [ "${DRY_RUN:-0}" -eq 0 ] && [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
    printf "  - Run %bmake release%b to publish your changes.\n" "${GREEN:-}" "${NC:-}"
    printf "  - Run %bgit push%b to upload changes to the remote repository.\n" "${GREEN:-}" "${NC:-}"
  fi
}

main "$@"
