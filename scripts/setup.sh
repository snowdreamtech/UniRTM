#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/setup.sh - Project Toolchain Setup
#
# Purpose:
#   Bootstraps the unirtm toolchain manager and installs all project tools
#   declared in .unirtm.toml via native unirtm install.
#
# Usage:
#   sh scripts/setup.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - Rule 01 (General, Network), Rule 08 (Dev Env).

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
# shellcheck source=/dev/null
. "${SCRIPT_DIR:-}/lib/common.sh"

# Purpose: Displays usage information.
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Bootstraps the unirtm toolchain and installs all project tools from .unirtm.toml.

Options:
  --dry-run        Preview what will be installed without making changes.
  -v, --verbose    Enable verbose output.
  -q, --quiet      Suppress informational output.
  -h, --help       Show this help message.

EOF
}

# Purpose: Main entry point. Bootstraps unirtm then delegates to unirtm install.
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  parse_common_args "$@"

  log_info "🚀 Setting up project toolchain via unirtm..."

  # 3. Bootstrap Toolchain Manager
  bootstrap_unirtm || {
    log_error "Failed to bootstrap unirtm. Please install it manually."
    exit 1
  }

  # 4. Install all tools declared in .unirtm.toml
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_info "[DRY-RUN] Would run: unirtm install"
  else
    optimize_network
    export GIT_PROTOCOL=version=2
    export UNIRTM_GIT_ALWAYS_USE_GIX=0
    log_info "Installing tools from .unirtm.toml..."
    "${_G_UNIRTM_BIN:-unirtm}" install

    # CI PATH Persistence: ensure installed tool paths are available to subsequent steps
    if is_ci_env; then
      log_info "[CI-PATH] Persisting unirtm paths to CI..."
      [ -d "${_G_UNIRTM_BIN_BASE:-}" ] && _persist_path_to_ci "${_G_UNIRTM_BIN_BASE:-}"
      [ -d "${_G_UNIRTM_SHIMS_BASE:-}" ] && _persist_path_to_ci "${_G_UNIRTM_SHIMS_BASE:-}"
    fi
  fi

  log_success "\n✨ Setup complete!"

  if [ "${DRY_RUN:-0}" -eq 0 ] && [ "${_IS_TOP_LEVEL:-true}" = "true" ]; then
    printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
    printf "  - Run %bmake verify%b to ensure environment health.\n" "${GREEN:-}" "${NC:-}"
  fi
}

main "$@"
