#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/verify.sh - Full Project Verification Orchestrator
#
# Purpose:
#   Unifies linting and security auditing into a single, standardized
#   verification workflow. Ensures the project is stable before PRs.
#
# Usage:
#   sh scripts/verify.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General), Rule 03 (Architecture), Rule 08 (Dev Env).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Atomic execution of the full verification suite.
#   - Integrated reporting for CI/CD environments.

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# Purpose: Displays usage information for the verification orchestrator.
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Executes the full project verification suite (lint, audit).

Options:
  --dry-run        Preview verification steps without execution.
  -v, --verbose    Enable verbose output for all sub-scripts.
  -q, --quiet      Suppress orchestration details.
  -h, --help       Show this help message.

EOF
}

# Purpose: Main entry point for the verification engine.
#          Coordinates lint and audit scripts.
# Params:
#   $@ - Command line arguments
# Examples:
#   main --verbose
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  parse_common_args "$@"

  log_info "Starting Full Project Verification..."

  local _EXIT_VERIFY=0

  # 3. Standardized Linting (Pre-commit)
  log_info "\n── Phase 1: Static Analysis (Lint) ──"
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_success "DRY-RUN: Would run unirtm run lint"
  else
    command unirtm run lint || _EXIT_VERIFY=$?
  fi

  # 5. Security Audit
  log_info "\n── Phase 2: Security Audit ──"
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_success "DRY-RUN: Would run unirtm run audit"
  else
    command unirtm run audit || _EXIT_VERIFY=$?
  fi

  # 7. Final Status Report
  if [ "${_EXIT_VERIFY:-}" -eq 0 ]; then
    log_success "\n✨ Full verification completed successfully!"
  else
    log_error "\n❌ Verification failed during one or more phases (Status: ${_EXIT_VERIFY:-})."
    exit "${_EXIT_VERIFY:-}"
  fi
}

main "$@"
