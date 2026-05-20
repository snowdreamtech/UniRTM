#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/update.sh - Toolchain Update Manager
#
# Purpose:
#   Standardizes the maintenance of global and project-local development toolsets.
#   Orchestrates updates for Brew, NPM, Python, and Git Hooks in a unified CLI.
#
# Usage:
#   sh scripts/update.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (Network), Rule 05 (Dependencies), Rule 08 (Dev Env).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Unified reporting for multi-stack toolchain maintenance.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"
. "${SCRIPT_DIR:-}/lib/github-api-info.sh"

# ── Configuration ────────────────────────────────────────────────────────────

# Purpose: Displays usage information for the update manager.
# Examples:
#   show_help
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Standardizes updating of global and project tools.

Options:
  --dry-run        Preview updates without applying them.
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

EOF
}

# Purpose: Internal helper to execute and report a standard update workflow.
# Params:
#   $1 - Manager/Category name (e.g., "Manager", "Project", "Lint Tool")
#   $2 - Tool name (e.g., "pnpm", "Homebrew")
#   $3 - Command to execute (e.g., "brew update")
#   $4 - Version command (optional, e.g., "$(get_version brew)")
_execute_update() {
  local _CATEGORY="${1:-}"
  local _TOOL="${2:-}"
  local _CMD="${3:-}"
  local _VERSION_CMD="${4:-}"

  local _T0
  _T0=$(date +%s)

  log_info "Updating ${_TOOL:-}..."
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "${_CATEGORY:-}" "${_TOOL:-}" "⚖️ Previewed" "-" "0"
  else
    if run_quiet eval "${_CMD:-}"; then
      log_summary "${_CATEGORY:-}" "${_TOOL:-}" "✅ Updated" "${_VERSION_CMD:-"-"}" "$(($(date +%s) - _T0))"
    else
      log_summary "${_CATEGORY:-}" "${_TOOL:-}" "❌ Failed" "-" "$(($(date +%s) - _T0))"
    fi
  fi
}

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Updates global Node.js manager if applicable.
# Examples:
#   update_node_manager_global
update_node_manager_global() {
  case "${NPM:-}" in
  pnpm)
    if command -v corepack >/dev/null 2>&1 && pnpm self-update --help 2>&1 | grep -q "corepack" >/dev/null 2>&1; then
      _execute_update "Manager" "pnpm" "corepack prepare pnpm@latest --activate" "$(get_version pnpm)"
    else
      _execute_update "Manager" "pnpm" "pnpm self-update" "$(get_version pnpm)"
    fi
    ;;
  npm)
    # Most npm versions are updated via npm itself, but since we use unirtm,
    # it's usually better to let unirtm handle it. We skip global self-update for npm here.
    log_debug "Skipping global self-update for $NPM (managed via unirtm)."
    ;;
  yarn)
    if command -v corepack >/dev/null 2>&1; then
      _execute_update "Manager" "yarn" "corepack prepare yarn@latest --activate" "$(get_version yarn)"
    else
      # Berry/Modern yarn uses 'set version latest' for self-updates
      _execute_update "Manager" "yarn" "yarn set version latest" "$(get_version yarn)"
    fi
    ;;
  bun)
    _execute_update "Manager" "${NPM:-}" "$NPM upgrade" "$(get_version "${NPM:-}" "--version")"
    ;;
  esac
}

# Purpose: Updates project dependencies using the detected manager.
# Examples:
#   update_node_project_deps
update_node_project_deps() {
  if [ -f "${PACKAGE_JSON:-}" ]; then
    _execute_update "Project" "$NPM-deps" "$NPM update" ""
  fi
}

# Purpose: Updates pip and project dependencies within the virtual environment.
# Examples:
#   update_python_venv
update_python_venv() {
  if [ -d "${VENV:-}" ] && [ -x "$VENV/bin/pip" ]; then
    local _CMD_PY="\"$VENV/bin/pip\" install --upgrade pip"
    if [ -f "${REQUIREMENTS_TXT:-}" ]; then
      _CMD_PY="${_CMD_PY:-} && \"$VENV/bin/pip\" install -r \"$REQUIREMENTS_TXT\" --upgrade"
    fi
    _execute_update "Project" "Python-Venv" "${_CMD_PY:-}" "\$(get_version \"$VENV/bin/pip\")"
  fi
}

# Purpose: Updates Homebrew formulae and casks.
# Examples:
#   update_homebrew
update_homebrew() {
  if command -v brew >/dev/null 2>&1; then
    _execute_update "Manager" "Homebrew" "brew update" ""
  fi
}

# Purpose: Updates MacPorts and outdated ports.
# Examples:
#   update_macports
update_macports() {
  if command -v port >/dev/null 2>&1; then
    _execute_update "Manager" "MacPorts" "sudo port selfupdate && sudo port -N upgrade outdated" ""
  fi
}

# Purpose: Updates Rubocop gem if installed.
# Examples:
#   update_ruby_gems
update_ruby_gems() {
  if command -v gem >/dev/null 2>&1 && gem list rubocop -i >/dev/null 2>&1; then
    _execute_update "Lint Tool" "Rubocop" "gem update rubocop --user-install --no-document --quiet" "$(get_version rubocop)"
  fi
}

