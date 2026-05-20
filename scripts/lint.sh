#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/lint.sh - Unified Quality Orchestrator
#
# Purpose:
#   Wraps pre-commit hooks and specialized language linters into a single CLI.
#   Enforces consistent code quality standards across the entire project.
#
# Usage:
#   sh scripts/lint.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General), Rule 02 (Coding Style), Rule 06 (CI/Testing).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Orchestrated linting for all supported language stacks.
#   - CI-optimized execution with strict error checking.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# Purpose: Displays usage information for the quality orchestrator.
# Examples:
#   show_help
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Unified project linter for Shell, Python, Node.js, Go, Rust, and more (20+ stacks).

Options:
  --fix            Try to auto-fix linting issues.
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

EOF
}

# Purpose: Orchestrates pre-commit hooks execution across all files.
# Params:
#   $1 - Optional fix flag (--fix)
# Examples:
#   run_pre_commit_lint "--fix"
run_pre_commit_lint() {
  local _LV_FIX="${1:-}"
  local _PRE_COMMIT_LNT
  _PRE_COMMIT_LNT=$(resolve_bin "pre-commit") || true

  if [ -z "${_PRE_COMMIT_LNT:-}" ]; then
    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_warn "DRY-RUN: pre-commit not found. Using placeholder for preview."
      _PRE_COMMIT_LNT="pre-commit"
    else
      log_error "Error: pre-commit not found. Please run 'unirtm install' first."
      exit 1
    fi
  fi

  # Run on all files
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_success "DRY-RUN: Would run ${_PRE_COMMIT_LNT:-} --all-files"
  elif [ -n "${_LV_FIX:-}" ]; then
    log_info "Running pre-commit with auto-fix enabled..."
    # Many pre-commit hooks auto-fix by default when they can
    "${_PRE_COMMIT_LNT:-}" run --all-files
  else
    "${_PRE_COMMIT_LNT:-}" run --all-files
  fi
}

# Purpose: Main entry point for the quality orchestration engine.
# Params:
#   $@ - Command line arguments
# Examples:
#   main --fix
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  local _FIX_LNT=""
  local _arg_lnt
  for _arg_lnt in "$@"; do
    case "${_arg_lnt:-}" in
    --fix) _FIX_LNT="--fix" ;;
    -q | --quiet | -v | --verbose | --dry-run | -h | --help) ;;
    esac
  done
  parse_common_args "$@"

  init_summary_table "Project Quality Audit Summary"
  local _T0_LNT
  _T0_LNT=$(date +%s)

  # 🔍 Starting Unified Project Linter...
  log_info "🔍 Starting Unified Project Linter...\n"

  # Shift-Left Governance:
  # Instead of hardcoded skips in lint.sh, we rely on scripts/lib/lint-wrapper.sh
  # which handles missing binaries gracefully (Warn, Skip) in local environments.
  # This rewards developers with "forced local installs" as per Rule 03/08.

  local _L_OK=0
  local _PC_VER
  _PC_VER=$(get_version "pre-commit")

  # Pass 1: Run lint with auto-fix enabled
  log_info "── Pass 1: Lint check with auto-fix ──"
  if run_pre_commit_lint "--fix"; then
    _L_OK=1
    log_summary "Quality" "pre-commit" "✅ Passed" "${_PC_VER:--}" "$(($(date +%s) - _T0_LNT))"
  else
    log_summary "Quality" "pre-commit (Pass 1)" "⚠️  Failed" "${_PC_VER:--}" "$(($(date +%s) - _T0_LNT))"

    # Pass 2: Run lint again to verify
    log_warn "\n⚠️  First pass failed. Running second check..."
    log_info "── Pass 2: Final verification ──"
    local _T2_LNT
    _T2_LNT=$(date +%s)
    if run_pre_commit_lint ""; then
      _L_OK=1
      log_summary "Quality" "pre-commit (Pass 2)" "✅ Passed" "${_PC_VER:--}" "$(($(date +%s) - _T2_LNT))"
      log_success "\n✨ Linting passed on second check!"
    else
      log_summary "Quality" "pre-commit (Pass 2)" "❌ Failed" "${_PC_VER:--}" "$(($(date +%s) - _T2_LNT))"
      log_error "\n❌ Linting failed! Manual fixes required."
    fi
  fi

  if [ "${_L_OK:-}" -eq 1 ]; then
    log_success "\n✨ Linting complete!"
  else
    log_error "\n❌ Linting failed! Please fix the errors above."
  fi
  finalize_summary_table

  # 5. Standardized Next Actions
  if [ "${DRY_RUN:-0}" -eq 0 ] && [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    if [ "${_L_OK:-}" -eq 1 ]; then
      printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
      printf "  - Run %bmake test%b to execute functional test suites.\n" "${GREEN:-}" "${NC:-}"
      printf "  - Run %bmake verify%b to ensure full project stability.\n" "${GREEN:-}" "${NC:-}"
    else
      printf "\n%bRecommended Actions:%b\n" "${YELLOW:-}" "${NC:-}"
      printf "  - Review the errors above and fix them manually.\n"
      printf "  - Some issues cannot be auto-fixed and require manual intervention.\n"
      printf "  - Run %bunirtm install%b if tools are missing.\n" "${GREEN:-}" "${NC:-}"
    fi
  fi

  # Exit with failure if linting failed
  if [ "${_L_OK:-}" -eq 0 ]; then
    exit 1
  fi
}

main "$@"
