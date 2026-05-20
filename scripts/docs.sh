#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/docs.sh - Documentation Lifecycle Manager
#
# Purpose:
#   Unified entrance for VitePress development, artifact building, and previews.
#   Streamlines the maintenance and publication of project documentation.
#
# Usage:
#   sh scripts/docs.sh [OPTIONS] [COMMAND]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General, Network), Rule 12 (Docs).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Automated VitePress installation checks.
#   - Environment-aware routing for local and CI docs.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# Purpose: Displays usage information for the documentation manager.
# Examples:
#   show_help
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS] [COMMAND]

Manages the VitePress documentation site.

Commands:
  dev              Start VitePress development server (default).
  build            Build the documentation site.
  preview          Preview the production build.

Options:
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

Environment Variables:
  NPM              NPM client (detected: $NPM)
  DOCS_DIR         Documentation directory (default: docs)

EOF
}

# Purpose: Main entry point for the documentation management engine.
#          Routes to appropriate VitePress commands based on user input.
# Params:
#   $@ - Command line arguments and optional command
# Examples:
#   main build
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  local _COMMAND_DOC="dev"
  local _arg_doc
  for _arg_doc in "$@"; do
    case "${_arg_doc:-}" in
    dev | build | preview) _COMMAND_DOC="${_arg_doc:-}" ;;
    esac
  done
  parse_common_args "$@"

  log_info "📖 Documentation Manager (${_COMMAND_DOC:-})...\n"

  # 3. Dependency checks — skipped in dry-run (environment may not have docs/ or vitepress)
  if [ "${DRY_RUN:-0}" -eq 0 ]; then
    if [ ! -d "${DOCS_DIR:-}" ]; then
      log_error "Error: Documentation directory '$DOCS_DIR' not found."
      exit 1
    fi

    if ! resolve_bin "${NPM:-}" >/dev/null 2>&1; then
      log_error "Error: $NPM not found. Please install it to build documentation."
      exit 1
    fi

    # 4. Install documentation dependencies if package.json exists
    if [ -f "$DOCS_DIR/package.json" ]; then
      log_info "Installing documentation dependencies..."
      # Use pnpm directly with frozen lockfile for faster installation
      # Skip if dependencies are already up to date
      if [ ! -d "$DOCS_DIR/node_modules" ] || [ "$DOCS_DIR/package.json" -nt "$DOCS_DIR/node_modules" ]; then
        pnpm --dir "${DOCS_DIR:-}" install --frozen-lockfile --prefer-offline
      else
        log_info "Dependencies are up to date, skipping install"
      fi
    fi

    # 5. Resolve VitePress (check docs/node_modules/.bin via extra search paths)
    local _VITEPRESS_BIN
    _VITEPRESS_BIN=$(resolve_bin "vitepress" "${DOCS_DIR:-}/node_modules/.bin") || true

    if [ -z "${_VITEPRESS_BIN:-}" ]; then
      log_error "Error: vitepress not found. Please run 'unirtm install' first."
      exit 1
    fi
  fi

  # 5. Execute VitePress
  case "${_COMMAND_DOC:-}" in
  dev)
    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_success "DRY-RUN: Would start VitePress dev server on $DOCS_DIR"
    else
      local _VITEPRESS_BIN
      _VITEPRESS_BIN=$(resolve_bin "vitepress" "${DOCS_DIR:-}/node_modules/.bin") || true
      log_info "Starting development server..."
      "${_VITEPRESS_BIN:-}" dev "${DOCS_DIR:-}"
    fi
    ;;
  build)
    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_success "DRY-RUN: Would build VitePress site from $DOCS_DIR"
    else
      # VitePress binary was already resolved and dependencies installed in the checks section
      local _VITEPRESS_BIN
      _VITEPRESS_BIN=$(resolve_bin "vitepress" "${DOCS_DIR:-}/node_modules/.bin") || true
      log_info "Building documentation site..."
      "${_VITEPRESS_BIN:-}" build "${DOCS_DIR:-}"
      log_success "\n✨ Build complete! Artifacts are in $DOCS_DIR/.vitepress/dist"
    fi
    ;;
  preview)
    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_success "DRY-RUN: Would preview VitePress site in $DOCS_DIR"
    else
      local _VITEPRESS_BIN
      _VITEPRESS_BIN=$(resolve_bin "vitepress" "${DOCS_DIR:-}/node_modules/.bin") || true
      log_info "Previewing production build..."
      "${_VITEPRESS_BIN:-}" preview "${DOCS_DIR:-}"
    fi
    ;;
  esac

  # 5. Standardized Next Actions
  if [ "${DRY_RUN:-0}" -eq 0 ] && [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
    if [ "${_COMMAND_DOC:-}" = "build" ]; then
      printf "  - Run %bmake release%b to publish the documentation and project.\n" "${GREEN:-}" "${NC:-}"
    else
      printf "  - Run %bmake docs build%b to generate production-ready documentation.\n" "${GREEN:-}" "${NC:-}"
    fi
  fi
}

main "$@"