# Purpose: Updates pre-commit hooks specified in the project configuration.
# Examples:
#   update_pre_commit
update_pre_commit() {
  local _BIN_PC=""
  if [ -x "$VENV/bin/pre-commit" ]; then
    _BIN_PC="$VENV/bin/pre-commit"
  elif command -v pre-commit >/dev/null 2>&1; then _BIN_PC="pre-commit"; fi

  if [ -n "${_BIN_PC:-}" ] && [ -f ".pre-commit-config.yaml" ]; then
    _execute_update "Other" "Hooks" "\"${_BIN_PC:-}\" autoupdate" "\$(get_version \"${_BIN_PC:-}\")"
  fi
}

# Purpose: Updates Go project dependencies.
# Examples:
#   update_go_mod
update_go_mod() {
  if [ -f "go.mod" ]; then
    if command -v go >/dev/null 2>&1; then
      _execute_update "Project" "Go-Mod" "go get -u ./... && go mod tidy" "$(get_version go)"
    else
      log_warn "go.mod found but go command is missing. Skipping."
      log_summary "Project" "Go-Mod" "⚠️ Missing" "-" "0"
    fi
  fi
}

# Purpose: Updates Rust project dependencies.
# Examples:
#   update_cargo_deps
update_cargo_deps() {
  if [ -f "Cargo.toml" ]; then
    if command -v cargo >/dev/null 2>&1; then
      _execute_update "Project" "Cargo-Deps" "cargo update" "$(get_version cargo)"
    else
      log_warn "Cargo.toml found but cargo command is missing. Skipping."
      log_summary "Project" "Cargo-Deps" "⚠️ Missing" "-" "0"
    fi
  fi
}

# Purpose: Upgrades UniRTM-managed tool versions in registry and config.
# Examples:
#   update_unirtm_tool_versions
update_unirtm_tool_versions() {
  if [ -f "${SCRIPT_DIR:-}/update-tools.sh" ]; then
    log_info "Upgrading UniRTM tool versions..."
    # Delegate to update-tools.sh, passing through relevant flags
    sh "${SCRIPT_DIR:-}/update-tools.sh" "$@" || log_warn "UniRTM tool version upgrade failed."
  fi
}

# Purpose: Main entry point for the toolchain update manager.
# Params:
#   $@ - Command line arguments
# Examples:
#   main --dry-run
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  parse_common_args "$@"

  local _START_TIME_M
  _START_TIME_M=$(date +%s)

  # Initialize Summary File correctly
  init_summary_table "Update Execution Summary"

  # Initialize Summary Legend (Only once per CI Job or first call)
  if [ "${_UPDATE_SUMMARY_INITIALIZED:-false}" != "true" ] && ! check_ci_summary "### Update Execution Summary"; then
    {
      printf "### Update Execution Summary\n\n"
    } >>"${CI_STEP_SUMMARY:-}"
    [ -n "${GITHUB_ENV:-}" ] && echo "_UPDATE_SUMMARY_INITIALIZED=true" >>"${GITHUB_ENV:-}"
    export _UPDATE_SUMMARY_INITIALIZED=true
  fi

  # Provide table header if not already present
  if [ "${_SUMMARY_TABLE_HEADER_SENTINEL:-false}" != "true" ] && ! check_ci_summary "| Category | Module | Status |"; then
    {
      printf "| Category | Module | Status | Version | Time |\n"
      printf "| :--- | :--- | :--- | :--- | :--- |\n"
    } >>"${CI_STEP_SUMMARY:-}"
    [ -n "${GITHUB_ENV:-}" ] && echo "_SUMMARY_TABLE_HEADER_SENTINEL=true" >>"${GITHUB_ENV:-}"
    export _SUMMARY_TABLE_HEADER_SENTINEL=true
  fi

  update_homebrew
  update_macports
  update_node_manager_global
  update_node_project_deps
  update_python_venv
  update_go_mod
  update_cargo_deps
  update_ruby_gems
  update_pre_commit
  update_unirtm_tool_versions "$@"

  # Optional: run npm update if defined
  run_npm_script "update"

  # Final Output Management
  if [ "${_IS_TOP_LEVEL:-true}" = "true" ]; then
    local _TOTAL_DUR_M
    _TOTAL_DUR_M=$(($(date +%s) - _START_TIME_M))
    printf "\n**Total Duration: %ss**\n" "${_TOTAL_DUR_M:-}" >>"${CI_STEP_SUMMARY:-}"

    # Add GitHub API rate limit info
    append_github_api_info

    printf "\n"
    finalize_summary_table
    log_info "\n✨ Update complete!"
  fi

  if [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    log_success "\n✨ All tools and dependencies updated successfully!"

    # 6. Standardized Next Actions
    if [ "${DRY_RUN:-0}" -eq 0 ]; then
      printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
      printf "  - Run %bunirtm install%b to synchronize project dependencies.\n" "${GREEN:-}" "${NC:-}"
      printf "  - Run %bmake verify%b to ensure environment health and stability.\n" "${GREEN:-}" "${NC:-}"
    fi
  fi
}

main "$@"
