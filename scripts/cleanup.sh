#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/cleanup.sh - Deep Project Sanitizer
#
# Purpose:
#   Thoroughly removes build artifacts, temporary files, and caches across all platforms.
#   Restores the project to a clean, "fresh-start" state safely.
#
# Usage:
#   sh scripts/cleanup.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General), Rule 03 (Architecture).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Cross-stack cleanup (node, python, go, rust, iac).
#   - Safe dry-run support for destructive operations.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Displays usage information for the cleanup script.
# Examples:
#   show_help
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Removes temporary files, build artifacts, and caches across all supported stacks.

Options:
  --dry-run        Preview files to be deleted without removing them.
  -q, --quiet      Only show errors.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

EOF
}

# Purpose: Safe wrapper for removing project artifacts.
# Params:
#   $1 - Absolute or relative path to the item.
#   $2 - Human-readable description of the item.
# Examples:
#   clean_item "dist" "Build distribution"
clean_item() {
  local _PATH_CLN="${1:-}"
  local _DESC_CLN="${2:-}"

  if [ -e "${_PATH_CLN:-}" ]; then
    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_success "DRY-RUN: Would remove ${_DESC_CLN:-} (${_PATH_CLN:-})"
    else
      log_info "Removing ${_DESC_CLN:-} (${_PATH_CLN:-})..."
      rm -rf "${_PATH_CLN:-}"
    fi
  else
    log_debug "Skipping ${_DESC_CLN:-} (${_PATH_CLN:-}) - Not found."
  fi
}

# Purpose: Main entry point for the deep project cleanup.
# Params:
#   $@ - Command line arguments
# Examples:
#   main --dry-run
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  parse_common_args "$@"

  log_info "🧹 Starting deep project cleanup...\n"

  # 2. General Artifacts
  clean_item "dist" "Build distribution"
  clean_item "build" "Build artifacts"
  clean_item "out" "Output directory"
  clean_item "bin" "Binary output"
  clean_item "target" "Language target directory"
  clean_item ".coverage" "Coverage report"
  clean_item "coverage.xml" "Coverage XML"

  # 3. Language Specific Caches
  log_info "\n📦 Cleaning language-specific caches..."

  # Python
  clean_item ".pytest_cache" "Pytest cache"
  clean_item ".ruff_cache" "Ruff cache"
  clean_item ".mypy_cache" "Mypy cache"
  clean_item ".venv" "Python virtual environment"

  # Finding and removing __pycache__ and .pyc files
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_success "DRY-RUN: Would search and remove all __pycache__ and *.pyc files."
  else
    # Using POSIX compliant find
    find . -type d -name "__pycache__" ! -path "*/.*" -exec rm -rf {} + 2>/dev/null || true
    find . -type f -name "*.pyc" ! -path "*/.*" -exec rm -f {} + 2>/dev/null || true
  fi

  # Node.js
  # We don't remove node_modules by default, only build output
  clean_item ".next" "Next.js build"
  clean_item ".nuxt" "Nuxt.js build"
  clean_item ".output" "Vite/Nitro output"

  # Rust / Go / Java
  # Handled in General Artifacts (target, bin)

  # 4. OS Specific Caches
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_success "DRY-RUN: Would remove .DS_Store and Thumbs.db files."
  else
    # Using POSIX compliant find
    find . -type f -name ".DS_Store" ! -path "*/.*" -exec rm -f {} + 2>/dev/null || true
    find . -type f -name "Thumbs.db" ! -path "*/.*" -exec rm -f {} + 2>/dev/null || true
  fi

  log_success "\n✨ Cleanup complete!"

  # 5. Standardized Next Actions
  if [ "${DRY_RUN:-0}" -eq 0 ] && [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
    printf "  - Run %bunirtm install%b to re-initialize the project environment.\n" "${GREEN:-}" "${NC:-}"
  fi
}

main "$@"
